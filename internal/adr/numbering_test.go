package adr

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("placeholder"), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	entries, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries, got %v", entries)
	}
}

func TestScan_MissingDirectory(t *testing.T) {
	entries, err := Scan(filepath.Join(t.TempDir(), "nonexistent"))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected nil entries for missing dir, got %v", entries)
	}
}

func TestScan_SortsByNumberAscending(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0003-do-no-harm-policy.md")
	writeFile(t, dir, "0001-record-architecture-decisions.md")
	writeFile(t, dir, "0002-agents-md-as-canonical.md")

	entries, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, want := range []int{1, 2, 3} {
		if entries[i].Number != want {
			t.Errorf("entries[%d].Number = %d, want %d", i, entries[i].Number, want)
		}
	}
}

func TestScan_SkipsNonMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-valid.md")
	writeFile(t, dir, "README.md")
	writeFile(t, dir, "draft-not-yet-numbered.md")
	writeFile(t, dir, "0002.md")                // missing slug
	writeFile(t, dir, "001-too-short.md")       // 3 digits
	writeFile(t, dir, "00012-too-long.md")      // 5 digits
	writeFile(t, dir, "0003-Uppercase-Bad.md")  // capital letters
	writeFile(t, dir, "0004-trailing-dash-.md") // technically allowed by pattern; allowed

	entries, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	want := []int{1, 4}
	if len(entries) != len(want) {
		t.Fatalf("got entries %v, want numbers %v", entries, want)
	}
	for i, n := range want {
		if entries[i].Number != n {
			t.Errorf("entries[%d].Number = %d, want %d", i, entries[i].Number, n)
		}
	}
}

func TestNext_StartsAtOneWhenEmpty(t *testing.T) {
	n, err := Next(t.TempDir())
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if n != 1 {
		t.Fatalf("Next on empty dir = %d, want 1", n)
	}
}

func TestNext_AdvancesPastMaximum(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 8; i++ {
		writeFile(t, dir, Filename(i, "slug"))
	}
	n, err := Next(dir)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if n != 9 {
		t.Fatalf("Next = %d, want 9", n)
	}
}

func TestNext_AdvancesPastGaps(t *testing.T) {
	// Gap between 3 and 7 must not return 4 — we never recycle numbers.
	dir := t.TempDir()
	writeFile(t, dir, "0001-a.md")
	writeFile(t, dir, "0002-b.md")
	writeFile(t, dir, "0003-c.md")
	writeFile(t, dir, "0007-g.md")

	n, err := Next(dir)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if n != 8 {
		t.Fatalf("Next with gap = %d, want 8", n)
	}
}

func TestNext_AdvancesPastDuplicateNumbers(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-first.md")
	writeFile(t, dir, "0001-also-first.md")
	writeFile(t, dir, "0002-second.md")

	n, err := Next(dir)
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if n != 3 {
		t.Fatalf("Next with duplicates = %d, want 3", n)
	}
}

func TestFilename(t *testing.T) {
	got := Filename(7, "no-generic-design-md")
	want := "0007-no-generic-design-md.md"
	if got != want {
		t.Fatalf("Filename = %q, want %q", got, want)
	}
}

func TestPath(t *testing.T) {
	got := Path("docs/adr", 12, "new-thing")
	want := filepath.Join("docs/adr", "0012-new-thing.md")
	if got != want {
		t.Fatalf("Path = %q, want %q", got, want)
	}
}
