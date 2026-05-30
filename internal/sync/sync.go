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
	"github.com/shigindo-inc/aikata/internal/glob"
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
	// StatusOwned means the path matched the project's `sync.own` glob
	// list (ADR 0025 D2). aikata treats the file as user-owned: it is
	// never rendered-compared, conflict-markered, overwritten, or
	// tracked in the manifest.
	StatusOwned Status = "owned"
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
//
// The Override* fields express one-off CLI scope overrides (ADR 0013).
// Each is nil when the matching CLI flag was not passed; non-nil
// values take precedence over both the manifest and `aikata.yaml`
// for the current invocation only. The values are intentionally not
// written back to disk — to make a change persistent, use
// `aikata add` / hand-edit `aikata.yaml`.
type Options struct {
	Root       string
	DryRun     bool
	Rebaseline bool
	// Reseed re-anchors an existing `.aikata/manifest.yaml` to the
	// current upstream rendering and exits without running the merge
	// (ADR 0025 D4). Only the manifest is written; no source file is
	// touched. Unlike --rebaseline (which seeds a *missing* manifest),
	// --reseed deliberately overwrites an existing one.
	Reseed bool
	Stdout io.Writer
	Stderr io.Writer

	OverridePreset       *string
	OverrideLang         *string
	OverrideStacks       *[]string
	OverrideWithMonorepo *bool
}

// ErrNoManifest signals that the project has never been initialized by
// aikata, or predates the v0.5 manifest schema. The cobra layer maps
// this to an actionable message: run `aikata sync --rebaseline` to
// seed a manifest from current on-disk state. The wording emphasises
// that `--rebaseline` is non-destructive (only the manifest is written)
// to avoid the v0.6.0 footgun where users feared their source files
// would be overwritten.
var ErrNoManifest = errors.New(
	"sync: .aikata/manifest.yaml not found (project predates v0.5).\n" +
		"  Run `aikata sync --rebaseline` to record current on-disk files as the\n" +
		"  baseline. This is non-destructive — only the manifest is written;\n" +
		"  no source files are modified. After that, `aikata sync` will pull\n" +
		"  upstream template updates via 3-way merge")

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

	// Per ADR 0011 D3, sync owns the schema migration: read via
	// LoadMigrated so any forward-migrations registered with
	// `config.aikataYamlMigrators` run before the merge sees the
	// payload. v0.5 ships with zero registered migrations, so the
	// path is effectively a pass-through; v2+ schemas slot in as
	// one-row additions to the registry.
	cfg, migrated, err := config.LoadMigrated(opts.Root)
	if err != nil {
		if errors.Is(err, config.ErrFutureSchema) {
			return RunResult{}, fmt.Errorf("sync: %w; upgrade aikata via `aikata update --check`", err)
		}
		return RunResult{}, fmt.Errorf("sync: load config: %w", err)
	}
	if migrated {
		if _, werr := fmt.Fprintln(opts.Stderr, "sync: migrated .aikata/aikata.yaml to the current schema version"); werr != nil {
			return RunResult{}, werr
		}
	}

	// Manifest may be absent (project predates v0.5). --rebaseline is
	// the only opt-in escape hatch; without it, refuse to proceed so
	// the user has a chance to read the docs before any merge runs.
	// The rebaseline path itself is handled below (after upstream is
	// rendered) and is intentionally non-destructive: it writes the
	// manifest from current disk state and exits without invoking the
	// 3-way merge. See ADR 0011 D4.
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

	// Determine which preset / lang / optional components / stacks
	// to render. The hierarchy is documented in ADR 0013:
	// CLI overrides > manifest > `.aikata/aikata.yaml` > defaults.
	preset, lang, withFlags, stacks, workflows := derivePlan(ancestor, cfg, manifestPresent, overrides{
		Preset:       opts.OverridePreset,
		Lang:         opts.OverrideLang,
		Stacks:       opts.OverrideStacks,
		WithMonorepo: opts.OverrideWithMonorepo,
	})

	scaffoldOpts := scaffold.Options{
		ProjectName:   cfg.Project.Name,
		Preset:        preset,
		TargetDir:     opts.Root, // not read or written by Render
		Lang:          lang,
		Stacks:        stacks,
		Workflows:     workflows,
		AITools:       append([]string(nil), cfg.AITools...),
		WithMemory:    withFlags.WithMemory,
		WithUI:        withFlags.WithUI,
		WithAPI:       withFlags.WithAPI,
		WithTDD:       withFlags.WithTDD,
		WithChangelog: withFlags.WithChangelog,
		WithMonorepo:  withFlags.WithMonorepo,
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

	// User-owned paths (ADR 0025 D2): a path matching `sync.own` is
	// never rendered-compared, conflict-markered, overwritten, or
	// tracked in the manifest. The matcher is shared with
	// `doctor.exclude` (internal/glob).
	isOwned := func(path string) bool { return glob.MatchAny(cfg.Sync.Own, path) }

	// --reseed re-anchors an existing manifest to the current upstream
	// rendering and exits without running the merge (ADR 0025 D4).
	// Manifest-only write; no source file is touched. Honored whether
	// or not a manifest is already present.
	if opts.Reseed {
		var result RunResult
		if !opts.DryRun {
			if err := saveManifestFromUpstream(opts.Root, preset, lang, upstream, isOwned); err != nil {
				return RunResult{}, err
			}
		}
		result.Notes = append(result.Notes,
			"Manifest re-seeded from the current upstream rendering. Source files were not modified.")
		return result, nil
	}

	// Rebaseline path (manifest absent + --rebaseline): write a
	// manifest whose ancestor hashes are the current *upstream*
	// rendering — the same manifest `aikata init` would have written
	// for a fresh project at this aikata version — and exit without
	// running the merge. No source file is touched.
	//
	// Why "upstream rendering" and not "current on-disk bytes":
	// if the ancestor were the on-disk bytes, the very next sync
	// would see `current == ancestor` and treat upstream-only changes
	// as auto-applicable, overwriting the user's customisations. By
	// seeding ancestor = upstream, the user's existing customisations
	// register as `user-only-edit` on the next sync and are preserved.
	// Conceptually: rebaseline pretends `aikata init` ran just now,
	// without touching anything that's already on disk. This is a
	// deliberate refinement of ADR 0011 D4's literal wording.
	if !manifestPresent {
		var result RunResult
		if !opts.DryRun {
			if err := saveManifestFromUpstream(opts.Root, preset, lang, upstream, isOwned); err != nil {
				return RunResult{}, err
			}
		}
		result.Notes = append(result.Notes,
			"Manifest seeded from current upstream rendering. Source files were not modified.",
			"Local customisations will appear as `user-only-edit` on the next `aikata sync` and be preserved.")
		return result, nil
	}

	// --rebaseline against a project that already has a manifest is a
	// no-op for re-seeding (ADR 0025 D4): say so explicitly instead of
	// silently running a normal merge, and point at --reseed. The merge
	// still proceeds below.
	if opts.Rebaseline {
		if _, werr := fmt.Fprintln(opts.Stderr,
			"sync: rebaseline skipped: .aikata/manifest.yaml already present "+
				"(use --reseed to re-anchor it to the current upstream rendering)"); werr != nil {
			return RunResult{}, werr
		}
	}

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
		// Owned paths bypass the 3-way merge entirely (ADR 0025 D2).
		if isOwned(path) {
			result.Files = append(result.Files, FileResult{Path: path, Status: StatusOwned})
			result.NoChange++
			continue
		}
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
	//
	// ADR 0025 D1: the new ancestor is the in-memory *upstream*
	// rendering, NOT a re-read of the post-merge on-disk bytes. This
	// unifies the post-clean-run path with the rebaseline / reseed
	// principle (ADR 0011 "ancestor = upstream rendering"). For a
	// `user-only-edit` file the upstream rendering equals the *old*
	// ancestor, so the user's divergence is preserved across unlimited
	// syncs instead of being absorbed as the ancestor and overwritten
	// next run. Recording the upstream rendering for a `user-deleted`
	// path also keeps its manifest entry, so a respected deletion is
	// not silently re-created (ADR 0019).
	if result.Conflicts == 0 {
		if err := saveManifestFromUpstream(opts.Root, preset, lang, upstream, isOwned); err != nil {
			return result, err
		}
	} else {
		result.Notes = append(result.Notes,
			"Manifest not updated because conflicts remain. Resolve them and re-run `aikata sync`.")
	}

	return result, nil
}

// saveManifestFromUpstream writes `.aikata/manifest.yaml` from the
// in-memory upstream rendering — the single ancestor-choice principle
// shared by rebaseline, --reseed, and the post-clean-run regeneration
// (ADR 0025 D1). Owned paths (ADR 0025 D2) and the config / manifest
// files themselves are excluded from the recorded set.
func saveManifestFromUpstream(root, preset, lang string, upstream map[string]string, isOwned func(string) bool) error {
	manifestRel := config.PrimaryDir + "/" + config.ManifestFilename
	configRel := config.PrimaryDir + "/" + config.Filename
	excludes := []string{manifestRel, configRel}
	for path := range upstream {
		if isOwned(path) {
			excludes = append(excludes, path)
		}
	}
	newManifest := config.BuildManifest(preset, lang, upstream, excludes)
	if err := config.SaveManifest(root, newManifest); err != nil {
		return fmt.Errorf("sync: save manifest: %w", err)
	}
	return nil
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
