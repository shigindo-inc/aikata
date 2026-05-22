package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// newInitCmd builds the `aikata init` subcommand.
//
// MVP scope (Task 4):
//   - `--preset minimal` is the only fully implemented preset.
//   - Interactive prompts are not implemented; --no-interactive is required.
//   - Output directory is always the current working directory.
//
// Future tasks will: add `standard` preset (Task 5), `--with-memory`
// (Task 5A), and an interactive flow (Task 6).
func newInitCmd() *cobra.Command {
	var (
		preset        string
		name          string
		noInteractive bool
		force         bool
		dryRun        bool
		lang          string
		withMemory    bool
	)

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a new aikata project in the current directory",
		Long: "Generate a coherent set of markdown documents for a new project.\n\n" +
			"In non-interactive mode (--no-interactive, required for v0.1), the project\n" +
			"name must be supplied either as the positional argument or via --name.\n" +
			"The target directory is always the current working directory.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && name == "" {
				name = args[0]
			}

			// Interactive when the user did not pass --no-interactive
			// AND stdin is attached to a real terminal. Piped or
			// redirected stdin auto-falls-back to non-interactive so
			// CI invocations don't hang waiting for input.
			interactive := !noInteractive && isTTYFunc()
			if interactive {
				result, err := runPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptResult{
					Name:       name,
					Preset:     preset,
					WithMemory: withMemory,
				})
				if err != nil {
					return &ExitError{Code: 2, Err: err}
				}
				name = result.Name
				preset = result.Preset
				withMemory = result.WithMemory
			}

			if name == "" {
				return &ExitError{
					Code: 2,
					Err:  errors.New("project name is required: pass it as the positional arg, --name, or via the interactive prompt"),
				}
			}

			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("init: getwd: %w", err)
			}

			opts := scaffold.Options{
				ProjectName: name,
				Preset:      preset,
				TargetDir:   target,
				Lang:        lang,
				Force:       force,
				DryRun:      dryRun,
				WithMemory:  withMemory,
				Stacks:      stacksForPreset(preset),
				Stdout:      cmd.OutOrStdout(),
			}

			if err := scaffold.Run(opts); err != nil {
				if errors.Is(err, scaffold.ErrTargetDirNotEmpty) {
					return &ExitError{Code: 2, Err: err}
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "standard", "preset name (minimal | standard | flutter)")
	cmd.Flags().StringVar(&name, "name", "", "project name; overrides the positional argument when both are given")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "skip interactive prompts (required in v0.1)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files in a non-empty target directory")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan without writing files")
	cmd.Flags().StringVar(&lang, "lang", "en", "document language (en | ja)")
	cmd.Flags().BoolVar(&withMemory, "with-memory", false, "include long-term agent memory under docs/memory/ (ADR 0004)")

	return cmd
}

// stacksForPreset returns the stack identifiers implied by a preset
// name. Stack-flavored presets like `flutter` set their own stack
// implicitly; `minimal` / `standard` ship without one.
func stacksForPreset(preset string) []string {
	switch preset {
	case "flutter":
		return []string{"flutter"}
	default:
		return nil
	}
}
