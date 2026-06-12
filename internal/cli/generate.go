package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/generate"
)

// newGenerateCmd builds the `aikata generate` subcommand.
//
// Reads the project config (`.aikata/aikata.yaml`), looks up each
// enabled AI tool's Provider, and writes the produced artifacts under
// cwd. Existing files are overwritten — that is the explicit
// contract; generated artifacts are disposable per ADR 0002.
func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Regenerate per-AI-tool configuration files from canonical sources",
		Long: "Reads the aikata config (.aikata/aikata.yaml) in the current directory\n" +
			"and emits per-AI-tool artifacts (CLAUDE.md, etc.) from the canonical AGENTS.md. Existing files are\n" +
			"overwritten; generated artifacts are disposable (ADR 0002).",
		Args: cobra.NoArgs,
		// Rebuild the doc map as a final, isolated step (ADR 0044 D5),
		// decoupled from the per-tool provider work above: a provider
		// failure returns before this runs, and a map failure here never
		// changes generate's exit status.
		PostRunE: docMapPostRun,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("generate: getwd: %w", err)
			}
			cfgPath, err := config.Resolve(target)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return &ExitError{
						Code: 2,
						Err:  fmt.Errorf("%s not found; run `aikata init` first or cd into an aikata project", config.PrimaryPath(target)),
					}
				}
				return fmt.Errorf("generate: locate config: %w", err)
			}
			body, err := os.ReadFile(cfgPath)
			if err != nil {
				return fmt.Errorf("generate: read %s: %w", cfgPath, err)
			}
			cfg, err := config.Unmarshal(body)
			if err != nil {
				return &ExitError{Code: 2, Err: err}
			}
			ctx := generate.Context{
				TargetDir: target,
				Project:   cfg,
			}
			counts, err := generate.Run(ctx)
			if err != nil {
				if errors.Is(err, generate.ErrUnknownAITool) {
					return &ExitError{
						Code: 2,
						Err:  fmt.Errorf("%w (known: %s)", err, strings.Join(generate.KnownTools(), ", ")),
					}
				}
				return err
			}
			for _, name := range cfg.AITools {
				if counts[name] == 0 {
					if _, err := fmt.Fprintf(cmd.ErrOrStderr(),
						"[%s] no files generated (reads AGENTS.md directly)\n", name); err != nil {
						return fmt.Errorf("generate: write noop notice: %w", err)
					}
				}
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "generated artifacts for: %s\n", strings.Join(cfg.AITools, ", ")); err != nil {
				return fmt.Errorf("generate: write status: %w", err)
			}
			return nil
		},
	}
	return cmd
}
