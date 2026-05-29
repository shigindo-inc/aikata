package components

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func seedAIToolConfig(t *testing.T, root, name string, tools []string) {
	t.Helper()
	cfg := config.Default(name, "en")
	cfg.AITools = tools
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}
}

func TestAIToolAdd_HappyPathAppendsToConfig(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"claude"})
	var stdout bytes.Buffer
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"cursor"},
		Stdout:      &stdout,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	cfg, err := config.Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !contains(cfg.AITools, "cursor") {
		t.Errorf("expected cfg.AITools to include cursor; got %v", cfg.AITools)
	}
	if !contains(cfg.AITools, "claude") {
		t.Errorf("expected cfg.AITools to retain claude; got %v", cfg.AITools)
	}
}

func TestAIToolAdd_SortsResult(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"cursor"})
	if err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"claude"},
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	cfg, _ := config.Load(tmp)
	if got, want := cfg.AITools, []string{"claude", "cursor"}; !equalSlice(got, want) {
		t.Errorf("expected sorted AITools=%v; got %v", want, got)
	}
}

func TestAIToolAdd_UnknownToolIsRejected(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"claude"})
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"copilot"},
		Stdout:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for unknown ai-tool")
	}
	if !strings.Contains(err.Error(), "copilot") || !strings.Contains(err.Error(), "claude") {
		t.Errorf("expected error to name unknown tool and known set; got %v", err)
	}
}

func TestAIToolAdd_AlreadyEnabledIsNoOp(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"claude", "cursor"})
	var stderr bytes.Buffer
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"cursor"},
		Stdout:      &bytes.Buffer{},
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !strings.Contains(stderr.String(), "already enabled") {
		t.Errorf("expected idempotent notice; got %q", stderr.String())
	}
	cfg, _ := config.Load(tmp)
	if got, want := len(cfg.AITools), 2; got != want {
		t.Errorf("expected AITools length unchanged at %d; got %d (%v)", want, got, cfg.AITools)
	}
}

func TestAIToolAdd_RequiresName(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"claude"})
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
	})
	if err == nil {
		t.Fatal("expected error when tool name is missing")
	}
}

func TestAIToolAdd_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	seedAIToolConfig(t, tmp, "demo", []string{"claude"})
	var stdout bytes.Buffer
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"cursor"},
		DryRun:      true,
		Stdout:      &stdout,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	cfg, _ := config.Load(tmp)
	if contains(cfg.AITools, "cursor") {
		t.Errorf("expected cfg.AITools not to be modified after --dry-run; got %v", cfg.AITools)
	}
	if !strings.Contains(stdout.String(), "cursor") {
		t.Errorf("expected dry-run plan to mention the tool; got %q", stdout.String())
	}
}

func TestAIToolAdd_MissingConfigIsErrorWithGuidance(t *testing.T) {
	tmp := t.TempDir()
	err := AITool.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"cursor"},
		Stdout:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
	if !strings.Contains(err.Error(), "aikata init") {
		t.Errorf("expected actionable hint to run aikata init; got %v", err)
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
