package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionCmd_KnownShells(t *testing.T) {
	cases := []struct {
		shell       string
		wantSubstr  string
		description string
	}{
		{"bash", "bash completion V2 for aikata", "bash V2 header"},
		{"zsh", "#compdef aikata", "zsh compdef directive"},
		{"fish", "fish completion for aikata", "fish header"},
		{"powershell", "powershell completion for aikata", "powershell header"},
	}
	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			cmd := newRootCmd("0.0.1-test")
			var out, errBuf bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errBuf)
			cmd.SetArgs([]string{"completion", tc.shell})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("completion %s: unexpected error: %v\nstderr: %s", tc.shell, err, errBuf.String())
			}
			got := out.String()
			if !strings.Contains(got, tc.wantSubstr) {
				t.Fatalf("completion %s: stdout does not contain %q (%s)\nfirst 200 chars: %q",
					tc.shell, tc.wantSubstr, tc.description, truncate(got, 200))
			}
		})
	}
}

func TestCompletionCmd_UnknownShellFails(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"completion", "tcsh"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("completion tcsh: expected non-nil error for unknown shell, got nil\nstdout: %s\nstderr: %s",
			out.String(), errBuf.String())
	}
}

func TestCompletionCmd_RequiresExactlyOneArg(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"completion"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("completion (no arg): expected non-nil error, got nil")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
