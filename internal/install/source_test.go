package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_LdflagOverride(t *testing.T) {
	original := ldflagSource
	t.Cleanup(func() { ldflagSource = original })

	ldflagSource = "github-release"
	if got := Detect(""); got != SourceGitHubRelease {
		t.Errorf("Detect() with ldflag = %q, want %q", got, SourceGitHubRelease)
	}

	ldflagSource = " INSTALL-SCRIPT "
	// Whitespace + casing should still resolve via normalizeSource.
	if got := Detect(""); got != SourceInstallScript {
		t.Errorf("Detect() with case-and-whitespace ldflag = %q, want %q", got, SourceInstallScript)
	}

	ldflagSource = "bogus-channel"
	// Unknown ldflag → fall through to metadata file / build info /
	// SourceUnknown.
	got := Detect("")
	if got == SourceUnknown {
		// Expected on test machines without a metadata file or go-install build info.
		return
	}
	// On a machine where build info or sibling file happens to be present,
	// the fall-through path is also acceptable.
	t.Logf("Detect() fell through to %q (acceptable)", got)
}

func TestDetect_MetadataFile(t *testing.T) {
	original := ldflagSource
	t.Cleanup(func() { ldflagSource = original })
	ldflagSource = "" // force fall-through to metadata path

	tmp := t.TempDir()
	exePath := filepath.Join(tmp, "aikata")
	if err := os.WriteFile(exePath, []byte("fake binary"), 0o755); err != nil {
		t.Fatalf("write fake exe: %v", err)
	}
	metaPath := filepath.Join(tmp, MetadataFilename)
	if err := os.WriteFile(metaPath, []byte("install-script\nversion=v0.6.0\n"), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	if got := Detect(exePath); got != SourceInstallScript {
		t.Errorf("Detect with metadata = %q, want %q", got, SourceInstallScript)
	}
}

func TestDetect_MetadataFileMissing_FallsThrough(t *testing.T) {
	original := ldflagSource
	t.Cleanup(func() { ldflagSource = original })
	ldflagSource = ""

	tmp := t.TempDir()
	exePath := filepath.Join(tmp, "aikata")
	if err := os.WriteFile(exePath, []byte("fake binary"), 0o755); err != nil {
		t.Fatalf("write fake exe: %v", err)
	}
	// No metadata file written.
	got := Detect(exePath)
	// With no ldflag and no metadata, Detect drops to build info (which
	// in `go test` reports the test binary path — not a typical aikata
	// `go install`) or returns SourceUnknown. Both outcomes are valid
	// proofs that the metadata path didn't false-positive.
	if got == SourceInstallScript {
		t.Errorf("Detect with missing metadata = %q, want non-install-script", got)
	}
}

func TestNormalizeSource(t *testing.T) {
	cases := map[string]Source{
		"":                 "",
		"  ":               "",
		"github-release":   SourceGitHubRelease,
		"GitHub-Release":   SourceGitHubRelease,
		"install-script":   SourceInstallScript,
		" INSTALL-SCRIPT ": SourceInstallScript,
		"homebrew":         SourceHomebrew,
		"npm":              SourceNpm,
		"go-install":       SourceGoInstall,
		"bogus":            "",
	}
	for in, want := range cases {
		if got := normalizeSource(in); got != want {
			t.Errorf("normalizeSource(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstLine(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"abc":        "abc",
		"abc\n":      "abc",
		"abc\ndef":   "abc",
		"\ndef":      "",
		"a=b\nc=d\n": "a=b",
	}
	for in, want := range cases {
		if got := firstLine(in); got != want {
			t.Errorf("firstLine(%q) = %q, want %q", in, got, want)
		}
	}
}
