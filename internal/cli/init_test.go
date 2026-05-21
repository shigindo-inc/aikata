package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// chdir switches the test process into dir for the lifetime of t. It is
// not parallel-safe — tests in this package therefore must not call
// t.Parallel. (Go 1.24's testing.T.Chdir would be ideal but go.mod
// targets 1.21.)
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func runInit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestInit_RequiresNoInteractiveFlag(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "minimal")
	if err == nil {
		t.Fatalf("expected error without --no-interactive")
	}
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if ee.Code != 2 {
		t.Errorf("exit code = %d, want 2", ee.Code)
	}
}

func TestInit_RequiresProjectName(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "--no-interactive", "--preset", "minimal")
	if err == nil {
		t.Fatalf("expected error when name is missing")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got: %v", err)
	}
}

func TestInit_PositionalNameGeneratesFiles(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	out, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive")
	if err != nil {
		t.Fatalf("init: %v (out: %s)", err, out)
	}
	for _, name := range []string{"README.md", "AGENTS.md", "SPEC.md"} {
		if _, err := os.Stat(filepath.Join(tmp, name)); err != nil {
			t.Errorf("%s missing: %v", name, err)
		}
	}
	readme, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "samplekata") {
		t.Errorf("README does not contain project name: %s", readme)
	}
}

func TestInit_NameFlagOverridesPositional(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "ignored", "--name", "fromFlag", "--preset", "minimal", "--no-interactive")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "fromFlag") {
		t.Errorf("README does not contain --name value: %s", readme)
	}
	if strings.Contains(string(readme), "ignored") {
		t.Errorf("README contains positional value that should have been overridden")
	}
}

func TestInit_DryRunWritesNothing(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	out, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive", "--dry-run")
	if err != nil {
		t.Fatalf("init dry-run: %v", err)
	}
	if !strings.Contains(out, "Would write 3 file(s)") {
		t.Errorf("dry-run output missing summary: %q", out)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run should not write any files: %v", entries)
	}
}

func TestInit_NonEmptyDirWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "preexisting.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive")
	if err == nil {
		t.Fatalf("expected error in non-empty dir without --force")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got: %v", err)
	}
	if !errors.Is(err, scaffold.ErrTargetDirNotEmpty) {
		t.Fatalf("expected ErrTargetDirNotEmpty in cause chain, got: %v", err)
	}
}

func TestRootCmdShowsInitInHelp(t *testing.T) {
	// Ensure newInitCmd has been wired into the root command.
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(buf.String(), "init") {
		t.Errorf("help does not mention `init`:\n%s", buf.String())
	}
}
