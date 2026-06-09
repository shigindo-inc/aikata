package cli

import (
	"github.com/spf13/cobra"
)

// Execute builds the root command tree and runs it. The version string
// flows in from cmd/aikata/main.go so the binary's version stays a build
// concern, not a code constant.
func Execute(version string) error {
	root := newRootCmd(version)
	return root.Execute()
}

// newRootCmd constructs the root `aikata` command. The post-init
// surface follows the v0.7.1 purpose-based split (ADR 0017), extended
// with the `fill` completion verb (ADR 0042):
//   - `fill` writes any missing canonical document into the current
//     repository without overwriting existing files (adoption/completion).
//   - `enable <capability>` persists project features (memory, ui,
//     monorepo, stack <name>, ai-tool <name>, ...).
//   - `new <artifact>` creates one-off authoring scaffolds (adr).
//   - `sync` maintains the already-declared surface.
//   - `generate` re-renders per-AI-tool artifacts from AGENTS.md.
//
// There is no `add` parent — see ADR 0017 for the rationale on why
// the pre-v1.0 surface drops it without an alias.
func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aikata",
		Short: "Scaffold AI-readable markdown documents for projects",
		Long: "aikata is a lightweight CLI that, in a single command, scaffolds " +
			"a project with markdown documents and per-AI-tool configuration files " +
			"designed for the AI-coding era. See https://github.com/shigindo-inc/aikata.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	// `--version` is provided by cobra automatically when Version is set;
	// we customize the output template to match Go conventions:
	//   aikata version 0.0.1-dev
	cmd.SetVersionTemplate("aikata version {{.Version}}\n")

	// Hide cobra's auto-registered `completion` command so the in-tree
	// implementation (with shell-specific activation examples in its
	// Long help) is the only one users discover via `aikata --help`.
	cmd.CompletionOptions.HiddenDefaultCmd = true

	// Subcommands. Keep this list short; each subcommand owns its own file.
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newFillCmd())
	cmd.AddCommand(newEnableCmd())
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDescribeCmd())
	cmd.AddCommand(newUpdateCmd(version))
	cmd.AddCommand(newSyncCmd())

	return cmd
}
