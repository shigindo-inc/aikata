package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestSave_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	cfg := Default("demo", "en")
	cfg.Stacks = []string{"flutter"}

	if err := Save(tmp, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(PrimaryPath(tmp)); err != nil {
		t.Fatalf("expected primary config to exist: %v", err)
	}

	got, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, cfg) {
		t.Errorf("round-trip diff:\n got=%+v\nwant=%+v", got, cfg)
	}
}

func TestSave_OverwritesExistingPrimary(t *testing.T) {
	tmp := t.TempDir()
	first := Default("first", "en")
	if err := Save(tmp, first); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	second := Default("second", "en")
	if err := Save(tmp, second); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	got, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Project.Name != "second" {
		t.Errorf("expected overwrite to win, got %q", got.Project.Name)
	}
}

func TestSave_CreatesPrimaryDir(t *testing.T) {
	tmp := t.TempDir()
	if _, err := os.Stat(filepath.Join(tmp, PrimaryDir)); !os.IsNotExist(err) {
		t.Fatalf("test setup: .aikata/ already exists: %v", err)
	}
	if err := Save(tmp, Default("demo", "en")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if info, err := os.Stat(filepath.Join(tmp, PrimaryDir)); err != nil || !info.IsDir() {
		t.Errorf("expected .aikata/ to be a directory after Save (err=%v info=%v)", err, info)
	}
}

func TestSave_AtomicLeavesNoPartialOnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only directory permission semantics differ on Windows")
	}
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, PrimaryDir), 0o555); err != nil {
		t.Fatalf("mkdir read-only .aikata: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(tmp, PrimaryDir), 0o755) })

	if err := Save(tmp, Default("demo", "en")); err == nil {
		t.Fatal("expected Save to fail on read-only .aikata/")
	}
	entries, err := os.ReadDir(filepath.Join(tmp, PrimaryDir))
	if err != nil {
		t.Fatalf("read .aikata/: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no partial files after failure; got %d entries", len(entries))
	}
}

func TestLoad_NotExistMapsToOsError(t *testing.T) {
	tmp := t.TempDir()
	_, err := Load(tmp)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

func TestLoad_AIOnlyReturnsErrNotExist(t *testing.T) {
	tmp := t.TempDir()
	cfg := Default("legacy-project", "en")
	body, err := Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	path := filepath.Join(tmp, ".ai", Filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write .ai/aikata.yaml: %v", err)
	}

	if _, err := Load(tmp); !os.IsNotExist(err) {
		t.Fatalf("Load: err = %v, want not-exist", err)
	}
}

func TestLoad_IgnoresAIWhenPrimaryExists(t *testing.T) {
	tmp := t.TempDir()
	primary := Default("primary-project", "en")
	if err := Save(tmp, primary); err != nil {
		t.Fatalf("Save primary: %v", err)
	}
	aiDir := filepath.Join(tmp, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	legacy := Default("legacy-project", "en")
	body, err := Marshal(legacy)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(aiDir, Filename), body, 0o644); err != nil {
		t.Fatalf("write .ai/aikata.yaml: %v", err)
	}

	got, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Project.Name != "primary-project" {
		t.Errorf("Project.Name = %q, want primary-project", got.Project.Name)
	}
}
