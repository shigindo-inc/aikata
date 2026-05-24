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
		fix    bool
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check consistency of the aikata project in the current directory",
		Long: "Run a suite of read-only checks against the project at the\n" +
			"current working directory. Errors set exit code 3; warnings\n" +
			"and infos do not. Pass --fix to apply the trivially-fixable\n" +
			"subset (stale `updated:` bumps and missing-frontmatter scaffolds).\n" +
			"Combine with --dry-run to preview the fix without writing.",
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
			if err := doctor.Format(cmd.OutOrStdout(), issues); err != nil {
				return err
			}

			if fix {
				if dryRun {
					n := countFixableIssues(issues)
					fmt.Fprintf(cmd.OutOrStdout(),
						"\n--fix --dry-run: would attempt to fix %d issue(s); no files written.\n", n)
				} else {
					res, ferr := doctor.Fix(opts, issues)
					if ferr != nil {
						return ferr
					}
					fmt.Fprintf(cmd.OutOrStdout(),
						"\nFixed %d issue(s) in %d file(s); %d issue(s) had no auto-fix.\n",
						res.Fixed, len(res.Files), res.Skipped)
					// Re-run so the exit code reflects the post-fix state.
					issues, err = doctor.Run(opts)
					if err != nil {
						return err
					}
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
