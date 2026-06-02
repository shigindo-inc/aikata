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

func TestInit_AutoFallsBackToNonInteractiveWithoutTTY(t *testing.T) {
	// Force the TTY detector to report "no TTY" so we are exercising
	// the auto-fallback branch rather than relying on whatever stdin
	// `go test` happens to attach to this test (macOS in particular
	// can leave os.Stdin in character-device mode under `go test`).
	prev := isTTYFunc
	isTTYFunc = func() bool { return false }
	t.Cleanup(func() { isTTYFunc = prev })

	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "minimal")
	if err != nil {
		t.Fatalf("expected auto-fallback to succeed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "README.md")); err != nil {
		t.Errorf("README.md missing after auto-fallback init: %v", err)
	}
}

func TestInit_InteractivePromptHappyPath(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	// Force the TTY check to true and feed answers via the cobra
	// command's input reader.
	prev := isTTYFunc
	isTTYFunc = func() bool { return true }
	t.Cleanup(func() { isTTYFunc = prev })

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Name, scope, stack, language, ai-tools, with-memory, then trailing
	// `n` answers for with-ui / with-api / with-tdd / with-changelog /
	// monorepo / prompts.
	cmd.SetIn(strings.NewReader("interactiveproj\nminimal\nnone\nen\nclaude\nn\nn\nn\nn\nn\nn\nn\nn\n"))
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("interactive init: %v (out: %s)", err, out.String())
	}
	body, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(body), "interactiveproj") {
		t.Errorf("README does not carry interactive name:\n%s", body)
	}
	// minimal preset, so .aikata/aikata.yaml must NOT exist.
	if _, err := os.Stat(filepath.Join(tmp, ".aikata", "aikata.yaml")); !os.IsNotExist(err) {
		t.Errorf("minimal preset should not produce .aikata/aikata.yaml: %v", err)
	}
}

func TestInit_InteractiveAcceptsDefaults(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	prev := isTTYFunc
	isTTYFunc = func() bool { return true }
	t.Cleanup(func() { isTTYFunc = prev })

	cmd := newInitCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// Provide name, then blank lines to accept scope / stack / language /
	// ai-tools / memory / ui / api / tdd / changelog / monorepo / prompts /
	// env defaults.
	cmd.SetIn(strings.NewReader("defaultproj\n" + strings.Repeat("\n", 12)))
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("interactive init (defaults): %v (out: %s)", err, out.String())
	}
	// Default preset is standard, so .aikata/aikata.yaml must exist.
	if _, err := os.Stat(filepath.Join(tmp, ".aikata", "aikata.yaml")); err != nil {
		t.Errorf("default preset (standard) should produce .aikata/aikata.yaml: %v", err)
	}
	// Default with-memory is false.
	if _, err := os.Stat(filepath.Join(tmp, "docs", "memory")); !os.IsNotExist(err) {
		t.Errorf("default with-memory is false; docs/memory/ must not exist: %v", err)
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
	if !strings.Contains(out, "Would write 4 file(s)") {
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

// TestInit_NonEmptyDirWithoutForce pins the adoption fallback end to end
// (ADR 0037 D4): init in a non-empty directory succeeds, writes the
// proposal under .aikata-proposed/, leaves the project root untouched,
// prints an actionable notice, and refuses a second run.
func TestInit_NonEmptyDirWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "preexisting.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	chdir(t, tmp)
	out, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive")
	if err != nil {
		t.Fatalf("init should succeed via the proposal fallback, got: %v", err)
	}
	if !strings.Contains(out, ".aikata-proposed/") {
		t.Errorf("expected a proposal notice mentioning .aikata-proposed/, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".aikata-proposed", "AGENTS.md")); err != nil {
		t.Errorf("expected .aikata-proposed/AGENTS.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "AGENTS.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("AGENTS.md should not land in the project root: %v", err)
	}

	// A second run must refuse rather than clobber the prior proposal.
	_, err = runInit(t, "samplekata", "--preset", "minimal", "--no-interactive")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2 on collision, got: %v", err)
	}
	if !errors.Is(err, scaffold.ErrProposalExists) {
		t.Fatalf("expected ErrProposalExists in cause chain, got: %v", err)
	}
}

func TestInit_StandardPresetWritesAikataYaml(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive")
	if err != nil {
		t.Fatalf("init standard: %v", err)
	}
	yamlPath := filepath.Join(tmp, ".aikata", "aikata.yaml")
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read aikata.yaml: %v", err)
	}
	if !strings.Contains(string(body), "name: samplekata") {
		t.Errorf("aikata.yaml does not contain project name: %s", body)
	}
}

func TestInit_WithMemoryFlagProducesMemoryDir(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive", "--with-memory")
	if err != nil {
		t.Fatalf("init --with-memory: %v", err)
	}
	for _, rel := range []string{
		"docs/memory/README.md",
		"docs/memory/user.md",
		"docs/memory/feedback.md",
		"docs/memory/project.md",
		"docs/memory/reference.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected %s after --with-memory: %v", rel, err)
		}
	}
}

func TestInit_WithoutMemoryFlagOmitsMemoryDir(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive")
	if err != nil {
		t.Fatalf("init standard: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs", "memory")); !os.IsNotExist(err) {
		t.Errorf("docs/memory/ must NOT exist without --with-memory: %v", err)
	}
}

func TestInit_DefaultPresetIsStandard(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--no-interactive")
	if err != nil {
		t.Fatalf("init with default preset: %v", err)
	}
	// The standard preset places aikata.yaml under .aikata/, which the
	// minimal preset does not produce.
	if _, err := os.Stat(filepath.Join(tmp, ".aikata", "aikata.yaml")); err != nil {
		t.Errorf("expected .aikata/aikata.yaml under default preset (standard), got %v", err)
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
