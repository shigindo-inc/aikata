package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmdVersionFlag(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	want := "aikata version 0.0.1-test\n"
	if got != want {
		t.Fatalf("version output mismatch:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestRootCmdHelpMentionsAikata(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "aikata") {
		t.Fatalf("help output does not mention aikata:\n%s", buf.String())
	}
}
