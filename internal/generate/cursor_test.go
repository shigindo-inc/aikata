package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func TestCursorProvider_Name(t *testing.T) {
	if got := (CursorProvider{}).Name(); got != "cursor" {
		t.Errorf("Name() = %q, want %q", got, "cursor")
	}
}

func TestRun_CursorRequiresAgentsMD(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("samplekata", "en")
	cfg.AITools = []string{"cursor"}
	_, err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err == nil {
		t.Fatalf("expected error when AGENTS.md missing")
	}
	if !strings.Contains(err.Error(), "AGENTS.md") {
		t.Errorf("error should mention AGENTS.md, got: %v", err)
	}
}

func TestRun_CursorProducesMDC(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	cfg.AITools = []string{"cursor"}
	counts, err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if counts["cursor"] != 1 {
		t.Errorf("counts[cursor] = %d, want 1", counts["cursor"])
	}
	mdcPath := filepath.Join(tmp, ".cursor", "rules", "main.mdc")
	body, err := os.ReadFile(mdcPath)
	if err != nil {
		t.Fatalf("read main.mdc: %v", err)
	}
	out := string(body)
	for _, needle := range []string{
		"description: Read AGENTS.md",
		"alwaysApply: true",
		"AGENTS.md",
		"samplekata",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("main.mdc missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_CursorOSSReadiness(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	cfg.AITools = []string{"cursor"}
	if _, err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".cursor", "rules", "main.mdc"))
	if err != nil {
		t.Fatalf("read main.mdc: %v", err)
	}
	if strings.Contains(string(body), "/Users/") {
		t.Errorf("main.mdc contains a /Users/ path (OSS leak)")
	}
	for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
		if strings.Contains(string(body), sec) {
			t.Errorf("main.mdc contains secret-like pattern %q", sec)
		}
	}
}
