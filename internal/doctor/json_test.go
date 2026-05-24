package doctor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// decodedReport mirrors jsonReport for the test layer. It exists so
// the test asserts against a concrete struct rather than a generic
// map/interface{} decode — that path triggers a CWE-502 lint and, more
// importantly, would let typo'd field names pass silently.
type decodedReport struct {
	Version int            `json:"version"`
	Issues  []decodedIssue `json:"issues"`
	Summary decodedSummary `json:"summary"`
}

type decodedIssue struct {
	Level   string `json:"level"`
	File    string `json:"file"`
	Line    *int   `json:"line,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type decodedSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

func TestFormatJSON_SchemaShape(t *testing.T) {
	issues := []Issue{
		{Level: LevelError, File: "SPEC.md", Line: 12, Code: "frontmatter.missing-key.version", Message: "frontmatter missing required key \"version\""},
		{Level: LevelWarning, File: "GLOSSARY.md", Code: "updated.stale", Message: "updated 2020-01-01 is more than 365 days old"},
		{Level: LevelInfo, File: "docs/adr", Message: "ADR number 0002 is unused (gap below 0008)"},
	}
	var buf bytes.Buffer
	if err := FormatJSON(&buf, issues); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	var got decodedReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if got.Version != JSONSchemaVersion {
		t.Errorf("Version = %d, want %d", got.Version, JSONSchemaVersion)
	}
	if got.Summary.Errors != 1 || got.Summary.Warnings != 1 || got.Summary.Info != 1 {
		t.Errorf("Summary = %+v, want errors=1 warnings=1 info=1", got.Summary)
	}
	if len(got.Issues) != 3 {
		t.Fatalf("Issues len = %d, want 3", len(got.Issues))
	}
	// Spot-check omitempty: the LevelInfo issue has neither Line
	// nor Code, so both fields should disappear from the JSON output
	// rather than appearing as zero values.
	info := got.Issues[2]
	if info.Line != nil {
		t.Errorf("line should be omitted when zero, got: %+v", info)
	}
	if info.Code != "" {
		t.Errorf("code should be omitted when empty, got: %+v", info)
	}
}

func TestFormatJSON_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatJSON(&buf, nil); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	out := buf.String()
	// Issues must be the empty array, not null — downstream `jq -e '.issues | length'`
	// should work without a null-guard.
	if !strings.Contains(out, `"issues": []`) {
		t.Errorf("expected empty array, got:\n%s", out)
	}
	if !strings.Contains(out, `"errors": 0`) {
		t.Errorf("expected zero errors in summary, got:\n%s", out)
	}
}

func TestFormatJSON_IsValidJSON(t *testing.T) {
	issues := []Issue{
		{Level: LevelError, File: "x", Line: 1, Code: "c", Message: "m"},
	}
	var buf bytes.Buffer
	if err := FormatJSON(&buf, issues); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	// Round-trip through the concrete decoder to assert validity
	// without using interface{} (which would trigger CWE-502 lints).
	var got decodedReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(got.Issues) != 1 {
		t.Errorf("round-trip lost issue, got %+v", got)
	}
}
