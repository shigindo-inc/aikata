package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/release"
)

// envReleaseEndpoint is a test-only seam that points the update check
// at an httptest.Server instead of the live GitHub API. Production
// callers leave this env var unset; release.Client.Endpoint then
// falls back to the canonical endpoint baked into the release package.
const envReleaseEndpoint = "AIKATA_RELEASE_ENDPOINT_FOR_TEST"

// newUpdateCmd builds `aikata update`. v0.4.2 ships only `--check`;
// self-update (binary overwrite) is scheduled for v0.6 alongside the
// installer-source metadata layer. The bare `aikata update` therefore
// prints a notice and exits 0 — non-zero would surprise users who
// reach for `update` expecting Claude Code-style behaviour.
//
// The version flowing in comes from cmd/aikata/main.effectiveVersion,
// which yields either the GoReleaser ldflags tag, the
// runtime/debug.ReadBuildInfo module version, or the dev sentinel.
func newUpdateCmd(version string) *cobra.Command {
	var (
		check   bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for or apply aikata CLI updates",
		Long: "Compare the running aikata binary against the latest published\n" +
			"GitHub Release.\n\n" +
			"Today only `--check` is wired up; it reads the GitHub Releases\n" +
			"API and reports whether a newer version is available. Self-update\n" +
			"(installing the new binary in place) is scheduled for v0.6, once\n" +
			"the installer-source metadata layer lands so aikata can pick the\n" +
			"safe upgrade path per install channel (ADR 0009).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if !check {
				return printUpdateGuidance(out)
			}
			client := &release.Client{
				UserAgent: "aikata/" + version,
				Endpoint:  os.Getenv(envReleaseEndpoint),
			}
			result, err := client.CheckLatest(context.Background(), version)
			if err != nil {
				return fmt.Errorf("update: %w", err)
			}
			if jsonOut {
				return writeUpdateJSON(out, result)
			}
			return writeUpdateText(out, result)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "check for a newer release without applying it")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

// printUpdateGuidance is the response to bare `aikata update`. The
// message points users at `--check` and explains why self-update is
// not yet available so support burden does not accrue on this
// transitional release.
func printUpdateGuidance(out io.Writer) error {
	const msg = "aikata self-update is not implemented yet (planned v0.6).\n" +
		"Run `aikata update --check` to see if a newer release is available\n" +
		"and upgrade instructions for your install path.\n"
	_, err := io.WriteString(out, msg)
	return err
}

// writeUpdateText renders the human-readable form of a check result.
// The upgrade guidance block lists every install path aikata supports
// at v0.4.2 (go install / curl|sh / manual download). Adding brew or
// npm is a v0.6 task — channel ships first, instruction lands with it.
func writeUpdateText(out io.Writer, r release.Result) error {
	switch r.Status {
	case release.StatusUpToDate:
		_, err := fmt.Fprintf(out, "aikata %s is up to date.\n", r.Current)
		return err
	case release.StatusDevBuild:
		_, err := fmt.Fprintf(out,
			"aikata is running a development build (%s); latest release is %s.\n%s",
			r.Current, r.Latest, upgradeGuidance(r))
		return err
	case release.StatusAhead:
		_, err := fmt.Fprintf(out,
			"aikata %s is newer than the latest published release (%s); no action needed.\n",
			r.Current, r.Latest)
		return err
	case release.StatusUpdateAvailable:
		_, err := fmt.Fprintf(out,
			"%s (current) -> %s (latest available)\n\n%s",
			r.Current, r.Latest, upgradeGuidance(r))
		return err
	default:
		// Defensive: an unknown status from a future release.Client
		// version should not crash the cli; surface the raw shape.
		_, err := fmt.Fprintf(out,
			"aikata: unexpected status %q (current=%s latest=%s)\n",
			r.Status, r.Current, r.Latest)
		return err
	}
}

func upgradeGuidance(r release.Result) string {
	releaseURL := r.ReleaseURL
	if releaseURL == "" {
		releaseURL = "https://github.com/shigindo-inc/aikata/releases/latest"
	}
	return "To upgrade, pick the path matching how you installed aikata:\n" +
		"  go install github.com/shigindo-inc/aikata/cmd/aikata@latest\n" +
		"  curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh\n" +
		"  manual download: " + releaseURL + "\n"
}

// updateJSONReport mirrors the envelope shape used by doctor / list /
// describe (`version: 1`, plus a `kind` tag) so downstream consumers
// can dispatch on `kind` without sniffing field names.
type updateJSONReport struct {
	Version    int            `json:"version"`
	Kind       string         `json:"kind"`
	Current    string         `json:"current"`
	Latest     string         `json:"latest"`
	Status     release.Status `json:"status"`
	ReleaseURL string         `json:"release_url"`
}

const updateJSONSchemaVersion = 1

func writeUpdateJSON(out io.Writer, r release.Result) error {
	report := updateJSONReport{
		Version:    updateJSONSchemaVersion,
		Kind:       "update-check",
		Current:    r.Current,
		Latest:     r.Latest,
		Status:     r.Status,
		ReleaseURL: r.ReleaseURL,
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
