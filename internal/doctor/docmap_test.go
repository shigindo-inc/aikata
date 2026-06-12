package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/docmap"
)

// seedProject creates a minimal aikata project: an `.aikata/aikata.yaml`
// marker (so config.Resolve succeeds) plus one canonical document.
func seedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeProjectFile(t, dir, ".aikata/aikata.yaml", "version: 2\n")
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
