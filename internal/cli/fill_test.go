package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runFill executes `aikata fill` (via its cobra command) in the current
// working directory, returning combined stdout/stderr.
func runFill(t *testing.T) (string, error) {
	t.Helper()
	cmd := newFillCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	return buf.String(), err
}

// TestFill_RegisteredInRoot pins the wiring: `fill` must be reachable
// from the root command (parity with TestRootCmdShowsInitInHelp).
func TestFill_RegisteredInRoot(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var found bool
	for _, c := range cmd.Commands() {
		if c.Name() == "fill" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("fill subcommand is not registered on the root command")
	}
}

// TestFill_RejectsArgs guards the zero-config contract: fill takes no
// positional arguments.
func TestFill_RejectsArgs(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	cmd := newFillCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"unexpected"})
	if err := cmd.Execute(); err == nil {
		t.Errorf("fill should reject positional arguments, got nil error")
	}
}

// TestFill_AdoptsUnmanagedRepo exercises the command end-to-end: a repo
// with only a hand-written AGENTS.md is adopted, the missing standard
// docs are written, and the hand-written file is preserved.
func TestFill_AdoptsUnmanagedRepo(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	handwritten := "# my AGENTS\nrules\n"
	if err := os.WriteFile(filepath.Join(tmp, "AGENTS.md"), []byte(handwritten), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	out, err := runFill(t)
	if err != nil {
		t.Fatalf("fill: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Adopted this repository") {
		t.Errorf("expected adoption summary, got: %q", out)
	}
	for _, want := range []string{"SPEC.md", "ARCHITECTURE.md"} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected %s to be written: %v", want, err)
		}
	}
	if got, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md")); err != nil || string(got) != handwritten {
		t.Errorf("hand-written AGENTS.md must be preserved; err=%v content=%q", err, string(got))
	}
	if _, err := os.Stat(filepath.Join(tmp, ".aikata", "aikata.yaml")); err != nil {
		t.Errorf("expected the repo to become aikata-managed: %v", err)
	}
}
