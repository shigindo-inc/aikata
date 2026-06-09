package fill

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// run is a test helper that invokes fill.Run against root with output
// discarded, failing the test on error.
func run(t *testing.T, root string) Result {
	t.Helper()
	res, err := Run(Options{Root: root, Stdout: io.Discard, Stderr: io.Discard})
	if err != nil {
		t.Fatalf("fill.Run: %v", err)
	}
	return res
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestFill_AdoptsUnmanagedRepo covers the originating use case: a repo
// with a hand-written AGENTS.md and nothing else gets the missing
// standard documents written, the hand-written file is preserved, and the
// repo becomes aikata-managed.
func TestFill_AdoptsUnmanagedRepo(t *testing.T) {
	tmp := t.TempDir()
	handwritten := "---\ntitle: handwritten\n---\n# Custom AGENTS\nmy rules\n"
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte(handwritten), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	res := run(t, tmp)

	if res.Managed {
		t.Errorf("expected Managed=false for an unmanaged repo")
	}
	if res.Preset != "standard" {
		t.Errorf("expected standard scope for an unmanaged repo, got %q", res.Preset)
	}
	// The standard-only docs must be written; the hand-written AGENTS.md
	// must be skipped (preserved).
	for _, want := range []string{"SPEC.md", "ARCHITECTURE.md", "GLOSSARY.md", "ROADMAP.md"} {
		if !contains(res.Written, want) {
			t.Errorf("expected %s in Written, got %v", want, res.Written)
		}
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected %s on disk: %v", want, err)
		}
	}
	if !contains(res.Skipped, "AGENTS.md") {
		t.Errorf("expected AGENTS.md in Skipped, got %v", res.Skipped)
	}
	if got, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md")); err != nil || string(got) != handwritten {
		t.Errorf("AGENTS.md must be left untouched; err=%v content=%q", err, string(got))
	}
	// The repo is now aikata-managed.
	for _, p := range []string{".aikata/aikata.yaml", ".aikata/manifest.yaml"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(p))); err != nil {
			t.Errorf("expected %s after adoption: %v", p, err)
		}
	}
}

// TestFill_ManifestRecordsUpstreamNotDisk pins the merge-safety
// invariant: the manifest ancestor for a hand-edited file is the UPSTREAM
// rendering, not the on-disk content. That is what makes the next
// `aikata sync` classify the user's edits as `user-only-edit` (preserve)
// rather than overwriting them.
func TestFill_ManifestRecordsUpstreamNotDisk(t *testing.T) {
	tmp := t.TempDir()
	handwritten := "# totally custom AGENTS that differs from the template\n"
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte(handwritten), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	run(t, tmp)

	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	var agentsHash string
	for _, f := range m.Files {
		if f.Path == "AGENTS.md" {
			agentsHash = f.SHA256
		}
	}
	if agentsHash == "" {
		t.Fatalf("manifest has no AGENTS.md entry: %+v", m.Files)
	}
	if agentsHash == config.HashContent([]byte(handwritten)) {
		t.Errorf("manifest recorded the on-disk (hand-written) hash; expected the upstream rendering hash so sync preserves user edits")
	}
}

// TestFill_Idempotent verifies a second run on a complete project writes
// nothing.
func TestFill_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte("# x\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	run(t, tmp)
	res := run(t, tmp)
	if len(res.Written) != 0 {
		t.Errorf("second fill should write nothing, wrote %v", res.Written)
	}
	if len(res.Skipped) == 0 {
		t.Errorf("second fill should report the existing files as skipped")
	}
}

// TestFill_RestoresDeletedDocInManagedProject covers the managed path: a
// standard project that lost a canonical doc gets exactly that doc back,
// leaving every other file untouched, and is reported as Managed.
func TestFill_RestoresDeletedDocInManagedProject(t *testing.T) {
	tmp := t.TempDir()
	if err := scaffold.Run(scaffold.Options{
		ProjectName: "managedproj",
		Preset:      "standard",
		TargetDir:   tmp,
		Lang:        "en",
	}); err != nil {
		t.Fatalf("scaffold standard: %v", err)
	}

	before, err := os.ReadFile(filepath.Join(tmp, "SPEC.md"))
	if err != nil {
		t.Fatalf("read SPEC before: %v", err)
	}
	if err := os.Remove(filepath.Join(tmp, "ARCHITECTURE.md")); err != nil {
		t.Fatalf("delete ARCHITECTURE.md: %v", err)
	}

	res := run(t, tmp)

	if !res.Managed {
		t.Errorf("expected Managed=true for an already-scaffolded project")
	}
	if len(res.Written) != 1 || res.Written[0] != "ARCHITECTURE.md" {
		t.Errorf("expected only ARCHITECTURE.md restored, got %v", res.Written)
	}
	if _, err := os.Stat(filepath.Join(tmp, "ARCHITECTURE.md")); err != nil {
		t.Errorf("ARCHITECTURE.md should be restored: %v", err)
	}
	// An untouched file is byte-identical to before.
	if after, err := os.ReadFile(filepath.Join(tmp, "SPEC.md")); err != nil || string(after) != string(before) {
		t.Errorf("SPEC.md must be left untouched; err=%v", err)
	}
}
