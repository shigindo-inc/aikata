package doctor

import (
	"fmt"
	"io"
	"strings"
)

// ContextJSONSchemaVersion is the schema number embedded in
// FormatContextJSON output. It versions the `kind: "context"` report
// separately from the checks report (JSONSchemaVersion).
const ContextJSONSchemaVersion = 1

type jsonContextFile struct {
	Path        string `json:"path"`
	Bytes       int    `json:"bytes"`
	OutsideRepo bool   `json:"outside_repository"`
	ImportedBy  string `json:"imported_by,omitempty"`
}

type jsonContextTool struct {
	Tool       string            `json:"tool"`
	Files      []jsonContextFile `json:"files"`
	TotalBytes int               `json:"total_bytes"`
	Notes      []string          `json:"notes"`
}

type jsonContextReport struct {
	Version        int               `json:"version"`
	Kind           string            `json:"kind"`
	Directory      string            `json:"directory"`
	RepositoryRoot string            `json:"repository_root,omitempty"`
	Tools          []jsonContextTool `json:"tools"`
	Issues         []jsonIssue       `json:"issues"`
	Summary        jsonSummary       `json:"summary"`
}

// FormatContextJSON writes report as JSON to w. It keeps the checks
// report's envelope (`version`, `issues`, `summary`) and adds `kind`,
// the inspected `directory`, the `repository_root` (omitted outside a
// repository), and one `tools` entry per AI tool:
//
//	{
//	  "version": 1,
//	  "kind": "context",
//	  "directory": "/work/parent/child",
//	  "repository_root": "/work/parent/child",
//	  "tools": [
//	    {
//	      "tool": "claude",
//	      "files": [
//	        {"path": "../CLAUDE.md", "bytes": 120, "outside_repository": true},
//	        {"path": "../AGENTS.md", "bytes": 8700, "outside_repository": true, "imported_by": "../CLAUDE.md"}
//	      ],
//	      "total_bytes": 8820,
//	      "notes": []
//	    }
//	  ],
//	  "issues": [...],
//	  "summary": {"errors": 0, "warnings": 0, "info": 1}
//	}
//
// File paths are relative to `directory` in slash form, like `file`
// in issues.
func FormatContextJSON(w io.Writer, report ContextReport) error {
	wire, summary := toJSONIssues(report.Issues)
	out := jsonContextReport{
		Version:        ContextJSONSchemaVersion,
		Kind:           "context",
		Directory:      report.Dir,
		RepositoryRoot: report.RepoRoot,
		Tools:          make([]jsonContextTool, 0, len(report.Tools)),
		Issues:         wire,
		Summary:        summary,
	}
	for _, tc := range report.Tools {
		jt := jsonContextTool{
			Tool:  tc.Tool,
			Files: make([]jsonContextFile, 0, len(tc.Files)),
			Notes: append([]string{}, tc.Notes...),
		}
		for _, f := range tc.Files {
			jf := jsonContextFile{
				Path:        displayPath(report.Dir, f.Path),
				Bytes:       f.Bytes,
				OutsideRepo: f.OutsideRepo,
			}
			if f.ImportedBy != "" {
				jf.ImportedBy = displayPath(report.Dir, f.ImportedBy)
			}
			jt.Files = append(jt.Files, jf)
			jt.TotalBytes += f.Bytes
		}
		out.Tools = append(out.Tools, jt)
	}
	return encodeJSON(w, out)
}

// FormatContext writes the human-readable rendering of report to w:
// the directory and repository root, then per tool its files in load
// order (imports indented under their importer) and notes, then the
// issues in Format's shape.
func FormatContext(w io.Writer, report ContextReport) error {
	if err := formatContext(w, report); err != nil {
		return fmt.Errorf("doctor: format context: %w", err)
	}
	return nil
}

func formatContext(w io.Writer, report ContextReport) error {
	repo := report.RepoRoot
	if repo == "" {
		repo = "none (no .git here or above)"
	}
	if _, err := fmt.Fprintf(w, "directory:  %s\nrepository: %s\n", report.Dir, repo); err != nil {
		return err
	}
	for _, tc := range report.Tools {
		if err := formatToolContext(w, report.Dir, tc); err != nil {
			return err
		}
	}
	if len(report.Issues) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return Format(w, report.Issues)
}

func formatToolContext(w io.Writer, dir string, tc ToolContext) error {
	total := 0
	for _, f := range tc.Files {
		total += f.Bytes
	}
	header := "no files listed"
	if len(tc.Files) > 0 {
		header = fmt.Sprintf("%d file(s), %d bytes", len(tc.Files), total)
	}
	if _, err := fmt.Fprintf(w, "\n%s: %s\n", tc.Tool, header); err != nil {
		return err
	}
	type row struct {
		path, bytes, flag string
	}
	depth := map[string]int{}
	rows := make([]row, 0, len(tc.Files))
	pathWidth, bytesWidth := 0, 0
	for _, f := range tc.Files {
		if f.ImportedBy != "" {
			depth[f.Path] = depth[f.ImportedBy] + 1
		}
		r := row{
			path:  strings.Repeat("  ", depth[f.Path]+1) + displayPath(dir, f.Path),
			bytes: fmt.Sprint(f.Bytes),
		}
		if f.OutsideRepo {
			r.flag = "  outside repository"
		}
		pathWidth = max(pathWidth, len(r.path))
		bytesWidth = max(bytesWidth, len(r.bytes))
		rows = append(rows, r)
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-*s  %*s%s\n", pathWidth, r.path, bytesWidth, r.bytes, r.flag); err != nil {
			return err
		}
	}
	for _, n := range tc.Notes {
		if _, err := fmt.Fprintf(w, "  note: %s\n", n); err != nil {
			return err
		}
	}
	return nil
}
