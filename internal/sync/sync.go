// Package sync implements `aikata sync`: pull newer aikata template
// content into an existing project without losing user edits.
//
// The merge contract is documented in ADR 0011. In short:
//
//   - `.aikata/manifest.yaml` (written by `aikata init`) records the
//     SHA-256 of every template-derived file as it was originally
//     rendered. That is the common ancestor in a 3-way merge.
//   - sync re-renders upstream templates against the project's
//     current scaffold options, compares against on-disk content,
//     and decides per file whether the change is upstream-only
//     (auto-apply), user-only (preserve), agreement (no-op), or a
//     true conflict.
//   - File-level conflict markers (git-merge-style `<<<<<<<`,
//     `|||||||`, `=======`, `>>>>>>>`) are written back to the file
//     for the user to resolve. v0.5 wraps the whole file content;
//     line-level merge is a v0.5.x follow-up.
//   - On a clean (zero-conflict) run, the manifest is regenerated so
//     the next sync uses the just-applied state as the ancestor.
//
// `aikata generate` artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`,
// …) are explicitly out of scope here — see ADR 0011 D1.
package sync

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// Status enumerates the per-file outcome of one sync run. Values
// double as the JSON envelope's `status` field so external tools can
// branch on the same set.
type Status string

const (
	// StatusUnchanged means ancestor == current == upstream; no write.
	StatusUnchanged Status = "unchanged"
	// StatusUpstreamApplied means the user file matched the manifest
	// hash, so upstream changes were applied cleanly without losing
	// any local edits.
	StatusUpstreamApplied Status = "upstream-applied"
	// StatusUserOnlyEdit means the user edited the file but upstream
	// is unchanged from the manifest. The user's content is preserved
	// verbatim.
	StatusUserOnlyEdit Status = "user-only-edit"
	// StatusBothMatch means the user file happens to equal the
	// upstream rendering. No write is performed, but the new manifest
	// hash advances to upstream's.
	StatusBothMatch Status = "both-match"
	// StatusConflict means the user edited the file AND upstream
	// changed it AND the two diverge. File-level conflict markers
	// were written; manual resolution is required.
	StatusConflict Status = "conflict"
	// StatusUpstreamAdded means the file did not exist locally and
	// upstream now renders it. The file is auto-written.
	StatusUpstreamAdded Status = "upstream-added"
	// StatusUpstreamRemoved means the file was in the manifest and is
	// still on disk, but upstream no longer renders it. Left in place
	// so the user can choose; surfaced in the summary as info-level.
	StatusUpstreamRemoved Status = "upstream-removed"
	// StatusUserDeleted means the file was in the manifest, the user
	// removed it, and upstream still renders it. The deletion is
	// respected (we don't restore it) unless --restore-deletes is
	// passed (v0.5.x follow-up).
	StatusUserDeleted Status = "user-deleted"
)

// FileResult describes the outcome for one path in a sync run.
type FileResult struct {
	Path   string
	Status Status
}

// RunResult aggregates per-file outcomes plus the counts the cobra
// layer uses to format the human-readable summary.
type RunResult struct {
	Files     []FileResult
	Conflicts int
	Applied   int
	NoChange  int
	Notes     []string
}

// Options controls one sync invocation. Root must be the project
// directory; aikata reads `.aikata/aikata.yaml` and
// `.aikata/manifest.yaml` from it.
type Options struct {
	Root       string
	DryRun     bool
	Rebaseline bool
	Stdout     io.Writer
	Stderr     io.Writer
}

// ErrNoManifest signals that the project has never been initialized by
// aikata, or predates the v0.5 manifest schema. The cobra layer maps
// this to an actionable message: run `aikata sync --rebaseline` to
// seed a manifest from current on-disk state.
var ErrNoManifest = errors.New("sync: .aikata/manifest.yaml not found; run `aikata sync --rebaseline` once to seed it from the current project state")

// Run performs one sync invocation against opts.Root. The return value
// is non-nil even when conflicts were detected; callers map
// RunResult.Conflicts > 0 to a non-zero exit code rather than treating
// it as an error.
func Run(opts Options) (RunResult, error) {
	if opts.Root == "" {
		return RunResult{}, errors.New("sync: Root is required")
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	cfg, _, err := config.Load(opts.Root)
	if err != nil {
		return RunResult{}, fmt.Errorf("sync: load config: %w", err)
	}

	// Manifest may be absent. --rebaseline replaces the empty
	// manifest with one seeded from current state so subsequent
	// syncs can do a real 3-way merge.
	var ancestor config.Manifest
	manifestPresent := true
	loaded, err := config.LoadManifest(opts.Root)
	if errors.Is(err, fs.ErrNotExist) {
		manifestPresent = false
		if !opts.Rebaseline {
			return RunResult{}, ErrNoManifest
		}
	} else if err != nil {
		return RunResult{}, fmt.Errorf("sync: load manifest: %w", err)
	} else {
		ancestor = loaded
	}

	// Determine which preset / lang / optional components to render.
	// In normal mode, the manifest carries the original preset / lang
	// and the file set tells us which optional components were
	// enabled. In rebaseline mode we have no manifest, so we trust
	// `.aikata/aikata.yaml`'s project.lang and pick `standard` as the
	// safest default for the preset (the user can re-run with an
	// explicit preset flag in a v0.5.x follow-up).
	preset, lang, withFlags := derivePlan(ancestor, cfg, manifestPresent)

	scaffoldOpts := scaffold.Options{
		ProjectName:   cfg.Project.Name,
		Preset:        preset,
		TargetDir:     opts.Root, // not read or written by Render
		Lang:          lang,
		Stacks:        append([]string(nil), cfg.Stacks...),
		AITools:       append([]string(nil), cfg.AITools...),
		WithMemory:    withFlags.WithMemory,
		WithUI:        withFlags.WithUI,
		WithAPI:       withFlags.WithAPI,
		WithTDD:       withFlags.WithTDD,
		WithChangelog: withFlags.WithChangelog,
		Stdout:        opts.Stderr, // route any fallback notice to stderr
	}
	upstream, err := scaffold.Render(scaffoldOpts)
	if err != nil {
		return RunResult{}, fmt.Errorf("sync: render upstream: %w", err)
	}

	// The freshly built manifest comes back inside `upstream`; pull
	// it out so it doesn't get treated like a template-derived file
	// in the merge. We regenerate it after the merge succeeds.
	manifestRel := config.PrimaryDir + "/" + config.ManifestFilename
	configRel := config.PrimaryDir + "/" + config.Filename
	delete(upstream, manifestRel)
	delete(upstream, configRel)

	// Build a lookup of ancestor hashes keyed by path.
	ancestorHash := make(map[string]string, len(ancestor.Files))
	for _, f := range ancestor.Files {
		ancestorHash[f.Path] = f.SHA256
	}

	// Collect every path that appears in ancestor or upstream; sort
	// for stable iteration / output.
	paths := make(map[string]struct{}, len(ancestorHash)+len(upstream))
	for p := range ancestorHash {
		paths[p] = struct{}{}
	}
	for p := range upstream {
		paths[p] = struct{}{}
	}
	sortedPaths := make([]string, 0, len(paths))
	for p := range paths {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	var result RunResult
	// Track the post-merge content so we can rebuild the manifest if
	// the run completes cleanly.
	merged := make(map[string]string, len(upstream))
	for _, path := range sortedPaths {
		fileResult, mergedContent, err := classifyAndMerge(path, ancestorHash, upstream, opts.Root)
		if err != nil {
			return RunResult{}, err
		}
		result.Files = append(result.Files, fileResult)
		switch fileResult.Status {
		case StatusConflict:
			result.Conflicts++
		case StatusUpstreamApplied, StatusUpstreamAdded:
			result.Applied++
		default:
			result.NoChange++
		}
		if mergedContent != nil {
			merged[path] = string(mergedContent)
		}
	}

	if opts.DryRun {
		return result, nil
	}

	// Apply the merged file set to disk for entries with a buffered
	// merged content (auto-applied or conflict-markered).
	for path, body := range merged {
		full := filepath.Join(opts.Root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return result, fmt.Errorf("sync: mkdir %s: %w", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return result, fmt.Errorf("sync: write %s: %w", full, err)
		}
	}

	// Regenerate the manifest only when the run is conflict-free —
	// otherwise the ancestor would jump past the in-progress
	// resolution and confuse the next sync.
	if result.Conflicts == 0 {
		snapshot, err := postMergeSnapshot(opts.Root, upstream)
		if err != nil {
			return result, err
		}
		newManifest := config.BuildManifest(preset, lang, snapshot, []string{manifestRel, configRel})
		if err := config.SaveManifest(opts.Root, newManifest); err != nil {
			return result, fmt.Errorf("sync: save manifest: %w", err)
		}
	} else {
		result.Notes = append(result.Notes,
			"Manifest not updated because conflicts remain. Resolve them and re-run `aikata sync`.")
	}

	return result, nil
}

// classifyAndMerge implements the per-file decision matrix described
// at the top of this file. Returns the file's result, the bytes to
// write (nil = no write needed), and any I/O error encountered.
func classifyAndMerge(path string, ancestorHash map[string]string, upstream map[string]string, root string) (FileResult, []byte, error) {
	ancestor, hadAncestor := ancestorHash[path]
	upstreamBody, hasUpstream := upstream[path]
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	currentBody, currentErr := os.ReadFile(fullPath)
	hasCurrent := currentErr == nil
	if currentErr != nil && !errors.Is(currentErr, fs.ErrNotExist) {
		return FileResult{Path: path}, nil, fmt.Errorf("sync: read %s: %w", path, currentErr)
	}

	currentHash := ""
	if hasCurrent {
		currentHash = config.HashContent(currentBody)
	}
	upstreamHash := ""
	if hasUpstream {
		upstreamHash = config.HashContent([]byte(upstreamBody))
	}

	switch {
	case hadAncestor && hasCurrent && hasUpstream:
		return mergeThreeWay(path, ancestor, currentHash, upstreamHash, currentBody, upstreamBody)
	case hadAncestor && hasCurrent && !hasUpstream:
		return FileResult{Path: path, Status: StatusUpstreamRemoved}, nil, nil
	case hadAncestor && !hasCurrent && hasUpstream:
		return FileResult{Path: path, Status: StatusUserDeleted}, nil, nil
	case hadAncestor && !hasCurrent && !hasUpstream:
		return FileResult{Path: path, Status: StatusUserDeleted}, nil, nil
	case !hadAncestor && !hasCurrent && hasUpstream:
		return FileResult{Path: path, Status: StatusUpstreamAdded}, []byte(upstreamBody), nil
	case !hadAncestor && hasCurrent && hasUpstream:
		// File is new to upstream but the user already has one at the
		// same path. Treat as 3-way with an empty-string ancestor —
		// effectively a 2-way diff; matching content is a no-op, mismatch
		// is a conflict.
		if currentHash == upstreamHash {
			return FileResult{Path: path, Status: StatusBothMatch}, nil, nil
		}
		marked := conflictMarkers(string(currentBody), "", upstreamBody)
		return FileResult{Path: path, Status: StatusConflict}, []byte(marked), nil
	default:
		// Path appeared in our iteration set but isn't in ancestor
		// or upstream — only possible from a bug. Treat as unchanged.
		return FileResult{Path: path, Status: StatusUnchanged}, nil, nil
	}
}

// mergeThreeWay handles the case where the path is present in all
// three sources (ancestor, current, upstream).
func mergeThreeWay(path, ancestorHash, currentHash, upstreamHash string, currentBody []byte, upstreamBody string) (FileResult, []byte, error) {
	switch {
	case currentHash == ancestorHash && upstreamHash == ancestorHash:
		return FileResult{Path: path, Status: StatusUnchanged}, nil, nil
	case currentHash == ancestorHash && upstreamHash != ancestorHash:
		// User hasn't edited; upstream evolved. Apply upstream.
		return FileResult{Path: path, Status: StatusUpstreamApplied}, []byte(upstreamBody), nil
	case currentHash != ancestorHash && upstreamHash == ancestorHash:
		// Upstream unchanged; user edited. Preserve user content.
		return FileResult{Path: path, Status: StatusUserOnlyEdit}, nil, nil
	case currentHash == upstreamHash:
		// Both diverged from ancestor but ended up at the same place.
		return FileResult{Path: path, Status: StatusBothMatch}, nil, nil
	default:
		marked := conflictMarkers(string(currentBody), "", upstreamBody)
		return FileResult{Path: path, Status: StatusConflict}, []byte(marked), nil
	}
}

// conflictMarkers wraps three bodies (ours / ancestor / theirs) in
// git-merge-style markers. v0.5 operates at file granularity; a
// future revision may switch to line-level diff3 once a real-world
// pain point shows up.
func conflictMarkers(ours, ancestor, theirs string) string {
	var b strings.Builder
	b.WriteString("<<<<<<< current (your edits)\n")
	b.WriteString(ours)
	if !strings.HasSuffix(ours, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("||||||| ancestor (.aikata/manifest.yaml)\n")
	b.WriteString(ancestor)
	if ancestor != "" && !strings.HasSuffix(ancestor, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("=======\n")
	b.WriteString(theirs)
	if !strings.HasSuffix(theirs, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(">>>>>>> upstream\n")
	return b.String()
}

// postMergeSnapshot reads the current on-disk bytes for every path in
// upstream so the rebuilt manifest reflects the post-sync ancestor
// (which is exactly what's now on disk). Only called when conflicts
// == 0, so no file holds conflict markers at this point.
func postMergeSnapshot(root string, upstream map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(upstream))
	for path, body := range upstream {
		full := filepath.Join(root, filepath.FromSlash(path))
		got, err := os.ReadFile(full)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// User deleted the file and the run respected it
				// (StatusUserDeleted). Skip it from the new manifest
				// so the next sync sees the same state.
				continue
			}
			return nil, fmt.Errorf("sync: snapshot read %s: %w", path, err)
		}
		out[path] = string(got)
		// Keep `body` referenced so go vet doesn't complain about
		// the unused value when the file was deleted.
		_ = body
	}
	return out, nil
}
