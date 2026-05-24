package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shigindo-inc/aikata/internal/config"
)

func seedConfig(t *testing.T, root, name string, stacks []string) {
	t.Helper()
	cfg := config.Default(name, "en")
	cfg.Stacks = stacks
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}
}

func TestStacks_ContainsBundledIdentifiers(t *testing.T) {
	got := Stacks()
	want := map[string]bool{"flutter": false, "typescript": false}
	for _, s := range got {
		if _, ok := want[s]; ok {
			want[s] = true
		}
	}
	for k, found := range want {
		if !found {
			t.Errorf("Stacks() missing %q; got %v", k, got)
		}
	}
}

func TestStackAdd_HappyPathWritesFileAndUpdatesConfig(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", nil)
	var stdout bytes.Buffer
	ctx := AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"flutter"},
		Clock:       func() time.Time { return time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC) },
		Stdout:      &stdout,
	}
	if err := Stack.Add(ctx); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs/stacks/flutter.md")); err != nil {
		t.Errorf("expected docs/stacks/flutter.md to exist: %v", err)
	}
	cfg, _, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !contains(cfg.Stacks, "flutter") {
		t.Errorf("expected cfg.Stacks to include flutter; got %v", cfg.Stacks)
	}
}

func TestStackAdd_UnknownStackIsRejected(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", nil)
	err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"rust"},
		Stdout:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for unknown stack")
	}
	if !strings.Contains(err.Error(), "rust") || !strings.Contains(err.Error(), "flutter") {
		t.Errorf("expected error to name unknown stack and known set; got %v", err)
	}
}

func TestStackAdd_AlreadyPresentIsNoOp(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", []string{"flutter"})
	stackDir := filepath.Join(tmp, "docs", "stacks")
	if err := os.MkdirAll(stackDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stackDir, "flutter.md"), []byte("custom"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	var stderr bytes.Buffer
	err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"flutter"},
		Stdout:      &bytes.Buffer{},
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(stackDir, "flutter.md"))
	if string(body) != "custom" {
		t.Errorf("existing file was overwritten; got %q", string(body))
	}
	if !strings.Contains(stderr.String(), "already present") {
		t.Errorf("expected idempotent notice; got %q", stderr.String())
	}
}

func TestStackAdd_FileExistsButCfgMissingUpdatesCfgOnly(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", nil)
	stackDir := filepath.Join(tmp, "docs", "stacks")
	if err := os.MkdirAll(stackDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stackDir, "flutter.md"), []byte("custom"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"flutter"},
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(stackDir, "flutter.md"))
	if string(body) != "custom" {
		t.Errorf("user-customized file was overwritten; got %q", string(body))
	}
	cfg, _, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !contains(cfg.Stacks, "flutter") {
		t.Errorf("expected cfg.Stacks to be updated; got %v", cfg.Stacks)
	}
}

func TestStackAdd_RequiresName(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", nil)
	err := Stack.Add(AddContext{TargetDir: tmp, ProjectName: "demo", Lang: "en"})
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestStackAdd_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	seedConfig(t, tmp, "demo", nil)
	var stdout bytes.Buffer
	err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"typescript"},
		DryRun:      true,
		Stdout:      &stdout,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs/stacks/typescript.md")); !os.IsNotExist(err) {
		t.Errorf("expected docs/stacks/typescript.md not to exist after --dry-run; got err=%v", err)
	}
	cfg, _, _ := config.Load(tmp)
	if contains(cfg.Stacks, "typescript") {
		t.Errorf("expected cfg.Stacks not to be modified after --dry-run; got %v", cfg.Stacks)
	}
	if !strings.Contains(stdout.String(), "docs/stacks/typescript.md") {
		t.Errorf("expected dry-run plan to list target file; got %q", stdout.String())
	}
}

func TestStackAdd_MissingConfigIsErrorWithGuidance(t *testing.T) {
	tmp := t.TempDir()
	// No config seeded.
	err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"flutter"},
		Stdout:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
	if !strings.Contains(err.Error(), "aikata init") {
		t.Errorf("expected actionable hint to run aikata init; got %v", err)
	}
}
