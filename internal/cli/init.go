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
			if !noInteractive {
				return &ExitError{
					Code: 2,
					Err:  errors.New("interactive mode is not yet implemented (Task 6); rerun with --no-interactive"),
				}
			}
			if name == "" {
				return &ExitError{
					Code: 2,
					Err:  errors.New("project name is required: pass it as the positional arg or --name"),
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

	cmd.Flags().StringVar(&preset, "preset", "minimal", "preset name (currently: minimal)")
	cmd.Flags().StringVar(&name, "name", "", "project name; overrides the positional argument when both are given")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "skip interactive prompts (required in v0.1)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files in a non-empty target directory")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan without writing files")
	cmd.Flags().StringVar(&lang, "lang", "en", "document language (en | ja)")

	return cmd
}
