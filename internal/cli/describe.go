package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// newDescribeCmd builds `aikata describe`. v0.3.1 ships only the
// `preset <name>` form; describe-for-stack and describe-for-ai-tool
// arrive in v0.4 alongside `aikata add`. Keeping the subcommand tree
// open now (rather than a flat `describe-preset` verb) means later
// additions slot in without breaking the surface.
func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show the long-form description of an aikata identifier",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newDescribePresetCmd())
	return cmd
}

const describeJSONSchemaVersion = 1

type describeJSONReport struct {
	Version int               `json:"version"`
	Kind    string            `json:"kind"`
	Item    describeJSONEntry `json:"item"`
}

type describeJSONEntry struct {
	Name        string   `json:"name"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
	Languages   []string `json:"languages,omitempty"`
}

func newDescribePresetCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "preset <name>",
		Short: "Describe one init preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			info, ok := scaffold.FindPreset(name)
			if !ok {
				active := scaffold.ActivePresetNames()
				return &ExitError{
					Code: 2,
					Err:  fmt.Errorf("describe: unknown preset %q (known: %s)", name, strings.Join(active, ", ")),
				}
			}
			if jsonOut {
				report := describeJSONReport{
					Version: describeJSONSchemaVersion,
					Kind:    "preset",
					Item: describeJSONEntry{
						Name:        info.Name,
						Status:      info.Status,
						Description: info.Description,
						Languages:   append([]string(nil), info.Languages...),
					},
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			out := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(out, "preset: %s\n", info.Name); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "status: %s\n", info.Status); err != nil {
				return err
			}
			if len(info.Languages) > 0 {
				if _, err := fmt.Fprintf(out, "languages: %s\n", strings.Join(info.Languages, ", ")); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(out, "\n%s\n", info.Description); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}
