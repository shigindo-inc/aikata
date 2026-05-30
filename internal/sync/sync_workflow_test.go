package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/components"
)

// enableWorkflowGit runs the enable-tier workflow component against a
// scaffolded project, mirroring `aikata enable workflow git`.
func enableWorkflowGit(t *testing.T, root string) {
	t.Helper()
	if err := components.Workflow.Add(components.AddContext{
		TargetDir:   root,
		ProjectName: "samplekata",
		Lang:        "en",
		Args:        []string{"git"},
		// nil Clock -> time.Now, matching sync's own upstream rendering;
		// both render the {{now}} date for the same day so the guide
		// classifies as unchanged rather than a date-only re-render.
		Clock:  nil,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("enable workflow git: %v", err)
	}
}

func statusOf(res RunResult, path string) Status {
	for _, f := range res.Files {
		if f.Path == path {
			return f.Status
		}
	}
	return ""
}

// TestRun_WorkflowGuide_ParticipatesAsManagedDocument asserts that a
// guide enabled post-init via `aikata enable workflow git` is rendered
// into the sync upstream tree (ADR 0026): it classifies as `unchanged`
// (not `upstream-removed`) and its manifest entry survives repeated
// syncs, exactly like any other aikata-managed document.
func TestRun_WorkflowGuide_ParticipatesAsManagedDocument(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	enableWorkflowGit(t, root)

	guideRel := "docs/workflows/git.md"
	if manifestHash(t, root, guideRel) == "" {
		t.Fatal("guide missing from manifest after enable")
	}

	res1 := runSync(t, root, Options{})
	if got := statusOf(res1, guideRel); got != StatusUnchanged {
		t.Errorf("first sync: %s = %q, want %q", guideRel, got, StatusUnchanged)
	}
	if manifestHash(t, root, guideRel) == "" {
		t.Errorf("guide manifest entry must survive the first sync")
	}

	res2 := runSync(t, root, Options{})
	if got := statusOf(res2, guideRel); got != StatusUnchanged {
		t.Errorf("second sync: %s = %q, want %q", guideRel, got, StatusUnchanged)
	}
	if _, err := os.Stat(filepath.Join(root, guideRel)); err != nil {
		t.Errorf("guide must remain on disk: %v", err)
	}
}

// TestRun_WorkflowPointer_SurvivesSync is the AGENTS.md half of the ADR
// 0026 contract crossed with ADR 0025 D1: the pointer `enable workflow`
// injects into the canonical AGENTS.md is a user-only-edit (its ancestor
// stays at the upstream rendering because AGENTS.md is not re-recorded),
// so it survives unlimited syncs without being overwritten.
func TestRun_WorkflowPointer_SurvivesSync(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	enableWorkflowGit(t, root)

	agents := filepath.Join(root, "AGENTS.md")
	before, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(before), "docs/workflows/git.md") {
		t.Fatal("precondition: AGENTS.md should already carry the pointer after enable")
	}

	res := runSync(t, root, Options{})
	if got := statusOf(res, "AGENTS.md"); got != StatusUserOnlyEdit {
		t.Errorf("AGENTS.md = %q, want %q (pointer is a user-only-edit)", got, StatusUserOnlyEdit)
	}

	// Two syncs, the pointer must persist verbatim.
	runSync(t, root, Options{})
	after, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("read AGENTS.md (after): %v", err)
	}
	if strings.Count(string(after), "## Workflow") != 1 {
		t.Errorf("pointer must remain exactly once after syncs; got %d", strings.Count(string(after), "## Workflow"))
	}
	if !strings.Contains(string(after), "docs/workflows/git.md") {
		t.Errorf("pointer link must survive syncs")
	}
}
