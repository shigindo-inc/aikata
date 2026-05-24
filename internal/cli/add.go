package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
)

// newAddCmd builds `aikata add` and registers one leaf subcommand
// per entry in components.All(). New components show up here
// automatically — there is no per-component switch in this file
// (Open/Closed Principle).
//
// Contract: every leaf delegates to Component.Add. Argument
// validation lives inside the component so each one can shape its
// own error messages without forcing a uniform signature.
func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <component>",
		Short: "Add a component to an existing aikata project",
		Long: "Add a single component (memory, adr, stack, ...) to the aikata\n" +
			"project rooted at the current working directory. Run `aikata list\n" +
			"components` to see the available identifiers, or `aikata add\n" +
			"--help` for the bundled subcommands.",
		Args: cobra.MinimumNArgs(1),
		// The parent itself is a dispatcher; cobra prints help when called
		// with no leaf, and we surface a custom error when the leaf is
		// unknown so the user sees the registry list instead of cobra's
		// generic "unknown command" text.
		RunE: func(_ *cobra.Command, args []string) error {
			return unknownComponentError(args[0])
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	for _, c := range components.All() {
		cmd.AddCommand(newAddLeafCmd(c))
	}
	return cmd
}

// newAddLeafCmd wraps one Component as a cobra subcommand. The
// command name is c.Name(); positional arguments are passed through
// to Component.Add via AddContext.Args. The shared --dry-run flag is
// the only knob exposed at this layer.
func newAddLeafCmd(c components.Component) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   c.Name() + " [args...]",
		Short: c.Description(),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("add %s: getwd: %w", c.Name(), err)
			}
			cfg, _, err := config.Load(target)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return &ExitError{
						Code: 2,
						Err: fmt.Errorf("aikata config not found at %s; run `aikata init` first or cd into an aikata project",
							config.PrimaryPath(target)),
					}
				}
				return fmt.Errorf("add %s: load config: %w", c.Name(), err)
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

func unknownComponentError(name string) error {
	return &ExitError{
		Code: 2,
		Err: fmt.Errorf("aikata add: unknown component %q (known: %s)",
			name, formatComponentList(components.ActiveNames())),
	}
}

func formatComponentList(names []string) string {
	if len(names) == 0 {
		return "(none registered)"
	}
	out := names[0]
	for _, n := range names[1:] {
		out += ", " + n
	}
	return out
}
