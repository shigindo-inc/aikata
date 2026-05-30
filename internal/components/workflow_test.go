package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// seedAgents writes a minimal AGENTS.md so the workflow pointer
// injection has a canonical file to extend. The frontmatter mirrors what
// the preset templates render so doctor stays applicable in tests that
// run it.
func seedAgents(t *testing.T, root string) {
	t.Helper()
	body := "---\nproject: demo\nstatus: draft\nversion: 0.0.1\nupdated: 2026-05-24\naudience: agent\n---\n\n# Agent Instructions for demo\n\n## 1. Project overview\n\nSee SPEC.md.\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}
}

func TestWorkflow_Add_WritesGuideConfigManifest(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	if err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"git"},
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add workflow git: %v", err)
	}

	// (a) guide written
	guide := filepath.Join(tmp, "docs", "workflows", "git.md")
	if _, err := os.Stat(guide); err != nil {
		t.Errorf("docs/workflows/git.md missing: %v", err)
	}

	// (b) config persisted
	cfg, _, err := config.LoadMigrated(tmp)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !contains(cfg.Workflows, "git") {
		t.Errorf("aikata.yaml workflows should include git; got %v", cfg.Workflows)
	}

	// (c) manifest records the guide
	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if !manifestHasPath(m, "docs/workflows/git.md") {
		t.Errorf("manifest should record docs/workflows/git.md; got %v", manifestPaths(m))
	}
}

func TestWorkflow_Add_InjectsAgentsPointer(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")
	seedAgents(t, tmp)

	add := func() {
		t.Helper()
		if err := Workflow.Add(AddContext{
			TargetDir:   tmp,
			ProjectName: "demo",
			Lang:        "en",
			Args:        []string{"git"},
			Clock:       fixedClock,
			Stdout:      &bytes.Buffer{},
			Stderr:      &bytes.Buffer{},
		}); err != nil {
			t.Fatalf("Add workflow git: %v", err)
		}
	}

	add()
	body, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	got := string(body)
	if !strings.Contains(got, "docs/workflows/git.md") {
		t.Errorf("AGENTS.md should reference docs/workflows/git.md after enable; got:\n%s", got)
	}
	if strings.Count(got, "## Workflow") != 1 {
		t.Errorf("AGENTS.md should have exactly one ## Workflow section; got %d", strings.Count(got, "## Workflow"))
	}

	// Idempotent: a second enable must not duplicate the pointer.
	add()
	body2, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md (2): %v", err)
	}
	if strings.Count(string(body2), "## Workflow") != 1 {
		t.Errorf("second enable must not duplicate the ## Workflow pointer; got %d", strings.Count(string(body2), "## Workflow"))
	}
	if !bytes.Equal(body, body2) {
		t.Errorf("second enable must leave AGENTS.md byte-identical")
	}
}

// TestWorkflow_Add_Ja_RendersLocalizedGuideAndPointer exercises the ja
// path end to end: the ja guide template renders without error and the
// AGENTS.md pointer is Japanese (not the hardcoded English section), so a
// ja project's canonical file stays single-language.
func TestWorkflow_Add_Ja_RendersLocalizedGuideAndPointer(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "ja")
	// A Japanese canonical AGENTS.md, mirroring the ja preset.
	ja := "---\nproject: demo\nstatus: draft\nversion: 0.0.1\nupdated: 2026-05-24\naudience: agent\n---\n\n# demo のエージェント指示書\n"
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte(ja), 0o644); err != nil {
		t.Fatalf("seed ja AGENTS.md: %v", err)
	}

	if err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "ja",
		Args:        []string{"git"},
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add workflow git (ja): %v", err)
	}

	// Guide renders (ja template) with valid five-key frontmatter.
	guide, err := os.ReadFile(filepath.Join(tmp, "docs", "workflows", "git.md"))
	if err != nil {
		t.Fatalf("read ja guide: %v", err)
	}
	if !strings.Contains(string(guide), "audience: [human, agent]") {
		t.Errorf("ja guide should carry five-key frontmatter; got:\n%s", guide)
	}
	if !strings.Contains(string(guide), "Git ワークフロー") {
		t.Errorf("ja guide should render the Japanese heading; got:\n%s", guide)
	}

	// Pointer is localized: ja heading present, English heading absent.
	body, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	got := string(body)
	if !strings.Contains(got, "## ワークフロー (Workflow)") {
		t.Errorf("ja AGENTS.md should gain a Japanese ## ワークフロー pointer; got:\n%s", got)
	}
	if strings.Contains(got, "If present, follow") {
		t.Errorf("ja AGENTS.md must not contain the English pointer text; got:\n%s", got)
	}
	if !strings.Contains(got, "docs/workflows/git.md") {
		t.Errorf("ja pointer should link the guide; got:\n%s", got)
	}
}

func TestWorkflow_Add_NoAgentsFileIsGraceful(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")
	// No AGENTS.md on disk.

	if err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"git"},
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add workflow git without AGENTS.md should succeed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("enable must not create AGENTS.md when it is absent")
	}
}

func TestWorkflow_Add_UnknownDomain(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"jujutsu"},
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("unknown workflow domain should error")
	}
	if !strings.Contains(err.Error(), "unknown workflow") {
		t.Errorf("error should name the unknown workflow; got %v", err)
	}
}

func TestWorkflow_Add_NoArgs(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("enable workflow with no domain should error")
	}
}

func TestWorkflow_Add_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")
	seedAgents(t, tmp)

	ctx := AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"git"},
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	}
	if err := Workflow.Add(ctx); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	manifest1, err := os.ReadFile(filepath.Join(tmp, ".aikata", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	var stderr bytes.Buffer
	ctx.Stderr = &stderr
	if err := Workflow.Add(ctx); err != nil {
		t.Fatalf("second Add: %v", err)
	}
	if !strings.Contains(stderr.String(), "already present") {
		t.Errorf("second Add should print an idempotent notice; got %q", stderr.String())
	}
	manifest2, err := os.ReadFile(filepath.Join(tmp, ".aikata", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest (2): %v", err)
	}
	if !bytes.Equal(manifest1, manifest2) {
		t.Errorf("repeated Add should produce a byte-identical manifest")
	}
}

func TestWorkflow_Add_DryRun(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")
	seedAgents(t, tmp)

	var stdout bytes.Buffer
	if err := Workflow.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"git"},
		Clock:       fixedClock,
		DryRun:      true,
		Stdout:      &stdout,
		Stderr:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("dry-run Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs", "workflows", "git.md")); !os.IsNotExist(err) {
		t.Errorf("dry-run must not write the guide")
	}
	out := stdout.String()
	for _, want := range []string{"docs/workflows/git.md", "workflows:", "AGENTS.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run output should mention %q; got:\n%s", want, out)
		}
	}
}
