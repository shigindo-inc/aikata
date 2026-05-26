package sync

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// seedStandardProject scaffolds a fresh `standard` preset project into
// root and returns its render output for assertions. The test's caller
// owns root (typically t.TempDir()).
func seedStandardProject(t *testing.T, root string) {
	t.Helper()
	opts := scaffold.Options{
		ProjectName: "samplekata",
		Preset:      "standard",
		TargetDir:   root,
		Lang:        "en",
		Stdout:      ioDiscard{},
	}
	if err := scaffold.Run(opts); err != nil {
		t.Fatalf("seed scaffold: %v", err)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func TestRun_NoManifest_WithoutRebaseline_Errors(t *testing.T) {
	root := t.TempDir()
	// Just a config; no manifest.
	cfg := config.Default("samplekata", "en")
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	_, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if !errors.Is(err, ErrNoManifest) {
		t.Errorf("Run without manifest = %v, want ErrNoManifest", err)
	}
}

func TestRun_FreshProject_AllUnchanged(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 0 || result.Applied != 0 {
		t.Errorf("fresh project should be all-unchanged: %+v", result)
	}
	if result.NoChange == 0 {
		t.Errorf("expected unchanged > 0, got result %+v", result)
	}
}

func TestRun_UserOnlyEdit_Preserved(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// User edits SPEC.md.
	specPath := filepath.Join(root, "SPEC.md")
	userContent := []byte("# SPEC.md\n\nuser-only edit\n")
	if err := os.WriteFile(specPath, userContent, 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 0 {
		t.Errorf("expected no conflicts, got %+v", result)
	}
	got, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	if string(got) != string(userContent) {
		t.Errorf("user content should be preserved, got: %q", got)
	}
	// Find the SPEC.md entry and verify its status.
	for _, f := range result.Files {
		if f.Path == "SPEC.md" && f.Status != StatusUserOnlyEdit {
			t.Errorf("SPEC.md status = %q, want %q", f.Status, StatusUserOnlyEdit)
		}
	}
}

func TestRun_UpstreamApplied_WhenManifestStale(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Mutate the manifest to record a stale hash for AGENTS.md so the
	// merge thinks upstream evolved while the user did not edit.
	manifest, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for i, f := range manifest.Files {
		if f.Path == "AGENTS.md" {
			manifest.Files[i].SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
		}
	}
	if err := config.SaveManifest(root, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	// Note: this simulates "upstream changed" because the manifest's
	// recorded ancestor hash no longer matches both current and
	// upstream (both of which are still the live template content).
	// The current file does not match the (fake) ancestor, but it
	// does match upstream — that lands in StatusBothMatch.

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 0 {
		t.Errorf("expected no conflicts, got %+v", result)
	}
	var found bool
	for _, f := range result.Files {
		if f.Path == "AGENTS.md" {
			found = true
			if f.Status != StatusBothMatch {
				t.Errorf("AGENTS.md status = %q, want %q (current == upstream against fake ancestor)", f.Status, StatusBothMatch)
			}
		}
	}
	if !found {
		t.Errorf("AGENTS.md missing from result; got files: %+v", result.Files)
	}
}

func TestRun_TrueConflict_WritesMarkers(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// User edits SPEC.md AND we mark the manifest ancestor as stale —
	// upstream's current content differs from both the user content
	// and the manifest hash, producing a true conflict.
	specPath := filepath.Join(root, "SPEC.md")
	userContent := []byte("# SPEC.md\nlocally-divergent\n")
	if err := os.WriteFile(specPath, userContent, 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}
	manifest, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for i, f := range manifest.Files {
		if f.Path == "SPEC.md" {
			manifest.Files[i].SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
		}
	}
	if err := config.SaveManifest(root, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 1 {
		t.Errorf("expected 1 conflict, got %d (%+v)", result.Conflicts, result)
	}
	got, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	for _, marker := range []string{"<<<<<<<", "|||||||", "=======", ">>>>>>>"} {
		if !bytes.Contains(got, []byte(marker)) {
			t.Errorf("conflict marker %q missing from SPEC.md content:\n%s", marker, got)
		}
	}
	if !bytes.Contains(got, userContent) {
		t.Errorf("user content not present in conflict body:\n%s", got)
	}
}

func TestRun_DryRun_NoWrites(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// User edit so the run isn't a pure no-op.
	specPath := filepath.Join(root, "SPEC.md")
	before, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	// Tamper with manifest for conflict scenario.
	manifest, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for i, f := range manifest.Files {
		if f.Path == "SPEC.md" {
			manifest.Files[i].SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
		}
	}
	if err := config.SaveManifest(root, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	if err := os.WriteFile(specPath, []byte("# user-divergent\n"), 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}

	result, err := Run(Options{Root: root, DryRun: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts == 0 {
		t.Errorf("expected at least one conflict in dry-run, got %+v", result)
	}
	// SPEC.md on disk must still match what we wrote (user content),
	// not have conflict markers.
	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC.md after: %v", err)
	}
	if string(after) != "# user-divergent\n" {
		t.Errorf("dry-run modified SPEC.md: before=%q after=%q", before, after)
	}
}

func TestRun_Rebaseline_SeedsManifest(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Remove manifest to simulate a pre-v0.5 project.
	if err := os.Remove(config.ManifestPath(root)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}

	result, err := Run(Options{Root: root, Rebaseline: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run rebaseline: %v", err)
	}
	if result.Conflicts != 0 {
		t.Errorf("rebaseline of clean tree should not conflict: %+v", result)
	}
	// Manifest must be back on disk.
	if _, err := config.LoadManifest(root); err != nil {
		t.Errorf("manifest not restored after --rebaseline: %v", err)
	}
}

func TestInferFlags_PresenceBased(t *testing.T) {
	m := config.Manifest{Files: []config.ManifestFile{
		{Path: "AGENTS.md"},
		{Path: "docs/memory/README.md"},
		{Path: "UI.md"},
	}}
	flags := inferFlags(m)
	if !flags.WithMemory || !flags.WithUI {
		t.Errorf("inference incomplete: %+v", flags)
	}
	if flags.WithAPI || flags.WithTDD || flags.WithChangelog {
		t.Errorf("false positive on opt-in flags: %+v", flags)
	}
}
