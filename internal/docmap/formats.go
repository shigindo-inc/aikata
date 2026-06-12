package docmap

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
)

// Optional rendering leaf names under the aikata machine zone. yaml and
// md are the defaults (constants live in yaml.go / render.go); these are
// the config-gated extras (docmap-design.md §2).
const (
	JSONFilename    = "docmap.json"
	TextFilename    = "docmap.txt"
	MermaidFilename = "docmap.mmd"
)

// defaultFormats is the rendering set written when `docmap.formats` is
// unset: the data layer plus the readable Markdown view.
var defaultFormats = []string{"yaml", "md"}

// renderer pairs a format name with the leaf file it writes and the bytes
// it produces from the Map.
type renderer struct {
	name string
	file string
	body func(Map) ([]byte, error)
}

// renderers is the registry of every supported format. Generate consults
// it so adding a format is a single entry. yaml is always emitted (the
// data layer / freshness source), independent of the configured set.
var renderers = []renderer{
	{"yaml", YAMLFilename, MarshalYAML},
	{"md", MarkdownFilename, func(m Map) ([]byte, error) { return RenderMarkdown(m), nil }},
	{"json", JSONFilename, MarshalJSON},
	{"txt", TextFilename, func(m Map) ([]byte, error) { return RenderText(m), nil }},
	{"mmd", MermaidFilename, func(m Map) ([]byte, error) { return RenderMermaid(m), nil }},
}

// resolveFormats returns the ordered, de-duplicated set of formats to
// emit: the configured set (or the default) with yaml forced on, since
// the data layer must always exist for the doctor freshness check.
func resolveFormats(configured []string) []string {
	want := configured
	if len(want) == 0 {
		want = defaultFormats
	}
	out := []string{"yaml"}
	seen := map[string]struct{}{"yaml": {}}
	for _, f := range want {
		f = strings.ToLower(strings.TrimSpace(f))
		if _, dup := seen[f]; dup || f == "" {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

// MarshalJSON renders the Map as pretty-printed JSON. Field order follows
// the struct definition and Docs is already sorted, so output is
// deterministic.
func MarshalJSON(m Map) ([]byte, error) {
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("docmap: marshal docmap.json: %w", err)
	}
	return append(out, '\n'), nil
}

// RenderMermaid renders just the relationship graph as a standalone
// Mermaid document (no Markdown fences, no node-count degrade — a `.mmd`
// file is an explicit opt-in to the raw diagram). Returns a short comment
// when there are no inter-document links.
func RenderMermaid(m Map) []byte {
	g := mermaidGraph(m.Docs)
	if g == "" {
		return []byte("%% no inter-document links\n")
	}
	return []byte(g)
}

// RenderText renders a plain-text view: the document tree followed by a
// `path — summary` index, with no Markdown syntax. Useful for terminals
// and tools that do not render Markdown.
func RenderText(m Map) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "Doc map (generated %s)\n\n", m.Generated)
	root := buildTree(m.Docs)
	writeTree(&b, root, "")
	b.WriteString("\n")
	for _, d := range m.Docs {
		s := d.Summary
		if s == "" {
			s = path.Base(d.Path)
		}
		fmt.Fprintf(&b, "%s — %s\n", d.Path, s)
	}
	return []byte(b.String())
}

// pathFor returns the machine-zone path of the given leaf under root.
func pathFor(root, leaf string) string {
	return filepath.Join(root, config.PrimaryDir, leaf)
}
