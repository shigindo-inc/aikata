package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/install"
	"github.com/shigindo-inc/aikata/internal/release"
	"github.com/shigindo-inc/aikata/internal/selfupdate"
)

// envReleaseEndpoint is a test-only seam that points the update check
// at an httptest.Server instead of the live GitHub API. Production
// callers leave this env var unset; release.Client.Endpoint then
// falls back to the canonical endpoint baked into the release package.
const envReleaseEndpoint = "AIKATA_RELEASE_ENDPOINT_FOR_TEST"

// envDownloadBase is the companion test-only seam for the self-update
// (`--apply`) download path: it points release-asset downloads at an
// httptest.Server. Unset in production, where release.Client.DownloadBase
// falls back to the canonical GitHub Releases download URL.
const envDownloadBase = "AIKATA_DOWNLOAD_BASE_FOR_TEST"

// releasesLatestURL is the human-facing latest-release page, used in
// manual-download guidance for channels that cannot self-update.
const releasesLatestURL = "https://github.com/shigindo-inc/aikata/releases/latest"

// goInstallCmd is the channel-native upgrade for `go install` users.
const goInstallCmd = "go install github.com/shigindo-inc/aikata/cmd/aikata@latest"

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
		apply   bool
		jsonOut bool
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for or apply aikata CLI updates",
		Long: "Compare the running aikata binary against the latest published\n" +
			"GitHub Release.\n\n" +
			"`--check` reads the GitHub Releases API and reports whether a newer\n" +
			"version is available. `--apply` performs a native self-update for\n" +
			"the prebuilt-binary channels (install-script, github-release):\n" +
			"download, verify the SHA-256 against checksums.txt, and atomically\n" +
			"replace the running binary (ADR 0035). `go install` users are shown\n" +
			"the channel-native upgrade command; Homebrew / npm are deferred to\n" +
			"v0.9.9; Windows and package-manager installs get manual guidance.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if apply {
				return applyUpdate(out, version, install.Detect(""))
			}
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
	cmd.Flags().BoolVar(&apply, "apply", false, "download, verify, and install the latest release in place (native channels)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON report instead of the text format")
	return cmd
}

// applyUpdate routes `aikata update --apply` by install channel
// (ADR 0035 D1). Only the prebuilt-binary channels (install-script,
// github-release) perform an in-place swap; everything else gets an
// actionable message and exits 0. The detected source is injected so
// routing is unit-testable without faking the binary's install
// metadata.
func applyUpdate(out io.Writer, version string, src install.Source) error {
	switch src {
	case install.SourceHomebrew:
		_, err := fmt.Fprintf(out,
			"aikata was installed via Homebrew; upgrade with your package manager:\n"+
				"  brew upgrade aikata\n")
		return err
	case install.SourceNpm:
		_, err := fmt.Fprintf(out,
			"aikata was installed via npm; upgrade with your package manager:\n"+
				"  npm install -g aikata    # or: npx aikata@latest\n")
		return err
	case install.SourceGoInstall:
		_, err := fmt.Fprintf(out,
			"aikata was installed with `go install`; upgrade with the Go toolchain:\n"+
				"  "+goInstallCmd+"\n")
		return err
	case install.SourceUnknown:
		_, err := fmt.Fprintf(out,
			"aikata could not determine how it was installed; download the latest\n"+
				"release manually:\n  "+releasesLatestURL+"\n")
		return err
	}

	// install-script / github-release: in-place swap. Windows cannot
	// overwrite a running .exe, so gate it before any download.
	if runtime.GOOS == "windows" {
		_, err := fmt.Fprintf(out,
			"aikata self-update cannot replace a running .exe on Windows; download\n"+
				"the latest release manually:\n  "+releasesLatestURL+"\n"+
				"  or, with Go: "+goInstallCmd+"\n")
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("update: locate running binary: %w", err)
	}
	client := &release.Client{
		UserAgent:    "aikata/" + version,
		Endpoint:     os.Getenv(envReleaseEndpoint),
		DownloadBase: os.Getenv(envDownloadBase),
	}
	outcome, err := selfupdate.Apply(context.Background(), client, version, exe)
	if err != nil {
		if errors.Is(err, selfupdate.ErrPermission) {
			return fmt.Errorf("update: %w\n  re-run with sufficient permissions, or reinstall via the install script:\n  curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh", err)
		}
		return fmt.Errorf("update: %w", err)
	}
	return writeApplyOutcome(out, outcome)
}

// writeApplyOutcome renders the result of a self-update attempt.
func writeApplyOutcome(out io.Writer, o selfupdate.Outcome) error {
	switch o.Action {
	case selfupdate.ActionUpdated:
		_, err := fmt.Fprintf(out, "updated aikata %s -> %s\n", o.From, o.To)
		return err
	case selfupdate.ActionUpToDate:
		_, err := fmt.Fprintf(out, "aikata %s is up to date.\n", o.From)
		return err
	case selfupdate.ActionAhead:
		_, err := fmt.Fprintf(out,
			"aikata %s is newer than the latest published release (%s); no action needed.\n", o.From, o.To)
		return err
	case selfupdate.ActionDevBuild:
		_, err := fmt.Fprintf(out,
			"aikata is running a development build (%s); self-update applies to released binaries.\n"+
				"Build from source or install a release via:\n  "+releasesLatestURL+"\n", o.From)
		return err
	default:
		_, err := fmt.Fprintf(out, "aikata: unexpected self-update outcome %q\n", o.Action)
		return err
	}
}

// printUpdateGuidance is the response to bare `aikata update`. The
// message points users at `--check` and explains why self-update is
// not yet available so support burden does not accrue on this
// transitional release.
func printUpdateGuidance(out io.Writer) error {
	const msg = "Run `aikata update --check` to see if a newer release is available,\n" +
		"or `aikata update --apply` to install it in place (native channels;\n" +
		"`go install` / Homebrew / npm users are shown their channel's command).\n"
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
