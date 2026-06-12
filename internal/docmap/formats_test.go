package docmap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFormats(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{nil, []string{"yaml", "md"}},                            // default
		{[]string{"md"}, []string{"yaml", "md"}},                 // yaml forced on
		{[]string{"json", "json"}, []string{"yaml", "json"}},     // dedup
		{[]string{"MD", " mmd "}, []string{"yaml", "md", "mmd"}}, // normalised
	}
	for _, c := range cases {
		got := resolveFormats(c.in)
		if !equal(got, c.want) {
			t.Errorf("resolveFormats(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestGenerate_WritesConfiguredFormats(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "A.md", "# A\n[b](B.md)\n")
	writeFile(t, root, "B.md", "# B\n")

	if err := Generate(Options{
		TargetDir: root,
		Now:       fixedClock,
		Formats:   []string{"md", "json", "txt", "mmd"},
	}); err != nil {
		t.Fatal(err)
	}
	// yaml (always), md, json, txt, mmd all present.
	for _, leaf := range []string{"docmap.yaml", "docmap.md", "docmap.json", "docmap.txt", "docmap.mmd"} {
		if _, err := os.Stat(filepath.Join(root, ".aikata", leaf)); err != nil {
			t.Errorf("expected %s to be written: %v", leaf, err)
		}
	}
	// .mmd is raw mermaid (no fences) with the A→B edge.
	mmd, err := os.ReadFile(filepath.Join(root, ".aikata", "docmap.mmd"))
	if err != nil {
		t.Fatal(err)
	}
	if string(mmd[:8]) != "graph LR" {
		t.Errorf("docmap.mmd should start with raw mermaid, got: %q", string(mmd))
	}
}

func TestGenerate_DefaultOmitsOptionalFormats(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "A.md", "# A\n")
	if err := Generate(Options{TargetDir: root, Now: fixedClock}); err != nil {
		t.Fatal(err)
	}
	for _, leaf := range []string{"docmap.yaml", "docmap.md"} {
		if _, err := os.Stat(filepath.Join(root, ".aikata", leaf)); err != nil {
			t.Errorf("default should write %s: %v", leaf, err)
		}
	}
	for _, leaf := range []string{"docmap.json", "docmap.txt", "docmap.mmd"} {
		if _, err := os.Stat(filepath.Join(root, ".aikata", leaf)); err == nil {
			t.Errorf("default should not write %s", leaf)
		}
	}
}

func TestOptionsFor_ReadsConfig(t *testing.T) {
	root := t.TempDir()
	cfg := "version: 2\n" +
		"project:\n  name: x\n  lang: en\n" +
		"components: {}\n" +
		"docmap:\n" +
		"  formats: [yaml, md, json]\n" +
		"  targets: [\"docs/**/*.md\"]\n" +
		"  exclude: [\"docs/scratch/**\"]\n"
	writeFile(t, root, ".aikata/aikata.yaml", cfg)

	opts := OptionsFor(root, []string{"AGENTS.md"})
	if opts.TargetDir != root {
		t.Errorf("TargetDir = %q", opts.TargetDir)
	}
	if !equal(opts.Formats, []string{"yaml", "md", "json"}) {
		t.Errorf("Formats = %v", opts.Formats)
	}
	if !equal(opts.Targets, []string{"docs/**/*.md"}) {
		t.Errorf("Targets = %v", opts.Targets)
	}
	if !equal(opts.Exclude, []string{"docs/scratch/**"}) {
		t.Errorf("Exclude = %v", opts.Exclude)
	}
	if !equal(opts.ManagedGlobs, []string{"AGENTS.md"}) {
		t.Errorf("ManagedGlobs = %v", opts.ManagedGlobs)
	}
}

func TestOptionsFor_NoConfigFallsBack(t *testing.T) {
	root := t.TempDir()
	opts := OptionsFor(root, nil)
	if len(opts.Targets) != 0 || len(opts.Formats) != 0 {
		t.Errorf("expected empty (default) options without config, got %#v", opts)
	}
}
