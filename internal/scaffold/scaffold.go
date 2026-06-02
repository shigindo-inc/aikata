package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/managed"
	"github.com/shigindo-inc/aikata/internal/templates"
)

// Options controls a single `aikata init` (and later `aikata add`) run.
type Options struct {
	// ProjectName is interpolated into templates as {{.ProjectName}}.
	ProjectName string
	// Preset is the preset name (e.g. "minimal"). Required.
	Preset string
	// TargetDir is the directory where files will be written. Required.
	TargetDir string
	// Lang is the document language ("en" or "ja"). Defaults to "en".
	Lang string
	// Force allows writing into a non-empty TargetDir.
	Force bool
	// DryRun prints the plan to Stdout without writing.
	DryRun bool
	// WithMemory provisions the optional long-term agent memory slot
	// under docs/memory/ (ADR 0004). Templates branch on {{.WithMemory}}
	// to add memory cross-references when this is true.
	WithMemory bool
	// WithUI, WithAPI, WithTDD, WithChangelog gate the single-file
	// optional components (UI.md / API.md / docs/testing.md /
	// CHANGELOG.md). Each component owns its own rendering; preset
	// templates do not branch on these flags (v0.4.1 ships them as
	// independent files only).
	WithUI        bool
	WithAPI       bool
	WithTDD       bool
	WithChangelog bool
	// WithMonorepo enables the v0.6 monorepo layout: emits
	// `docs/monorepo.md`, `apps/README.md`, and `apps/_example/AGENTS.md`,
	// and flips `features.monorepo` to true in `.aikata/aikata.yaml`.
	// Per-app `AGENTS.md` files are user-managed; aikata does not
	// regenerate them.
	WithMonorepo bool
	// WithPrompts provisions the opt-in reusable-prompt library at
	// docs/prompts.md (ADR 0034). Single-file component; preset
	// templates do not branch on this flag. Off by default — the file
	// was a default scaffold through v0.9.1 and is now opt-in.
	WithPrompts bool
	// WithEnv provisions the opt-in environment-variable template at
	// .env.example (ADR 0037). Single-file component. Off by default —
	// the file was a default scaffold through v0.9.6 and is now opt-in.
	// Note: the scaffolded .gitignore always ignores the `.env` /
	// `.env.local` secret files regardless of this flag (a minimal
	// secret baseline, ADR 0037 D2); only the example file is opt-in.
	WithEnv bool
	// Stacks lists stack identifiers (e.g. "flutter") the project opts
	// into. Templates branch on {{range .Stacks}} to include
	// docs/stacks/<stack>.md cross-references; the values also flow
	// into .aikata/aikata.yaml's `stacks:` field for downstream tools.
	Stacks []string
	// Workflows lists workflow-guide domains (e.g. "git") the project
	// opts into (ADR 0026). Each renders docs/workflows/<domain>.md.
	// There is no init flag for workflows in v0.8.4 — they are enabled
	// post-init via `aikata enable workflow <domain>` — but `aikata sync`
	// populates this from `.aikata/aikata.yaml` `workflows:` so an
	// enabled guide is rendered into the upstream tree and classified
	// like any other aikata-managed document instead of upstream-removed.
	Workflows []string
	// AITools lists the AI-tool identifiers the project enables in its
	// initial `.aikata/aikata.yaml`. Empty defaults to `["claude"]` for
	// backward compatibility with v0.2 init behaviour.
	AITools []string
	// Clock is the time source for template helpers; nil = time.Now.
	Clock templates.Clock
	// Stdout receives dry-run output and progress messages.
	// nil = os.Stdout.
	Stdout io.Writer
}

// proposedDirName is the adoption-fallback subdirectory. When
// `aikata init` runs against a non-empty directory without --force, the
// full scaffold is rendered here instead of the project root so nothing
// existing is touched (SPEC §4.1, ADR 0037 D4).
const proposedDirName = ".aikata-proposed"

// ErrProposalExists signals that the adoption fallback could not write
// because a non-empty `.aikata-proposed/` already exists. aikata refuses
// to overwrite a prior proposal silently; the user reviews and removes it
// first. The cli layer maps this to exit code 2.
var ErrProposalExists = errors.New("scaffold: .aikata-proposed/ already exists and is not empty (review and remove it, or pass --force to write into the project root)")

// Run scaffolds the requested preset into TargetDir. Generation is
// all-or-nothing: every template is rendered into memory first; only if
// every render succeeds do we touch the filesystem.
//
// When TargetDir is non-empty and Force is false, Run does not error:
// it renders the full scaffold under TargetDir/.aikata-proposed/ instead
// (the adoption fallback, SPEC §4.1 / ADR 0037 D4) and prints an
// actionable notice. An existing non-empty proposal directory is refused
// with ErrProposalExists rather than overwritten.
func Run(opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}

	nonEmpty, err := isNonEmpty(opts.TargetDir)
	if err != nil {
		return err
	}
	proposalMode := nonEmpty && !opts.Force

	writeRoot := opts.TargetDir
	if proposalMode {
		writeRoot = filepath.Join(opts.TargetDir, proposedDirName)
	}

	rendered, err := renderInto(opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return printDryRun(opts.Stdout, writeRoot, rendered)
	}

	if proposalMode {
		existing, err := isNonEmpty(writeRoot)
		if err != nil {
			return err
		}
		if existing {
			return ErrProposalExists
		}
	}

	if err := writeAll(writeRoot, rendered); err != nil {
		return err
	}

	if proposalMode {
		return printProposalNotice(opts.Stdout)
	}
	return nil
}

// printProposalNotice tells the user where the adoption fallback wrote
// the proposed scaffold and how to adopt it. Mirrors the steps in
// docs/adoption.md.
func printProposalNotice(w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}
	_, err := fmt.Fprintf(w,
		"Target directory is not empty; wrote the proposed scaffold to %s/ instead.\n"+
			"Review it, merge what you want into the project root, then remove %s/.\n"+
			"(Re-run with --force to write directly into the project root.)\n",
		proposedDirName, proposedDirName)
	return err
}

// Render returns the in-memory file set that Run would write, without
// touching the filesystem and without enforcing the
// "target directory empty" rule. `aikata sync` uses this entry point
// to re-render upstream templates against the current project's
// scaffold options so the merge has a fresh "theirs" side. Force and
// DryRun in opts are ignored by Render; TargetDir is not read or
// written.
func Render(opts Options) (map[string]string, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return renderInto(opts)
}

// renderInto is the shared rendering core for Run and Render. It does
// every step Run performs except the empty-directory precheck and the
// final writeAll / dry-run. The returned map is keyed by
// target-relative path and includes `.aikata/manifest.yaml`.
func renderInto(opts Options) (map[string]string, error) {
	langDir, err := resolveLangDir("presets/"+opts.Preset, opts.Lang, opts.Stdout)
	if err != nil {
		return nil, err
	}
	files, err := listTemplateFiles(langDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("scaffold: preset %q has no templates", opts.Preset)
	}

	rendered, err := renderAll(files, langDir, opts)
	if err != nil {
		return nil, err
	}

	// Preset-specific post-render artifacts that aren't expressed as
	// markdown templates. Keep this list short — when it grows past two
	// or three branches, lift it into internal/presets.
	if err := addPresetArtifacts(opts, rendered); err != nil {
		return nil, err
	}

	// Opt-in optional components. Each renderer lives in
	// internal/components so the init-time path and the corresponding
	// `aikata add <name>` command share one code path. The table grows
	// when a new optional component lands; scaffold itself stays
	// component-agnostic.
	sfp := components.SingleFileParams{
		Lang:        opts.Lang,
		ProjectName: opts.ProjectName,
		Clock:       opts.Clock,
	}
	optionalSpecs := []struct {
		enabled bool
		render  func() (map[string]string, error)
	}{
		{opts.WithMemory, func() (map[string]string, error) {
			return components.RenderMemory(components.MemoryParams{
				Lang:        opts.Lang,
				ProjectName: opts.ProjectName,
				Clock:       opts.Clock,
			})
		}},
		{opts.WithUI, func() (map[string]string, error) { return components.RenderUI(sfp) }},
		{opts.WithAPI, func() (map[string]string, error) { return components.RenderAPI(sfp) }},
		{opts.WithTDD, func() (map[string]string, error) { return components.RenderTDD(sfp) }},
		{opts.WithChangelog, func() (map[string]string, error) { return components.RenderChangelog(sfp) }},
		{opts.WithMonorepo, func() (map[string]string, error) {
			return components.RenderMonorepo(components.MonorepoParams{
				Lang:        opts.Lang,
				ProjectName: opts.ProjectName,
				Clock:       opts.Clock,
			})
		}},
		{opts.WithPrompts, func() (map[string]string, error) { return components.RenderPrompts(sfp) }},
		{opts.WithEnv, func() (map[string]string, error) { return components.RenderEnv(sfp) }},
	}
	for _, spec := range optionalSpecs {
		if !spec.enabled {
			continue
		}
		files, err := spec.render()
		if err != nil {
			return nil, fmt.Errorf("scaffold: %w", err)
		}
		for k, v := range files {
			rendered[k] = v
		}
	}

	// Opt-in workflow guides (ADR 0026). Like stacks this is a list
	// axis, so it lives outside the boolean optionalSpecs table. Each
	// enabled domain renders docs/workflows/<domain>.md via the same
	// helper `aikata enable workflow` uses.
	for _, domain := range opts.Workflows {
		wf, err := components.RenderWorkflowGuide(domain, sfp)
		if err != nil {
			return nil, fmt.Errorf("scaffold: %w", err)
		}
		for k, v := range wf {
			rendered[k] = v
		}
	}

	// Build the sync manifest (.aikata/manifest.yaml) from the
	// just-rendered file set, excluding mutable / circular entries.
	// ADR 0011 D4: `aikata sync` reads this as the common ancestor
	// for the 3-way merge. The manifest itself is added to `rendered`
	// so writeAll and printDryRun treat it like any other emitted
	// file.
	manifestRel := config.PrimaryDir + "/" + config.ManifestFilename
	configRel := config.PrimaryDir + "/" + config.Filename
	manifest := config.BuildManifest(opts.Preset, opts.Lang, rendered, []string{manifestRel, configRel})
	manifestBytes, err := config.MarshalManifest(manifest)
	if err != nil {
		return nil, err
	}
	rendered[manifestRel] = string(manifestBytes)
	return rendered, nil
}

// addPresetArtifacts injects non-template files that a preset is
// expected to ship alongside its markdown set. The standard and
// stack-flavored presets emit a struct-driven `.aikata/aikata.yaml`
// so downstream tooling (aikata generate, doctor) has structured
// config.
func addPresetArtifacts(opts Options, rendered map[string]string) error {
	if opts.Preset == "standard" || opts.Preset == "flutter" || opts.Preset == "typescript" {
		cfg := config.Default(opts.ProjectName, opts.Lang)
		if len(opts.Stacks) > 0 {
			cfg.Stacks = append([]string(nil), opts.Stacks...)
		}
		if len(opts.AITools) > 0 {
			cfg.AITools = append([]string(nil), opts.AITools...)
		}
		// Schema-v2: persist optional components as first-class fields
		// so post-init commands and `aikata sync` read intent from one
		// declarative source instead of inferring from manifest paths.
		cfg.Components.Memory = opts.WithMemory
		cfg.Components.UI = opts.WithUI
		cfg.Components.API = opts.WithAPI
		cfg.Components.TDD = opts.WithTDD
		cfg.Components.Changelog = opts.WithChangelog
		cfg.Components.Monorepo = opts.WithMonorepo
		cfg.Components.Prompts = opts.WithPrompts
		cfg.Components.Env = opts.WithEnv
		buf, err := config.Marshal(cfg)
		if err != nil {
			return err
		}
		rendered[config.PrimaryDir+"/"+config.Filename] = string(buf)
	}
	return nil
}

func (o *Options) validate() error {
	if o.Preset == "" {
		return errors.New("scaffold: preset is required")
	}
	if o.ProjectName == "" {
		return errors.New("scaffold: project name is required")
	}
	if o.TargetDir == "" {
		return errors.New("scaffold: target directory is required")
	}
	if o.Lang == "" {
		o.Lang = "en"
	}
	return nil
}

// listTemplateFiles returns every `*.md.tmpl` file under presetDir in
// the embedded FS, in deterministic order.
func listTemplateFiles(presetDir string) ([]string, error) {
	root, err := templates.FS()
	if err != nil {
		return nil, err
	}
	var out []string
	err = fs.WalkDir(root, presetDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".tmpl") {
			return nil
		}
		out = append(out, p)
		return nil
	})
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("scaffold: unknown preset %q", filepath.Base(presetDir))
		}
		return nil, fmt.Errorf("scaffold: walk %s: %w", presetDir, err)
	}
	sort.Strings(out)
	return out, nil
}

// isNonEmpty reports whether dir exists and contains at least one entry.
// A missing directory counts as empty.
func isNonEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("scaffold: read dir %s: %w", dir, err)
	}
	return len(entries) > 0, nil
}

// templateData is the single source of truth for what {{.foo}} fields
// templates can reference. New fields go here, not inline at the call
// sites.
func templateData(opts Options) map[string]any {
	return map[string]any{
		"ProjectName": opts.ProjectName,
		"Lang":        opts.Lang,
		"Preset":      opts.Preset,
		"WithMemory":  opts.WithMemory,
		"Stacks":      opts.Stacks,
	}
}

// renderAll evaluates every template up front. The returned map is
// keyed by the *target* path relative to TargetDir (e.g. "README.md"),
// not the template path.
func renderAll(files []string, presetDir string, opts Options) (map[string]string, error) {
	data := templateData(opts)
	rendered := make(map[string]string, len(files))
	for _, tmplPath := range files {
		content, err := templates.Render(tmplPath, data, opts.Clock)
		if err != nil {
			return nil, err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(tmplPath, presetDir+"/"), ".tmpl")
		rendered[rel] = content
	}
	return rendered, nil
}

// resolveLangDir picks the language-scoped directory under base via
// templates.LangDir and emits the fallback notice to stdout when the
// requested language is unavailable. Returns the embed path to the
// chosen lang subdirectory (e.g. "presets/standard/en").
func resolveLangDir(base, lang string, stdout io.Writer) (string, error) {
	langDir, fellBack, err := templates.LangDir(base, lang)
	if err != nil {
		return "", fmt.Errorf("scaffold: %w", err)
	}
	if fellBack {
		if stdout == nil {
			stdout = os.Stderr
		}
		if _, werr := fmt.Fprintf(stdout,
			"scaffold: language %q not available for %s, falling back to en\n", lang, base); werr != nil {
			return "", fmt.Errorf("scaffold: write fallback notice: %w", werr)
		}
	}
	return langDir, nil
}

// writeAll persists rendered content to TargetDir. Files inherit 0644
// and intermediate directories are created with 0755. Atomic per-file
// via the standard library; a partial failure may leave some files
// already written (acceptable for v0.1 MVP, see SPEC §5.1).
//
// Special-case `.gitignore` (and similar files governed by ADR 0018):
// when the target already exists, the file is preserved and the
// aikata-owned portion is merged in place via the managed-block
// writer. Other paths use the simple overwrite path. The managed list
// is intentionally narrow today — extend it only after the merge
// rules are pinned for each new target (ROADMAP v0.7.x).
func writeAll(targetDir string, rendered map[string]string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("scaffold: mkdir %s: %w", targetDir, err)
	}
	// Iterate in deterministic order so any partial failure leaves a
	// predictable on-disk state.
	keys := sortedKeys(rendered)
	for _, rel := range keys {
		full := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("scaffold: mkdir %s: %w", filepath.Dir(full), err)
		}
		body, werr := contentForWrite(full, rel, rendered[rel])
		if werr != nil {
			return fmt.Errorf("scaffold: prepare %s: %w", full, werr)
		}
		if err := os.WriteFile(full, body, 0o644); err != nil {
			return fmt.Errorf("scaffold: write %s: %w", full, err)
		}
	}
	return nil
}

// contentForWrite returns the bytes writeAll should persist for one
// rendered (path, content) pair. Managed-append paths (today
// `.gitignore`, see managed.IsAppendPath) always carry the aikata
// marker block on disk so init and `aikata sync` share one
// representation (ADR 0038): a fresh write is the framed standalone
// block, and an existing file is merged in place so the user's content
// outside the markers survives (ADR 0018). For everything else the
// rendered content is returned verbatim.
func contentForWrite(fullPath, rel, rendered string) ([]byte, error) {
	if !managed.IsAppendPath(rel) {
		return []byte(rendered), nil
	}
	existing, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Fresh write — frame the block so the on-disk file always
			// carries the markers (the canonical representation `aikata
			// sync` refreshes; ADR 0038).
			return managed.Frame([]byte(rendered)), nil
		}
		return nil, fmt.Errorf("read existing %s: %w", fullPath, err)
	}
	merged, err := managed.ApplyBlock(existing, []byte(rendered))
	if err != nil {
		return nil, fmt.Errorf("merge managed block in %s: %w", fullPath, err)
	}
	return merged, nil
}

func printDryRun(w io.Writer, target string, rendered map[string]string) error {
	if w == nil {
		w = os.Stdout
	}
	if _, err := fmt.Fprintf(w, "Would write %d file(s) under %s:\n", len(rendered), target); err != nil {
		return err
	}
	for _, rel := range sortedKeys(rendered) {
		if _, err := fmt.Fprintf(w, "  %s\n", rel); err != nil {
			return err
		}
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
