package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_PrimaryOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, PrimaryDir, Filename), "version: 1\n")

	path, isLegacy, err := Resolve(root)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if isLegacy {
		t.Errorf("isLegacy = true, want false")
	}
	if want := PrimaryPath(root); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolve_LegacyOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, LegacyDir, Filename), "version: 1\n")

	path, isLegacy, err := Resolve(root)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !isLegacy {
		t.Errorf("isLegacy = false, want true")
	}
	if want := LegacyPath(root); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolve_BothPresent_PrefersPrimary(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, PrimaryDir, Filename), "version: 1\n# primary\n")
	mustWrite(t, filepath.Join(root, LegacyDir, Filename), "version: 1\n# legacy\n")

	path, isLegacy, err := Resolve(root)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if isLegacy {
		t.Errorf("isLegacy = true, want false (primary wins)")
	}
	if want := PrimaryPath(root); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolve_NeitherPresent_ReturnsErrNotExist(t *testing.T) {
	root := t.TempDir()

	_, _, err := Resolve(root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Resolve: err = %v, want fs.ErrNotExist", err)
	}
}

func TestPrimaryAndLegacyPathLayout(t *testing.T) {
	root := "/tmp/project"
	if got, want := PrimaryPath(root), filepath.Join("/tmp/project", ".aikata", "aikata.yaml"); got != want {
		t.Errorf("PrimaryPath = %q, want %q", got, want)
	}
	if got, want := LegacyPath(root), filepath.Join("/tmp/project", ".ai", "aikata.yaml"); got != want {
		t.Errorf("LegacyPath = %q, want %q", got, want)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
