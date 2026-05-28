package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestListCapabilities_TextOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "capabilities"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list capabilities: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "memory\n") {
		t.Errorf("list capabilities: expected memory in output:\n%s", got)
	}
	if strings.Contains(got, "adr") {
		t.Errorf("list capabilities should not include adr (it's an artifact, not a capability):\n%s", got)
	}
}

func TestListCapabilities_JSONOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "capabilities", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list capabilities --json: %v\nstderr: %s", err, errBuf.String())
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
	if report.Kind != "capabilities" {
		t.Errorf("kind = %q, want capabilities", report.Kind)
	}
	if report.Summary.Total != len(report.Items) {
		t.Errorf("summary.total %d != len(items) %d", report.Summary.Total, len(report.Items))
	}
	hasMemory := false
	hasMonorepo := false
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
		if it.Name == "monorepo" {
			hasMonorepo = true
		}
		if it.Name == "adr" {
			t.Errorf("list capabilities should not include adr; got %+v", it)
		}
	}
	if !hasMemory {
		t.Errorf("expected memory in items; got %+v", report.Items)
	}
	if !hasMonorepo {
		t.Errorf("expected monorepo in items (v0.7.1 addition); got %+v", report.Items)
	}
}

func TestListArtifacts_TextOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "artifacts"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list artifacts: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "adr\n") {
		t.Errorf("list artifacts: expected adr in output:\n%s", got)
	}
	if strings.Contains(got, "memory") {
		t.Errorf("list artifacts should not include memory (it's a capability):\n%s", got)
	}
}

func TestListArtifacts_JSONOutput(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "artifacts", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list artifacts --json: %v\nstderr: %s", err, errBuf.String())
	}

	var report struct {
		Version int    `json:"version"`
		Kind    string `json:"kind"`
		Items   []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode json: %v\noutput: %s", err, out.String())
	}
	if report.Kind != "artifacts" {
		t.Errorf("kind = %q, want artifacts", report.Kind)
	}
	hasAdr := false
	for _, it := range report.Items {
		if it.Name == "adr" {
			hasAdr = true
		}
	}
	if !hasAdr {
		t.Errorf("expected adr in items; got %+v", report.Items)
	}
}
