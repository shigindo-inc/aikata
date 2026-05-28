package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// writeAikataYamlV2 seeds a minimal v2 aikata.yaml under root for the
// enable / new tests. v1 callers use config.Default() directly when
// they need to test the v1 -> v2 lazy migration path; this helper
// writes the modern shape.
func writeAikataYamlV2(t *testing.T, root, name string) {
	t.Helper()
	cfg := config.Default(name, "en")
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("seed aikata.yaml: %v", err)
	}
}

// runEnableFromRoot runs `aikata enable <args...>` with TargetDir set
// to the given root by changing the process working directory.
func runEnableFromRoot(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"enable"}, args...))
	cerr := cmd.Execute()
	return out.String(), errBuf.String(), cerr
}

// TestEnableMemory_HappyPath verifies that `aikata enable memory`
// renders the memory templates, records them in the manifest, and
// flips `components.memory: true` in aikata.yaml (the schema-v2
// invariant).
func TestEnableMemory_HappyPath(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")

	_, _, err := runEnableFromRoot(t, root, "memory")
	if err != nil {
		t.Fatalf("enable memory: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "docs", "memory", "user.md")); statErr != nil {
		t.Errorf("docs/memory/user.md missing: %v", statErr)
	}

	cfg, _, err := config.LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !cfg.Components.Memory {
		t.Errorf("Components.Memory should be true after enable memory, got %+v", cfg.Components)
	}
}

// TestEnableMonorepo_HappyPath asserts the v0.7.1-introduced
// monorepo capability flips the schema-v2 flag and renders the
// monorepo scaffold files.
func TestEnableMonorepo_HappyPath(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")

	_, _, err := runEnableFromRoot(t, root, "monorepo")
	if err != nil {
		t.Fatalf("enable monorepo: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "docs", "monorepo.md")); statErr != nil {
		t.Errorf("docs/monorepo.md missing: %v", statErr)
	}

	cfg, _, err := config.LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !cfg.Components.Monorepo {
		t.Errorf("Components.Monorepo should be true after enable monorepo, got %+v", cfg.Components)
	}
}

// TestEnableMemory_OnV1ConfigMigrates verifies the lazy migration
// path: an enable on a v1 config triggers v1 -> v2 migration via
// LoadMigrated and then flips components.memory.
func TestEnableMemory_OnV1ConfigMigrates(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".aikata"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	v1 := []byte("version: 1\nproject:\n  name: legacy\n  lang: en\nai_tools:\n  - claude\nfeatures:\n  tdd: true\n  monorepo: false\n  obsidian_hints: false\n")
	if err := os.WriteFile(filepath.Join(root, ".aikata", "aikata.yaml"), v1, 0o644); err != nil {
		t.Fatalf("write v1: %v", err)
	}

	if _, _, err := runEnableFromRoot(t, root, "memory"); err != nil {
		t.Fatalf("enable memory on v1: %v", err)
	}

	cfg, _, err := config.LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if cfg.Version != config.Version {
		t.Errorf("expected v2 after migration, got %d", cfg.Version)
	}
	if !cfg.Components.Memory {
		t.Errorf("Components.Memory should be true, got %+v", cfg.Components)
	}
	if !cfg.Components.TDD {
		t.Errorf("Components.TDD should have been lifted from features.tdd, got %+v", cfg.Components)
	}
}

// TestEnable_WithoutConfig_Exit2 mirrors the pre-v0.7.1
// TestAddMemory_WithoutConfig_Exit2 behaviour: enabling outside an
// aikata project errors with exit code 2.
func TestEnable_WithoutConfig_Exit2(t *testing.T) {
	root := t.TempDir()
	// No aikata.yaml seeded.
	_, _, err := runEnableFromRoot(t, root, "memory")
	if err == nil {
		t.Fatalf("expected error when aikata config missing")
	}
	var exit *ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exit.Code != 2 {
		t.Errorf("exit code = %d, want 2", exit.Code)
	}
}

// TestEnable_UnknownCapability_Exit2 confirms the parent dispatcher
// surfaces a registry-aware error for an unknown leaf name.
func TestEnable_UnknownCapability_Exit2(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")
	_, _, err := runEnableFromRoot(t, root, "nope-not-a-capability")
	if err == nil {
		t.Fatalf("expected error for unknown capability")
	}
	if !strings.Contains(err.Error(), "unknown capability") {
		t.Errorf("expected message to mention 'unknown capability'; got %q", err.Error())
	}
}
