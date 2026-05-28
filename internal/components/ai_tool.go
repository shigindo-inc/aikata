package components

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/generate"
)

// aiToolComponent enables one AI-tool target in an existing project's
// aikata.yaml. The component emits no markdown of its own — running
// `aikata generate` after `aikata add ai-tool <name>` is what
// materializes the per-tool artifact (e.g. `.cursor/rules/main.mdc`).
// Use the init-time `--ai-tools` flag for new projects; this command
// is the post-init counterpart.
type aiToolComponent struct{}

// AITool is the singleton registered in the components registry.
var AITool Component = aiToolComponent{}

func (aiToolComponent) Name() string { return "ai-tool" }
func (aiToolComponent) Description() string {
	return "Enable an AI-tool target in aikata.yaml (post-init counterpart of --ai-tools)."
}
func (aiToolComponent) Status() string { return StatusActive }

// Add appends ctx.Args[0] to aikata.yaml's `ai_tools:` list when the
// tool is registered with `internal/generate` and not already enabled.
// Re-running on an already-enabled tool prints a notice on Stderr and
// returns nil. Unknown tool names are refused with a hint listing the
// known set.
func (aiToolComponent) Add(ctx AddContext) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("components: ai-tool: tool name is required (usage: aikata add ai-tool <name>; known: %s)",
			strings.Join(generate.KnownTools(), ", "))
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: ai-tool: target directory is required")
	}
	name := strings.ToLower(strings.TrimSpace(ctx.Args[0]))
	if name == "" {
		return fmt.Errorf("components: ai-tool: tool name cannot be blank")
	}
	if !aiToolKnown(name) {
		return fmt.Errorf("components: ai-tool: unknown ai-tool %q (known: %s)",
			name, strings.Join(generate.KnownTools(), ", "))
	}

	// LoadMigrated picks up any pending v1 -> v2 schema migration (ADR
	// 0016) so the subsequent config.Save lands at the current schema.
	cfg, _, err := config.LoadMigrated(ctx.TargetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("components: ai-tool: aikata config not found at %s (run `aikata init` first)",
				config.PrimaryPath(ctx.TargetDir))
		}
		return fmt.Errorf("components: ai-tool: load config: %w", err)
	}

	if contains(cfg.AITools, name) {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: ai-tool %q already enabled in %s; nothing to do\n",
			name, config.PrimaryPath(ctx.TargetDir)); werr != nil {
			return werr
		}
		return nil
	}

	if ctx.DryRun {
		if _, werr := fmt.Fprintf(stdout(ctx),
			"Would add %q to %s ai_tools:\n", name, config.PrimaryPath(ctx.TargetDir)); werr != nil {
			return werr
		}
		return nil
	}

	cfg.AITools = append(cfg.AITools, name)
	sort.Strings(cfg.AITools)
	if err := config.Save(ctx.TargetDir, cfg); err != nil {
		return fmt.Errorf("components: ai-tool: save config: %w", err)
	}
	if _, werr := fmt.Fprintf(stdout(ctx),
		"updated %s ai_tools: += %q\n", config.PrimaryPath(ctx.TargetDir), name); werr != nil {
		return werr
	}
	return nil
}

func aiToolKnown(name string) bool {
	for _, t := range generate.KnownTools() {
		if t == name {
			return true
		}
	}
	return false
}
