package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestListComponents_TextOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "components"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list components: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "memory\n") {
		t.Errorf("list components: expected memory in output:\n%s", got)
	}
}

func TestListComponents_JSONOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "components", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list components --json: %v\nstderr: %s", err, errBuf.String())
	}

	var report struct {
		Version int    `json:"version"`
		Kind    string `json:"kind"`
		Items   []struct {
			Name        string `json:"name"`
			Status      string `json:"status"`
			Description string `json:"description"`
		} `json:"items"`
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode json: %v\noutput: %s", err, out.String())
	}
	if report.Version != 1 {
		t.Errorf("version = %d, want 1", report.Version)
	}
	if report.Kind != "components" {
		t.Errorf("kind = %q, want components", report.Kind)
	}
	if report.Summary.Total != len(report.Items) {
		t.Errorf("summary.total %d != len(items) %d", report.Summary.Total, len(report.Items))
	}
	hasMemory := false
	for _, it := range report.Items {
		if it.Name == "memory" {
			hasMemory = true
			if it.Status != "active" {
				t.Errorf("memory.status = %q, want active", it.Status)
			}
			if it.Description == "" {
				t.Errorf("memory.description should be non-empty")
			}
		}
	}
	if !hasMemory {
		t.Errorf("expected memory in items; got %+v", report.Items)
	}
}
