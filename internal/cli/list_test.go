package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestListPresets_TextOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "presets"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list presets: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"minimal\n", "standard\n", "flutter\n", "typescript\n", "extended (reserved)\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("list presets: stdout missing %q\nfull: %s", want, got)
		}
	}
}

type decodedListReport struct {
	Version int `json:"version"`
	Kind    string
	Items   []struct {
		Name      string   `json:"name"`
		Status    string   `json:"status"`
		Languages []string `json:"languages"`
	} `json:"items"`
	Summary struct {
		Total int `json:"total"`
	} `json:"summary"`
}

func TestListPresets_JSONOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "presets", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list presets --json: %v\nstderr: %s", err, errBuf.String())
	}
	var report decodedListReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode JSON: %v\nstdout: %s", err, out.String())
	}
	if report.Version != 1 {
		t.Fatalf("schema version = %d, want 1", report.Version)
	}
	if report.Kind != "presets" {
		t.Fatalf("kind = %q, want presets", report.Kind)
	}
	if report.Summary.Total != len(report.Items) {
		t.Fatalf("summary.total = %d != len(items) = %d", report.Summary.Total, len(report.Items))
	}
	if report.Summary.Total < 5 {
		t.Fatalf("summary.total = %d, want >= 5", report.Summary.Total)
	}
}

func TestListStacks(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "stacks"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list stacks: %v", err)
	}
	for _, want := range []string{"flutter\n", "typescript\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("list stacks: missing %q\nfull: %s", want, out.String())
		}
	}
}

func TestListAITools(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "ai-tools"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list ai-tools: %v", err)
	}
	got := out.String()
	for _, want := range []string{"claude\n", "codex\n", "cursor\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("list ai-tools: missing %q\nfull: %s", want, got)
		}
	}
	// Order must be alphabetical: claude, codex, cursor.
	idxClaude := strings.Index(got, "claude")
	idxCodex := strings.Index(got, "codex")
	idxCursor := strings.Index(got, "cursor")
	if idxClaude >= idxCodex || idxCodex >= idxCursor {
		t.Errorf("list ai-tools: not alphabetically sorted\nfull: %s", got)
	}
}
