package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDescribePreset_ActiveText(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"describe", "preset", "minimal"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("describe preset minimal: %v\nstderr: %s", err, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"preset: minimal", "status: active", "languages: en, ja", "Minimal preset"} {
		if !strings.Contains(got, want) {
			t.Errorf("describe preset minimal: missing %q\nfull: %s", want, got)
		}
	}
}

func TestDescribePreset_ReservedText(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"describe", "preset", "extended"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("describe preset extended: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "status: reserved") {
		t.Errorf("describe preset extended: expected reserved status\nfull: %s", got)
	}
	if strings.Contains(got, "languages:") {
		t.Errorf("describe preset extended: reserved preset should not list languages\nfull: %s", got)
	}
}

type decodedDescribeReport struct {
	Version int    `json:"version"`
	Kind    string `json:"kind"`
	Item    struct {
		Name        string   `json:"name"`
		Status      string   `json:"status"`
		Description string   `json:"description"`
		Languages   []string `json:"languages"`
	} `json:"item"`
}

func TestDescribePreset_JSON(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"describe", "preset", "standard", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("describe preset standard --json: %v", err)
	}
	var report decodedDescribeReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("decode JSON: %v\nstdout: %s", err, out.String())
	}
	if report.Version != 1 || report.Kind != "preset" || report.Item.Name != "standard" || report.Item.Status != "active" {
		t.Fatalf("unexpected JSON shape: %+v", report)
	}
	if report.Item.Description == "" {
		t.Fatalf("expected non-empty description; got %+v", report.Item)
	}
}

func TestDescribePreset_UnknownExits2(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"describe", "preset", "does-not-exist"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("describe preset does-not-exist: expected error, got nil")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 2 {
		t.Fatalf("exit code = %d, want 2", exitErr.Code)
	}
}
