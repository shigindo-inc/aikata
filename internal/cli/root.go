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

// newRootCmd constructs the root `aikata` command. Subcommands (init,
// add, doctor, generate, update, list) will be attached here as they are
// implemented in later phases.
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

	// Subcommands. Keep this list short; each subcommand owns its own file.
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newGenerateCmd())

	return cmd
}
