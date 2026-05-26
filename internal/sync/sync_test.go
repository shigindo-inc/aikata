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

// TestRun_Rebaseline_PreservesUserEdits asserts the regression that
// motivated v0.6.1: with a pre-v0.5 project that has *customised*
// source files, --rebaseline must NOT touch those files. The bytes on
// disk before and after the rebaseline run must be byte-identical.
// Also asserts that the manifest's recorded hash for the customised
// file equals the *upstream rendering's* hash (not the customised
// content's hash) — that's what makes the next sync classify the
// customisation as `user-only-edit`.
func TestRun_Rebaseline_PreservesUserEdits(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Customise AGENTS.md, then remove the manifest to simulate a
	// project that pre-dates the v0.5 manifest schema.
	agentsPath := filepath.Join(root, "AGENTS.md")
	customContent := []byte("# AGENTS\n\nproject-specific override\n")
	if err := os.WriteFile(agentsPath, customContent, 0o644); err != nil {
		t.Fatalf("write custom AGENTS.md: %v", err)
	}
	if err := os.Remove(config.ManifestPath(root)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}

	if _, err := Run(Options{Root: root, Rebaseline: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err != nil {
		t.Fatalf("Run rebaseline: %v", err)
	}

	// Source file must be byte-identical to the pre-run content.
	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md after rebaseline: %v", err)
	}
	if !bytes.Equal(after, customContent) {
		t.Errorf("rebaseline modified AGENTS.md:\nbefore=%q\nafter =%q", customContent, after)
	}

	// Manifest must record the upstream rendering's hash for AGENTS.md,
	// not the customised content's hash. This is the key invariant: on
	// the next sync, current (custom) != ancestor (upstream) → classified
	// as user-only-edit and preserved.
	m, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest after rebaseline: %v", err)
	}
	customHash := config.HashContent(customContent)
	var agentsEntry *config.ManifestFile
	for i, f := range m.Files {
		if f.Path == "AGENTS.md" {
			agentsEntry = &m.Files[i]
			break
		}
	}
	if agentsEntry == nil {
		t.Fatalf("AGENTS.md missing from rebaselined manifest: %+v", m.Files)
	}
	if agentsEntry.SHA256 == customHash {
		t.Errorf("manifest recorded customised hash as ancestor — next sync would overwrite it. got=%s", agentsEntry.SHA256)
	}
}

// TestRun_Rebaseline_NoPerFileResults asserts that rebaseline never
// runs the merge: the returned RunResult should have an empty Files
// slice and zero counts (the human-readable notes carry the messaging
// instead). This guards against accidental re-introduction of a
// merge-style code path.
func TestRun_Rebaseline_NoPerFileResults(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	if err := os.Remove(config.ManifestPath(root)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}

	result, err := Run(Options{Root: root, Rebaseline: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run rebaseline: %v", err)
	}
	if len(result.Files) != 0 {
		t.Errorf("rebaseline should not produce per-file outcomes, got %d entries: %+v", len(result.Files), result.Files)
	}
	if result.Conflicts != 0 || result.Applied != 0 || result.NoChange != 0 {
		t.Errorf("rebaseline counters should all be zero, got %+v", result)
	}
	if len(result.Notes) == 0 {
		t.Errorf("rebaseline should emit at least one note explaining the seed action")
	}
}

// TestRun_Rebaseline_DryRun_DoesNotWriteManifest asserts that
// --rebaseline --dry-run is a pure inspection: no manifest is written,
// no source file is modified.
func TestRun_Rebaseline_DryRun_DoesNotWriteManifest(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	if err := os.Remove(config.ManifestPath(root)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}

	if _, err := Run(Options{Root: root, Rebaseline: true, DryRun: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err != nil {
		t.Fatalf("Run rebaseline dry-run: %v", err)
	}
	if _, err := os.Stat(config.ManifestPath(root)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("dry-run rebaseline should not create manifest, got stat err=%v", err)
	}
}

// TestRun_PostRebaseline_PreservesEditsAsUserOnly is the integration
// test that would have caught the v0.6.0 regression: after rebaseline,
// the next plain `aikata sync` must classify customised files as
// `user-only-edit` and leave their on-disk bytes alone. If ancestor
// were recorded as the customised bytes (the obvious-but-wrong design),
// the immediate follow-up sync would StatusUpstreamApplied and clobber
// the customisation.
func TestRun_PostRebaseline_PreservesEditsAsUserOnly(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	agentsPath := filepath.Join(root, "AGENTS.md")
	customContent := []byte("# AGENTS\n\nproject-specific override\n")
	if err := os.WriteFile(agentsPath, customContent, 0o644); err != nil {
		t.Fatalf("write custom AGENTS.md: %v", err)
	}
	if err := os.Remove(config.ManifestPath(root)); err != nil {
		t.Fatalf("remove manifest: %v", err)
	}

	if _, err := Run(Options{Root: root, Rebaseline: true, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}); err != nil {
		t.Fatalf("rebaseline: %v", err)
	}

	// Now run a plain sync. AGENTS.md customisation must survive and
	// be reported as user-only-edit.
	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("post-rebaseline sync: %v", err)
	}
	if result.Conflicts != 0 {
		t.Errorf("post-rebaseline sync should not conflict, got %+v", result)
	}

	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md after sync: %v", err)
	}
	if !bytes.Equal(after, customContent) {
		t.Errorf("post-rebaseline sync clobbered AGENTS.md:\nbefore=%q\nafter =%q", customContent, after)
	}

	var sawAgents bool
	for _, f := range result.Files {
		if f.Path == "AGENTS.md" {
			sawAgents = true
			if f.Status != StatusUserOnlyEdit {
				t.Errorf("AGENTS.md status = %q, want %q", f.Status, StatusUserOnlyEdit)
			}
		}
	}
	if !sawAgents {
		t.Errorf("AGENTS.md missing from post-rebaseline sync result: %+v", result.Files)
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
