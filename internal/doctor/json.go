package doctor

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONSchemaVersion is the schema number embedded in FormatJSON output.
// Bump this when the JSON shape changes in a backward-incompatible way
// so downstream tooling can detect old aikata binaries.
const JSONSchemaVersion = 1

// jsonIssue is the wire shape of a single Issue in JSON output. It
// omits empty Code so older tooling can ignore the field without
// failing on null vs. string mismatches.
type jsonIssue struct {
	Level   string `json:"level"`
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type jsonSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

type jsonReport struct {
	Version int          `json:"version"`
	Issues  []jsonIssue  `json:"issues"`
	Summary jsonSummary  `json:"summary"`
}

// FormatJSON writes a machine-readable rendering of issues to w. The
// schema is intentionally flat so it composes well with `jq`:
//
//	{
//	  "version": 1,
//	  "issues": [
//	    {"level": "error", "file": "SPEC.md", "line": 12, "code": "frontmatter.missing-key.version", "message": "..."}
//	  ],
//	  "summary": {"errors": 1, "warnings": 0, "info": 0}
//	}
//
// Empty `line` and `code` are omitted so consumers do not need to
// distinguish "missing field" from "explicit zero/empty".
func FormatJSON(w io.Writer, issues []Issue) error {
	report := jsonReport{
		Version: JSONSchemaVersion,
		Issues:  make([]jsonIssue, 0, len(issues)),
	}
	for _, iss := range issues {
		report.Issues = append(report.Issues, jsonIssue{
			Level:   iss.Level.String(),
			File:    iss.File,
			Line:    iss.Line,
			Code:    iss.Code,
			Message: iss.Message,
		})
		switch iss.Level {
		case LevelError:
			report.Summary.Errors++
		case LevelWarning:
			report.Summary.Warnings++
		case LevelInfo:
			report.Summary.Info++
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("doctor: format json: %w", err)
	}
	return nil
}
