package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func TestCodexProvider_Name(t *testing.T) {
	if got := (CodexProvider{}).Name(); got != "codex" {
		t.Errorf("Name() = %q, want %q", got, "codex")
	}
}

func TestRun_CodexIsNoOp(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	cfg.AITools = []string{"codex"}
	counts, err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if counts["codex"] != 0 {
		t.Errorf("counts[codex] = %d, want 0 (codex reads AGENTS.md directly)", counts["codex"])
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("read tmp: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "AGENTS.md" {
			t.Errorf("codex provider wrote unexpected file: %s", e.Name())
		}
	}
}

func TestRun_MixedClaudeCursorCodex(t *testing.T) {
	tmp := t.TempDir()
	seedAgentsMD(t, tmp)
	cfg := config.Default("samplekata", "en")
	cfg.AITools = []string{"claude", "cursor", "codex"}
	counts, err := Run(Context{TargetDir: tmp, Project: cfg, Clock: fixedClock()})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := map[string]int{"claude": 1, "cursor": 1, "codex": 0}
	for name, n := range want {
		if counts[name] != n {
			t.Errorf("counts[%s] = %d, want %d", name, counts[name], n)
		}
	}
	for _, rel := range []string{"CLAUDE.md", filepath.Join(".cursor", "rules", "main.mdc")} {
		if _, err := os.Stat(filepath.Join(tmp, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}
}
