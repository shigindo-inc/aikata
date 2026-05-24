package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixedDate() time.Time { return time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC) }

func writeMD(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestFix_MissingFrontmatterAddsBlock(t *testing.T) {
	tmp := t.TempDir()
	writeMD(t, filepath.Join(tmp, "NOTE.md"), "# Body\n")

	res, err := Fix(Options{TargetDir: tmp, Now: fixedDate()}, []Issue{
		{Level: LevelError, File: "NOTE.md", Code: "frontmatter.missing"},
	})
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if res.Fixed != 1 || len(res.Files) != 1 {
		t.Fatalf("FixResult = %+v, want Fixed=1 Files=[NOTE.md]", res)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, "NOTE.md"))
	want := "---\nproject: TODO\nstatus: draft\nversion: 0.0.0\nupdated: 2026-05-24\naudience: [human, agent]\n---\n\n# Body\n"
	if string(got) != want {
		t.Errorf("file contents:\n--- got ---\n%s\n--- want ---\n%s", string(got), want)
	}
}

func TestFix_MissingKeyInjectsBeforeClosingFence(t *testing.T) {
	tmp := t.TempDir()
	// SPEC.md missing `updated`.
	original := "---\nproject: sample\nstatus: draft\nversion: 0.0.1\naudience: [human, agent]\n---\n\n# Spec\n"
	writeMD(t, filepath.Join(tmp, "SPEC.md"), original)

	res, err := Fix(Options{TargetDir: tmp, Now: fixedDate()}, []Issue{
		{Level: LevelError, File: "SPEC.md", Code: "frontmatter.missing-key.updated"},
	})
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if res.Fixed != 1 {
		t.Fatalf("Fixed = %d, want 1", res.Fixed)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, "SPEC.md"))
	if !strings.Contains(string(got), "updated: 2026-05-24") {
		t.Errorf("expected updated key injection, got:\n%s", string(got))
	}
	// Body must still be intact.
	if !strings.HasSuffix(string(got), "---\n\n# Spec\n") {
		t.Errorf("body suffix mangled:\n%s", string(got))
	}
}

func TestFix_StaleUpdatedBumpsToToday(t *testing.T) {
	tmp := t.TempDir()
	original := "---\nproject: sample\nstatus: draft\nversion: 0.0.1\nupdated: 2020-01-01\naudience: [human, agent]\n---\n\n# Glossary\n"
	writeMD(t, filepath.Join(tmp, "GLOSSARY.md"), original)

	res, err := Fix(Options{TargetDir: tmp, Now: fixedDate()}, []Issue{
		{Level: LevelWarning, File: "GLOSSARY.md", Code: "updated.stale"},
	})
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if res.Fixed != 1 {
		t.Fatalf("Fixed = %d, want 1", res.Fixed)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, "GLOSSARY.md"))
	if strings.Contains(string(got), "2020-01-01") {
		t.Errorf("old date still present:\n%s", string(got))
	}
	if !strings.Contains(string(got), "updated: 2026-05-24") {
		t.Errorf("expected today's date, got:\n%s", string(got))
	}
}

func TestFix_SkipsUnknownCodes(t *testing.T) {
	tmp := t.TempDir()
	writeMD(t, filepath.Join(tmp, "X.md"), "untouched")

	res, err := Fix(Options{TargetDir: tmp, Now: fixedDate()}, []Issue{
		{Level: LevelInfo, File: "X.md", Code: "some.unknown.code"},
		{Level: LevelInfo, File: "X.md"}, // empty code
	})
	if err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if res.Fixed != 0 || res.Skipped != 2 {
		t.Fatalf("FixResult = %+v, want Fixed=0 Skipped=2", res)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, "X.md"))
	if string(got) != "untouched" {
		t.Errorf("file modified unexpectedly: %q", string(got))
	}
}

func TestFix_IsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	writeMD(t, filepath.Join(tmp, "NOTE.md"), "# Body\n")
	opts := Options{TargetDir: tmp, Now: fixedDate()}
	issues := []Issue{{Level: LevelError, File: "NOTE.md", Code: "frontmatter.missing"}}

	if _, err := Fix(opts, issues); err != nil {
		t.Fatalf("first Fix: %v", err)
	}
	// Second invocation should not append a duplicate block.
	if _, err := Fix(opts, issues); err != nil {
		t.Fatalf("second Fix: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(tmp, "NOTE.md"))
	if strings.Count(string(got), "---") != 2 {
		t.Errorf("expected exactly one frontmatter block (2 fences), got:\n%s", string(got))
	}
}

// End-to-end: Run, then Fix, then Run again — the second Run should
// see fewer fixable issues.
func TestFix_ReducesIssuesEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// Make SPEC.md frontmatter stale and drop `version:` from GLOSSARY.md.
	stale := "---\nproject: sample\nstatus: draft\nversion: 0.0.1\nupdated: 2020-01-01\naudience: [human, agent]\n---\n\n# Spec\n"
	if err := os.WriteFile(filepath.Join(tmp, "SPEC.md"), []byte(stale), 0o644); err != nil {
		t.Fatalf("rewrite SPEC.md: %v", err)
	}

	opts := Options{TargetDir: tmp, Now: fixedDate()}
	before, err := Run(opts)
	if err != nil {
		t.Fatalf("Run before: %v", err)
	}
	beforeFixable := countFixable(before)
	if beforeFixable == 0 {
		t.Fatalf("test setup did not produce fixable issues; got: %+v", before)
	}

	if _, err := Fix(opts, before); err != nil {
		t.Fatalf("Fix: %v", err)
	}

	after, err := Run(opts)
	if err != nil {
		t.Fatalf("Run after: %v", err)
	}
	if countFixable(after) >= beforeFixable {
		t.Errorf("expected fewer fixable issues after Fix; before=%d after=%d", beforeFixable, countFixable(after))
	}
}

func countFixable(issues []Issue) int {
	n := 0
	for _, iss := range issues {
		if iss.Code == "" {
			continue
		}
		for _, c := range FixableCodes {
			if iss.Code == c || (strings.HasPrefix(c, "frontmatter.missing-key.") && iss.Code == c) {
				n++
				break
			}
		}
	}
	return n
}
