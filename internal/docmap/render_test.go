package docmap

import (
	"fmt"
	"strings"
	"testing"
)

func TestRenderMarkdown_Sections(t *testing.T) {
	m := Map{
		Version:   Version,
		Generated: "2026-06-12",
		Docs: []Doc{
			{Path: "AGENTS.md", Title: "AGENTS", Status: "draft", Updated: "2026-06-01",
				Managed: true, Links: []string{"SPEC.md", "docs/adr/0001-x.md"}},
			{Path: "NOTES.md", Title: "Notes", Summary: "Scratch notes.", Managed: false},
			{Path: "SPEC.md", Title: "SPEC", Summary: "What and why.", Status: "draft", Managed: true},
			{Path: "docs/adr/0001-x.md", Title: "ADR 0001", Summary: "First decision.", Managed: true},
		},
	}
	out := string(RenderMarkdown(m))

	for _, want := range []string{
		"# Doc map",
		"_generated: 2026-06-12_",
		"## Documents",
		"## Relationships",
		"## Index",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n---\n%s", want, out)
		}
	}

	// Tree: a directory node and an external marker.
	if !strings.Contains(out, "docs/") || !strings.Contains(out, "adr/") {
		t.Errorf("tree missing directory nodes:\n%s", out)
	}
	if !strings.Contains(out, "(external)") {
		t.Errorf("external marker missing for NOTES.md:\n%s", out)
	}

	// Graph: Mermaid with the AGENTS→SPEC edge (NOTES isolated, not a node).
	if !strings.Contains(out, "```mermaid") {
		t.Errorf("expected mermaid graph:\n%s", out)
	}
	if strings.Contains(out, `"NOTES.md"`) {
		t.Errorf("isolated NOTES.md should not be a graph node:\n%s", out)
	}

	// Index: every doc gets a path → summary line; SPEC uses its summary.
	if !strings.Contains(out, "- `SPEC.md` — What and why.") {
		t.Errorf("index line for SPEC missing:\n%s", out)
	}
}

func TestRenderMarkdown_Deterministic(t *testing.T) {
	m := Map{Generated: "2026-06-12", Docs: []Doc{
		{Path: "A.md", Title: "A", Managed: true, Links: []string{"B.md"}},
		{Path: "B.md", Title: "B", Managed: true, Links: []string{"A.md"}},
	}}
	first := string(RenderMarkdown(m))
	second := string(RenderMarkdown(m))
	if first != second {
		t.Fatal("RenderMarkdown is not deterministic")
	}
}

func TestRenderMarkdown_NoLinks(t *testing.T) {
	m := Map{Generated: "2026-06-12", Docs: []Doc{
		{Path: "A.md", Title: "A", Managed: true},
	}}
	out := string(RenderMarkdown(m))
	if !strings.Contains(out, "_No inter-document links._") {
		t.Errorf("expected no-links note:\n%s", out)
	}
	if strings.Contains(out, "```mermaid") {
		t.Errorf("should not emit empty mermaid graph:\n%s", out)
	}
}

func TestRenderMarkdown_DegradeToAdjacencyList(t *testing.T) {
	// Build a star: hub links to N spokes so participating nodes exceed
	// the threshold and the graph degrades to an adjacency list.
	n := degradeThreshold + 5
	docs := make([]Doc, 0, n+1)
	links := make([]string, 0, n)
	for i := 0; i < n; i++ {
		p := fmt.Sprintf("spoke-%02d.md", i)
		docs = append(docs, Doc{Path: p, Title: p, Managed: true})
		links = append(links, p)
	}
	docs = append(docs, Doc{Path: "hub.md", Title: "hub", Managed: true, Links: links})

	out := string(RenderMarkdown(Map{Generated: "2026-06-12", Docs: docs}))
	if strings.Contains(out, "```mermaid") {
		t.Errorf("expected adjacency-list degrade, got mermaid:\n%s", out[:400])
	}
	if !strings.Contains(out, "- `hub.md` → `spoke-00.md`") {
		t.Errorf("adjacency list missing hub line:\n%s", out[:400])
	}
}
