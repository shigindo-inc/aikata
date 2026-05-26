package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runDoctor(t *testing.T) (string, error) {
	t.Helper()
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	return buf.String(), err
}

func TestDoctor_HealthyAfterInitExitsZero(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	out, err := runDoctor(t)
	if err != nil {
		t.Fatalf("doctor on fresh standard preset should succeed, got %v\n%s", err, out)
	}
}

func TestDoctor_BrokenAGENTSLinkSetsExitCodeThree(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Append a known-bad relative link to AGENTS.md.
	agentsPath := filepath.Join(tmp, "AGENTS.md")
	body, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read agents: %v", err)
	}
	body = append(body, []byte("\n\n[broken](./no-such-file.md)\n")...)
	if err := os.WriteFile(agentsPath, body, 0o644); err != nil {
		t.Fatalf("write agents: %v", err)
	}

	out, err := runDoctor(t)
	if err == nil {
		t.Fatalf("doctor should return error for broken link, got nil\n%s", out)
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("expected ExitError code 3, got %v", err)
	}
	if !strings.Contains(out, "broken link") {
		t.Errorf("expected broken-link diagnostic in output:\n%s", out)
	}
}

// TestDoctor_StrictTreatsWarningsAsErrors covers the v0.5+ dogfood
// gate. A project with a stale `updated:` frontmatter triggers a
// warning-level finding; --strict must turn that into exit 3 while
// the plain run continues to exit 0.
func TestDoctor_StrictTreatsWarningsAsErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Plant a stale `updated:` value on README.md so the doctor's
	// `updated.stale` warning fires (>365 days from any plausible
	// current date). The exact `updated:` value scaffold writes
	// depends on the test's wall clock, so rewrite the frontmatter
	// rather than pattern-match on the date.
	readmePath := filepath.Join(tmp, "README.md")
	stale := []byte("---\nproject: samplekata\nstatus: draft\nversion: 0.0.1\nupdated: 2000-01-01\naudience: [human, agent]\n---\n\n# README.md\n")
	if err := os.WriteFile(readmePath, stale, 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	// Without --strict, warnings do not affect exit code.
	if _, err := runDoctor(t); err != nil {
		t.Errorf("non-strict doctor should not fail on a warning, got: %v", err)
	}

	// With --strict, the warning bumps exit to 3.
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--strict"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("strict doctor should fail on a warning, got nil\n%s", buf.String())
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("strict doctor: expected ExitError code 3, got %v", err)
	}
	if !strings.Contains(err.Error(), "warning") {
		t.Errorf("strict error message should mention warnings: %v", err)
	}
}

func TestRootCmdShowsDoctorInHelp(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(buf.String(), "doctor") {
		t.Errorf("help does not mention `doctor`:\n%s", buf.String())
	}
}
