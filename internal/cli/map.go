package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/docmap"
	"github.com/shigindo-inc/aikata/internal/doctor"
)

// newMapCmd builds the `aikata map` subcommand: the explicit way to
// rebuild the document map (ADR 0044). The map is also rebuilt
// automatically as an isolated final step of init / fill / enable /
// sync / generate (see docMapPostRun), so this verb is mainly for
// regenerating after hand edits to documents.
//
// The map catalogs documents only — no aikata config is required, so the
// verb runs in any repository.
func newMapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "map",
		Short: "Rebuild the document map (.aikata/docmap.{yaml,md})",
		Long: "Scans the project's Markdown documents and writes a machine-derived\n" +
			"map of the document set — inventory, cross-references, freshness, and a\n" +
			"managed/external split — to .aikata/docmap.yaml (data) and\n" +
			".aikata/docmap.md (readable view). Documents only; no source code is read\n" +
			"(ADR 0044).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("map: getwd: %w", err)
			}
			if err := rebuildDocMap(target); err != nil {
				return fmt.Errorf("map: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "wrote %s and %s\n",
				docmap.YAMLPath(target), docmap.MarkdownPath(target)); err != nil {
				return fmt.Errorf("map: write status: %w", err)
			}
			return nil
		},
	}
}

// rebuildDocMap regenerates the document map under targetDir. The
// managed/external flag reuses doctor's managed-surface globs so the two
// surfaces agree; docmap itself never imports doctor (the dependency runs
// one way so doctor's freshness check can import docmap).
func rebuildDocMap(targetDir string) error {
	return docmap.Generate(docmap.OptionsFor(targetDir, doctor.ManagedIncludeGlobs(targetDir)))
}

// docMapPostRun is the isolated lifecycle hook wired onto init / fill /
// enable / sync / generate as PostRunE (PersistentPostRunE for enable's
// subcommands). cobra runs it only after the command's own RunE
// succeeds, so the map is rebuilt against the just-updated document set.
//
// It is deliberately best-effort: a rebuild failure is reported but never
// changes the command's exit status, keeping the map step decoupled from
// the command that triggered it (ADR 0044 D5 / docmap-design.md §5).
//
// The hook is suppressed in two cases so it never writes when the command
// itself wrote nothing: a `--dry-run` invocation, and any directory that
// is not yet an established aikata project (notably `init`'s proposal
// mode, which writes `.aikata-proposed/` rather than `.aikata/`). The
// explicit `aikata map` verb has neither guard — it catalogs documents in
// any directory.
func docMapPostRun(cmd *cobra.Command, _ []string) error {
	if f := cmd.Flags().Lookup("dry-run"); f != nil && f.Value.String() == "true" {
		return nil
	}
	target, err := os.Getwd()
	if err != nil {
		return nil
	}
	if _, err := config.Resolve(target); err != nil {
		return nil
	}
	if err := rebuildDocMap(target); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: doc map rebuild failed: %v\n", err)
	}
	return nil
}
