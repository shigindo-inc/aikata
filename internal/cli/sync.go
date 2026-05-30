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
		dryRun       bool
		rebaseline   bool
		reseed       bool
		jsonOut      bool
		preset       string
		lang         string
		stacks       []string
		withMonorepo bool
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
			"Scope is derived from the manifest, then `.aikata/aikata.yaml`,\n" +
			"then the CLI override flags (`--preset`, `--lang`, `--stack`,\n" +
			"`--with-monorepo`). Overrides apply to this invocation only —\n" +
			"they are not written back to either file. See ADR 0013 for\n" +
			"the hierarchy and rationale.\n\n" +
			"See docs/adr/0011-aikata-sync-design.md for the merge contract.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("sync: getwd: %w", err)
			}

			// Capture only the flags the user actually changed so we
			// can pass them through as overrides. `cmd.Flags().Changed`
			// is the way to distinguish "user passed --foo=false" from
			// "user did not pass --foo", which BoolVar alone cannot.
			var (
				presetPtr       *string
				langPtr         *string
				stacksPtr       *[]string
				withMonorepoPtr *bool
			)
			if cmd.Flags().Changed("preset") {
				presetPtr = &preset
			}
			if cmd.Flags().Changed("lang") {
				langPtr = &lang
			}
			if cmd.Flags().Changed("stack") {
				stacksPtr = &stacks
			}
			if cmd.Flags().Changed("with-monorepo") {
				withMonorepoPtr = &withMonorepo
			}

			result, err := sync.Run(sync.Options{
				Root:                 target,
				DryRun:               dryRun,
				Rebaseline:           rebaseline,
				Reseed:               reseed,
				Stdout:               cmd.OutOrStdout(),
				Stderr:               cmd.ErrOrStderr(),
				OverridePreset:       presetPtr,
				OverrideLang:         langPtr,
				OverrideStacks:       stacksPtr,
				OverrideWithMonorepo: withMonorepoPtr,
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
	cmd.Flags().BoolVar(&rebaseline, "rebaseline", false, "seed a missing manifest from the current upstream rendering (non-destructive: no source files are modified). No-op with a notice if a manifest already exists — use --reseed")
	cmd.Flags().BoolVar(&reseed, "reseed", false, "re-anchor an existing manifest to the current upstream rendering and exit (manifest-only write; no source files are modified)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON envelope (shape: {version: 1, kind: \"sync\", ...})")
	// Scope override flags (ADR 0013). Each one is a transient lens
	// over this invocation only; the manifest and `aikata.yaml` are
	// never rewritten by sync to record an override.
	cmd.Flags().StringVar(&preset, "preset", "", "override the preset for this run (minimal | standard | flutter | typescript)")
	cmd.Flags().StringVar(&lang, "lang", "", "override the rendering language for this run (en | ja)")
	cmd.Flags().StringSliceVar(&stacks, "stack", nil, "override the stack list for this run; repeatable (e.g. --stack flutter --stack typescript)")
	cmd.Flags().BoolVar(&withMonorepo, "with-monorepo", false, "override the monorepo opt-in for this run (use --with-monorepo=false to force-disable)")
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
