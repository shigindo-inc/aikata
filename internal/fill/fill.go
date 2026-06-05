// Package fill implements `aikata fill`: top up a repository's canonical
// document set by writing only the files that are missing, never
// overwriting anything that already exists.
//
// fill is the low-friction adoption / completion verb. It serves the gap
// that neither `init`, `sync`, nor `enable` covers (see Q-INTEROP-03 and
// ADR 0042):
//
//   - `init` scaffolds a NEW project; in a non-empty directory it diverts
//     the whole tree to `.aikata-proposed/` for manual merge.
//   - `sync` pulls upstream template *changes* and deliberately respects
//     user deletions (a doc the user removed is not restored).
//   - `enable` adds a single *capability*, and requires an existing
//     `.aikata/aikata.yaml`.
//
// fill instead makes the repository a complete aikata project: it renders
// the canonical set the project's scope defines and writes every file
// that is absent, leaving hand-edited files untouched. It is option-free
// and idempotent — a second run on a complete project writes nothing.
//
// Scope resolution (no CLI overrides, so simpler than sync.derivePlan):
//
//   - Managed/partial project (a `.aikata/manifest.yaml` and/or
//     `aikata.yaml` is present): preset/lang come from the manifest when
//     present (else default), and the optional-component / stack / AI-tool
//     set comes from `aikata.yaml`.
//   - Unmanaged repository (no `.aikata/`): default to scope=standard,
//     lang=en, project name = the working-directory basename. standard is
//     init's own default; fill is non-destructive, so unwanted documents
//     can simply be deleted afterward.
//
// The manifest is (re)built from the rendered (upstream) hashes via
// components.RecordInManifest, which merges into any existing manifest. A
// hand-written file's ancestor is therefore recorded as the *upstream*
// rendering, so the next `aikata sync` classifies the user's content as
// `user-only-edit` and preserves it rather than overwriting. Because fill
// writes every missing rendered file, nothing is recorded-but-absent, so
// the "absent reads as user-deleted" footgun of a bare rebaseline cannot
// occur here.
package fill

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/managed"
	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// Options controls one fill invocation. Root must be the project
// directory. fill takes no behavioural flags by design (the command is
// zero-config); Stdout receives the human-readable summary.
type Options struct {
	Root   string
	Stdout io.Writer
	Stderr io.Writer
}

// Result reports what one fill invocation did. Paths are project-relative
// and slash-separated, sorted for stable output.
type Result struct {
	// Managed is true when the project already had aikata config or a
	// manifest before this run (i.e. fill topped up an existing project
	// rather than adopting an unmanaged repository).
	Managed bool
	// Preset / Lang are the resolved scaffold scope fill rendered.
	Preset string
	Lang   string
	// Written lists files fill created; Skipped lists rendered files that
	// already existed and were left untouched.
	Written []string
	Skipped []string
}

// Run performs one fill invocation against opts.Root.
func Run(opts Options) (Result, error) {
	if opts.Root == "" {
		return Result{}, errors.New("fill: Root is required")
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	scaffoldOpts, isManaged, err := resolvePlan(opts.Root)
	if err != nil {
		return Result{}, err
	}
	// Route any scaffold fallback notices (e.g. missing ja templates) to
	// stderr so they don't pollute the summary.
	scaffoldOpts.Stdout = opts.Stderr

	rendered, err := scaffold.Render(scaffoldOpts)
	if err != nil {
		return Result{}, fmt.Errorf("fill: render scaffold: %w", err)
	}

	// scaffold.Render bakes in `.aikata/manifest.yaml`; we rebuild the
	// manifest ourselves (merge-safe) via RecordInManifest, so drop the
	// rendered copy. `.aikata/aikata.yaml` stays in the write set so an
	// unmanaged repo gets a config written (WriteIfMissing skips it when a
	// managed project already has one).
	manifestRel := config.PrimaryDir + "/" + config.ManifestFilename
	configRel := config.PrimaryDir + "/" + config.Filename
	delete(rendered, manifestRel)

	// Managed-append paths (today `.gitignore`, ADR 0018/0038) are NOT
	// write-if-missing: their aikata marker block is merged in place even
	// when the file already exists, because the merge is non-destructive
	// (user lines outside the markers survive). Writing them via
	// scaffold.ContentForWrite — the same path `init` uses — keeps the
	// on-disk representation identical to init's (framed) and adds aikata's
	// ignore entries to an adopted repo's existing .gitignore at fill time
	// rather than deferring them to the next sync. The remaining documents
	// use write-if-missing so a hand-edited doc is never touched.
	managedWritten, managedSkipped, err := applyManagedAppend(opts.Root, rendered)
	if err != nil {
		return Result{}, err
	}
	normal := make(map[string]string, len(rendered))
	for path, body := range rendered {
		if !managed.IsAppendPath(path) {
			normal[path] = body
		}
	}

	// Partition by on-disk presence BEFORE writing — this is the only
	// reliable way to attribute each path to "written" vs "skipped".
	// (After WriteIfMissing every rendered path exists, and a pre-existing
	// file whose bytes happen to equal the upstream rendering would be
	// indistinguishable from one fill just wrote.)
	written, skipped, err := partitionByPresence(opts.Root, normal)
	if err != nil {
		return Result{}, err
	}
	written = append(written, managedWritten...)
	skipped = append(skipped, managedSkipped...)
	sort.Strings(written)
	sort.Strings(skipped)

	if _, _, err := components.WriteIfMissing(opts.Root, normal); err != nil {
		return Result{}, fmt.Errorf("fill: %w", err)
	}

	// Record the canonical document set (everything except the config
	// file itself, which the manifest never tracks) so a subsequent
	// `aikata sync` recognises these as aikata-managed. RecordInManifest
	// requires `aikata.yaml` to exist; WriteIfMissing above has already
	// created it for a freshly-adopted repo, so this is safe to call now.
	docSet := make(map[string]string, len(rendered))
	for path, body := range rendered {
		if path == configRel {
			continue
		}
		docSet[path] = body
	}
	if err := components.RecordInManifest(opts.Root, docSet); err != nil {
		return Result{}, fmt.Errorf("fill: %w", err)
	}

	res := Result{
		Managed: isManaged,
		Preset:  scaffoldOpts.Preset,
		Lang:    scaffoldOpts.Lang,
		Written: written,
		Skipped: skipped,
	}
	if err := printSummary(opts.Stdout, res); err != nil {
		return res, err
	}
	return res, nil
}

// applyManagedAppend writes the managed-append entries (e.g. `.gitignore`)
// in rendered, merging the aikata marker block in place via
// scaffold.ContentForWrite. A path whose merged content equals what is
// already on disk is reported as skipped (idempotent); otherwise it is
// written (created or block-refreshed). Non-managed paths are ignored.
func applyManagedAppend(root string, rendered map[string]string) (written, skipped []string, err error) {
	for path, body := range rendered {
		if !managed.IsAppendPath(path) {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(path))
		before, readErr := os.ReadFile(full)
		existed := readErr == nil
		if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
			return nil, nil, fmt.Errorf("fill: read %s: %w", path, readErr)
		}
		out, cErr := scaffold.ContentForWrite(full, path, body)
		if cErr != nil {
			return nil, nil, fmt.Errorf("fill: %w", cErr)
		}
		if existed && bytes.Equal(before, out) {
			skipped = append(skipped, path)
			continue
		}
		if mkErr := os.MkdirAll(filepath.Dir(full), 0o755); mkErr != nil {
			return nil, nil, fmt.Errorf("fill: mkdir %s: %w", filepath.Dir(full), mkErr)
		}
		if wErr := os.WriteFile(full, out, 0o644); wErr != nil {
			return nil, nil, fmt.Errorf("fill: write %s: %w", path, wErr)
		}
		written = append(written, path)
	}
	return written, skipped, nil
}

// partitionByPresence sorts rendered paths into those absent on disk
// (will be written) and those already present (will be skipped),
// inspecting the filesystem before any write happens. Returned slices are
// sorted for stable output.
func partitionByPresence(root string, rendered map[string]string) (written, skipped []string, err error) {
	for path := range rendered {
		full := filepath.Join(root, filepath.FromSlash(path))
		_, statErr := os.Stat(full)
		switch {
		case statErr == nil:
			skipped = append(skipped, path)
		case errors.Is(statErr, fs.ErrNotExist):
			written = append(written, path)
		default:
			return nil, nil, fmt.Errorf("fill: stat %s: %w", path, statErr)
		}
	}
	sort.Strings(written)
	sort.Strings(skipped)
	return written, skipped, nil
}

// resolvePlan derives the scaffold.Options fill should render and reports
// whether the project was already aikata-managed. See the package doc for
// the resolution hierarchy.
func resolvePlan(root string) (scaffold.Options, bool, error) {
	preset := "standard"
	lang := "en"
	name := filepath.Base(root)

	manifest, mErr := config.LoadManifest(root)
	if mErr != nil && !errors.Is(mErr, fs.ErrNotExist) {
		return scaffold.Options{}, false, fmt.Errorf("fill: load manifest: %w", mErr)
	}
	hasManifest := mErr == nil

	cfg, cErr := config.Load(root)
	if cErr != nil && !errors.Is(cErr, fs.ErrNotExist) {
		return scaffold.Options{}, false, fmt.Errorf("fill: load config: %w", cErr)
	}
	hasConfig := cErr == nil

	managed := hasManifest || hasConfig

	if hasManifest {
		if manifest.Preset != "" {
			preset = manifest.Preset
		}
		if manifest.Lang != "" {
			lang = manifest.Lang
		}
	}

	opts := scaffold.Options{
		Preset:    preset,
		TargetDir: root,
		Lang:      lang,
	}
	if hasConfig {
		if cfg.Project.Name != "" {
			name = cfg.Project.Name
		}
		// When there is no manifest to anchor the language, honour the
		// config's project.lang before falling back to the default.
		if !hasManifest && cfg.Project.Lang != "" {
			opts.Lang = cfg.Project.Lang
		}
		opts.Stacks = append([]string(nil), cfg.Stacks...)
		opts.Workflows = append([]string(nil), cfg.Workflows...)
		opts.AITools = append([]string(nil), cfg.AITools...)
		opts.WithMemory = cfg.Components.Memory
		opts.WithUI = cfg.Components.UI
		opts.WithAPI = cfg.Components.API
		opts.WithTDD = cfg.Components.TDD
		opts.WithChangelog = cfg.Components.Changelog
		opts.WithMonorepo = cfg.Components.Monorepo
		opts.WithPrompts = cfg.Components.Prompts
		opts.WithEnv = cfg.Components.Env
		// Honour the pre-v2 features map too, mirroring sync.derivePlan,
		// so legacy projects that predate the components.* schema still
		// re-render their opted-in docs.
		if cfg.Features != nil {
			if cfg.Features["monorepo"] {
				opts.WithMonorepo = true
			}
			if cfg.Features["tdd"] {
				opts.WithTDD = true
			}
		}
	}
	opts.ProjectName = name
	return opts, managed, nil
}

// printSummary writes a human-readable report of what fill did.
func printSummary(w io.Writer, res Result) error {
	if len(res.Written) == 0 {
		_, err := fmt.Fprintf(w,
			"Nothing to fill: the %s document set is already complete (%d file(s) present).\n",
			res.Preset, len(res.Skipped))
		return err
	}

	lead := "Filled in"
	if !res.Managed {
		lead = fmt.Sprintf("Adopted this repository as an aikata project (scope: %s) and filled in", res.Preset)
	}
	if _, err := fmt.Fprintf(w, "%s %d missing file(s); left %d existing file(s) untouched:\n",
		lead, len(res.Written), len(res.Skipped)); err != nil {
		return err
	}
	for _, p := range res.Written {
		if _, err := fmt.Fprintf(w, "  + %s\n", p); err != nil {
			return err
		}
	}
	if !res.Managed {
		if _, err := fmt.Fprintf(w,
			"\nReview the new documents and prune any that don't fit this project.\n"+
				"Run `aikata generate` to emit per-AI-tool configs (CLAUDE.md, …).\n"); err != nil {
			return err
		}
	}
	return nil
}
