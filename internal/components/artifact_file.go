package components

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// oneOffArtifact is the shared implementation behind one-off
// `aikata new` artifacts that render a single template to a fixed
// target path (app-icon, mascot). It differs from singleFile
// capabilities in two deliberate ways (ADR 0031 D2):
//
//   - it refuses to clobber an existing file (errors rather than
//     silently skipping) — a one-off artifact is stamped once and then
//     rewritten heavily by the project, so a second `new` is almost
//     always a mistake;
//   - it records nothing in `.aikata/manifest.yaml` and flips no
//     schema-v2 flag, so after stamping the project owns the file and
//     `aikata sync` never restores or merges it.
type oneOffArtifact struct {
	name       string // artifact name as seen by `aikata new`
	desc       string // one-line description for `aikata list artifacts`
	targetPath string // relative to TargetDir, e.g. "docs/design/app-icon-concepts.md"
	tmplBase   string // embed base, e.g. "components/app-icon" (lang appended at render time)
	tmplName   string // template file name under <tmplBase>/<lang>/
}

func (a oneOffArtifact) Name() string        { return a.name }
func (a oneOffArtifact) Description() string { return a.desc }
func (a oneOffArtifact) Status() string      { return StatusActive }

// Add renders the artifact template for ctx.Lang and writes it to
// targetPath under ctx.TargetDir, refusing to overwrite an existing
// file. Failure leaves the filesystem unchanged.
func (a oneOffArtifact) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: %s: project name is required", a.name)
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: %s: target directory is required", a.name)
	}

	full := filepath.Join(ctx.TargetDir, filepath.FromSlash(a.targetPath))
	if _, statErr := os.Stat(full); statErr == nil {
		return fmt.Errorf(
			"components: %s: %s already exists (a one-off artifact is stamped once and then project-owned; remove or rename it to regenerate)",
			a.name, a.targetPath)
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("components: %s: stat %s: %w", a.name, full, statErr)
	}

	langDir, _, err := templates.LangDir(a.tmplBase, ctx.Lang)
	if err != nil {
		return fmt.Errorf("components: %s: %w", a.name, err)
	}
	tmplPath := langDir + "/" + a.tmplName
	data := map[string]any{
		"ProjectName": ctx.ProjectName,
		"Lang":        ctx.Lang,
	}
	content, err := templates.Render(tmplPath, data, ctx.Clock)
	if err != nil {
		return fmt.Errorf("components: %s: render %s: %w", a.name, tmplPath, err)
	}

	if ctx.DryRun {
		if _, werr := fmt.Fprintf(stdout(ctx), "Would write %s\n", a.targetPath); werr != nil {
			return werr
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("components: %s: mkdir %s: %w", a.name, filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		return fmt.Errorf("components: %s: write %s: %w", a.name, full, err)
	}
	if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", a.targetPath); werr != nil {
		return werr
	}
	return nil
}
