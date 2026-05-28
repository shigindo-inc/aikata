package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
)

// newEnableCmd builds `aikata enable` and registers one leaf
// subcommand per entry in components.Capabilities(). Each leaf
// persists a durable project capability — schema-v2 component flags
// (memory, ui, api, tdd, changelog, monorepo), stack registrations,
// and AI-tool enrolments. New capabilities show up here automatically
// (Open/Closed Principle).
//
// Per ADR 0017 there is no pre-release `add` alias; users who relied
// on `aikata add <component>` migrate to `aikata enable <capability>`
// or `aikata new <artifact>` depending on intent.
func newEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <capability>",
		Short: "Persist a project capability (memory, monorepo, stack <name>, ai-tool <name>, ...)",
		Long: "Enable a durable project capability and persist the choice in\n" +
			".aikata/aikata.yaml. The leaf subcommand renders any files\n" +
			"associated with the capability (idempotent, never overwrites\n" +
			"user customisations), records them in .aikata/manifest.yaml,\n" +
			"and flips the corresponding schema-v2 components.* flag or\n" +
			"appends to ai_tools: / stacks: as appropriate.\n\n" +
			"Run `aikata list capabilities` to see the available identifiers.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return unknownCapabilityError(args[0])
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	for _, c := range components.Capabilities() {
		cmd.AddCommand(newEnableLeafCmd(c))
	}
	return cmd
}

// newEnableLeafCmd wraps one capability Component as a cobra
// subcommand. Identical shape to the pre-v0.7.1 add-leaf so behaviour
// is unchanged — only the parent name (`enable` vs `add`) differs.
func newEnableLeafCmd(c components.Component) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   c.Name() + " [args...]",
		Short: c.Description(),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("enable %s: getwd: %w", c.Name(), err)
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
				return fmt.Errorf("enable %s: load config: %w", c.Name(), err)
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

func unknownCapabilityError(name string) error {
	return &ExitError{
		Code: 2,
		Err: fmt.Errorf("aikata enable: unknown capability %q (known: %s)",
			name, formatNameList(components.ActiveCapabilityNames())),
	}
}
