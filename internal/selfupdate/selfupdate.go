// Package selfupdate implements the binary-swap half of
// `aikata update --apply` for the prebuilt-binary channels
// (install-script, github-release). It downloads the release archive
// for the host platform, verifies it against checksums.txt, extracts
// the aikata binary, and atomically replaces the running executable.
//
// Safety model: ADR 0035. The download/verify HTTP boundary lives in
// internal/release; this package owns extraction and the atomic
// in-place replace. exePath is always supplied by the caller (never
// discovered with os.Executable inside this package) so tests can never
// risk the test runner.
package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/shigindo-inc/aikata/internal/release"
)

// binaryName is the GoReleaser `binary:` name shipped inside every
// archive (no platform suffix on the POSIX swap path).
const binaryName = "aikata"

// maxBinaryBytes caps the extracted binary size, mirroring the archive
// cap in internal/release. Defends against a tar entry that claims an
// absurd size.
const maxBinaryBytes = 64 << 20

// Action enumerates what Apply did, so the CLI can phrase the result.
type Action string

const (
	// ActionUpdated means the running binary was replaced.
	ActionUpdated Action = "updated"
	// ActionUpToDate means the running version is already the latest.
	ActionUpToDate Action = "up-to-date"
	// ActionAhead means the running version is newer than the latest
	// published release (no action taken).
	ActionAhead Action = "ahead"
	// ActionDevBuild means the running binary is a development build;
	// self-update is refused.
	ActionDevBuild Action = "dev-build"
)

// Outcome is the structured result of Apply.
type Outcome struct {
	Action Action
	From   string // running version
	To     string // latest version (may equal From for up-to-date)
}

// ErrPermission signals that the running binary could not be replaced
// because its directory is not writable (e.g. a root-owned install
// prefix). The CLI maps this to an actionable message.
var ErrPermission = errors.New("selfupdate: permission denied replacing the binary")

// HostAssetName returns the GoReleaser archive name for the running
// platform at the given version. runtime.GOOS / runtime.GOARCH already
// equal GoReleaser's {{.Os}}/{{.Arch}} template values, so no mapping
// is needed. Windows uses .zip; everything else .tar.gz — though the
// CLI gates Windows out before Apply is reached (ADR 0035 D3).
func HostAssetName(version string) string {
	ver := version
	if len(ver) > 0 && (ver[0] == 'v' || ver[0] == 'V') {
		ver = ver[1:]
	}
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("%s_%s_%s_%s.%s", binaryName, ver, runtime.GOOS, runtime.GOARCH, ext)
}

// Apply checks for a newer release and, when one exists, downloads it,
// verifies it, and replaces the binary at exePath. It performs no
// channel routing or OS gating — the caller (cli/update.go) is
// responsible for restricting Apply to swappable channels on POSIX.
//
// Download happens only when the check reports update-available; all
// other statuses return an Outcome with no network artifact written
// (ADR 0035 D4).
func Apply(ctx context.Context, client *release.Client, current, exePath string) (Outcome, error) {
	res, err := client.CheckLatest(ctx, current)
	if err != nil {
		return Outcome{}, fmt.Errorf("selfupdate: %w", err)
	}
	out := Outcome{From: res.Current, To: res.Latest}
	switch res.Status {
	case release.StatusDevBuild:
		out.Action = ActionDevBuild
		return out, nil
	case release.StatusUpToDate:
		out.Action = ActionUpToDate
		return out, nil
	case release.StatusAhead:
		out.Action = ActionAhead
		return out, nil
	case release.StatusUpdateAvailable:
		// proceed below
	default:
		return out, fmt.Errorf("selfupdate: unexpected release status %q", res.Status)
	}

	tmp, err := os.MkdirTemp("", "aikata-selfupdate-*")
	if err != nil {
		return out, fmt.Errorf("selfupdate: temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	asset := HostAssetName(res.Latest)
	assetPath, checksumsPath, err := client.DownloadReleaseAsset(ctx, res.Latest, asset, tmp)
	if err != nil {
		return out, fmt.Errorf("selfupdate: download: %w", err)
	}
	// Verify the ARCHIVE before extracting anything — checksums.txt
	// lists archive hashes, not the inner binary (ADR 0035 D2).
	if err := release.VerifySHA256(assetPath, checksumsPath); err != nil {
		return out, fmt.Errorf("selfupdate: %w", err)
	}
	bin, err := extractBinary(assetPath)
	if err != nil {
		return out, fmt.Errorf("selfupdate: %w", err)
	}
	if err := replaceExecutable(exePath, bin); err != nil {
		return out, err
	}
	out.Action = ActionUpdated
	return out, nil
}

// extractBinary returns the contents of the tar entry named exactly
// `aikata` (root, no directory prefix) from a .tar.gz archive. The
// archive also bundles LICENSE / README / CHANGELOG, so the lookup is
// by name, not "the only file".
func extractBinary(archivePath string) ([]byte, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive %s: %w", archivePath, err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gunzip %s: %w", archivePath, err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar %s: %w", archivePath, err)
		}
		if filepath.Clean(hdr.Name) != binaryName {
			continue
		}
		body, err := io.ReadAll(io.LimitReader(tr, maxBinaryBytes+1))
		if err != nil {
			return nil, fmt.Errorf("read %q from archive: %w", binaryName, err)
		}
		if int64(len(body)) > maxBinaryBytes {
			return nil, fmt.Errorf("extracted %q exceeds the %d-byte cap", binaryName, maxBinaryBytes)
		}
		return body, nil
	}
	return nil, fmt.Errorf("archive %s does not contain %q", filepath.Base(archivePath), binaryName)
}

// replaceExecutable atomically replaces the file at exePath with
// content. It writes a temp file in the same directory (so os.Rename is
// a same-filesystem atomic move), sets it executable, and renames it
// over exePath. On POSIX, replacing a running binary this way is safe:
// the live process keeps the old inode. Any failure removes the temp
// file so nothing is left behind; a permission error is wrapped as
// ErrPermission.
func replaceExecutable(exePath string, content []byte) error {
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".aikata-update-*")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("%w: cannot write to %s", ErrPermission, dir)
		}
		return fmt.Errorf("selfupdate: create temp in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("selfupdate: write new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("selfupdate: close new binary: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		cleanup()
		return fmt.Errorf("selfupdate: chmod new binary: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		cleanup()
		if os.IsPermission(err) {
			return fmt.Errorf("%w: cannot replace %s", ErrPermission, exePath)
		}
		return fmt.Errorf("selfupdate: replace %s: %w", exePath, err)
	}
	return nil
}
