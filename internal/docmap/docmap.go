// Package docmap builds the doc-cartography artifact decided in
// ADR 0044: a machine-derived map of a project's *document set* —
// inventory, cross-references, freshness, and a managed/external split.
// The design note is docs/decisions/docmap-design.md.
//
// Build produces the data layer (Map / docmap.yaml). The readable view
// (docmap.md) is rendered from the same Map in a later phase. The
// builder reads document text only — front-matter, headings, the first
// lines, and a link regex (via internal/docmeta) — so the stack-agnostic
// core is preserved (ADR 0044 D4): no source code is read.
package docmap

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shigindo-inc/aikata/internal/docmeta"
	"github.com/shigindo-inc/aikata/internal/glob"
)

// Version is the schema version of `.aikata/docmap.yaml`. Bumped only on
// incompatible shape changes.
const Version = 1

// defaultTargets is the default catalog breadth (Q-DOCMAP-01): every
// Markdown file, with the per-document `managed` flag distinguishing the
// aikata-managed surface from external documents. The built-in scan
// skips (vendored deps, build outputs, machine/scratch areas, generated
// AI-tool artifacts) are shared with doctor via docmeta.DefaultSkipDirs /
// DefaultSkipFiles so the two surfaces cannot drift.
var defaultTargets = []string{"**/*.md"}

// Options carries the inputs Build needs.
type Options struct {
	// TargetDir is the project root to scan.
	TargetDir string
	// Now sets the map's `generated:` date. The zero value resolves to
	// time.Now(); tests inject a fixed clock for deterministic golden
	// output (only the date is recorded, never the time).
	Now time.Time
	// Targets are the catalog globs (slash-form, relative to TargetDir).
	// Empty means defaultTargets ("**/*.md").
	Targets []string
	// Exclude lists additional catalog skips (user config, P5). The
	// built-in directory and AI-tool-artifact skips always apply on top.
	Exclude []string
	// Formats selects which renderings Generate writes. The data layer
	// (docmap.yaml) is always written regardless; empty means the default
	// view set [yaml, md]. Recognised: yaml, md, json, txt, mmd.
	Formats []string
	// ManagedGlobs marks which tracked documents count as aikata-managed
	// (the `managed: true` flag). Callers pass doctor.ManagedIncludeGlobs
	// so the flag matches the doctor surface; docmap does not import
	// doctor, which keeps the dependency one-directional (doctor's
	// freshness check imports docmap, not the reverse).
	ManagedGlobs []string
}

// Doc is one catalogued document. All fields are mechanically derived;
// Links holds only references that resolve to another document inside the
// tracked set (sorted, de-duplicated) so the graph has no dangling edges.
type Doc struct {
	Path    string   `yaml:"path"`
	Title   string   `yaml:"title"`
	Summary string   `yaml:"summary"`
	Status  string   `yaml:"status,omitempty"`
	Updated string   `yaml:"updated,omitempty"`
	Managed bool     `yaml:"managed"`
	Links   []string `yaml:"links,omitempty"`
}

// Map is the data layer of the doc map — the single source of the map's
// truth, from which every rendering (docmap.yaml, docmap.md, optional
// formats) is derived.
type Map struct {
	Version   int    `yaml:"version"`
	Generated string `yaml:"generated"`
	Docs      []Doc  `yaml:"docs"`
}

// Build scans TargetDir and returns the document Map. Docs are sorted by
// path for deterministic, diff-stable output. The error return is
// reserved for I/O failures that prevent the scan from running.
func Build(opts Options) (Map, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	targets := opts.Targets
	if len(targets) == 0 {
		targets = defaultTargets
	}

	paths, bodies, err := scan(opts.TargetDir, targets, opts.Exclude)
	if err != nil {
		return Map{}, err
	}

	// The tracked set is every path that survived the scan; it bounds
	// which links become edges (no dangling edges, design §2.1).
	tracked := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		tracked[p] = struct{}{}
	}

	docs := make([]Doc, 0, len(paths))
	for _, rel := range paths {
		body := bodies[rel]
		lines, offset := docmeta.ParseFrontmatter(body)
		docs = append(docs, Doc{
			Path:    rel,
			Title:   title(lines, body, offset, rel),
			Summary: summary(lines, body, offset, rel),
			Status:  frontmatterScalar(lines, "status"),
			Updated: updated(lines, opts.TargetDir, rel),
			Managed: glob.MatchAny(opts.ManagedGlobs, rel),
			Links:   edges(body, rel, tracked),
		})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })

	return Map{
		Version:   Version,
		Generated: now.Format("2006-01-02"),
		Docs:      docs,
	}, nil
}

// scan walks dir and returns the tracked Markdown paths (slash-form,
// relative to dir, sorted) plus their bodies. Built-in directory and
// file skips always apply; a path is tracked only if it matches a target
// glob and no exclude glob.
func scan(dir string, targets, exclude []string) (paths []string, bodies map[string][]byte, err error) {
	bodies = make(map[string][]byte)
	walkErr := filepath.WalkDir(dir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if _, skip := docmeta.DefaultSkipDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		if _, skip := docmeta.DefaultSkipFiles[d.Name()]; skip {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if !glob.MatchAny(targets, relSlash) {
			return nil
		}
		if glob.MatchAny(exclude, relSlash) {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		paths = append(paths, relSlash)
		bodies[relSlash] = body
		return nil
	})
	if walkErr != nil {
		return nil, nil, walkErr
	}
	sort.Strings(paths)
	return paths, bodies, nil
}

// edges resolves the document's relative links against its own directory
// and keeps only those landing on another tracked document (excluding a
// self-link). The result is sorted and de-duplicated.
func edges(body []byte, fromRel string, tracked map[string]struct{}) []string {
	dir := path.Dir(fromRel) // slash-form; "." for root documents
	seen := make(map[string]struct{})
	var out []string
	for _, link := range docmeta.ExtractLinks(body) {
		target := link.Target
		if dir != "." {
			target = path.Join(dir, target)
		} else {
			target = path.Clean(target)
		}
		if target == fromRel {
			continue
		}
		if _, ok := tracked[target]; !ok {
			continue
		}
		if _, dup := seen[target]; dup {
			continue
		}
		seen[target] = struct{}{}
		out = append(out, target)
	}
	sort.Strings(out)
	return out
}

// frontmatterScalar returns a trimmed top-level scalar from the parsed
// front-matter, or "" when absent.
func frontmatterScalar(lines []string, key string) string {
	v, _, ok := docmeta.FrontmatterValue(lines, key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

// updated returns the document's `updated:` front-matter value when
// present, else the file's modification date (YYYY-MM-DD). A stat
// failure degrades to "" rather than failing the whole scan.
func updated(lines []string, targetDir, rel string) string {
	if v := frontmatterScalar(lines, "updated"); v != "" {
		return v
	}
	info, err := os.Stat(filepath.Join(targetDir, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	return info.ModTime().UTC().Format("2006-01-02")
}
