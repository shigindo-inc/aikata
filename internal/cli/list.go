package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// newListCmd builds the `aikata list` subcommand and its children
// (presets / stacks / ai-tools). Each child supports --json for
// machine-readable output; text output is human-curated.
//
// Discovery of preset / stack / ai-tool identifiers is the user-facing
// goal here; descriptions for individual presets live behind
// `aikata describe preset <name>` so this command stays scannable.
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known presets, stacks, or AI tools",
		Long: "Enumerate the identifiers aikata accepts at init time.\n" +
			"Use `aikata describe preset <name>` for the long-form description\n" +
			"of an individual preset.",
		Args: cobra.NoArgs,
	}
	cmd.AddCommand(newListPresetsCmd())
	cmd.AddCommand(newListStacksCmd())
	cmd.AddCommand(newListAIToolsCmd())
	cmd.AddCommand(newListComponentsCmd())
	return cmd
}

func newListComponentsCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "components",
		Short: "List components addable via `aikata add`",
		Long: "Enumerate every component the `aikata add` subcommand knows\n" +
			"about. Reserved entries are surfaced with a `(reserved)` suffix\n" +
			"in text output and a status field in --json so callers can\n" +
			"discover names held for a future release.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			all := components.All()
			items := make([]listItem, 0, len(all))
			for _, c := range all {
				items = append(items, listItem{
					Name:        c.Name(),
					Status:      c.Status(),
					Description: c.Description(),
				})
			}
			return writeList(cmd.OutOrStdout(), "components", items, jsonOut, formatComponentText)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

// formatComponentText mirrors formatPresetText — a `(reserved)`
// suffix marks non-active entries so the human output carries the
// same status signal the JSON envelope does.
func formatComponentText(out io.Writer, items []listItem) error {
	for _, it := range items {
		suffix := ""
		if it.Status == components.StatusReserved {
			suffix = " (reserved)"
		}
		if _, err := fmt.Fprintf(out, "%s%s\n", it.Name, suffix); err != nil {
			return err
		}
	}
	return nil
}

const listJSONSchemaVersion = 1

// listJSONReport is the top-level shape emitted by `--json`. Kept in
// lockstep with doctor --json's schema philosophy (versioned, flat
// items array).
type listJSONReport struct {
	Version int         `json:"version"`
	Kind    string      `json:"kind"`
	Items   []listItem  `json:"items"`
	Summary listSummary `json:"summary"`
}

type listItem struct {
	Name        string   `json:"name"`
	Status      string   `json:"status,omitempty"`
	Description string   `json:"description,omitempty"`
	Languages   []string `json:"languages,omitempty"`
}

type listSummary struct {
	Total int `json:"total"`
}

func newListPresetsCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "presets",
		Short: "List known init presets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items := make([]listItem, 0)
			for _, p := range scaffold.Presets() {
				items = append(items, listItem{
					Name:      p.Name,
					Status:    p.Status,
					Languages: append([]string(nil), p.Languages...),
				})
			}
			return writeList(cmd.OutOrStdout(), "presets", items, jsonOut, formatPresetText)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

func newListStacksCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "stacks",
		Short: "List known stack identifiers (paired with --preset)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items := make([]listItem, 0)
			for _, s := range scaffold.Stacks() {
				items = append(items, listItem{Name: s})
			}
			return writeList(cmd.OutOrStdout(), "stacks", items, jsonOut, formatPlainText)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

func newListAIToolsCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "ai-tools",
		Short: "List AI-tool identifiers accepted by --ai-tools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			names := make([]string, 0, len(validAITools))
			for n := range validAITools {
				names = append(names, n)
			}
			sort.Strings(names)
			items := make([]listItem, 0, len(names))
			for _, n := range names {
				items = append(items, listItem{Name: n})
			}
			return writeList(cmd.OutOrStdout(), "ai-tools", items, jsonOut, formatPlainText)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

// writeList renders items either as JSON (with the shared schema) or as
// human-readable text via the supplied formatter.
func writeList(out io.Writer, kind string, items []listItem, asJSON bool, textFormatter func(io.Writer, []listItem) error) error {
	if asJSON {
		report := listJSONReport{
			Version: listJSONSchemaVersion,
			Kind:    kind,
			Items:   items,
			Summary: listSummary{Total: len(items)},
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}
	return textFormatter(out, items)
}

// formatPresetText renders the preset list with a "(reserved)" suffix
// for non-active entries so the human-readable output mirrors the
// status flag carried in JSON.
func formatPresetText(out io.Writer, items []listItem) error {
	for _, it := range items {
		suffix := ""
		if it.Status == scaffold.StatusReserved {
			suffix = " (reserved)"
		}
		if _, err := fmt.Fprintf(out, "%s%s\n", it.Name, suffix); err != nil {
			return err
		}
	}
	return nil
}

// formatPlainText prints names only — used for stacks and ai-tools
// where there is no status distinction at the list level.
func formatPlainText(out io.Writer, items []listItem) error {
	for _, it := range items {
		if _, err := fmt.Fprintln(out, it.Name); err != nil {
			return err
		}
	}
	return nil
}
