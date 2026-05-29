package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
)

// newNewCmd builds `aikata new` and registers one leaf subcommand
// per entry in components.Artifacts(). Each leaf creates a one-off
// authoring scaffold (ADR file today; future artifact templates
// follow the same shape) without flipping any durable project flag.
// `aikata new adr "<title>"` is the post-v0.7.1 spelling for
// what used to be `aikata add adr`.
func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <artifact>",
		Short: "Create an authoring artifact (adr \"<title>\", ...)",
		Long: "Create a one-off authoring scaffold under the aikata project\n" +
			"rooted at the current working directory. Unlike `aikata enable`,\n" +
			"`new` does not flip any durable schema flag — it just stamps a\n" +
			"single file.\n\n" +
			"Run `aikata list artifacts` to see the available identifiers.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return unknownArtifactError(args[0])
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	for _, c := range components.Artifacts() {
		cmd.AddCommand(newNewLeafCmd(c))
	}
	return cmd
}

// newNewLeafCmd wraps one artifact Component as a cobra subcommand.
func newNewLeafCmd(c components.Component) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   c.Name() + " [args...]",
		Short: c.Description(),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("new %s: getwd: %w", c.Name(), err)
			}
			cfg, err := config.Load(target)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return &ExitError{
						Code: 2,
						Err: fmt.Errorf("aikata config not found at %s; run `aikata init` first or cd into an aikata project",
							config.PrimaryPath(target)),
					}
				}
				return fmt.Errorf("new %s: load config: %w", c.Name(), err)
			}
			ctx := components.AddContext{
				TargetDir:   target,
				ProjectName: cfg.Project.Name,
				Lang:        cfg.Project.Lang,
				Args:        args,
				DryRun:      dryRun,
				Stdout:      cmd.OutOrStdout(),
				Stderr:      cmd.ErrOrStderr(),
			}
			return c.Add(ctx)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan without writing files")
	return cmd
}

func unknownArtifactError(name string) error {
	return &ExitError{
		Code: 2,
		Err: fmt.Errorf("aikata new: unknown artifact %q (known: %s)",
			name, formatNameList(components.ActiveArtifactNames())),
	}
}

// formatNameList renders a sorted name slice as "a, b, c" or
// "(none registered)". Shared by both enable and new error paths.
func formatNameList(names []string) string {
	if len(names) == 0 {
		return "(none registered)"
	}
	out := names[0]
	for _, n := range names[1:] {
		out += ", " + n
	}
	return out
}
