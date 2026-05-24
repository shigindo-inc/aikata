package components

import (
	"fmt"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// singleFile is the shared implementation behind the four v0.4.1
// optional components (ui / api / tdd / changelog). Each one renders
// a single template file to a fixed target path under the project
// root, refuses to clobber an existing file, and is idempotent on
// re-run. Components that emit multiple files (memory) or no files
// at all (ai-tool) use their own implementation.
type singleFile struct {
	name       string // component name as seen by `aikata add`
	desc       string // one-line description shown in `aikata list components`
	targetPath string // relative to TargetDir, e.g. "UI.md" or "docs/testing.md"
	tmplBase   string // embed base, e.g. "components/ui" (lang appended at render time)
	tmplName   string // template file name under <tmplBase>/<lang>/, e.g. "ui.md.tmpl"
}

func (s singleFile) Name() string        { return s.name }
func (s singleFile) Description() string { return s.desc }
func (s singleFile) Status() string      { return StatusActive }

// Add renders the template for ctx.Lang and writes it to targetPath
// under ctx.TargetDir when no file already exists at that path.
// Re-running on a project where targetPath is already present prints
// a notice on Stderr and returns nil (the user may have customized
// the file; we never overwrite).
func (s singleFile) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: %s: project name is required", s.name)
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: %s: target directory is required", s.name)
	}
	rendered, err := s.render(ctx.Lang, ctx.ProjectName, ctx.Clock)
	if err != nil {
		return err
	}

	if ctx.DryRun {
		if _, werr := fmt.Fprintf(stdout(ctx), "Would write %s\n", s.targetPath); werr != nil {
			return werr
		}
		return nil
	}

	written, _, err := writeIfMissing(ctx.TargetDir, rendered)
	if err != nil {
		return err
	}
	if written == 0 {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: %s already present at %s; nothing to do\n", s.name, s.targetPath); werr != nil {
			return werr
		}
		return nil
	}
	if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", s.targetPath); werr != nil {
		return werr
	}
	return nil
}

// render produces the map[targetPath]content shape that both Add and
// the scaffold init-time path consume. The init-time path reaches
// this through the per-component Render<Name> shims exported by
// ui.go / api.go / tdd.go / changelog.go.
func (s singleFile) render(lang, projectName string, clock templates.Clock) (map[string]string, error) {
	if projectName == "" {
		return nil, fmt.Errorf("components: %s: project name is required", s.name)
	}
	langDir, _, err := templates.LangDir(s.tmplBase, lang)
	if err != nil {
		return nil, fmt.Errorf("components: %s: %w", s.name, err)
	}
	tmplPath := langDir + "/" + s.tmplName
	data := map[string]any{
		"ProjectName": projectName,
		"Lang":        lang,
	}
	content, err := templates.Render(tmplPath, data, clock)
	if err != nil {
		return nil, fmt.Errorf("components: %s: render %s: %w", s.name, tmplPath, err)
	}
	return map[string]string{s.targetPath: content}, nil
}
