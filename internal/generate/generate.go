package generate

import (
	"errors"
	"fmt"
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

// registry maps tool names ("claude", "cursor", …) to their Provider
// implementations. Kept unexported so callers go through Run / Get.
var registry = map[string]Provider{
	"claude": ClaudeProvider{},
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
// any disk write happens.
func Run(ctx Context) error {
	if ctx.TargetDir == "" {
		return errors.New("generate: target directory is required")
	}
	if len(ctx.Project.AITools) == 0 {
		return errors.New("generate: no AI tools enabled in .ai/aikata.yaml")
	}

	rendered := make(map[string]string)
	for _, name := range ctx.Project.AITools {
		provider, err := Get(name)
		if err != nil {
			return err
		}
		files, err := provider.Files(ctx)
		if err != nil {
			return fmt.Errorf("generate(%s): %w", name, err)
		}
		for rel, content := range files {
			rendered[rel] = content
		}
	}

	return writeAll(ctx.TargetDir, rendered)
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
