package components

import (
	"fmt"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// modelingComponent provides the opt-in document pair docs/usecases.md
// (behaviour) + docs/domain.md (structure). The two are one capability
// because they are read and edited as a pair: docs/domain.md links back
// to use-case IDs per field, so half the pair cannot discharge either
// side of that check.
//
// It cannot use singleFile (single targetPath by construction), so it
// follows memoryComponent's multi-file shape with fixed paths rather
// than a template-tree walk.
type modelingComponent struct{}

// Modeling is the singleton registered in the capabilities registry.
var Modeling Component = modelingComponent{}

func (modelingComponent) Name() string { return "modeling" }
func (modelingComponent) Description() string {
	return "Use-case ledger and domain model at docs/usecases.md + docs/domain.md."
}
func (modelingComponent) Status() string { return StatusActive }

// modelingFiles maps the target path to its template file name under
// components/modeling/<lang>/.
var modelingFiles = map[string]string{
	"docs/usecases.md": "usecases.md.tmpl",
	"docs/domain.md":   "domain.md.tmpl",
}

// ModelingParams carries the inputs RenderModeling needs. Reduced form
// of scaffold.Options so the init-time and enable-time paths share one
// renderer.
type ModelingParams struct {
	Lang        string
	ProjectName string
	Clock       templates.Clock
}

// RenderModeling returns the pair keyed by target path relative to the
// project root ("docs/usecases.md", "docs/domain.md").
func RenderModeling(p ModelingParams) (map[string]string, error) {
	if p.ProjectName == "" {
		return nil, fmt.Errorf("components: modeling: project name is required")
	}
	langDir, _, err := templates.LangDir("components/modeling", p.Lang)
	if err != nil {
		return nil, fmt.Errorf("components: modeling: %w", err)
	}
	data := map[string]any{
		"ProjectName": p.ProjectName,
		"Lang":        p.Lang,
	}
	out := make(map[string]string, len(modelingFiles))
	for rel, tmplName := range modelingFiles {
		tmplPath := langDir + "/" + tmplName
		content, rerr := templates.Render(tmplPath, data, p.Clock)
		if rerr != nil {
			return nil, fmt.Errorf("components: modeling: render %s: %w", tmplPath, rerr)
		}
		out[rel] = content
	}
	return out, nil
}

// Add renders the pair under ctx.TargetDir. Existing files are
// preserved, so re-running is idempotent and never clobbers a
// hand-edited document.
func (m modelingComponent) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: modeling: project name is required")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: modeling: target directory is required")
	}
	rendered, err := RenderModeling(ModelingParams{
		Lang:        ctx.Lang,
		ProjectName: ctx.ProjectName,
		Clock:       ctx.Clock,
	})
	if err != nil {
		return err
	}

	if ctx.DryRun {
		for _, rel := range sortedKeys(rendered) {
			if _, werr := fmt.Fprintf(stdout(ctx), "Would write %s\n", rel); werr != nil {
				return werr
			}
		}
		return nil
	}

	written, skipped, err := WriteIfMissing(ctx.TargetDir, rendered)
	if err != nil {
		return err
	}
	// ADR 0014: register both paths so the next `aikata sync` treats
	// them as aikata-managed templates.
	if err := RecordInManifest(ctx.TargetDir, rendered); err != nil {
		return err
	}
	if err := EnableComponentInConfig(ctx.TargetDir, "modeling"); err != nil {
		return err
	}
	if written == 0 {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: modeling already present (%d file(s)); nothing to do\n", skipped); werr != nil {
			return werr
		}
		return nil
	}
	for _, rel := range sortedKeys(rendered) {
		if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", rel); werr != nil {
			return werr
		}
	}
	return nil
}
