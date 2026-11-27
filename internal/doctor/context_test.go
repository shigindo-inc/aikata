package doctor

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/generate"
)

// contextTree builds the ADR 0051 layout under a fresh temp dir:
//
//	parent/            (no .git)
//	  CLAUDE.md        imports @AGENTS.md
//	  AGENTS.md        holds sharedRule
//	  child/           (.git)
//	    CLAUDE.md      imports @AGENTS.md
//	    AGENTS.md      repeats sharedRule
//
// It returns the parent and child paths with symlinks resolved, so they
// compare equal to the paths InspectContext reports.
func contextTree(t *testing.T) (parent, child string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	parent = filepath.Join(root, "parent")
	child = filepath.Join(parent, "child")
	writeFiles(t, map[string]string{
		filepath.Join(parent, "CLAUDE.md"):        "# Parent\n\n@AGENTS.md\n",
		filepath.Join(parent, "AGENTS.md"):        "# Parent rules\n\n" + sharedRule + "\n\nParent-only rule.\n",
		filepath.Join(child, ".git", "HEAD"):      "ref: refs/heads/main\n",
		filepath.Join(child, "CLAUDE.md"):         "# Child\n\n@AGENTS.md\n",
		filepath.Join(child, "AGENTS.md"):         "# Child rules\n\nChild-only rule.\n\n" + sharedRule + "\n",
		filepath.Join(root, "unrelated", "x.txt"): "outside the walk",
	})
	return parent, child
}

const sharedRule = "Never commit secrets; reference .env.example instead."

func writeFiles(t *testing.T, files map[string]string) {
	t.Helper()
	for path, body := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// fileRow is the comparable part of a ContextFile, with paths relative
// to a test root.
type fileRow struct {
	Path        string
	Outside     bool
	ImportedBy  string
	SizeMatches bool
}

func rows(t *testing.T, root string, files []ContextFile) []fileRow {
	t.Helper()
	out := []fileRow{}
	for _, f := range files {
		info, err := os.Stat(f.Path)
		if err != nil {
			t.Fatalf("listed file %s: %v", f.Path, err)
		}
		r := fileRow{Path: relTo(t, root, f.Path), Outside: f.OutsideRepo, SizeMatches: info.Size() == int64(f.Bytes)}
		if f.ImportedBy != "" {
			r.ImportedBy = relTo(t, root, f.ImportedBy)
		}
		out = append(out, r)
	}
	return out
}

func relTo(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(rel)
}

func toolByName(t *testing.T, report ContextReport, name string) ToolContext {
	t.Helper()
	for _, tc := range report.Tools {
		if tc.Tool == name {
			return tc
		}
	}
	t.Fatalf("no %s section in report: %+v", name, report.Tools)
	return ToolContext{}
}

func issuesWithCode(report ContextReport, code string) []Issue {
	var out []Issue
	for _, iss := range report.Issues {
		if iss.Code == code {
			out = append(out, iss)
		}
	}
	return out
}

func TestInspectContext_NestedRepositoryRows(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude", "codex", "cursor"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	if report.RepoRoot != child {
		t.Fatalf("RepoRoot = %q, want %q", report.RepoRoot, child)
	}

	tests := []struct {
		tool string
		want []fileRow
	}{
		{
			// Root first, crossing the git boundary; each import right
			// after the file that imports it, resolved next to it.
			tool: "claude",
			want: []fileRow{
				{Path: "CLAUDE.md", Outside: true, SizeMatches: true},
				{Path: "AGENTS.md", Outside: true, ImportedBy: "CLAUDE.md", SizeMatches: true},
				{Path: "child/CLAUDE.md", SizeMatches: true},
				{Path: "child/AGENTS.md", ImportedBy: "child/CLAUDE.md", SizeMatches: true},
			},
		},
		{
			// From the git root down only: the parent's AGENTS.md is
			// above it.
			tool: "codex",
			want: []fileRow{{Path: "child/AGENTS.md", SizeMatches: true}},
		},
		{tool: "cursor", want: []fileRow{}},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			got := rows(t, parent, toolByName(t, report, tt.tool).Files)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rows =\n%+v\nwant\n%+v", got, tt.want)
			}
		})
	}
	if notes := toolByName(t, report, "cursor").Notes; len(notes) != 1 || !strings.Contains(notes[0], "does not document") {
		t.Errorf("cursor notes = %q, want one note that parent handling is undocumented", notes)
	}
	if notes := strings.Join(toolByName(t, report, "codex").Notes, "\n"); !strings.Contains(notes, "inference") {
		t.Errorf("codex notes must mark the git-root boundary as an inference: %q", notes)
	}
}

func TestInspectContext_OutsideRepoIssueCarriesExcludeGlob(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude", "codex"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	got := issuesWithCode(report, CodeContextOutsideRepo)
	// One issue for the parent CLAUDE.md the walk found; its import is
	// covered by it rather than reported again.
	if len(got) != 1 {
		t.Fatalf("outside-repo issues = %+v, want exactly 1", got)
	}
	iss := got[0]
	if iss.Level != LevelInfo || iss.File != "../CLAUDE.md" {
		t.Errorf("issue = %+v, want info on ../CLAUDE.md", iss)
	}
	glob := `"` + filepath.ToSlash(filepath.Join(parent, "CLAUDE.md")) + `"`
	for _, want := range []string{glob, "claudeMdExcludes", ".claude/settings.local.json", "with 1 import(s)", "drops the whole file", "into a file it imports"} {
		if !strings.Contains(iss.Message, want) {
			t.Errorf("message missing %q:\n%s", want, iss.Message)
		}
	}
}

func TestInspectContext_DuplicateParagraphAcrossFiles(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude", "codex"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	got := issuesWithCode(report, CodeContextDuplicateParagraph)
	if len(got) != 1 {
		t.Fatalf("duplicate issues = %+v, want exactly 1", got)
	}
	iss := got[0]
	// sharedRule sits on line 5 of the child's AGENTS.md and line 3 of
	// the parent's; only claude loads both.
	if iss.Level != LevelInfo || iss.File != "AGENTS.md" || iss.Line != 5 {
		t.Errorf("issue = %+v, want info at AGENTS.md:5", iss)
	}
	for _, want := range []string{"1 paragraph(s)", "repeat verbatim from ../AGENTS.md (from line 3)", "load in claude"} {
		if !strings.Contains(iss.Message, want) {
			t.Errorf("message missing %q:\n%s", want, iss.Message)
		}
	}
	if strings.Contains(iss.Message, "codex") {
		t.Errorf("codex does not load the parent AGENTS.md, so it must not be named:\n%s", iss.Message)
	}
}

// TestInspectContext_DuplicateNamesEveryToolLoadingThePair: when two
// tools load the same pair, one issue names both.
func TestInspectContext_DuplicateNamesEveryToolLoadingThePair(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub")
	writeFiles(t, map[string]string{
		filepath.Join(root, ".git", "HEAD"): "x",
		filepath.Join(root, "CLAUDE.md"):    "@AGENTS.md\n",
		filepath.Join(root, "AGENTS.md"):    sharedRule + "\n",
		filepath.Join(sub, "CLAUDE.md"):     "@AGENTS.md\n",
		filepath.Join(sub, "AGENTS.md"):     sharedRule + "\n",
	})
	report, err := inspectContext(sub, []string{"claude", "codex"}, root)
	if err != nil {
		t.Fatal(err)
	}
	// The two CLAUDE.md files share only "@AGENTS.md", which loads a
	// different file in each directory, so it is not a repeat.
	got := issuesWithCode(report, CodeContextDuplicateParagraph)
	if len(got) != 1 || got[0].File != "AGENTS.md" || !strings.Contains(got[0].Message, "load in claude, codex") {
		t.Fatalf("duplicate issues = %+v, want one on AGENTS.md naming claude, codex", got)
	}
}

func TestInspectContext_CodexOverLimitWarns(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantWarn bool
	}{
		{name: "at the limit", size: 32 * 1024, wantWarn: false},
		{name: "one byte over", size: 32*1024 + 1, wantWarn: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			sub := filepath.Join(root, "sub")
			// The root file is 1 KiB of 64-byte lines; the rest goes in
			// sub, so the cut falls inside sub/AGENTS.md.
			line := strings.Repeat("r", 63) + "\n"
			rootBody := strings.Repeat(line, 16)
			subBody := strings.Repeat(line, (tt.size-len(rootBody))/64)
			subBody += strings.Repeat("s", tt.size-len(rootBody)-len(subBody))
			writeFiles(t, map[string]string{
				filepath.Join(root, ".git", "HEAD"): "x",
				filepath.Join(root, "AGENTS.md"):    rootBody,
				filepath.Join(sub, "AGENTS.md"):     subBody,
			})
			report, err := inspectContext(sub, []string{"codex"}, root)
			if err != nil {
				t.Fatal(err)
			}
			got := issuesWithCode(report, CodeContextCodexOverLimit)
			if !tt.wantWarn {
				if len(got) != 0 {
					t.Fatalf("unexpected warning: %+v", got)
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("over-limit issues = %+v, want 1", got)
			}
			iss := got[0]
			// 32 KiB - 1 KiB = 31744 bytes of sub = 496 full lines, so
			// the cap falls at the start of line 497.
			if iss.Level != LevelWarning || iss.File != "AGENTS.md" || iss.Line != 497 {
				t.Errorf("issue = %+v, want warning at AGENTS.md:497", iss)
			}
			if !strings.Contains(iss.Message, "32769 bytes") || !strings.Contains(iss.Message, "32768") {
				t.Errorf("message should give the total and the limit:\n%s", iss.Message)
			}
		})
	}
}

func TestInspectContext_CodexOnePerDirectory(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "override replaces AGENTS.md",
			files: map[string]string{"AGENTS.override.md": "override\n", "AGENTS.md": "base\n"},
			want:  []string{"AGENTS.override.md"},
		},
		{
			name:  "empty override falls through",
			files: map[string]string{"AGENTS.override.md": " \n\n", "AGENTS.md": "base\n"},
			want:  []string{"AGENTS.md"},
		},
		{
			name:  "empty AGENTS.md is skipped",
			files: map[string]string{"AGENTS.md": ""},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			files := map[string]string{filepath.Join(root, ".git", "HEAD"): "x"}
			for name, body := range tt.files {
				files[filepath.Join(root, name)] = body
			}
			writeFiles(t, files)
			report, err := inspectContext(root, []string{"codex"}, root)
			if err != nil {
				t.Fatal(err)
			}
			got := []string{}
			for _, f := range toolByName(t, report, "codex").Files {
				got = append(got, relTo(t, root, f.Path))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("codex files = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInspectContext_ClaudeImports(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		want      []string
		wantNotes []string
	}{
		{
			name: "local file follows CLAUDE.md in the same directory",
			files: map[string]string{
				"CLAUDE.local.md": "mine\n",
				"CLAUDE.md":       "team\n",
			},
			want: []string{"CLAUDE.md", "CLAUDE.local.md"},
		},
		{
			name: "relative to the importing file",
			files: map[string]string{
				"CLAUDE.md":       "@docs/a.md\n",
				"docs/a.md":       "@b.md\n",
				"docs/b.md":       "leaf\n",
				"b.md":            "wrong b\n",
				"docs/unused.md":  "not imported\n",
				"CLAUDE.local.md": "@docs/b.md\n",
			},
			// docs/b.md is listed once, at its first load.
			want: []string{"CLAUDE.md", "docs/a.md", "docs/b.md", "CLAUDE.local.md"},
		},
		{
			name: "four hops and no further",
			files: map[string]string{
				"CLAUDE.md": "@h1.md\n",
				"h1.md":     "@h2.md\n",
				"h2.md":     "@h3.md\n",
				"h3.md":     "@h4.md\n",
				"h4.md":     "@h5.md\n",
				"h5.md":     "fifth hop\n",
			},
			want: []string{"CLAUDE.md", "h1.md", "h2.md", "h3.md", "h4.md"},
		},
		{
			name: "code spans, fences, emails, and missing files are not imports",
			files: map[string]string{
				"CLAUDE.md": "Run `cat @span.md now` and ``x @span.md ``.\n" +
					"```\n@fenced.md\n```\n" +
					"~~~~md\n@tilde.md\n~~~~\n" +
					"mail me@real.md and @missing.md and @dir and @real.md/x\n" +
					"@real.md\n",
				"span.md":   "s\n",
				"fenced.md": "f\n",
				"tilde.md":  "t\n",
				"real.md":   "r\n",
				"dir/x.md":  "x\n",
			},
			want: []string{"CLAUDE.md", "real.md"},
		},
		{
			name: "escaped spaces, absolute paths, and cycles",
			files: map[string]string{
				"CLAUDE.md":          `@Design\ Docs/api.md` + "\n@ABS\n",
				"Design Docs/api.md": "@../CLAUDE.md\n",
				"abs.md":             "absolute\n",
			},
			want: []string{"CLAUDE.md", "Design Docs/api.md", "abs.md"},
		},
		{
			name: "home imports are named, not read",
			files: map[string]string{
				"CLAUDE.md": "intro\n@~/.claude/mine.md\n",
			},
			want:      []string{"CLAUDE.md"},
			wantNotes: []string{"@~/.claude/mine.md in CLAUDE.md:2 not followed"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			files := map[string]string{}
			for name, body := range tt.files {
				files[filepath.Join(root, filepath.FromSlash(name))] = strings.ReplaceAll(body, "@ABS", "@"+filepath.Join(root, "abs.md"))
			}
			writeFiles(t, files)
			report, err := inspectContext(root, []string{"claude"}, root)
			if err != nil {
				t.Fatal(err)
			}
			tc := toolByName(t, report, "claude")
			got := []string{}
			for _, f := range tc.Files {
				got = append(got, relTo(t, root, f.Path))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("claude files = %q, want %q", got, tt.want)
			}
			notes := strings.Join(tc.Notes, "\n")
			for _, want := range tt.wantNotes {
				if !strings.Contains(notes, want) {
					t.Errorf("notes missing %q:\n%s", want, notes)
				}
			}
		})
	}
}

func TestInspectContext_ExternalImportNote(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	// The parent's @AGENTS.md resolves outside the child directory.
	notes := strings.Join(toolByName(t, report, "claude").Notes, "\n")
	if !strings.Contains(notes, "1 import(s) resolve outside this directory") {
		t.Errorf("notes should count the external import:\n%s", notes)
	}
}

func TestInspectContext_NoRepository(t *testing.T) {
	parent, _ := contextTree(t)
	report, err := inspectContext(parent, []string{"claude", "codex"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	if report.RepoRoot != "" {
		t.Fatalf("RepoRoot = %q, want none", report.RepoRoot)
	}
	for _, tc := range report.Tools {
		for _, f := range tc.Files {
			if f.OutsideRepo {
				t.Errorf("%s: %s flagged outside a repository that does not exist", tc.Tool, f.Path)
			}
		}
	}
	// Outside a repository Codex checks only the current directory.
	if got := rows(t, parent, toolByName(t, report, "codex").Files); len(got) != 1 || got[0].Path != "AGENTS.md" {
		t.Errorf("codex rows = %+v, want just AGENTS.md", got)
	}
	if got := issuesWithCode(report, CodeContextOutsideRepo); len(got) != 0 {
		t.Errorf("outside-repo issues without a repository: %+v", got)
	}
}

func TestInspectContext_GitFileMarksWorktreeRoot(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(root, "main", ".claude", "worktrees", "wt")
	writeFiles(t, map[string]string{
		filepath.Join(root, "main", ".git", "HEAD"): "x",
		filepath.Join(wt, ".git"):                   "gitdir: ../../../.git/worktrees/wt\n",
	})
	report, err := inspectContext(wt, nil, root)
	if err != nil {
		t.Fatal(err)
	}
	if report.RepoRoot != wt {
		t.Errorf("RepoRoot = %q, want the worktree %q", report.RepoRoot, wt)
	}
}

// TestInspectContext_LinkedWorktreeStopsClaudeWalk: Claude Code's
// upward walk stops at a repository root whose `.git` is a file (a
// linked worktree) but crosses one whose `.git` is a directory
// (measured 2026-11-27 with /memory).
func TestInspectContext_LinkedWorktreeStopsClaudeWalk(t *testing.T) {
	tests := []struct {
		name        string
		gitPath     string
		wantOutside []string
	}{
		{name: ".git file", gitPath: ".git", wantOutside: nil},
		{name: ".git directory", gitPath: filepath.Join(".git", "HEAD"), wantOutside: []string{"CLAUDE.md", "AGENTS.md"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			parent := filepath.Join(root, "parent")
			child := filepath.Join(parent, "child")
			writeFiles(t, map[string]string{
				filepath.Join(parent, "CLAUDE.md"): "@AGENTS.md\n",
				filepath.Join(parent, "AGENTS.md"): "Parent rule.\n",
				filepath.Join(child, tt.gitPath):   "gitdir: ../.git/worktrees/child\n",
				filepath.Join(child, "CLAUDE.md"):  "Child rule.\n",
			})
			report, err := inspectContext(child, []string{"claude"}, root)
			if err != nil {
				t.Fatal(err)
			}
			if report.RepoRoot != child {
				t.Fatalf("RepoRoot = %q, want %q", report.RepoRoot, child)
			}
			var outside []string
			for _, f := range toolByName(t, report, "claude").Files {
				if f.OutsideRepo {
					outside = append(outside, relTo(t, parent, f.Path))
				}
			}
			if !reflect.DeepEqual(outside, tt.wantOutside) {
				t.Errorf("outside rows = %q, want %q", outside, tt.wantOutside)
			}
			if n := len(issuesWithCode(report, CodeContextOutsideRepo)); n != len(tt.wantOutside)/2 {
				t.Errorf("outside-repo issues = %d, want %d", n, len(tt.wantOutside)/2)
			}
		})
	}
}

func TestInspectContext_UnknownTool(t *testing.T) {
	_, err := inspectContext(t.TempDir(), []string{"windsurf"}, os.TempDir())
	if !errors.Is(err, ErrUnknownContextTool) {
		t.Fatalf("err = %v, want ErrUnknownContextTool", err)
	}
}

// TestContextToolRules_CoverEveryGenerator pins the rule tables to the
// tools `aikata generate` knows, so a new tool cannot ship without
// saying what it loads.
func TestContextToolRules_CoverEveryGenerator(t *testing.T) {
	got := make([]string, 0, len(contextToolRules))
	for name := range contextToolRules {
		got = append(got, name)
	}
	sort.Strings(got)
	if want := generate.KnownTools(); !reflect.DeepEqual(got, want) {
		t.Errorf("contextToolRules keys = %q, want generate.KnownTools() %q", got, want)
	}
}

func TestParagraphs_SkipsMetadataAndStructure(t *testing.T) {
	body := "---\nproject: x\nstatus: draft\n---\n\n# Title\n\n---\n\nFirst rule\nwraps here.  \r\n\n## Heading\nwith body\n" +
		"\n@AGENTS.md @docs/x.md\n\nSee @AGENTS.md first.\n"
	got := paragraphs([]byte(body))
	want := []contextParagraph{
		{text: "First rule\nwraps here.", line: 10},
		{text: "## Heading\nwith body", line: 13},
		{text: "See @AGENTS.md first.", line: 18},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paragraphs = %+v, want %+v", got, want)
	}
}

func TestExcludeGlob_EscapesMetacharacters(t *testing.T) {
	path := filepath.Join(string(filepath.Separator)+"work", "a[1]", "b*{x}", "CLAUDE.md")
	got := excludeGlob(path)
	want := `/work/a\[1\]/b\*\{x\}/CLAUDE.md`
	if filepath.Separator == '\\' {
		want = `\work/a\[1\]/b\*\{x\}/CLAUDE.md`
	}
	if got != want {
		t.Errorf("excludeGlob = %q, want %q", got, want)
	}
}

type decodedContextReport struct {
	Version        int    `json:"version"`
	Kind           string `json:"kind"`
	Directory      string `json:"directory"`
	RepositoryRoot string `json:"repository_root"`
	Tools          []struct {
		Tool  string `json:"tool"`
		Files []struct {
			Path        string `json:"path"`
			Bytes       int    `json:"bytes"`
			OutsideRepo bool   `json:"outside_repository"`
			ImportedBy  string `json:"imported_by"`
		} `json:"files"`
		TotalBytes int      `json:"total_bytes"`
		Notes      []string `json:"notes"`
	} `json:"tools"`
	Issues  []decodedIssue `json:"issues"`
	Summary decodedSummary `json:"summary"`
}

func TestFormatContextJSON_Shape(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude", "cursor"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := FormatContextJSON(&buf, report); err != nil {
		t.Fatal(err)
	}
	var got decodedContextReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if got.Version != ContextJSONSchemaVersion || got.Kind != "context" || got.Directory != child || got.RepositoryRoot != child {
		t.Errorf("envelope = %+v", got)
	}
	if len(got.Tools) != 2 || got.Tools[0].Tool != "claude" || got.Tools[1].Tool != "cursor" {
		t.Fatalf("tools = %+v, want claude then cursor in config order", got.Tools)
	}
	claude := got.Tools[0]
	if len(claude.Files) != 4 {
		t.Fatalf("claude files = %+v", claude.Files)
	}
	imported := claude.Files[1]
	if imported.Path != "../AGENTS.md" || imported.ImportedBy != "../CLAUDE.md" || !imported.OutsideRepo {
		t.Errorf("imported row = %+v", imported)
	}
	sum := 0
	for _, f := range claude.Files {
		sum += f.Bytes
	}
	if claude.TotalBytes != sum {
		t.Errorf("total_bytes = %d, want %d", claude.TotalBytes, sum)
	}
	if got.Tools[1].Files == nil || got.Tools[1].Notes == nil {
		t.Errorf("files and notes must be arrays, not null: %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"imported_by"`) || strings.Count(buf.String(), `"imported_by"`) != 2 {
		t.Errorf("imported_by should appear only on the 2 imported rows:\n%s", buf.String())
	}
	if got.Summary.Info != len(got.Issues) || got.Summary.Warnings != 0 {
		t.Errorf("summary = %+v for %d info issues", got.Summary, len(got.Issues))
	}
}

func TestFormatContext_Text(t *testing.T) {
	parent, child := contextTree(t)
	report, err := inspectContext(child, []string{"claude", "codex", "cursor"}, filepath.Dir(parent))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := FormatContext(&buf, report); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"repository: " + child,
		"claude: 4 file(s)",
		"  ../CLAUDE.md",
		"    ../AGENTS.md", // imports indent under their importer
		"outside repository",
		"codex: 1 file(s)",
		"cursor: no files listed",
		"../CLAUDE.md: [info]",
		"AGENTS.md:5: [info]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q:\n%s", want, out)
		}
	}
}
