package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// TestEnableWorkflowGit_HappyPath verifies the full CLI surface
// `aikata enable workflow git` (ADR 0026): the guide is written, the
// `workflows:` list axis records git, the manifest tracks the guide, and
// the canonical AGENTS.md gains a single pointer section.
func TestEnableWorkflowGit_HappyPath(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")
	// A canonical AGENTS.md so the pointer injection has a target.
	agents := "---\nproject: tester\nstatus: draft\nversion: 0.0.1\nupdated: 2026-05-24\naudience: agent\n---\n\n# Agent Instructions\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(agents), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	if _, _, err := runEnableFromRoot(t, root, "workflow", "git"); err != nil {
		t.Fatalf("enable workflow git: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "workflows", "git.md")); err != nil {
		t.Errorf("docs/workflows/git.md missing: %v", err)
	}

	cfg, _, err := config.LoadMigrated(root)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	found := false
	for _, w := range cfg.Workflows {
		if w == "git" {
			found = true
		}
	}
	if !found {
		t.Errorf("aikata.yaml workflows should include git; got %v", cfg.Workflows)
	}

	body, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if strings.Count(string(body), "## Workflow") != 1 {
		t.Errorf("AGENTS.md should carry exactly one ## Workflow pointer; got:\n%s", body)
	}
	if !strings.Contains(string(body), "docs/workflows/git.md") {
		t.Errorf("AGENTS.md pointer should link the guide; got:\n%s", body)
	}
}

// TestEnableWorkflow_UnknownDomain asserts an unknown domain is a clean
// error naming the known set, mirroring `enable stack <unknown>`.
func TestEnableWorkflow_UnknownDomain(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")

	_, _, err := runEnableFromRoot(t, root, "workflow", "mercurial")
	if err == nil {
		t.Fatal("unknown workflow domain should error")
	}
	if !strings.Contains(err.Error(), "unknown workflow") {
		t.Errorf("error should name the unknown workflow; got %v", err)
	}
}
