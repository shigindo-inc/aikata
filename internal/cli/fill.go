package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/fill"
)

// newFillCmd builds `aikata fill`: a zero-config verb that tops up the
// current repository's canonical document set, writing only the files
// that are missing and never overwriting anything that already exists
// (ADR 0042). It is the low-friction counterpart to the three existing
// post-init/adoption paths:
//
//   - `init` scaffolds a NEW project (and diverts to `.aikata-proposed/`
//     in a non-empty directory);
//   - `enable` adds a single capability and requires existing config;
//   - `sync` pulls upstream template changes and respects deletions.
//
// fill takes no flags by design — scope is inferred from the project
// (`.aikata/manifest.yaml` + `aikata.yaml` when present) or defaults to
// the standard scope for an unmanaged repository.
func newFillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fill",
		Short: "Write any missing canonical documents (never overwrites existing files)",
		Long: "Top up the current repository's canonical document set: render the\n" +
			"documents this project's scope defines and write only the ones that\n" +
			"are missing. Existing files — including hand-edited ones — are never\n" +
			"overwritten, and the run is idempotent.\n\n" +
			"Scope is inferred with no flags: from .aikata/ when the project is\n" +
			"already aikata-managed, otherwise the standard scope is used and the\n" +
			"repository is adopted as an aikata project (project name = directory\n" +
			"name). Because fill never overwrites, prune any document that does not\n" +
			"fit afterward.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("fill: getwd: %w", err)
			}
			_, err = fill.Run(fill.Options{
				Root:   target,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			})
			return err
		},
	}
}
