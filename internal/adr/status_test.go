package adr

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseStatus(t *testing.T) {
	cases := []struct {
		name, body string
		ok         bool
		class      Class
		line       int
		links      []int
	}{
		{name: "english list item", body: "# T\n\n- **Status**: Accepted\n", ok: true, class: ClassEffective, line: 3},
		{name: "bold value", body: "- **Status**: **Accepted** — note\n", ok: true, class: ClassEffective, line: 1},
		{name: "japanese key, full-width colon", body: "> **状態**：active\n", ok: true, class: ClassEffective, line: 1},
		{
			name: "superseded with a link", ok: true, class: ClassSuperseded, line: 1, links: []int{80},
			body: "> **状態**: superseded by [ADR 0080](./0080-trip-places.md)（2026-10-17）\n",
		},
		{
			name: "partial supersession stays effective and keeps links", ok: true, class: ClassEffective, line: 1, links: []int{115, 103},
			body: "> **状態**: active（一部失効。[ADR 0115](./0115-read-only.md) で失効。決定 4 は ADR 0103 が置き換えた）\n",
		},
		{name: "first word is not a decision word", body: "> **状態**: 一部 active。\n", ok: true, class: ClassUnknown, line: 1},
		{name: "rejected", body: "- **Status**: Rejected\n", ok: true, class: ClassRetired, line: 1},
		{name: "frontmatter is not read", body: "---\nstatus: accepted\n---\n\n# T\n", ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st, ok := ParseStatus([]byte(tc.body))
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if !ok {
				return
			}
			if st.Class != tc.class || st.Line != tc.line || !reflect.DeepEqual(st.Links, tc.links) {
				t.Errorf("got class=%s line=%d links=%v, want class=%s line=%d links=%v",
					st.Class, st.Line, st.Links, tc.class, tc.line, tc.links)
			}
		})
	}
}

func TestDecisionWord(t *testing.T) {
	for _, v := range []string{"proposed", "Accepted", "active", "superseded", "deprecated", "rejected"} {
		if _, ok := DecisionWord(v); !ok {
			t.Errorf("DecisionWord(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"draft", "stable", "", "superseded by ADR 0001"} {
		if _, ok := DecisionWord(v); ok {
			t.Errorf("DecisionWord(%q) = true, want false", v)
		}
	}
}

func TestInventory(t *testing.T) {
	root := t.TempDir()
	write(t, root, "docs/adr/0001-first.md", "# ADR 0001 — First\n\n- **Status**: Accepted\n")
	write(t, root, "docs/adr/0002-second.md", "# ADR 0002 — Second\n\n- **Status**: Superseded by [ADR 0003](./0003-third.md)\n")
	write(t, root, "docs/adr/0003-third.md", "# ADR 0003 — Third\n\nReplaces ADR 0002. Mentions ADR 0003 itself.\n")
	write(t, root, "AGENTS.md", "See [ADR 1](docs/adr/0001-first.md) and ADR 0001 again.\n")
	write(t, root, "src/notes.txt", "ADR 0001 in a non-Markdown file is not counted.\n")

	got, err := Inventory(root)
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	type row struct {
		title     string
		hasStatus bool
		class     Class
		links     []int
		refs      int
	}
	want := []row{
		{"ADR 0001 — First", true, ClassEffective, nil, 1},
		{"ADR 0002 — Second", true, ClassSuperseded, []int{3}, 1},
		{"ADR 0003 — Third", false, "", nil, 1},
	}
	for i, w := range want {
		r := got[i]
		g := row{r.Title, r.HasStatus, r.Status.Class, r.Status.Links, r.DocRefs}
		if !reflect.DeepEqual(g, w) {
			t.Errorf("record %d = %+v, want %+v", i, g, w)
		}
	}
}
