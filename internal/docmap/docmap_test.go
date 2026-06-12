package docmap

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeFile creates dir tree and writes a file under root.
func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixedClock is an arbitrary deterministic date for golden stability.
var fixedClock = time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)

func buildTestMap(t *testing.T, root string) Map {
	t.Helper()
	m, err := Build(Options{
		TargetDir:    root,
		Now:          fixedClock,
		ManagedGlobs: []string{"AGENTS.md", "SPEC.md", "docs/adr/**"},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return m
}

func docByPath(m Map, p string) (Doc, bool) {
	for _, d := range m.Docs {
		if d.Path == p {
			return d, true
		}
	}
	return Doc{}, false
}

func TestBuild_CoreFields(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "AGENTS.md",
		"---\nproject: aikata\nstatus: draft\nupdated: 2026-06-01\n---\n"+
			"# AGENTS — Operating rules\n\n> The canonical operating rules.\n\n"+
			"See [spec](./SPEC.md) and [adr](docs/adr/0001-x.md) and [ext](https://x.test).\n")
	writeFile(t, root, "SPEC.md",
		"---\nproject: aikata\nstatus: draft\n---\n# SPEC\n\nWhat and why of aikata.\n")
	writeFile(t, root, "docs/adr/0001-x.md",
		"# ADR 0001\n\nFirst decision.\n") // no frontmatter
	writeFile(t, root, "NOTES.md",
		"# Notes\n\nScratch.\n[dangling](./does-not-exist.md)\n") // external, dangling link

	m := buildTestMap(t, root)

	if m.Version != Version {
		t.Errorf("version = %d", m.Version)
	}
	if m.Generated != "2026-06-12" {
		t.Errorf("generated = %q", m.Generated)
	}
	if len(m.Docs) != 4 {
		t.Fatalf("expected 4 docs, got %d: %#v", len(m.Docs), m.Docs)
	}

	agents, _ := docByPath(m, "AGENTS.md")
	if agents.Title != "AGENTS — Operating rules" {
		t.Errorf("AGENTS title = %q", agents.Title)
	}
	if agents.Summary != "The canonical operating rules." {
		t.Errorf("AGENTS summary = %q", agents.Summary)
	}
	if agents.Updated != "2026-06-01" {
		t.Errorf("AGENTS updated = %q", agents.Updated)
	}
	if !agents.Managed {
		t.Errorf("AGENTS should be managed")
	}
	// Links resolve into the tracked set; the https link is dropped.
	if want := []string{"SPEC.md", "docs/adr/0001-x.md"}; !equal(agents.Links, want) {
		t.Errorf("AGENTS links = %#v, want %#v", agents.Links, want)
	}

	// SPEC: summary falls through to first paragraph after H1 (no blockquote).
	spec, _ := docByPath(m, "SPEC.md")
	if spec.Summary != "What and why of aikata." {
		t.Errorf("SPEC summary = %q", spec.Summary)
	}

	// ADR has no frontmatter: title/summary from headings, updated from mtime.
	adr, _ := docByPath(m, "docs/adr/0001-x.md")
	if adr.Title != "ADR 0001" {
		t.Errorf("ADR title = %q", adr.Title)
	}
	if adr.Summary != "First decision." {
		t.Errorf("ADR summary = %q", adr.Summary)
	}
	if adr.Status != "" {
		t.Errorf("ADR status = %q, want empty", adr.Status)
	}
	if !adr.Managed {
		t.Errorf("ADR under docs/adr/** should be managed")
	}

	// NOTES is external and its only link dangles → no edges.
	notes, _ := docByPath(m, "NOTES.md")
	if notes.Managed {
		t.Errorf("NOTES should be external (unmanaged)")
	}
	if len(notes.Links) != 0 {
		t.Errorf("NOTES links = %#v, want none (dangling dropped)", notes.Links)
	}
}

func TestBuild_Deterministic(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "A.md", "# A\n[b](B.md)\n")
	writeFile(t, root, "B.md", "# B\n[a](A.md)\n")

	m1 := buildTestMap(t, root)
	m2 := buildTestMap(t, root)
	b1, err := MarshalYAML(m1)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := MarshalYAML(m2)
	if err != nil {
		t.Fatal(err)
	}
	if string(b1) != string(b2) {
		t.Fatalf("non-deterministic output:\n%s\n---\n%s", b1, b2)
	}
	// Docs sorted by path.
	if m1.Docs[0].Path != "A.md" || m1.Docs[1].Path != "B.md" {
		t.Fatalf("docs not sorted: %#v", m1.Docs)
	}
}

func TestBuild_ExcludesMachineZoneAndArtifacts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "README.md", "# R\n")
	writeFile(t, root, "CLAUDE.md", "generated\n")              // AI-tool artifact
	writeFile(t, root, ".aikata/docmap.md", "# self\n")         // machine zone
	writeFile(t, root, "node_modules/pkg/README.md", "# dep\n") // vendored

	m := buildTestMap(t, root)
	if len(m.Docs) != 1 || m.Docs[0].Path != "README.md" {
		t.Fatalf("expected only README.md, got %#v", m.Docs)
	}
}

func TestBuild_TargetsAndExclude(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "keep.md", "# keep\n")
	writeFile(t, root, "drop.md", "# drop\n")
	m, err := Build(Options{
		TargetDir: root,
		Now:       fixedClock,
		Exclude:   []string{"drop.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Docs) != 1 || m.Docs[0].Path != "keep.md" {
		t.Fatalf("exclude not applied: %#v", m.Docs)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
