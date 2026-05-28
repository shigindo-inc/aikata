package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runNewFromRoot runs `aikata new <args...>` with the process cwd at
// root, matching the production code path.
func runNewFromRoot(t *testing.T, root string, args ...string) (string, string, error) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"new"}, args...))
	cerr := cmd.Execute()
	return out.String(), errBuf.String(), cerr
}

// TestNewAdr_HappyPath verifies that `aikata new adr "<title>"`
// stamps a docs/adr/NNNN-*.md file in the seeded project.
func TestNewAdr_HappyPath(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")

	_, _, err := runNewFromRoot(t, root, "adr", "First New Style")
	if err != nil {
		t.Fatalf("new adr: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(root, "docs", "adr"))
	if err != nil {
		t.Fatalf("readdir docs/adr: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one ADR file under docs/adr")
	}
}

// TestNew_WithoutConfig_Exit2 mirrors the enable equivalent: running
// outside an aikata project surfaces an ExitError with code 2.
func TestNew_WithoutConfig_Exit2(t *testing.T) {
	root := t.TempDir()
	_, _, err := runNewFromRoot(t, root, "adr", "Bare")
	if err == nil {
		t.Fatalf("expected error when aikata config missing")
	}
	var exit *ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exit.Code != 2 {
		t.Errorf("exit code = %d, want 2", exit.Code)
	}
}

// TestNew_UnknownArtifact_Exit2 confirms the parent dispatcher
// rejects unknown artifact names.
func TestNew_UnknownArtifact_Exit2(t *testing.T) {
	root := t.TempDir()
	writeAikataYamlV2(t, root, "tester")
	_, _, err := runNewFromRoot(t, root, "nope-not-an-artifact")
	if err == nil {
		t.Fatalf("expected error for unknown artifact")
	}
	if !strings.Contains(err.Error(), "unknown artifact") {
		t.Errorf("expected message to mention 'unknown artifact'; got %q", err.Error())
	}
}
