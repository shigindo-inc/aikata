package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/doctor"
)

// newDoctorCmd builds the `aikata doctor` subcommand. v0.2 is
// read-only; `--fix` is reserved for v0.3 per ADR plans.
func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check consistency of the aikata project in the current directory",
		Long: "Run a suite of read-only checks against the project at the\n" +
			"current working directory. Errors set exit code 3; warnings\n" +
			"and infos do not. The set of checks is fixed for v0.2; see\n" +
			"ROADMAP.md for the v0.3 --fix and --json plans.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("doctor: getwd: %w", err)
			}
			issues, err := doctor.Run(doctor.Options{TargetDir: target})
			if err != nil {
				return err
			}
			if err := doctor.Format(cmd.OutOrStdout(), issues); err != nil {
				return err
			}
			if doctor.HasErrors(issues) {
				return &ExitError{Code: 3, Err: fmt.Errorf("doctor: %d issue(s) at error level", countErrors(issues))}
			}
			return nil
		},
	}
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
