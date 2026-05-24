package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// writeAikataYaml seeds the target dir with a minimal valid project
// config so `aikata add` subcommands have something to load. The
// `chdir` helper used by these tests is defined in init_test.go.
func writeAikataYaml(t *testing.T, root, name string) {
	t.Helper()
	if err := config.Save(root, config.Default(name, "en")); err != nil {
		t.Fatalf("seed config: %v", err)
	}
}

func TestAddRoot_NoArgs_PrintsHelpExit(t *testing.T) {
	root := newRootCmd("test")
	root.SetArgs([]string{"add"})
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetOut(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no subcommand is provided")
	}
}

func TestAddRoot_UnknownComponent_Exit2(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	writeAikataYaml(t, tmp, "demo")

	root := newRootCmd("test")
	root.SetArgs([]string{"add", "nope"})
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetOut(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
	// cobra surfaces unknown subcommands itself before our RunE is
	// reached, so this asserts the user sees something actionable.
	if !strings.Contains(err.Error(), "nope") && !strings.Contains(stderr.String(), "nope") {
		t.Errorf("expected error to mention the unknown name; err=%v stderr=%q", err, stderr.String())
	}
}

func TestAddMemory_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	writeAikataYaml(t, tmp, "demo")

	root := newRootCmd("test")
	root.SetArgs([]string{"add", "memory"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs/memory/user.md")); err != nil {
		t.Errorf("expected docs/memory/user.md to exist after add memory: %v", err)
	}
}

func TestAddMemory_WithoutConfig_Exit2(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	// No config seeded.

	root := newRootCmd("test")
	root.SetArgs([]string{"add", "memory"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Errorf("expected ExitError with Code=2, got %#v", err)
	}
}

func TestAddMemory_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	writeAikataYaml(t, tmp, "demo")

	root := newRootCmd("test")
	root.SetArgs([]string{"add", "memory", "--dry-run"})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs/memory")); !os.IsNotExist(err) {
		t.Errorf("expected docs/memory/ not to exist after --dry-run; got err=%v", err)
	}
	if !strings.Contains(stdout.String(), "docs/memory/user.md") {
		t.Errorf("expected dry-run plan to list memory files; got %q", stdout.String())
	}
}
