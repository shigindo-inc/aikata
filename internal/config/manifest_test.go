package config

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashContent_Stable(t *testing.T) {
	got := HashContent([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("HashContent(%q) = %q, want %q", "hello", got, want)
	}
}

func TestBuildManifest_Deterministic(t *testing.T) {
	rendered := map[string]string{
		"AGENTS.md":             "alpha",
		"SPEC.md":               "beta",
		"docs/adr/0001-x.md":    "gamma",
		".aikata/aikata.yaml":   "version: 1",
		".aikata/manifest.yaml": "should-be-skipped",
	}
	excludes := []string{".aikata/aikata.yaml", ".aikata/manifest.yaml"}

	m := BuildManifest("standard", "en", rendered, excludes)

	if m.Version != ManifestVersion {
		t.Errorf("Version = %d, want %d", m.Version, ManifestVersion)
	}
	if m.Preset != "standard" || m.Lang != "en" {
		t.Errorf("preset/lang mismatch: %+v", m)
	}
	if len(m.Files) != 3 {
		t.Fatalf("Files length = %d, want 3 (config + manifest excluded); got %+v", len(m.Files), m.Files)
	}
	// Verify sorted ascending by Path.
	wantPaths := []string{"AGENTS.md", "SPEC.md", "docs/adr/0001-x.md"}
	for i, want := range wantPaths {
		if m.Files[i].Path != want {
			t.Errorf("Files[%d].Path = %q, want %q", i, m.Files[i].Path, want)
		}
	}
	// SHA256 fields are non-empty hex.
	for _, f := range m.Files {
		if len(f.SHA256) != 64 {
			t.Errorf("Files[%s].SHA256 length = %d, want 64", f.Path, len(f.SHA256))
		}
	}
}

func TestManifest_Roundtrip(t *testing.T) {
	original := BuildManifest("flutter", "ja", map[string]string{
		"AGENTS.md": "hello",
		"SPEC.md":   "world",
	}, nil)
	body, err := MarshalManifest(original)
	if err != nil {
		t.Fatalf("MarshalManifest: %v", err)
	}
	got, err := UnmarshalManifest(body)
	if err != nil {
		t.Fatalf("UnmarshalManifest: %v", err)
	}
	if got.Version != original.Version || got.Preset != original.Preset || got.Lang != original.Lang {
		t.Errorf("scalar mismatch: got %+v want %+v", got, original)
	}
	if len(got.Files) != len(original.Files) {
		t.Fatalf("Files length mismatch: got %d want %d", len(got.Files), len(original.Files))
	}
	for i := range got.Files {
		if got.Files[i] != original.Files[i] {
			t.Errorf("Files[%d] = %+v, want %+v", i, got.Files[i], original.Files[i])
		}
	}
}

func TestUnmarshalManifest_MissingVersionIsError(t *testing.T) {
	body := []byte("preset: standard\nlang: en\nfiles: []\n")
	_, err := UnmarshalManifest(body)
	if err == nil {
		t.Fatalf("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error should mention version: %v", err)
	}
}

func TestUnmarshalManifest_GarbledYAML(t *testing.T) {
	_, err := UnmarshalManifest([]byte("not: yaml: at all: : :"))
	if err == nil {
		t.Fatalf("expected parse error for garbled YAML")
	}
}

func TestSaveAndLoadManifest(t *testing.T) {
	root := t.TempDir()
	m := BuildManifest("standard", "en", map[string]string{
		"AGENTS.md": "content",
	}, nil)
	if err := SaveManifest(root, m); err != nil {
		t.Fatalf("SaveManifest: %v", err)
	}
	// File should land under .aikata/manifest.yaml.
	expectedPath := filepath.Join(root, PrimaryDir, ManifestFilename)
	if expectedPath != ManifestPath(root) {
		t.Errorf("ManifestPath = %q, want %q", ManifestPath(root), expectedPath)
	}
	got, err := LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got.Preset != "standard" || len(got.Files) != 1 || got.Files[0].Path != "AGENTS.md" {
		t.Errorf("LoadManifest roundtrip drift: %+v", got)
	}
}

func TestLoadManifest_MissingReturnsErrNotExist(t *testing.T) {
	root := t.TempDir()
	_, err := LoadManifest(root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("LoadManifest on empty dir = %v, want fs.ErrNotExist", err)
	}
}
