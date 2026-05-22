package generate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/templates"
)

// Context carries the inputs every Provider needs to produce its
// per-AI-tool artifacts. Computed once by the cli layer and reused
// across providers in a single `aikata generate` run.
type Context struct {
	// TargetDir is the project root where the canonical AGENTS.md
	// lives and where generated files will be written.
	TargetDir string
	// Project mirrors the values in `.ai/aikata.yaml`.
	Project config.AikataYaml
	// Clock injects the time source for golden-stable rendering;
	// nil = time.Now via templates.Helpers.
	Clock templates.Clock
}

// Lang returns the document language for this run, defaulting to "en"
// when the project did not pin one. Providers use it to pick the
// language-scoped template under ai_tools/<tool>/<lang>/.
func (c Context) Lang() string {
	if c.Project.Project.Lang == "" {
		return "en"
	}
	return c.Project.Project.Lang
}

// Provider produces a set of files for one AI tool. Returned map keys
// are paths relative to ctx.TargetDir (e.g. "CLAUDE.md",
// ".cursor/rules/aikata.mdc"); values are file contents.
type Provider interface {
	Name() string
	Files(ctx Context) (map[string]string, error)
}

// ErrUnknownAITool is returned when `.ai/aikata.yaml` enables a tool
// for which no Provider is registered. The cli maps this to exit code 2.
var ErrUnknownAITool = errors.New("generate: unknown ai_tool")

// resolveLangTemplate picks the language-scoped template under base,
// falling back to "en" when the requested lang directory is missing
// in the embedded FS. The return value is always a usable embed path;
// when both lang and en are missing the caller's templates.Render
// surfaces the not-found error.
func resolveLangTemplate(base, lang, filename string) string {
	if lang == "" {
		lang = "en"
	}
	root, err := templates.FS()
	if err == nil {
		if _, statErr := fs.Stat(root, base+"/"+lang+"/"+filename); statErr == nil {
			return base + "/" + lang + "/" + filename
		}
	}
	return base + "/en/" + filename
}

// registry maps tool names ("claude", "cursor", …) to their Provider
// implementations. Kept unexported so callers go through Run / Get.
var registry = map[string]Provider{
	"claude": ClaudeProvider{},
	"codex":  CodexProvider{},
	"cursor": CursorProvider{},
}

// Get returns the Provider for the given name. Returns
// ErrUnknownAITool when the name is not registered.
func Get(name string) (Provider, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownAITool, name)
	}
	return p, nil
}

// KnownTools returns the names of every registered Provider, sorted
// lexicographically for stable diagnostics.
func KnownTools() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Run materializes every enabled provider's files to ctx.TargetDir.
// Files are written all-or-nothing: every provider must succeed before
// any disk write happens. The returned map reports how many files each
// provider produced, so the cli layer can surface no-op providers
// (e.g. codex, which reads AGENTS.md directly).
func Run(ctx Context) (map[string]int, error) {
	if ctx.TargetDir == "" {
		return nil, errors.New("generate: target directory is required")
	}
	if len(ctx.Project.AITools) == 0 {
		return nil, errors.New("generate: no AI tools enabled in .ai/aikata.yaml")
	}

	rendered := make(map[string]string)
	counts := make(map[string]int, len(ctx.Project.AITools))
	for _, name := range ctx.Project.AITools {
		provider, err := Get(name)
		if err != nil {
			return nil, err
		}
		files, err := provider.Files(ctx)
		if err != nil {
			return nil, fmt.Errorf("generate(%s): %w", name, err)
		}
		counts[name] = len(files)
		for rel, content := range files {
			rendered[rel] = content
		}
	}

	if err := writeAll(ctx.TargetDir, rendered); err != nil {
		return nil, err
	}
	return counts, nil
}

func writeAll(targetDir string, rendered map[string]string) error {
	for rel, content := range rendered {
		full := filepath.Join(targetDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("generate: mkdir %s: %w", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return fmt.Errorf("generate: write %s: %w", full, err)
		}
	}
	return nil
}
