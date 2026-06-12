package docmap

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
)

// MarkdownFilename is the leaf name of the doc map's readable view under
// the aikata machine zone (`.aikata/docmap.md`).
const MarkdownFilename = "docmap.md"

// degradeThreshold is the node count above which the relationship graph
// falls back from a Mermaid diagram to a flat adjacency list, which stays
// readable when a diagram would not (Q-DOCMAP-03; provisional, made
// configurable in P5). "Nodes" counts documents that participate in at
// least one edge.
const degradeThreshold = 40

// MarkdownPath returns the path of the doc map readable view under root.
func MarkdownPath(root string) string {
	return filepath.Join(root, config.PrimaryDir, MarkdownFilename)
}

// RenderMarkdown produces the `.aikata/docmap.md` readable view entirely
// from the Map (no filesystem reads), so the data layer remains the
// single source of truth (ADR 0044 D2). Output is deterministic: every
// section derives from the already-sorted Docs slice and re-sorts any
// derived collection, so identical inputs yield byte-identical output.
func RenderMarkdown(m Map) []byte {
	var b strings.Builder
	b.WriteString("# Doc map\n\n")
	b.WriteString("> Machine-generated inventory of this project's document set " +
		"(ADR 0044). Derived from `.aikata/docmap.yaml`; do not edit by hand — " +
		"`aikata map` and the doctor freshness check keep it current.\n\n")
	fmt.Fprintf(&b, "_generated: %s_\n\n", m.Generated)

	renderTree(&b, m.Docs)
	renderGraph(&b, m.Docs)
	renderIndex(&b, m.Docs)

	return []byte(b.String())
}

// renderTree writes the tracked documents as a directory tree, each leaf
// annotated with `title · status · updated` and an `(external)` marker
// for unmanaged documents.
func renderTree(b *strings.Builder, docs []Doc) {
	b.WriteString("## Documents\n\n```\n")
	root := buildTree(docs)
	writeTree(b, root, "")
	b.WriteString("```\n\n")
}

// treeNode is a directory or file node in the document tree.
type treeNode struct {
	name     string
	doc      *Doc // non-nil for file leaves
	children map[string]*treeNode
}

func newTreeNode(name string) *treeNode {
	return &treeNode{name: name, children: map[string]*treeNode{}}
}

// buildTree groups the sorted docs into a nested directory structure.
func buildTree(docs []Doc) *treeNode {
	root := newTreeNode("")
	for i := range docs {
		segs := strings.Split(docs[i].Path, "/")
		cur := root
		for j, seg := range segs {
			child, ok := cur.children[seg]
			if !ok {
				child = newTreeNode(seg)
				cur.children[seg] = child
			}
			if j == len(segs)-1 {
				child.doc = &docs[i]
			}
			cur = child
		}
	}
	return root
}

// writeTree renders node's children with box-drawing connectors.
// Directories are listed before files at each level, each group sorted by
// name, for a stable, readable order.
func writeTree(b *strings.Builder, node *treeNode, prefix string) {
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		ci, cj := node.children[names[i]], node.children[names[j]]
		di, dj := len(ci.children) > 0, len(cj.children) > 0
		if di != dj {
			return di // directories first
		}
		return names[i] < names[j]
	})
	for i, name := range names {
		child := node.children[name]
		last := i == len(names)-1
		connector, childPrefix := "├── ", prefix+"│   "
		if last {
			connector, childPrefix = "└── ", prefix+"    "
		}
		if len(child.children) > 0 {
			fmt.Fprintf(b, "%s%s%s/\n", prefix, connector, name)
			writeTree(b, child, childPrefix)
			continue
		}
		fmt.Fprintf(b, "%s%s%s%s\n", prefix, connector, name, leafAnnotation(child.doc))
	}
}

// leafAnnotation renders the trailing ` — title · status · updated` (and
// `(external)`) describing a file leaf.
func leafAnnotation(d *Doc) string {
	if d == nil {
		return ""
	}
	meta := []string{}
	if d.Title != "" {
		meta = append(meta, d.Title)
	}
	if d.Status != "" {
		meta = append(meta, d.Status)
	}
	if d.Updated != "" {
		meta = append(meta, d.Updated)
	}
	if !d.Managed {
		meta = append(meta, "(external)")
	}
	if len(meta) == 0 {
		return ""
	}
	return " — " + strings.Join(meta, " · ")
}

// edge is a directed doc→doc link.
type edge struct{ from, to string }

// collectEdges returns every doc→doc edge (sorted) and the set of
// documents that participate in at least one edge.
func collectEdges(docs []Doc) (edges []edge, participating map[string]struct{}) {
	participating = map[string]struct{}{}
	for _, d := range docs {
		for _, to := range d.Links {
			edges = append(edges, edge{d.Path, to})
			participating[d.Path] = struct{}{}
			participating[to] = struct{}{}
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		return edges[i].to < edges[j].to
	})
	return edges, participating
}

// mermaidGraph renders the edges as a Mermaid `graph LR` body (including
// the `graph LR` header, no code fences). Returns "" when there are no
// edges. Node ids are assigned from sorted participating paths so the
// output is deterministic.
func mermaidGraph(docs []Doc) string {
	edges, participating := collectEdges(docs)
	if len(edges) == 0 {
		return ""
	}
	nodes := make([]string, 0, len(participating))
	for p := range participating {
		nodes = append(nodes, p)
	}
	sort.Strings(nodes)
	id := make(map[string]string, len(nodes))
	for i, p := range nodes {
		id[p] = fmt.Sprintf("n%d", i)
	}
	var b strings.Builder
	b.WriteString("graph LR\n")
	for _, p := range nodes {
		fmt.Fprintf(&b, "  %s[\"%s\"]\n", id[p], p)
	}
	for _, e := range edges {
		fmt.Fprintf(&b, "  %s --> %s\n", id[e.from], id[e.to])
	}
	return b.String()
}

// renderGraph writes the doc→doc relationship section: a Mermaid diagram
// when the participating-node count is within the threshold, otherwise a
// flat adjacency list.
func renderGraph(b *strings.Builder, docs []Doc) {
	b.WriteString("## Relationships\n\n")

	_, participating := collectEdges(docs)
	if len(participating) == 0 {
		b.WriteString("_No inter-document links._\n\n")
		return
	}

	if len(participating) > degradeThreshold {
		// Degrade: adjacency list, grouped by source (docs already sorted).
		for _, d := range docs {
			if len(d.Links) == 0 {
				continue
			}
			fmt.Fprintf(b, "- `%s` → %s\n", d.Path, strings.Join(backtickList(d.Links), ", "))
		}
		b.WriteString("\n")
		return
	}

	b.WriteString("```mermaid\n")
	b.WriteString(mermaidGraph(docs))
	b.WriteString("```\n\n")
}

// renderIndex writes the path → summary list, the compression payoff:
// reading the map alone conveys the gist of every document.
func renderIndex(b *strings.Builder, docs []Doc) {
	b.WriteString("## Index\n\n")
	for _, d := range docs {
		s := d.Summary
		if s == "" {
			s = path.Base(d.Path)
		}
		fmt.Fprintf(b, "- `%s` — %s\n", d.Path, s)
	}
	b.WriteString("\n")
}

// backtickList wraps each item in backticks.
func backtickList(items []string) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = "`" + it + "`"
	}
	return out
}
