package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// ClaudeProvider emits CLAUDE.md from the canonical AGENTS.md. In v0.1
// it produces a thin wrapper that defers to AGENTS.md and lists the
// project's top-level docs in the Read order section.
//
// Claude-only extensions (Skills, hooks, fast-mode references) are
// expected to land via templates/data/ai_tools/claude/extensions/ in
// a later release; see ADR 0002 Q-DESIGN-01.
type ClaudeProvider struct{}

// Name implements Provider.
func (ClaudeProvider) Name() string { return "claude" }

// Files implements Provider. The returned map always contains exactly
// `CLAUDE.md`. Render-time data adapts to which canonical docs the
// project actually ships, so the generated Read order never references
// a missing file.
func (ClaudeProvider) Files(ctx Context) (map[string]string, error) {
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
	path := resolveLangTemplate("ai_tools/claude", ctx.Lang(), "CLAUDE.md.tmpl")
	content, err := templates.Render(path, data, ctx.Clock)
	if err != nil {
		return nil, err
	}
	return map[string]string{"CLAUDE.md": content}, nil
}
