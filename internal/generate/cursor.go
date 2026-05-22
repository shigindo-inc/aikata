package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// CursorProvider emits a thin `.cursor/rules/main.mdc` that defers to
// the canonical AGENTS.md. Cursor reads AGENTS.md natively in 2026, so
// the .mdc is a discoverability wrapper with `alwaysApply: true` rather
// than a duplicated rule set. The richer split into glob-scoped MDC
// files (per-file-pattern rules) is deferred to v0.3 (Q-DESIGN-08).
type CursorProvider struct{}

// Name implements Provider.
func (CursorProvider) Name() string { return "cursor" }

// Files implements Provider. Always emits exactly one file:
// `.cursor/rules/main.mdc`. Read-order rendering mirrors ClaudeProvider
// so the wrapper adapts to whichever canonical docs the project ships.
func (CursorProvider) Files(ctx Context) (map[string]string, error) {
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(ctx.TargetDir, filepath.FromSlash(rel)))
		return err == nil
	}

	if !has("AGENTS.md") {
		return nil, fmt.Errorf("AGENTS.md not found in %s (run `aikata init` first)", ctx.TargetDir)
	}

	data := map[string]any{
		"ProjectName":      ctx.Project.Project.Name,
		"HasReadme":        has("README.md"),
		"HasSpec":          has("SPEC.md"),
		"HasArchitecture":  has("ARCHITECTURE.md"),
		"HasGlossary":      has("GLOSSARY.md"),
		"HasOpenQuestions": has("docs/decisions/open-questions.md"),
		"HasMemory":        has("docs/memory"),
	}
	path := resolveLangTemplate("ai_tools/cursor", ctx.Lang(), "main.mdc.tmpl")
	content, err := templates.Render(path, data, ctx.Clock)
	if err != nil {
		return nil, err
	}
	return map[string]string{".cursor/rules/main.mdc": content}, nil
}
