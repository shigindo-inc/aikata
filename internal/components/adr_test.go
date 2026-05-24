package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAdrAdd_HappyPathWritesNNNNFile(t *testing.T) {
	tmp := t.TempDir()
	var stdout bytes.Buffer
	ctx := AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"Use", "Go", "modules"},
		Clock:       func() time.Time { return time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC) },
		Stdout:      &stdout,
	}
	if err := Adr.Add(ctx); err != nil {
		t.Fatalf("Add: %v", err)
	}
	want := filepath.Join(tmp, "docs", "adr", "0001-use-go-modules.md")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected %s to exist: %v", want, err)
	}
	body, _ := os.ReadFile(want)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "# ADR 0001 - Use Go modules") {
		t.Errorf("missing canonical heading in body:\n%s", bodyStr)
	}
	if !strings.Contains(bodyStr, "project: demo") {
		t.Errorf("missing project frontmatter:\n%s", bodyStr)
	}
	if !strings.Contains(stdout.String(), "docs/adr/0001-use-go-modules.md") {
		t.Errorf("expected stdout to mention the path; got %q", stdout.String())
	}
}

func TestAdrAdd_NumbersIncrementPastExisting(t *testing.T) {
	tmp := t.TempDir()
	adrDir := filepath.Join(tmp, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Seed with two existing ADRs so Next == 3.
	for _, name := range []string{"0001-foo.md", "0002-bar.md"} {
		if err := os.WriteFile(filepath.Join(adrDir, name), []byte("stub"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	err := Adr.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"third"},
		Stdout:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(adrDir, "0003-third.md")); err != nil {
		t.Errorf("expected 0003-third.md to exist: %v", err)
	}
}

func TestAdrAdd_DuplicateSlugIsRejected(t *testing.T) {
	tmp := t.TempDir()
	adrDir := filepath.Join(tmp, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(adrDir, "0007-use-go-modules.md"), []byte("stub"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	err := Adr.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"use", "go", "modules"},
		Stdout:      &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for duplicate slug")
	}
	if !strings.Contains(err.Error(), "0007") {
		t.Errorf("expected error to cite the conflicting ADR; got %v", err)
	}
}

func TestAdrAdd_RequiresTitle(t *testing.T) {
	if err := Adr.Add(AddContext{TargetDir: t.TempDir(), ProjectName: "demo"}); err == nil {
		t.Fatal("expected error when title is missing")
	}
}

func TestAdrAdd_BlankSlugRejected(t *testing.T) {
	err := Adr.Add(AddContext{
		TargetDir:   t.TempDir(),
		ProjectName: "demo",
		Args:        []string{"!!!"},
	})
	if err == nil {
		t.Fatal("expected error when title produces an empty slug")
	}
	if !strings.Contains(err.Error(), "empty slug") {
		t.Errorf("expected empty-slug message; got %v", err)
	}
}

func TestAdrAdd_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := Adr.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Args:        []string{"dry", "run"},
		DryRun:      true,
		Stdout:      &stdout,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs", "adr")); !os.IsNotExist(err) {
		t.Errorf("expected docs/adr/ not to exist after --dry-run; got err=%v", err)
	}
	if !strings.Contains(stdout.String(), "0001-dry-run.md") {
		t.Errorf("expected dry-run plan to list the target path; got %q", stdout.String())
	}
}

func TestAdrAdd_JaLangPicksJaTemplate(t *testing.T) {
	tmp := t.TempDir()
	err := Adr.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "ja",
		Args:        []string{"japanese", "adr"},
		Stdout:      &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(tmp, "docs", "adr", "0001-japanese-adr.md"))
	if !strings.Contains(string(body), "ステータス") {
		t.Errorf("expected ja template to render Japanese headings; got:\n%s", string(body))
	}
}
