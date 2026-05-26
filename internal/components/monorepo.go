package components

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/shigindo-inc/aikata/internal/templates"
)

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
