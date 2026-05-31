package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/release"
)

const testLatest = "v9.9.9"

// buildArchive returns a .tar.gz containing an `aikata` entry with the
// given binary bytes plus a decoy LICENSE entry (so extraction must
// select by name, not "the only file").
func buildArchive(t *testing.T, binary []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range []struct {
		name string
		body []byte
	}{
		{"LICENSE", []byte("MIT decoy")},
		{"aikata", binary},
	} {
		if err := tw.WriteHeader(&tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body))}); err != nil {
			t.Fatalf("tar header %s: %v", e.name, err)
		}
		if _, err := tw.Write(e.body); err != nil {
			t.Fatalf("tar write %s: %v", e.name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

// newServer wires an httptest.Server that answers the latest-release
// API plus the asset and checksums download paths. checksumsBody lets a
// test serve a tampered checksums.txt.
func newServer(t *testing.T, archive []byte, checksumsBody string) (*release.Client, string) {
	t.Helper()
	asset := HostAssetName(testLatest)
	mux := http.NewServeMux()
	latestJSON := fmt.Sprintf(`{"tag_name":%q,"html_url":"https://example.test/r"}`, testLatest)
	mux.HandleFunc("/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(latestJSON))
	})
	mux.HandleFunc("/dl/"+testLatest+"/"+asset, func(w http.ResponseWriter, _ *http.Request) {
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write(archive)
	})
	mux.HandleFunc("/dl/"+testLatest+"/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(checksumsBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := &release.Client{
		Endpoint:     srv.URL + "/releases/latest",
		DownloadBase: srv.URL + "/dl",
	}
	return client, asset
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// TestApply_VerifiesBeforeSwap is the security-critical pair: a tampered
// archive (checksums.txt hash does not match) must error AND leave the
// target binary unchanged.
func TestApply_TamperedArchive_NoSwap(t *testing.T) {
	newBin := []byte("THE-NEW-BINARY-BYTES")
	archive := buildArchive(t, newBin)
	// Wrong hash on purpose.
	asset := HostAssetName(testLatest)
	tampered := "deadbeef" + strings.Repeat("0", 56) + "  " + asset + "\n"
	client, _ := newServer(t, archive, tampered)

	exe := filepath.Join(t.TempDir(), "aikata")
	const original = "ORIGINAL-BINARY"
	if err := os.WriteFile(exe, []byte(original), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := Apply(context.Background(), client, "v0.0.1", exe)
	if err == nil {
		t.Fatal("expected a checksum-mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("expected checksum mismatch error, got: %v", err)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != original {
		t.Errorf("target binary was modified despite verify failure: %q", got)
	}
}

func TestApply_ValidArchive_SwapsBinary(t *testing.T) {
	newBin := []byte("THE-NEW-BINARY-BYTES")
	archive := buildArchive(t, newBin)
	asset := HostAssetName(testLatest)
	checksums := sha256Hex(archive) + "  " + asset + "\n"
	client, _ := newServer(t, archive, checksums)

	exe := filepath.Join(t.TempDir(), "aikata")
	if err := os.WriteFile(exe, []byte("ORIGINAL-BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, err := Apply(context.Background(), client, "v0.0.1", exe)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if out.Action != ActionUpdated {
		t.Errorf("Action = %q, want updated", out.Action)
	}
	if out.To != testLatest {
		t.Errorf("To = %q, want %q", out.To, testLatest)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != string(newBin) {
		t.Errorf("binary not swapped: got %q, want %q", got, newBin)
	}
}

func TestApply_DevBuild_Refuses(t *testing.T) {
	archive := buildArchive(t, []byte("x"))
	client, asset := newServer(t, archive, sha256Hex(archive)+"  "+HostAssetName(testLatest)+"\n")
	_ = asset
	exe := filepath.Join(t.TempDir(), "aikata")
	_ = os.WriteFile(exe, []byte("ORIGINAL"), 0o755)

	out, err := Apply(context.Background(), client, "0.0.1-dev", exe)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if out.Action != ActionDevBuild {
		t.Errorf("Action = %q, want dev-build", out.Action)
	}
	if got, _ := os.ReadFile(exe); string(got) != "ORIGINAL" {
		t.Errorf("dev build should not be swapped; got %q", got)
	}
}

func TestApply_UpToDate_NoSwap(t *testing.T) {
	archive := buildArchive(t, []byte("x"))
	client, _ := newServer(t, archive, sha256Hex(archive)+"  "+HostAssetName(testLatest)+"\n")
	exe := filepath.Join(t.TempDir(), "aikata")
	_ = os.WriteFile(exe, []byte("ORIGINAL"), 0o755)

	out, err := Apply(context.Background(), client, testLatest, exe)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if out.Action != ActionUpToDate {
		t.Errorf("Action = %q, want up-to-date", out.Action)
	}
	if got, _ := os.ReadFile(exe); string(got) != "ORIGINAL" {
		t.Errorf("up-to-date should not swap; got %q", got)
	}
}

func TestHostAssetName(t *testing.T) {
	got := HostAssetName("v1.2.3")
	if !strings.HasPrefix(got, "aikata_1.2.3_") {
		t.Errorf("HostAssetName = %q, want aikata_1.2.3_<os>_<arch>.<ext>", got)
	}
	if !strings.HasSuffix(got, ".tar.gz") && !strings.HasSuffix(got, ".zip") {
		t.Errorf("HostAssetName = %q, missing archive extension", got)
	}
}

func TestReplaceExecutable_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory-permission semantics; chmod does not gate writes on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root ignores directory permissions")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "aikata")
	if err := os.WriteFile(exe, []byte("ORIGINAL"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Make the directory read-only so the temp-file create / rename fails.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := replaceExecutable(exe, []byte("NEW"))
	if err == nil {
		t.Fatal("expected a permission error")
	}
	if !errors.Is(err, ErrPermission) {
		t.Errorf("expected ErrPermission, got: %v", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "ORIGINAL" {
		t.Errorf("target changed despite permission failure: %q", got)
	}
}
