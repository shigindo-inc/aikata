package components

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// memoryComponent provides the docs/memory/ scaffolding (ADR 0004).
// The same renderer is invoked from `aikata init --with-memory`
// (preset-time) and `aikata add memory` (post-init), so there is no
// path where the two flows can drift.
type memoryComponent struct{}

// Memory is the singleton instance registered in the components
// registry. Exported so scaffold.Run can reach RenderMemory through
// the same value the registry uses.
var Memory Component = memoryComponent{}

func (memoryComponent) Name() string { return "memory" }
func (memoryComponent) Description() string {
	return "Long-term agent memory under docs/memory/ (ADR 0004)."
}
func (memoryComponent) Status() string { return StatusActive }

// Add renders the memory templates for ctx.Lang under
// ctx.TargetDir/docs/memory/. Existing files are preserved so the
// command is idempotent — re-running prints a notice on Stderr and
// returns nil. Failure leaves any files already written in place
// (best-effort; rerun is safe).
func (memoryComponent) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: memory: project name is required")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: memory: target directory is required")
	}
	rendered, err := RenderMemory(MemoryParams{
		Lang:        ctx.Lang,
		ProjectName: ctx.ProjectName,
		Clock:       ctx.Clock,
	})
	if err != nil {
		return err
	}

	if ctx.DryRun {
		return printMemoryPlan(ctx.Stdout, ctx.TargetDir, rendered)
	}

	written, skipped, err := writeIfMissing(ctx.TargetDir, rendered)
	if err != nil {
		return err
	}
	if written == 0 {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: memory already present (%d file(s) under docs/memory/); nothing to do\n", skipped); werr != nil {
			return werr
		}
	}
	return nil
}

// MemoryParams carries the inputs RenderMemory needs. It is a
// reduced form of scaffold.Options so the two call paths (init-time
// and add-time) share one renderer.
type MemoryParams struct {
	// Lang selects the locale subdirectory (templates/data/memory/<lang>/).
	// Empty / unknown values fall back to "en".
	Lang string
	// ProjectName populates {{.ProjectName}} in memory templates.
	// Required.
	ProjectName string
	// Clock supplies the template helper {{now}}. nil = time.Now.
	Clock templates.Clock
}

// RenderMemory returns docs/memory/*.md keyed by target path
// (relative to project root, e.g. "docs/memory/user.md"). The keys
// match what scaffold.Run expects, so the init-time integration is a
// direct map-merge. Callers that want to surface a lang fallback
// notice should call templates.LangDir("memory", lang) first.
func RenderMemory(p MemoryParams) (map[string]string, error) {
	if p.ProjectName == "" {
		return nil, fmt.Errorf("components: memory: project name is required")
	}
	root, err := templates.FS()
	if err != nil {
		return nil, err
	}
	langDir, _, err := templates.LangDir("memory", p.Lang)
	if err != nil {
		return nil, err
	}
	data := map[string]any{
		"ProjectName": p.ProjectName,
		"Lang":        p.Lang,
	}
	out := map[string]string{}
	err = fs.WalkDir(root, langDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		content, rerr := templates.Render(path, data, p.Clock)
		if rerr != nil {
			return rerr
		}
		rel := "docs/memory/" + strings.TrimSuffix(strings.TrimPrefix(path, langDir+"/"), ".tmpl")
		out[rel] = content
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("components: memory: render %s: %w", langDir, err)
	}
	return out, nil
}

// writeIfMissing writes each (rel, content) pair under targetDir only
// when the destination does not already exist. Returns counts of
// (written, skipped). Intermediate directories are created with 0755.
func writeIfMissing(targetDir string, rendered map[string]string) (written, skipped int, err error) {
	for rel, content := range rendered {
		full := filepath.Join(targetDir, filepath.FromSlash(rel))
		if _, statErr := os.Stat(full); statErr == nil {
			skipped++
			continue
		} else if !os.IsNotExist(statErr) {
			return written, skipped, fmt.Errorf("components: stat %s: %w", full, statErr)
		}
		if mkErr := os.MkdirAll(filepath.Dir(full), 0o755); mkErr != nil {
			return written, skipped, fmt.Errorf("components: mkdir %s: %w", filepath.Dir(full), mkErr)
		}
		if wErr := os.WriteFile(full, []byte(content), 0o644); wErr != nil {
			return written, skipped, fmt.Errorf("components: write %s: %w", full, wErr)
		}
		written++
	}
	return written, skipped, nil
}

func printMemoryPlan(w io.Writer, targetDir string, rendered map[string]string) error {
	if w == nil {
		w = os.Stdout
	}
	keys := sortedKeys(rendered)
	if _, err := fmt.Fprintf(w, "Would write %d file(s) under %s:\n", len(keys), targetDir); err != nil {
		return err
	}
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "  %s\n", k); err != nil {
			return err
		}
	}
	return nil
}
