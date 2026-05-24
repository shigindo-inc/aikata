package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/doctor"
)

// newDoctorCmd builds the `aikata doctor` subcommand. v0.3 introduces
// `--fix` (apply the trivially-fixable subset) and the related
// `--dry-run` flag that surfaces what `--fix` would do without
// touching the filesystem.
func newDoctorCmd() *cobra.Command {
	var (
		fix     bool
		dryRun  bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check consistency of the aikata project in the current directory",
		Long: "Run a suite of read-only checks against the project at the\n" +
			"current working directory. Errors set exit code 3; warnings\n" +
			"and infos do not. Pass --fix to apply the trivially-fixable\n" +
			"subset (stale `updated:` bumps and missing-frontmatter scaffolds).\n" +
			"Combine with --dry-run to preview the fix without writing.\n" +
			"Pass --json to emit a machine-readable report instead of the\n" +
			"human-readable text format.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("doctor: getwd: %w", err)
			}
			opts := doctor.Options{TargetDir: target}
			issues, err := doctor.Run(opts)
			if err != nil {
				return err
			}

			if fix {
				// With --json, the text-format pre-print would corrupt
				// the JSON stream. Apply the fix silently and emit the
				// post-fix report instead.
				if !jsonOut {
					if err := doctor.Format(cmd.OutOrStdout(), issues); err != nil {
						return err
					}
				}
				if dryRun {
					if !jsonOut {
						n := countFixableIssues(issues)
						if _, werr := fmt.Fprintf(cmd.OutOrStdout(),
							"\n--fix --dry-run: would attempt to fix %d issue(s); no files written.\n", n); werr != nil {
							return werr
						}
					}
				} else {
					res, ferr := doctor.Fix(opts, issues)
					if ferr != nil {
						return ferr
					}
					if !jsonOut {
						if _, werr := fmt.Fprintf(cmd.OutOrStdout(),
							"\nFixed %d issue(s) in %d file(s); %d issue(s) had no auto-fix.\n",
							res.Fixed, len(res.Files), res.Skipped); werr != nil {
							return werr
						}
					}
					issues, err = doctor.Run(opts)
					if err != nil {
						return err
					}
				}
			} else if !jsonOut {
				if err := doctor.Format(cmd.OutOrStdout(), issues); err != nil {
					return err
				}
			}

			if jsonOut {
				if err := doctor.FormatJSON(cmd.OutOrStdout(), issues); err != nil {
					return err
				}
			}

			if doctor.HasErrors(issues) {
				return &ExitError{Code: 3, Err: fmt.Errorf("doctor: %d issue(s) at error level", countErrors(issues))}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Apply auto-fixes for the trivially-fixable subset of issues")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "With --fix, show what would change but do not write files")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report to stdout instead of the text format")
	return cmd
}

func countErrors(issues []doctor.Issue) int {
	n := 0
	for _, iss := range issues {
		if iss.Level == doctor.LevelError {
			n++
		}
	}
	return n
}

func countFixableIssues(issues []doctor.Issue) int {
	known := make(map[string]struct{}, len(doctor.FixableCodes))
	for _, c := range doctor.FixableCodes {
		known[c] = struct{}{}
	}
	n := 0
	for _, iss := range issues {
		if iss.Code == "" {
			continue
		}
		if _, ok := known[iss.Code]; ok {
			n++
		}
	}
	return n
}
