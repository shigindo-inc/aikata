package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestManagedIncludeGlobs_StaticSetMatchesCanonicalAndManagedDirs checks
// that, with no manifest present, the managed include set matches the
// canonical top-level docs and the known docs/ subtrees but not
// arbitrary third-party Markdown (ADR 0037 D1).
func TestManagedIncludeGlobs_StaticSetMatchesCanonicalAndManagedDirs(t *testing.T) {
	globs := ManagedIncludeGlobs(t.TempDir()) // no .aikata/manifest.yaml

	managed := []string{
		"AGENTS.md",
		"SPEC.md",
		"README.md",
		"docs/adr/0001-record-architecture-decisions.md",
		"docs/memory/user.md",
		"docs/stacks/flutter.md",
		"docs/tasks/current.md",
		"apps/web/AGENTS.md",
	}
	for _, p := range managed {
		if !MatchAny(globs, p) {
			t.Errorf("managed path %q should be in scope", p)
		}
	}

	unmanaged := []string{
		"skills/foo/SKILL.md",
		"CONTRIBUTING.md",
		"SECURITY.md",
		"vendor/docs/readme.md",
		"docs/decisions/open-questions.md",
	}
	for _, p := range unmanaged {
		if MatchAny(globs, p) {
			t.Errorf("third-party path %q should be out of default scope", p)
		}
	}
}

// TestManagedIncludeGlobs_UnionsManifestEntries verifies that paths
// recorded in .aikata/manifest.yaml join the managed set, so adopted
// projects keep validating files outside the static name list.
func TestManagedIncludeGlobs_UnionsManifestEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".aikata"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	manifest := "version: 1\npreset: standard\nlang: en\nfiles:\n" +
		"    - path: docs/custom/handbook.md\n      sha256: abc\n"
	if err := os.WriteFile(filepath.Join(root, ".aikata", "manifest.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	globs := ManagedIncludeGlobs(root)
	if !MatchAny(globs, "docs/custom/handbook.md") {
		t.Errorf("manifest-tracked path should be in scope: %v", globs)
	}
}

func TestManagedIncludeGlobs_CoversModelingPair(t *testing.T) {
	globs := ManagedIncludeGlobs(t.TempDir())
	want := []string{"docs/usecases.md", "docs/domain.md"}
	for _, w := range want {
		found := false
		for _, g := range globs {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("managed globs missing %q; got %v", w, globs)
		}
	}
}
