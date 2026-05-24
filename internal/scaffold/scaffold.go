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

	"github.com/shigindo-inc/aikata/internal/config"
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
	// Stacks lists stack identifiers (e.g. "flutter") the project opts
	// into. Templates branch on {{range .Stacks}} to include
	// docs/stacks/<stack>.md cross-references; the values also flow
	// into .ai/aikata.yaml's `stacks:` field for downstream tools.
	Stacks []string
	// AITools lists the AI-tool identifiers the project enables in its
	// initial `.ai/aikata.yaml`. Empty defaults to `["claude"]` for
	// backward compatibility with v0.2 init behaviour.
	AITools []string
	// Clock is the time source for template helpers; nil = time.Now.
	Clock templates.Clock
	// Stdout receives dry-run output and progress messages.
	// nil = os.Stdout.
	Stdout io.Writer
}

// ErrTargetDirNotEmpty signals that TargetDir already contains files and
// Force is false. Callers (the cli layer) map this to exit code 2 per
// ARCHITECTURE.md §7.2.
var ErrTargetDirNotEmpty = errors.New("scaffold: target directory is not empty (use --force to overwrite)")

// Run scaffolds the requested preset into TargetDir. Generation is
// all-or-nothing: every template is rendered into memory first; only if
// every render succeeds do we touch the filesystem.
func Run(opts Options) error {
	if err := opts.validate(); err != nil {
		return err
	}

	presetDir, langDir, err := resolveLangDir("presets/"+opts.Preset, opts.Lang, opts.Stdout)
	if err != nil {
		return err
	}
	files, err := listTemplateFiles(langDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("scaffold: preset %q has no templates", opts.Preset)
	}
	// renderAll trims the lang-scoped prefix (e.g. "presets/minimal/en/")
	// to compute target paths, so files render to the same location
	// regardless of language.
	_ = presetDir

	nonEmpty, err := isNonEmpty(opts.TargetDir)
	if err != nil {
		return err
	}
	if nonEmpty && !opts.Force {
		return ErrTargetDirNotEmpty
	}

	rendered, err := renderAll(files, langDir, opts)
	if err != nil {
		return err
	}

	// Preset-specific post-render artifacts that aren't expressed as
	// markdown templates. Keep this list short — when it grows past two
	// or three branches, lift it into internal/presets.
	if err := addPresetArtifacts(opts, rendered); err != nil {
		return err
	}

	// Opt-in long-term memory slot (ADR 0004). Adds docs/memory/*.md
	// independent of preset choice.
	if err := addMemoryArtifacts(opts, rendered); err != nil {
		return err
	}

	if opts.DryRun {
		return printDryRun(opts.Stdout, opts.TargetDir, rendered)
	}

	return writeAll(opts.TargetDir, rendered)
}

// addPresetArtifacts injects non-template files that a preset is
// expected to ship alongside its markdown set. The standard and
// stack-flavored presets emit a struct-driven `.aikata/aikata.yaml`
// so downstream tooling (aikata generate, doctor) has structured
// config. The path moved from `.ai/` to `.aikata/` in v0.3.2 per
// ADR 0008; existing projects keep reading from `.ai/` via the
// resolver in internal/config.
func addPresetArtifacts(opts Options, rendered map[string]string) error {
	if opts.Preset == "standard" || opts.Preset == "flutter" || opts.Preset == "typescript" {
		cfg := config.Default(opts.ProjectName, opts.Lang)
		if len(opts.Stacks) > 0 {
			cfg.Stacks = append([]string(nil), opts.Stacks...)
		}
		if len(opts.AITools) > 0 {
			cfg.AITools = append([]string(nil), opts.AITools...)
		}
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

// addMemoryArtifacts renders the docs/memory/*.md slot when WithMemory
// is set. The memory templates live under `memory/<lang>/` in the
// embedded FS (separate from `presets/`) so they compose with every
// preset and language.
//
// Do-No-Harm (ADR 0003 / 0004): when WithMemory is false this is a
// no-op and no memory references appear in the rendered output.
func addMemoryArtifacts(opts Options, rendered map[string]string) error {
	if !opts.WithMemory {
		return nil
	}
	root, err := templates.FS()
	if err != nil {
		return err
	}
	_, langDir, err := resolveLangDir("memory", opts.Lang, opts.Stdout)
	if err != nil {
		return err
	}
	data := templateData(opts)
	return fs.WalkDir(root, langDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".tmpl") {
			return nil
		}
		content, err := templates.Render(p, data, opts.Clock)
		if err != nil {
			return err
		}
		rel := "docs/memory/" + strings.TrimSuffix(strings.TrimPrefix(p, langDir+"/"), ".tmpl")
		rendered[rel] = content
		return nil
	})
}

// resolveLangDir picks the language-scoped directory under base for
// Lang, falling back to "en" when Lang is empty or its directory is
// missing in the embedded FS. The fallback is reported via stdout (if
// supplied) so users discovering the missing translation see it.
//
// Returns (base, base+"/"+chosenLang).
func resolveLangDir(base, lang string, stdout io.Writer) (string, string, error) {
	root, err := templates.FS()
	if err != nil {
		return "", "", err
	}
	chosen := lang
	if chosen == "" {
		chosen = "en"
	}
	target := base + "/" + chosen
	if _, err := fs.Stat(root, target); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return "", "", fmt.Errorf("scaffold: stat %s: %w", target, err)
		}
		// Fall back to "en"; surface the substitution so users see it.
		if stdout == nil {
			stdout = os.Stderr
		}
		if chosen != "en" {
			if _, werr := fmt.Fprintf(stdout, "scaffold: language %q not available for %s, falling back to en\n", chosen, base); werr != nil {
				return "", "", fmt.Errorf("scaffold: write fallback notice: %w", werr)
			}
		}
		target = base + "/en"
		if _, err := fs.Stat(root, target); err != nil {
			return "", "", fmt.Errorf("scaffold: %s/en not found: %w", base, err)
		}
	}
	return base, target, nil
}

// writeAll persists rendered content to TargetDir. Files inherit 0644
// and intermediate directories are created with 0755. Atomic per-file
// via the standard library; a partial failure may leave some files
// already written (acceptable for v0.1 MVP, see SPEC §5.1).
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
		if err := os.WriteFile(full, []byte(rendered[rel]), 0o644); err != nil {
			return fmt.Errorf("scaffold: write %s: %w", full, err)
		}
	}
	return nil
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
