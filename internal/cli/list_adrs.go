package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/adr"
)

// adrItem is one entry of `aikata list adrs --json`. Every field is
// computed from document text; the adr-inventory skill adds the
// judgments (ADR 0052).
type adrItem struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Path   string `json:"path"`
	// Status is the raw text of the body status line; empty when absent.
	Status string    `json:"status"`
	Class  adr.Class `json:"class"`
	// Links are the ADR numbers the status line names.
	Links   []int `json:"links"`
	DocRefs int   `json:"doc_refs"`
	Bytes   int   `json:"bytes"`
}

type adrReport struct {
	Version int        `json:"version"`
	Kind    string     `json:"kind"`
	Items   []adrItem  `json:"items"`
	Summary adrSummary `json:"summary"`
}

type adrSummary struct {
	Total   int               `json:"total"`
	ByClass map[adr.Class]int `json:"by_class"`
}

// classMissing marks an ADR without a body status line in the report.
const classMissing adr.Class = "missing"

func newListADRsCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "adrs",
		Short: "List this project's ADRs with their parsed status",
		Long: "Read docs/adr in the current directory and report, per ADR, its\n" +
			"number, title, decision status and class (pending / effective /\n" +
			"superseded / retired / unknown, or missing without a status line),\n" +
			"the ADRs its status line names, and how many Markdown files refer\n" +
			"to it. Read-only. The adr-inventory skill consumes --json.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			records, err := adr.Inventory(root)
			if err != nil {
				return err
			}
			report := buildADRReport(records)
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			return writeADRText(cmd.OutOrStdout(), report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

func buildADRReport(records []adr.Record) adrReport {
	report := adrReport{
		Version: listJSONSchemaVersion,
		Kind:    "adrs",
		Items:   make([]adrItem, 0, len(records)),
		Summary: adrSummary{Total: len(records), ByClass: map[adr.Class]int{}},
	}
	for _, r := range records {
		item := adrItem{
			Number:  r.Number,
			Title:   r.Title,
			Path:    r.Path,
			Class:   classMissing,
			Links:   []int{},
			DocRefs: r.DocRefs,
			Bytes:   r.Bytes,
		}
		if r.HasStatus {
			item.Status = r.Status.Raw
			item.Class = r.Status.Class
			if r.Status.Links != nil {
				item.Links = r.Status.Links
			}
		}
		report.Summary.ByClass[item.Class]++
		report.Items = append(report.Items, item)
	}
	return report
}

func writeADRText(out io.Writer, report adrReport) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, it := range report.Items {
		if _, err := fmt.Fprintf(w, "%04d\t%s\t%d refs\t%s\n", it.Number, it.Class, it.DocRefs, it.Title); err != nil {
			return err
		}
	}
	return w.Flush()
}
