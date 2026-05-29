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

	path, err := Resolve(root)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := PrimaryPath(root); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolve_AIOnlyReturnsErrNotExist(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".ai", Filename), "version: 1\n")

	if _, err := Resolve(root); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Resolve: err = %v, want fs.ErrNotExist", err)
	}
}

func TestResolve_BothPresentIgnoresAI(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, PrimaryDir, Filename), "version: 1\n# primary\n")
	mustWrite(t, filepath.Join(root, ".ai", Filename), "version: 1\n# ignored\n")

	path, err := Resolve(root)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := PrimaryPath(root); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestResolve_NeitherPresent_ReturnsErrNotExist(t *testing.T) {
	root := t.TempDir()

	_, err := Resolve(root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Resolve: err = %v, want fs.ErrNotExist", err)
	}
}

func TestPrimaryPathLayout(t *testing.T) {
	root := "/tmp/project"
	if got, want := PrimaryPath(root), filepath.Join("/tmp/project", ".aikata", "aikata.yaml"); got != want {
		t.Errorf("PrimaryPath = %q, want %q", got, want)
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
