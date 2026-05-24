package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixedClock() time.Time { return time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC) }

func TestRenderMemory_RequiresProjectName(t *testing.T) {
	if _, err := RenderMemory(MemoryParams{Lang: "en"}); err == nil {
		t.Fatal("expected error when ProjectName is empty")
	}
}

func TestRenderMemory_EmitsCanonicalFiveFiles(t *testing.T) {
	files, err := RenderMemory(MemoryParams{
		Lang:        "en",
		ProjectName: "demo",
		Clock:       func() time.Time { return fixedClock() },
	})
	if err != nil {
		t.Fatalf("RenderMemory: %v", err)
	}
	want := []string{
		"docs/memory/README.md",
		"docs/memory/feedback.md",
		"docs/memory/project.md",
		"docs/memory/reference.md",
		"docs/memory/user.md",
	}
	for _, w := range want {
		if _, ok := files[w]; !ok {
			t.Errorf("missing key %q in rendered output; got keys %v", w, sortedKeys(files))
		}
	}
	user := files["docs/memory/user.md"]
	if !strings.Contains(user, "project: demo") {
		t.Errorf("expected frontmatter project: demo, got:\n%s", user)
	}
	if !strings.Contains(user, "updated: 2026-05-24") {
		t.Errorf("expected updated: 2026-05-24 from fixed clock, got:\n%s", user)
	}
}

func TestRenderMemory_FallsBackToEnForUnknownLang(t *testing.T) {
	files, err := RenderMemory(MemoryParams{
		Lang:        "xx",
		ProjectName: "demo",
		Clock:       func() time.Time { return fixedClock() },
	})
	if err != nil {
		t.Fatalf("RenderMemory: %v", err)
	}
	if _, ok := files["docs/memory/user.md"]; !ok {
		t.Errorf("expected en fallback to produce user.md; got keys %v", sortedKeys(files))
	}
}

func TestMemoryAdd_WritesFilesAndIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	var stderrBuf bytes.Buffer
	ctx := AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       func() time.Time { return fixedClock() },
		Stderr:      &stderrBuf,
	}
	if err := Memory.Add(ctx); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs/memory/user.md")); err != nil {
		t.Fatalf("expected user.md to exist after Add: %v", err)
	}

	// Second invocation is a no-op (idempotent) — files already exist.
	stderrBuf.Reset()
	if err := Memory.Add(ctx); err != nil {
		t.Fatalf("second Add: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "already present") {
		t.Errorf("expected idempotent notice on stderr; got %q", stderrBuf.String())
	}
}

func TestMemoryAdd_RequiresProjectName(t *testing.T) {
	tmp := t.TempDir()
	if err := Memory.Add(AddContext{TargetDir: tmp, Lang: "en"}); err == nil {
		t.Fatal("expected error when ProjectName is empty")
	}
}

func TestMemoryAdd_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	var stdoutBuf bytes.Buffer
	err := Memory.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		DryRun:      true,
		Stdout:      &stdoutBuf,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs")); !os.IsNotExist(err) {
		t.Errorf("expected docs/ not to exist after dry-run; got err=%v", err)
	}
	if !strings.Contains(stdoutBuf.String(), "docs/memory/user.md") {
		t.Errorf("expected dry-run plan to list user.md; got %q", stdoutBuf.String())
	}
}
