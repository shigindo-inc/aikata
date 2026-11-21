package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/docmap"
)

// seedProject creates a minimal aikata project with the doc map enabled:
// an `.aikata/aikata.yaml` (so config.Resolve succeeds) plus one canonical
// document.
func seedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeProjectFile(t, dir, ".aikata/aikata.yaml", "version: 2\ndocmap:\n  enabled: true\n")
	writeProjectFile(t, dir, "AGENTS.md",
		"---\nproject: x\nstatus: draft\nversion: 0.0.1\nupdated: 2026-06-01\naudience: [agent]\n---\n# AGENTS\n\n> Rules.\n")
	return dir
}

func writeProjectFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// genMap writes the doc map the same way the freshness check rebuilds it.
func genMap(t *testing.T, dir string) {
	t.Helper()
	if err := docmap.Generate(docmap.Options{
		TargetDir:    dir,
		ManagedGlobs: ManagedIncludeGlobs(dir),
	}); err != nil {
		t.Fatalf("generate map: %v", err)
	}
}

func TestCheckDocMap_FreshAfterGenerate(t *testing.T) {
	dir := seedProject(t)
	genMap(t, dir)
	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues for a fresh map, got %#v", issues)
	}
}

func TestCheckDocMap_MissingWhenNoMap(t *testing.T) {
	dir := seedProject(t)
	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Code != "docmap.missing" {
		t.Fatalf("expected docmap.missing warning, got %#v", issues)
	}
	if issues[0].Level != LevelWarning {
		t.Errorf("missing map should be a warning, not %v", issues[0].Level)
	}
}

func TestCheckDocMap_StaleAfterDocChange(t *testing.T) {
	dir := seedProject(t)
	genMap(t, dir)
	// Add a new document after the map was built → structural drift.
	writeProjectFile(t, dir, "SPEC.md",
		"---\nproject: x\nstatus: draft\nversion: 0.0.1\nupdated: 2026-06-01\naudience: [agent]\n---\n# SPEC\n\nWhat.\n")

	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Code != "docmap.stale" {
		t.Fatalf("expected docmap.stale warning, got %#v", issues)
	}
}

func TestCheckDocMap_DisabledWithoutFilesIsSilent(t *testing.T) {
	dir := seedProject(t)
	writeProjectFile(t, dir, ".aikata/aikata.yaml", "version: 2\n")
	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("a disabled map with no files should not be reported, got %#v", issues)
	}
}

func TestCheckDocMap_DisabledWithFilesIsUnmaintained(t *testing.T) {
	dir := seedProject(t)
	genMap(t, dir)
	writeProjectFile(t, dir, ".aikata/aikata.yaml", "version: 2\n")
	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Code != "docmap.unmaintained" || issues[0].Level != LevelWarning {
		t.Fatalf("expected one docmap.unmaintained warning, got %#v", issues)
	}
	res, err := Fix(Options{TargetDir: dir}, issues)
	if err != nil {
		t.Fatal(err)
	}
	if res.Fixed != 0 {
		t.Errorf("--fix must not delete or rebuild an unmaintained map, fixed %d", res.Fixed)
	}
}

func TestCheckDocMap_NoOpOutsideAikataProject(t *testing.T) {
	dir := t.TempDir() // no .aikata/aikata.yaml
	writeProjectFile(t, dir, "README.md", "# R\n")
	issues, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("non-aikata project should not be nagged, got %#v", issues)
	}
}

func TestFix_RegeneratesStaleDocMap(t *testing.T) {
	dir := seedProject(t)
	genMap(t, dir)
	writeProjectFile(t, dir, "SPEC.md",
		"---\nproject: x\nstatus: draft\nversion: 0.0.1\nupdated: 2026-06-01\naudience: [agent]\n---\n# SPEC\n\nWhat.\n")

	stale, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) != 1 {
		t.Fatalf("setup: expected one stale issue, got %#v", stale)
	}

	res, err := Fix(Options{TargetDir: dir}, stale)
	if err != nil {
		t.Fatalf("fix: %v", err)
	}
	if res.Fixed != 1 {
		t.Errorf("expected 1 fix, got %d (skipped %d)", res.Fixed, res.Skipped)
	}

	after, err := checkDocMap(Options{TargetDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 0 {
		t.Fatalf("map still stale after --fix: %#v", after)
	}
}

func TestCheckDocMap_RepairsConfiguredView(t *testing.T) {
	dir := seedProject(t)
	genMap(t, dir)
	path := docmap.MarkdownPath(dir)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	opts := Options{TargetDir: dir}
	issues, err := checkDocMap(opts)
	if err != nil || len(issues) != 1 || issues[0].Code != "docmap.missing" || issues[0].File != ".aikata/docmap.md" {
		t.Fatalf("missing Markdown view: issues=%#v, err=%v", issues, err)
	}
	if _, err := Fix(opts, issues); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, dir, ".aikata/docmap.md", "# Stale view")
	issues, err = checkDocMap(opts)
	if err != nil || len(issues) != 1 || issues[0].Code != "docmap.stale" || issues[0].File != ".aikata/docmap.md" {
		t.Fatalf("stale Markdown view: issues=%#v, err=%v", issues, err)
	}
}
