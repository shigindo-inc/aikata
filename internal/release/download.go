package release

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// defaultDownloadBase is the GitHub Releases download root for the
// aikata repository. Release assets live at
// `<base>/<tag>/<asset-name>`. Kept separate from the API endpoint
// (which lists the latest release) because downloads come from the
// release CDN, not the API host.
const defaultDownloadBase = "https://github.com/shigindo-inc/aikata/releases/download"

// maxAssetBytes caps a downloaded artifact so a malicious or broken
// server cannot exhaust memory / disk. aikata archives are well under a
// few MB; 64 MiB is a generous ceiling.
const maxAssetBytes = 64 << 20

// DownloadBase overrides the release-download root. Empty falls back to
// defaultDownloadBase. Tests point this at an httptest.Server so the
// self-update path can be exercised without hitting GitHub. It lives on
// Client alongside Endpoint so a single injected Client seams both the
// "check latest" and "download artifact" halves of self-update.
func (c *Client) downloadBase() string {
	if c.DownloadBase != "" {
		return c.DownloadBase
	}
	return defaultDownloadBase
}

// DownloadReleaseAsset downloads the named release asset and the
// release's checksums.txt for the given tag into destDir. It returns the
// on-disk paths of both files. The caller verifies the asset against the
// checksums before using it (see VerifySHA256); downloading and
// verifying are deliberately separate steps so the verify-before-use
// ordering is explicit at the call site.
func (c *Client) DownloadReleaseAsset(ctx context.Context, tag, assetName, destDir string) (assetPath, checksumsPath string, err error) {
	base := strings.TrimRight(c.downloadBase(), "/")
	assetURL := fmt.Sprintf("%s/%s/%s", base, tag, assetName)
	checksumsURL := fmt.Sprintf("%s/%s/checksums.txt", base, tag)

	assetPath = filepath.Join(destDir, assetName)
	checksumsPath = filepath.Join(destDir, "checksums.txt")

	if err := c.downloadTo(ctx, assetURL, assetPath); err != nil {
		return "", "", err
	}
	if err := c.downloadTo(ctx, checksumsURL, checksumsPath); err != nil {
		return "", "", err
	}
	return assetPath, checksumsPath, nil
}

// downloadTo fetches url into dest, bounded by maxAssetBytes. A non-2xx
// status is surfaced as an error (no silent partial writes).
func (c *Client) downloadTo(ctx context.Context, url, dest string) error {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("release: build request %s: %w", url, err)
	}
	userAgent := c.UserAgent
	if userAgent == "" {
		userAgent = "aikata-cli"
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("release: GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("release: %s returned %d: %s", url, resp.StatusCode, strings.TrimSpace(string(preview)))
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("release: create %s: %w", dest, err)
	}
	// Bound the copy; a stream longer than the cap is an error rather
	// than a silently truncated file.
	n, copyErr := io.Copy(f, io.LimitReader(resp.Body, maxAssetBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("release: download %s: %w", url, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(dest)
		return fmt.Errorf("release: close %s: %w", dest, closeErr)
	}
	if n > maxAssetBytes {
		_ = os.Remove(dest)
		return fmt.Errorf("release: %s exceeds the %d-byte asset size cap", url, maxAssetBytes)
	}
	return nil
}

// VerifySHA256 checks that the file at assetPath matches its entry in a
// GoReleaser-format checksums.txt (`<hex>  <filename>`, two spaces). The
// lookup is by the asset's base name. A missing entry or a mismatch is
// an error; this is the gate that must pass before an extracted binary
// is trusted (ADR 0035 D2).
func VerifySHA256(assetPath, checksumsPath string) error {
	want, err := checksumFor(checksumsPath, filepath.Base(assetPath))
	if err != nil {
		return err
	}
	f, err := os.Open(assetPath)
	if err != nil {
		return fmt.Errorf("release: open %s for hashing: %w", assetPath, err)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("release: hash %s: %w", assetPath, err)
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("release: checksum mismatch for %s: computed %s, expected %s",
			filepath.Base(assetPath), got, want)
	}
	return nil
}

// checksumFor returns the hex digest recorded for name in a
// GoReleaser-format checksums.txt. Each line is `<hex>␣␣<filename>`.
func checksumFor(checksumsPath, name string) (string, error) {
	f, err := os.Open(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("release: open checksums %s: %w", checksumsPath, err)
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[1] == name {
			return fields[0], nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("release: read checksums %s: %w", checksumsPath, err)
	}
	return "", fmt.Errorf("release: no checksum entry for %q in checksums.txt", name)
}
