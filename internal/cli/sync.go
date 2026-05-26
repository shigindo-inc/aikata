package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/sync"
)

// newSyncCmd wires `aikata sync` into the root command. The merge
// contract is documented in ADR 0011 and internal/sync; this cobra
// layer is just argument parsing, exit code mapping, and output
// formatting.
//
// Exit codes:
//
//   - 0: clean (no conflicts) or dry-run with conflicts reported but
//     no write attempted.
//   - 2: conflicts written to disk (mirrors `git merge --no-commit`'s
//     non-zero exit so CI loops can distinguish "merge needed" from
//     "merge failed").
//   - 1: I/O or load error before the merge even ran.
func newSyncCmd() *cobra.Command {
	var (
		dryRun     bool
		rebaseline bool
		jsonOut    bool
	)
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Pull newer aikata template content into the current project",
		Long: "Performs a 3-way diff-merge against `.aikata/manifest.yaml`\n" +
			"and the latest aikata templates. User edits are preserved;\n" +
			"upstream-only changes are auto-applied; true conflicts are\n" +
			"written back with git-style markers for manual resolution.\n\n" +
			"Run `aikata sync --rebaseline` once in projects that pre-date\n" +
			"the v0.5 manifest to seed the ancestor from current on-disk\n" +
			"state. Rebaseline is non-destructive: it only writes\n" +
			"`.aikata/manifest.yaml`; source files are not modified. A\n" +
			"subsequent `aikata sync` then performs the actual 3-way merge.\n\n" +
			"See docs/adr/0011-aikata-sync-design.md for the merge contract.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("sync: getwd: %w", err)
			}
			result, err := sync.Run(sync.Options{
				Root:       target,
				DryRun:     dryRun,
				Rebaseline: rebaseline,
				Stdout:     cmd.OutOrStdout(),
				Stderr:     cmd.ErrOrStderr(),
			})
			if err != nil {
				if errors.Is(err, sync.ErrNoManifest) {
					return &ExitError{Code: 2, Err: err}
				}
				return fmt.Errorf("sync: %w", err)
			}
			if jsonOut {
				if err := writeSyncJSON(cmd.OutOrStdout(), result); err != nil {
					return err
				}
			} else if err := writeSyncText(cmd.OutOrStdout(), result, dryRun); err != nil {
				return err
			}
			if result.Conflicts > 0 && !dryRun {
				return &ExitError{Code: 2, Err: fmt.Errorf("sync: %d conflict(s) require manual resolution", result.Conflicts)}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report the merge plan without writing to disk")
	cmd.Flags().BoolVar(&rebaseline, "rebaseline", false, "seed a missing manifest from current on-disk state (non-destructive: no source files are modified)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON envelope (shape: {version: 1, kind: \"sync\", ...})")
	return cmd
}

func writeSyncText(w io.Writer, r sync.RunResult, dryRun bool) error {
	header := "aikata sync"
	if dryRun {
		header = "aikata sync --dry-run"
	}
	if _, err := fmt.Fprintln(w, header); err != nil {
		return err
	}
	for _, f := range r.Files {
		if f.Status == sync.StatusUnchanged {
			continue
		}
		if _, err := fmt.Fprintf(w, "  %-22s %s\n", string(f.Status), f.Path); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "\nSummary: %d applied, %d conflict(s), %d unchanged.\n",
		r.Applied, r.Conflicts, r.NoChange); err != nil {
		return err
	}
	for _, note := range r.Notes {
		if _, err := fmt.Fprintln(w, note); err != nil {
			return err
		}
	}
	if r.Conflicts > 0 && !dryRun {
		if _, err := fmt.Fprintln(w, "Resolve the conflict markers in the listed files and re-run `aikata sync`."); err != nil {
			return err
		}
	}
	return nil
}

func writeSyncJSON(w io.Writer, r sync.RunResult) error {
	type jsonFile struct {
		Path   string `json:"path"`
		Status string `json:"status"`
	}
	type jsonSummary struct {
		Applied   int `json:"applied"`
		Conflicts int `json:"conflicts"`
		NoChange  int `json:"no_change"`
	}
	payload := struct {
		Version int         `json:"version"`
		Kind    string      `json:"kind"`
		Files   []jsonFile  `json:"files"`
		Summary jsonSummary `json:"summary"`
		Notes   []string    `json:"notes,omitempty"`
	}{
		Version: 1,
		Kind:    "sync",
		Summary: jsonSummary{Applied: r.Applied, Conflicts: r.Conflicts, NoChange: r.NoChange},
		Notes:   r.Notes,
	}
	for _, f := range r.Files {
		payload.Files = append(payload.Files, jsonFile{Path: f.Path, Status: string(f.Status)})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
