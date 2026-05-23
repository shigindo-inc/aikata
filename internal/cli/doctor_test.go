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
