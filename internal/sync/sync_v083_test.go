package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// manifestHash returns the recorded SHA-256 for path, or "" if absent.
func manifestHash(t *testing.T, root, path string) string {
	t.Helper()
	m, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for _, f := range m.Files {
		if f.Path == path {
			return f.SHA256
		}
	}
	return ""
}

func runSync(t *testing.T, root string, opts Options) RunResult {
	t.Helper()
	opts.Root = root
	if opts.Stdout == nil {
		opts.Stdout = &bytes.Buffer{}
	}
	if opts.Stderr == nil {
		opts.Stderr = &bytes.Buffer{}
	}
	res, err := Run(opts)
	if err != nil {
		t.Fatalf("sync.Run: %v", err)
	}
	return res
}

// TestRun_UserOnlyEdit_SurvivesTwoSyncs is the ADR 0025 D1 regression
// anchor: before the fix, a clean run recorded the user's on-disk bytes
// as the new ancestor, so the *next* sync reclassified the file as
// upstream-applied and silently overwrote the user's content. With D1
// the ancestor stays at the upstream rendering, so the edit survives
// unlimited syncs.
func TestRun_UserOnlyEdit_SurvivesTwoSyncs(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	upstreamHash := manifestHash(t, root, "SPEC.md")
	if upstreamHash == "" {
		t.Fatal("SPEC.md missing from seed manifest")
	}

	specPath := filepath.Join(root, "SPEC.md")
	userContent := []byte("# SPEC.md\n\nintentional fork\n")
	if err := os.WriteFile(specPath, userContent, 0o644); err != nil {
		t.Fatalf("write user edit: %v", err)
	}

	// Run 1.
	res1 := runSync(t, root, Options{})
	if res1.Conflicts != 0 {
		t.Fatalf("run 1 should be conflict-free: %+v", res1)
	}
	if h := manifestHash(t, root, "SPEC.md"); h != upstreamHash {
		t.Errorf("after run 1, SPEC.md ancestor = %q, want upstream hash %q (D1: ancestor must stay at upstream, not absorb user bytes)", h, upstreamHash)
	}

	// Run 2 — the run that used to overwrite.
	res2 := runSync(t, root, Options{})
	if res2.Conflicts != 0 {
		t.Fatalf("run 2 should be conflict-free: %+v", res2)
	}
	got, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read SPEC.md: %v", err)
	}
	if string(got) != string(userContent) {
		t.Errorf("user content was overwritten on the second sync (the data-loss bug); got:\n%s", got)
	}
	for _, f := range res2.Files {
		if f.Path == "SPEC.md" && f.Status != StatusUserOnlyEdit {
			t.Errorf("run 2 SPEC.md status = %q, want %q", f.Status, StatusUserOnlyEdit)
		}
	}
}

// TestRun_UserDeleted_StaysDeletedAcrossTwoSyncs covers the D1 side
// effect: recording the upstream rendering keeps the manifest entry for
// a user-deleted path, so the deletion is respected across syncs rather
// than being silently re-created (ADR 0019).
func TestRun_UserDeleted_StaysDeletedAcrossTwoSyncs(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Delete a managed file the user does not want.
	target := filepath.Join(root, "docs", "troubleshooting.md")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("precondition: %s should exist after seed: %v", target, err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove: %v", err)
	}

	res1 := runSync(t, root, Options{})
	if res1.Conflicts != 0 {
		t.Fatalf("run 1 conflicts: %+v", res1)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("run 1 re-created a user-deleted file: %v", err)
	}
	// The manifest must still carry the entry so run 2 sees user-deleted
	// rather than upstream-added.
	if manifestHash(t, root, "docs/troubleshooting.md") == "" {
		t.Errorf("D1: user-deleted path dropped from manifest; next sync would resurrect it")
	}

	res2 := runSync(t, root, Options{})
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("run 2 resurrected a user-deleted file: %v", err)
	}
	_ = res2
}

// TestRun_OwnedPath_SkippedAndUntracked covers ADR 0025 D2: a path
// under `sync.own` is reported `owned`, never overwritten, and excluded
// from the regenerated manifest.
func TestRun_OwnedPath_SkippedAndUntracked(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Declare README.md as owned.
	cfg, _, err := config.LoadMigrated(root)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	cfg.Sync.Own = []string{"README.md"}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	readme := filepath.Join(root, "README.md")
	forked := []byte("# My project\n\ncompletely rewritten\n")
	if err := os.WriteFile(readme, forked, 0o644); err != nil {
		t.Fatalf("fork README: %v", err)
	}

	res := runSync(t, root, Options{})
	if res.Conflicts != 0 {
		t.Fatalf("owned path must not conflict: %+v", res)
	}
	var sawOwned bool
	for _, f := range res.Files {
		if f.Path == "README.md" {
			if f.Status != StatusOwned {
				t.Errorf("README.md status = %q, want %q", f.Status, StatusOwned)
			}
			sawOwned = true
		}
	}
	if !sawOwned {
		t.Errorf("expected an owned result for README.md; files: %+v", res.Files)
	}
	got, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if string(got) != string(forked) {
		t.Errorf("owned README was modified:\n%s", got)
	}
	if manifestHash(t, root, "README.md") != "" {
		t.Errorf("owned path must be excluded from the manifest")
	}
}

// TestRun_Reseed_ReanchorsManifest covers ADR 0025 D4: --reseed rewrites
// the manifest from the current upstream rendering and touches no source
// file.
func TestRun_Reseed_ReanchorsManifest(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)
	upstreamHash := manifestHash(t, root, "SPEC.md")

	// Corrupt the manifest entry so we can prove reseed re-anchors it.
	m, err := config.LoadManifest(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for i := range m.Files {
		if m.Files[i].Path == "SPEC.md" {
			m.Files[i].SHA256 = "deadbeef"
		}
	}
	if err := config.SaveManifest(root, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	specPath := filepath.Join(root, "SPEC.md")
	before, _ := os.ReadFile(specPath)

	res := runSync(t, root, Options{Reseed: true})
	if len(res.Files) != 0 {
		t.Errorf("reseed should not produce per-file merge results, got %+v", res.Files)
	}
	if h := manifestHash(t, root, "SPEC.md"); h != upstreamHash {
		t.Errorf("reseed did not re-anchor SPEC.md: got %q, want %q", h, upstreamHash)
	}
	after, _ := os.ReadFile(specPath)
	if string(before) != string(after) {
		t.Errorf("reseed modified a source file")
	}
}

// TestRun_RebaselineWithManifest_EmitsNotice covers ADR 0025 D4: passing
// --rebaseline to a project that already has a manifest is no longer a
// silent no-op; it prints a notice pointing at --reseed and continues
// the normal merge.
func TestRun_RebaselineWithManifest_EmitsNotice(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	var stderr bytes.Buffer
	_, err := Run(Options{Root: root, Rebaseline: true, Stdout: &bytes.Buffer{}, Stderr: &stderr})
	if err != nil {
		t.Fatalf("sync.Run: %v", err)
	}
	if !strings.Contains(stderr.String(), "rebaseline skipped") {
		t.Errorf("expected a rebaseline-skipped notice on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "--reseed") {
		t.Errorf("notice should point at --reseed, got: %q", stderr.String())
	}
}
