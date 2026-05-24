package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newCompletionCmd builds the `aikata completion <shell>` subcommand.
// It wraps cobra's built-in completion generators with a custom Long
// help so the activation snippets stay close to the command itself.
//
// We surface our own implementation (rather than relying on cobra's
// auto-registered `completion` command) so the help text can include
// per-shell activation examples without the user having to consult
// cobra's docs.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <bash|zsh|fish|powershell>",
		Short: "Emit a shell completion script for aikata",
		Long: `Emit a shell completion script for aikata to standard output.

Bash:

    # Current shell only:
    source <(aikata completion bash)

    # System-wide (requires sudo on most systems):
    aikata completion bash > /etc/bash_completion.d/aikata

    # Per-user (with bash-completion installed):
    aikata completion bash > "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/aikata"

Zsh:

    # If shell completion is not already enabled, enable it once:
    #   echo "autoload -Uz compinit; compinit" >> ~/.zshrc
    aikata completion zsh > "${fpath[1]}/_aikata"

    # For the current shell only:
    source <(aikata completion zsh)

Fish:

    aikata completion fish | source

    # To load on every new shell:
    aikata completion fish > ~/.config/fish/completions/aikata.fish

PowerShell:

    aikata completion powershell | Out-String | Invoke-Expression

    # To load on every new shell, append the script's output to your
    # PowerShell profile (see $PROFILE).
`,
		Args:                  cobra.ExactArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(out)
			default:
				// cobra's ValidArgs already rejects unknown shells, but
				// keep an explicit fallback so the switch is exhaustive.
				fmt.Fprintf(os.Stderr, "unknown shell %q; expected bash|zsh|fish|powershell\n", args[0])
				return &ExitError{Code: 2, Err: fmt.Errorf("completion: unknown shell %q", args[0])}
			}
		},
	}
	return cmd
}
