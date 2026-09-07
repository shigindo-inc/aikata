package docmeta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	body := []byte("---\nproject: aikata\nstatus: draft\n---\n# Title\n\nBody.\n")
	lines, offset := ParseFrontmatter(body)
	if len(lines) != 2 {
		t.Fatalf("expected 2 frontmatter lines, got %d: %q", len(lines), lines)
	}
	if got := string(body[offset:]); got != "# Title\n\nBody.\n" {
		t.Fatalf("body after offset = %q", got)
	}
	if v, _, ok := FrontmatterValue(lines, "status"); !ok || v != "draft" {
		t.Fatalf("status = %q ok=%v", v, ok)
	}
	if _, _, ok := FrontmatterValue(lines, "missing"); ok {
		t.Fatalf("missing key reported present")
	}
}

func TestParseFrontmatter_None(t *testing.T) {
	lines, offset := ParseFrontmatter([]byte("# No frontmatter\n"))
	if lines != nil || offset != 0 {
		t.Fatalf("expected (nil,0), got (%v,%d)", lines, offset)
	}
}

func TestExtractLinks(t *testing.T) {
	body := []byte("See [a](./docs/a.md) and [abs](https://x.test) and [mail](mailto:x@y).\n" +
		"[frag](b.md#sec) [empty]() [q](c.md?x=1)\n")
	links := ExtractLinks(body)
	got := map[string]int{}
	for _, l := range links {
		got[l.Target] = l.Line
	}
	want := map[string]int{"docs/a.md": 1, "b.md": 2, "c.md": 2}
	if len(got) != len(want) {
		t.Fatalf("links = %#v", links)
	}
	for tgt, line := range want {
		if got[tgt] != line {
			t.Fatalf("target %q line = %d, want %d (%#v)", tgt, got[tgt], line, links)
		}
	}
}

func TestSkipDir_NameDenylist(t *testing.T) {
	if !SkipDir(".claude", "/tmp/proj/.claude", "/tmp/proj") {
		t.Fatal(".claude must be skipped (parity with .cursor)")
	}
	if !SkipDir(".cursor", "/tmp/proj/.cursor", "/tmp/proj") {
		t.Fatal(".cursor must stay skipped")
	}
}

func TestSkipDir_NestedVCSRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if SkipDir(filepath.Base(root), root, root) {
		t.Fatal("scan root must not be skipped for containing .git")
	}

	wt := filepath.Join(root, "linked-worktree")
	if err := os.Mkdir(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: /somewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !SkipDir("linked-worktree", wt, root) {
		t.Fatal("nested git worktree (.git file) must be skipped")
	}

	plain := filepath.Join(root, "docs")
	if err := os.Mkdir(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if SkipDir("docs", plain, root) {
		t.Fatal("ordinary subdirectory must not be skipped")
	}
}
