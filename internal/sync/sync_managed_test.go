package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/managed"
)

// gitignoreStatus returns the sync status recorded for `.gitignore` in
// a run result, or "" if the path was not classified.
func gitignoreStatus(r RunResult) Status {
	for _, f := range r.Files {
		if f.Path == ".gitignore" {
			return f.Status
		}
	}
	return ""
}

// TestRun_ManagedAppend_CleanNoOp pins that a freshly-scaffolded
// `.gitignore` (which now carries the managed-block markers, ADR 0038)
// is a no-op on sync: the block re-renders identically, so nothing is
// written and the status is unchanged.
func TestRun_ManagedAppend_CleanNoOp(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	before, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read seeded .gitignore: %v", err)
	}
	if !managed.HasBlock(before) {
		t.Fatalf("fresh .gitignore should carry managed markers:\n%s", before)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := gitignoreStatus(result); got != StatusUnchanged {
		t.Errorf(".gitignore status = %q, want %q", got, StatusUnchanged)
	}
	after, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf(".gitignore should be byte-identical after a clean sync:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestRun_ManagedAppend_RefreshesBlockPreservingUserLines pins ADR 0038:
// when the on-disk `.gitignore` carries a stale aikata block plus
// user-owned lines outside the markers, sync refreshes only the block
// (to the current upstream rendering) and byte-preserves the user lines
// — and never writes conflict markers, even though all three sides
// (ancestor / current / upstream) differ.
func TestRun_ManagedAppend_RefreshesBlockPreservingUserLines(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	gitignorePath := filepath.Join(root, ".gitignore")
	userTop := "# project-specific ignores\nsecret-local/\n"
	userBottom := "*.bak\n"
	stale := userTop + "\n" + string(managed.Frame([]byte("OUTDATED-AIKATA-BLOCK\n"))) + userBottom
	if err := os.WriteFile(gitignorePath, []byte(stale), 0o644); err != nil {
		t.Fatalf("write stale .gitignore: %v", err)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 0 {
		t.Fatalf("managed-append must never conflict, got %d conflicts: %+v", result.Conflicts, result)
	}
	if got := gitignoreStatus(result); got != StatusUpstreamApplied {
		t.Errorf(".gitignore status = %q, want %q", got, StatusUpstreamApplied)
	}

	got, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read merged .gitignore: %v", err)
	}
	out := string(got)
	// User-owned lines outside the markers survive.
	for _, needle := range []string{"# project-specific ignores", "secret-local/", "*.bak"} {
		if !strings.Contains(out, needle) {
			t.Errorf("user line %q dropped by managed-append refresh:\n%s", needle, out)
		}
	}
	// The stale block body is replaced with the current upstream block.
	if strings.Contains(out, "OUTDATED-AIKATA-BLOCK") {
		t.Errorf("stale aikata block was not refreshed:\n%s", out)
	}
	if !strings.Contains(out, "/.aikata-proposed/") {
		t.Errorf("refreshed block missing current aikata content:\n%s", out)
	}
	// Never conflict markers for a managed-append path.
	if strings.Contains(out, "<<<<<<<") || strings.Contains(out, ">>>>>>>") {
		t.Errorf(".gitignore must not carry conflict markers:\n%s", out)
	}
}

// TestRun_ManagedAppend_MigratesPristineLegacyFile pins the v0.9.8
// upgrade path (ADR 0038): a project scaffolded by v0.9.7-or-earlier has
// a MARKERLESS `.gitignore`. The first `aikata sync` after upgrade must
// migrate it to the framed form in place — NOT append a second framed
// copy (which a naive ApplyBlock would do, silently doubling every
// rule). The manifest ancestor hash is what proves the file is still
// pristine aikata output.
func TestRun_ManagedAppend_MigratesPristineLegacyFile(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	gitignorePath := filepath.Join(root, ".gitignore")
	framed, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read seeded .gitignore: %v", err)
	}
	// Reconstruct the pre-0.9.8 on-disk reality: the raw (markerless)
	// block body, plus a manifest whose `.gitignore` hash records that
	// raw body (that is exactly what old `init` wrote and tracked).
	startIdx := strings.Index(string(framed), managed.BlockStart+"\n")
	endIdx := strings.Index(string(framed), managed.BlockEnd)
	if startIdx < 0 || endIdx < 0 {
		t.Fatalf("seeded .gitignore missing markers:\n%s", framed)
	}
	rawBody := string(framed)[startIdx+len(managed.BlockStart)+1 : endIdx]
	if err := os.WriteFile(gitignorePath, []byte(rawBody), 0o644); err != nil {
		t.Fatalf("write legacy markerless .gitignore: %v", err)
	}
	if managed.HasBlock([]byte(rawBody)) {
		t.Fatalf("test setup wrong: legacy body should be markerless:\n%s", rawBody)
	}

	// Point the manifest's `.gitignore` ancestor hash at the raw body so
	// the file reads as pristine legacy output, not user content.
	manifest, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	rawHash := config.HashContent([]byte(rawBody))
	var pointed bool
	for i, f := range manifest.Files {
		if f.Path == ".gitignore" {
			manifest.Files[i].SHA256 = rawHash
			pointed = true
		}
	}
	if !pointed {
		t.Fatalf(".gitignore missing from manifest; cannot simulate legacy state")
	}
	if err := config.SaveManifest(root, manifest); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Conflicts != 0 {
		t.Fatalf("legacy migration must not conflict: %+v", result)
	}

	got, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read migrated .gitignore: %v", err)
	}
	out := string(got)
	// Exactly one managed block — no duplication.
	if n := strings.Count(out, managed.BlockStart); n != 1 {
		t.Errorf("migrated .gitignore must carry exactly one block, got %d:\n%s", n, out)
	}
	if !managed.HasBlock(got) {
		t.Errorf("migrated .gitignore should be framed:\n%s", out)
	}
	// No aikata rule appears twice (the duplication symptom).
	if strings.Count(out, "/.aikata-proposed/") != 1 {
		t.Errorf("rule duplicated after migration:\n%s", out)
	}
}

// TestRun_ManagedAppend_RespectsUserDeletion pins that a user who
// deletes `.gitignore` entirely is not re-created behind their back
// (ADR 0019 carries over to the managed-append branch).
func TestRun_ManagedAppend_RespectsUserDeletion(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	if err := os.Remove(filepath.Join(root, ".gitignore")); err != nil {
		t.Fatalf("remove .gitignore: %v", err)
	}

	result, err := Run(Options{Root: root, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := gitignoreStatus(result); got != StatusUserDeleted {
		t.Errorf(".gitignore status = %q, want %q", got, StatusUserDeleted)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf(".gitignore should stay deleted, got err=%v", err)
	}
}
