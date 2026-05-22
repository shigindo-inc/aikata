package generate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shigindo-inc/aikata/internal/config"
)

func fixedClock() func() time.Time {
	t := time.Date(2026, time.May, 22, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

// seedAgentsMD writes a minimal AGENTS.md so ClaudeProvider's existence
// guard passes. The body is irrelevant — generate doesn't read it.
func seedAgentsMD(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS\n"), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}
}

func TestGet_UnknownToolReturnsErr(t *testing.T) {
	_, err := Get("nope")
	if !errors.Is(err, ErrUnknownAITool) {
		t.Fatalf("expected ErrUnknownAITool, got %v", err)
	}
}

func TestGet_ClaudeIsRegistered(t *testing.T) {
	p, err := Get("claude")
	if err != nil {
		t.Fatalf("Get(claude): %v", err)
	}
	if p.Name() != "claude" {
		t.Errorf("Name() = %q, want %q", p.Name(), "claude")
	}
}

func TestKnownTools_IsSorted(t *testing.T) {
	got := KnownTools()
	if len(got) == 0 {
		t.Fatalf("KnownTools empty")
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Errorf("KnownTools not sorted: %v", got)
			break
		}
	}
}

func TestRun_RequiresTargetDir(t *testing.T) {
	err := Run(Context{Project: config.Default("x", "en")})
	if err == nil {
		t.Fatalf("expected error for empty TargetDir")
	}
}

func TestRun_RequiresEnabledAITools(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("x", "en")
	cfg.AITools = nil
	err := Run(Context{TargetDir: tmp, Project: cfg})
	if err == nil {
		t.Fatalf("expected error when ai_tools is empty")
	}
}

func TestRun_UnknownAIToolBubblesUp(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("x", "en")
	cfg.AITools = []string{"nope"}
	err := Run(Context{TargetDir: tmp, Project: cfg})
	if !errors.Is(err, ErrUnknownAITool) {
		t.Fatalf("expected ErrUnknownAITool wrapped, got %v", err)
	}
}

func TestRun_ClaudeRequiresAgentsMD(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("samplekata", "en")
	err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err == nil {
		t.Fatalf("expected error when AGENTS.md missing")
	}
	if !strings.Contains(err.Error(), "AGENTS.md") {
		t.Errorf("error should mention AGENTS.md, got: %v", err)
	}
}

func TestRun_ClaudeProducesCLAUDE(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	out := string(body)
	for _, needle := range []string{
		"AGENTS.md",
		"canonical",
		"aikata generate",
		"samplekata",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("CLAUDE.md missing %q:\n%s", needle, out)
		}
	}
	// Sections for absent files must not appear.
	for _, absent := range []string{"SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md"} {
		if strings.Contains(out, "["+absent+"]") {
			t.Errorf("CLAUDE.md references %s but the file does not exist in target", absent)
		}
	}
}

func TestRun_ClaudeReflectsAvailableDocs(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	for _, name := range []string{"README.md", "SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("# "+name+"\n"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	cfg := config.Default("samplekata", "en")
	if err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	out := string(body)
	for _, needle := range []string{"SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md", "README.md"} {
		if !strings.Contains(out, needle) {
			t.Errorf("CLAUDE.md should reference %s when it exists:\n%s", needle, out)
		}
	}
}

func TestRun_OSSReadiness(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	if err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if strings.Contains(string(body), "/Users/") {
		t.Errorf("CLAUDE.md contains a /Users/ path (OSS leak)")
	}
	for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
		if strings.Contains(string(body), sec) {
			t.Errorf("CLAUDE.md contains secret-like pattern %q", sec)
		}
	}
}
