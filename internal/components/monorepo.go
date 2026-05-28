package components

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// monorepoComponent provides the v0.6 monorepo scaffolding under
// `docs/monorepo.md` plus the `apps/` skeleton. The same renderer is
// invoked from `aikata init --monorepo` (preset-time) and
// `aikata enable monorepo` (post-init), so the two flows cannot
// drift. Schema-v2: enabling flips `components.monorepo` in
// .aikata/aikata.yaml (ADR 0016).
type monorepoComponent struct{}

// Monorepo is the singleton instance registered in the capabilities
// registry.
var Monorepo Component = monorepoComponent{}

func (monorepoComponent) Name() string { return "monorepo" }
func (monorepoComponent) Description() string {
	return "v0.6 monorepo layout: docs/monorepo.md + apps/ skeleton."
}
func (monorepoComponent) Status() string { return StatusActive }

// Add renders the monorepo files for ctx.Lang under ctx.TargetDir.
// Existing files are preserved so the command is idempotent. The
// schema-v2 `components.monorepo` flag is flipped to true.
func (monorepoComponent) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: monorepo: project name is required")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: monorepo: target directory is required")
	}
	rendered, err := RenderMonorepo(MonorepoParams{
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
	if err := RecordInManifest(ctx.TargetDir, rendered); err != nil {
		return err
	}
	if err := EnableComponentInConfig(ctx.TargetDir, "monorepo"); err != nil {
		return err
	}
	if written == 0 {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: monorepo already present (%d file(s) under docs/ and apps/); nothing to do\n", skipped); werr != nil {
			return werr
		}
	}
	return nil
}

// MonorepoParams carries the inputs RenderMonorepo needs. Mirrors the
// reduced shape used by MemoryParams / SingleFileParams.
type MonorepoParams struct {
	Lang        string
	ProjectName string
	Clock       templates.Clock
}

// RenderMonorepo returns the monorepo scaffold files keyed by their
// target-relative path: `docs/monorepo.md`, `apps/README.md`, and the
// per-app rule template at `apps/_example/AGENTS.md`. The init-time
// `--monorepo` flag calls this; a future `aikata add monorepo` (if it
// ever exists) would reuse the same renderer so the two flows can't
// drift.
func RenderMonorepo(p MonorepoParams) (map[string]string, error) {
	if p.ProjectName == "" {
		return nil, fmt.Errorf("components: monorepo: project name is required")
	}
	root, err := templates.FS()
	if err != nil {
		return nil, err
	}
	langDir, _, err := templates.LangDir("monorepo", p.Lang)
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
		rel := strings.TrimSuffix(strings.TrimPrefix(path, langDir+"/"), ".tmpl")
		out[rel] = content
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("components: monorepo: render %s: %w", langDir, err)
	}
	return out, nil
}
