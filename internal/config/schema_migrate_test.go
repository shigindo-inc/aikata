package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateAikataYaml_CurrentVersionPassesThrough(t *testing.T) {
	cfg := Default("samplekata", "en")
	body, err := Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, migrated, err := MigrateAikataYaml(body)
	if err != nil {
		t.Fatalf("MigrateAikataYaml: %v", err)
	}
	if migrated {
		t.Errorf("current-version payload should not report migrated=true")
	}
	if got.Project.Name != "samplekata" {
		t.Errorf("project name lost: %+v", got)
	}
}

func TestMigrateAikataYaml_FutureSchemaIsError(t *testing.T) {
	body := []byte("version: 999\nproject:\n  name: future\n  lang: en\n")
	_, _, err := MigrateAikataYaml(body)
	if !errors.Is(err, ErrFutureSchema) {
		t.Errorf("expected ErrFutureSchema, got %v", err)
	}
}

func TestMigrateAikataYaml_MissingMigratorIsError(t *testing.T) {
	// Hand-craft v0 input — there's no migrator for 0 → 1 because
	// v0 never shipped, so the registry refuses to advance.
	body := []byte("version: 0\nproject:\n  name: legacy\n  lang: en\n")
	_, _, err := MigrateAikataYaml(body)
	if err == nil {
		t.Fatalf("expected error for unknown source version")
	}
	if errors.Is(err, ErrFutureSchema) {
		t.Errorf("v0 should not be classified as future schema")
	}
}

func TestMigrateAikataYaml_V1ToV2_LiftsFeaturesAndPersists(t *testing.T) {
	// A v1 payload with `features.tdd: true` and
	// `features.monorepo: true` should migrate to v2, with both keys
	// lifted into the new `components:` block. The migrated bytes are
	// written back to disk so subsequent commands see the v2 shape.
	root := t.TempDir()
	primary := PrimaryPath(root)
	if err := os.MkdirAll(filepath.Dir(primary), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	payload := []byte(`version: 1
project:
  name: legacy
  lang: en
features:
  tdd: true
  monorepo: true
  obsidian_hints: false
`)
	if err := os.WriteFile(primary, payload, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, migrated, err := LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !migrated {
		t.Fatalf("v1 -> v2 should report migrated=true")
	}
	if cfg.Version != Version {
		t.Errorf("Version = %d, want %d (v2)", cfg.Version, Version)
	}
	if !cfg.Components.TDD {
		t.Errorf("Components.TDD should have been lifted from features.tdd: %+v", cfg.Components)
	}
	if !cfg.Components.Monorepo {
		t.Errorf("Components.Monorepo should have been lifted from features.monorepo: %+v", cfg.Components)
	}
	if cfg.Features["tdd"] {
		t.Errorf("features.tdd should be removed after lift: %+v", cfg.Features)
	}
	if cfg.Features["monorepo"] {
		t.Errorf("features.monorepo should be removed after lift: %+v", cfg.Features)
	}
	if got, want := cfg.Features["obsidian_hints"], false; got != want {
		t.Errorf("non-lifted features key should survive: features.obsidian_hints = %v, want %v", got, want)
	}

	// The persisted file must now report v2.
	persisted, err := os.ReadFile(primary)
	if err != nil {
		t.Fatalf("read persisted: %v", err)
	}
	if !strings.Contains(string(persisted), "version: 2") {
		t.Errorf("persisted file did not reach v2:\n%s", persisted)
	}
}

func TestMigrateAikataYaml_V1ToV2_IsIdempotent(t *testing.T) {
	// Running migration twice on the same root must be a no-op the
	// second time. The first LoadMigrated rewrites the file; the
	// second sees v2 and reports migrated=false.
	root := t.TempDir()
	primary := PrimaryPath(root)
	if err := os.MkdirAll(filepath.Dir(primary), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(primary, []byte("version: 1\nproject:\n  name: idem\n  lang: en\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, _, err := LoadMigrated(root); err != nil {
		t.Fatalf("LoadMigrated #1: %v", err)
	}
	_, migrated, err := LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated #2: %v", err)
	}
	if migrated {
		t.Errorf("second LoadMigrated should be a no-op; got migrated=true")
	}
}

func TestMigrateAikataYaml_V1ToV2_NoFeaturesBlock(t *testing.T) {
	// A v1 payload with no `features:` block at all must still
	// migrate cleanly, producing an all-false components block.
	payload := []byte("version: 1\nproject:\n  name: bare\n  lang: en\n")
	cfg, migrated, err := MigrateAikataYaml(payload)
	if err != nil {
		t.Fatalf("MigrateAikataYaml: %v", err)
	}
	if !migrated {
		t.Errorf("v1 payload should report migrated=true even without features")
	}
	if cfg.Components.TDD || cfg.Components.Monorepo {
		t.Errorf("absent features must produce all-false components: %+v", cfg.Components)
	}
}

func TestLoadMigrated_MissingConfigReturnsErrNotExist(t *testing.T) {
	root := t.TempDir()
	_, _, err := LoadMigrated(root)
	if err == nil {
		t.Fatalf("expected error on missing config")
	}
}
