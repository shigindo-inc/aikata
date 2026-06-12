package docmeta

import "testing"

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
