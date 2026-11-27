package doctor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// Issue codes emitted by InspectContext (ADR 0051).
const (
	CodeContextOutsideRepo        = "context.outside-repo"
	CodeContextDuplicateParagraph = "context.duplicate-paragraph"
	CodeContextCodexOverLimit     = "context.codex-over-limit"
)

// ErrUnknownContextTool is returned when InspectContext is asked about
// an AI tool that has no loading rules in contextToolRules.
var ErrUnknownContextTool = errors.New("doctor: no context loading rules for ai_tool")

// ContextFile is one instruction file an AI tool loads at session start.
type ContextFile struct {
	// Path is absolute and cleaned.
	Path string
	// Bytes is the file size on disk. aikata counts bytes, not tokens
	// (SPEC §8).
	Bytes int
	// OutsideRepo is true when the repository root is known and Path
	// lies outside it.
	OutsideRepo bool
	// ImportedBy is the absolute path of the file whose `@path` import
	// loaded this one. Empty for files the tool finds by name.
	ImportedBy string
}

// ToolContext is the load list of one AI tool, in load order: each
// file comes before the files it imports.
type ToolContext struct {
	Tool  string
	Files []ContextFile
	// Notes state what the list leaves out or assumes.
	Notes []string
}

// ContextReport is the result of InspectContext.
type ContextReport struct {
	// Dir is the inspected working directory, absolute with symlinks
	// resolved. Issue.File paths are relative to it.
	Dir string
	// RepoRoot is the nearest directory at or above Dir that holds a
	// `.git` file or directory; empty when there is none.
	RepoRoot string
	Tools    []ToolContext
	Issues   []Issue
	// contents holds each listed file's bytes as read, so the issue
	// checks work on the same data the rows describe.
	contents map[string][]byte
}

// dirWalk selects which directories a tool searches for instruction
// files. The zero value searches none.
type dirWalk int

const (
	walkNone dirWalk = iota
	// walkFromFilesystemRoot visits every directory from the
	// filesystem root down to the working directory, ignoring git
	// boundaries unless contextRules.stopAtLinkedWorktree applies.
	walkFromFilesystemRoot
	// walkFromRepoRoot visits the repository root down to the working
	// directory, or only the working directory outside a repository.
	walkFromRepoRoot
)

// contextRules describes how one AI tool picks the instruction files
// it loads at session start. Each tool has exactly one value in
// contextToolRules; every field cites the vendor document it encodes.
type contextRules struct {
	walk dirWalk
	// stopAtLinkedWorktree ends a walkFromFilesystemRoot walk at the
	// repository root when that root is a linked worktree (its `.git` is
	// a file).
	stopAtLinkedWorktree bool
	// names are looked up in each visited directory, in load order.
	names []string
	// firstPerDir keeps at most one file per directory: the first
	// name found (after skipEmpty).
	firstPerDir bool
	// skipEmpty ignores files that hold only whitespace.
	skipEmpty bool
	// importHops is how many `@path` import hops are followed from a
	// file found by name; 0 disables import parsing.
	importHops int
	// inputCap is the combined input size past which the tool drops
	// the rest; nil means no cap.
	inputCap *inputCap
	// excludeSetting is the tool's setting that skips a file by
	// absolute-path glob; empty when the tool has none.
	excludeSetting string
	notes          []string
}

// inputCap is a tool's limit on combined instruction bytes.
type inputCap struct {
	bytes int
	// setting names the tool option that changes the limit.
	setting string
	// code is the Issue.Code reported when the files pass the limit.
	code string
}

// claudeCodeRules: https://code.claude.com/docs/en/memory (read
// 2026-11-27). The walk and the import were also measured on
// 2026-11-27 with `/memory` in a child repository nested inside a
// parent directory: the parent's CLAUDE.md and its @-imported AGENTS.md
// loaded, then the child's CLAUDE.md and @AGENTS.md.
var claudeCodeRules = contextRules{
	// "Claude Code loads CLAUDE.md and CLAUDE.local.md from your current
	// working directory and every directory above it." / "content is
	// ordered from the filesystem root down to your working directory."
	// Measured: the walk crosses the git boundary.
	walk: walkFromFilesystemRoot,
	// Measured 2026-11-27 with `/memory`: a session started in a linked
	// worktree under the main checkout's .claude/worktrees/ did not load
	// the main checkout's CLAUDE.md, while a nested repository whose
	// `.git` is a directory loaded its parent's. That the walk also skips
	// directories above the main checkout is an inference, as is the same
	// stop for other `.git` files such as submodules.
	stopAtLinkedWorktree: true,
	// "Within each directory, CLAUDE.local.md is appended after
	// CLAUDE.md."
	names: []string{"CLAUDE.md", "CLAUDE.local.md"},
	// "Relative paths resolve relative to the file containing the
	// import ... with a maximum depth of four hops."
	importHops: 4,
	// "The claudeMdExcludes setting lets you skip specific files by path
	// or glob pattern ... matched against absolute file paths."
	excludeSetting: "claudeMdExcludes",
	notes: []string{
		// The same page also loads these; they are not measured yet,
		// so the list leaves them out and says so (ADR 0051 確度).
		"not listed: .claude/CLAUDE.md, .claude/rules/*.md, and the AGENTS.md Claude Code reads when no CLAUDE.md exists",
	},
}

// codexRules: https://developers.openai.com/codex/guides/agents-md
// (served from learn.chatgpt.com/docs/agent-configuration/agents-md,
// read 2026-11-27).
var codexRules = contextRules{
	// "Starting at the project root (typically the Git root), Codex
	// walks down to your current working directory." / "If Codex cannot
	// find a project root, it only checks the current directory."
	// Inference: Codex never reads above the git root; the docs do not
	// say so.
	walk: walkFromRepoRoot,
	// "Codex checks each directory in this order: AGENTS.override.md,
	// AGENTS.md" / "Codex includes at most one file per directory."
	names:       []string{"AGENTS.override.md", "AGENTS.md"},
	firstPerDir: true,
	// "Codex skips empty files". Treating whitespace-only as empty is
	// an inference.
	skipEmpty: true,
	// "stops adding files once the combined size reaches the limit
	// defined by project_doc_max_bytes (32 KiB by default)."
	inputCap: &inputCap{bytes: 32 * 1024, setting: "project_doc_max_bytes", code: CodeContextCodexOverLimit},
	notes: []string{
		"that Codex never reads above the repository root is an inference; its docs do not say",
		"project_doc_fallback_filenames and project_doc_max_bytes from your Codex config are not read; defaults assumed",
	},
}

// cursorRules: Cursor does not document whether rules in directories
// above the opened workspace apply (ADR 0051), so nothing is listed.
var cursorRules = contextRules{
	walk: walkNone,
	notes: []string{
		"Cursor does not document whether rules above the workspace apply; no files listed",
	},
}

// contextToolRules maps every ai_tool name aikata knows to its loading
// rules. A test pins its keys to generate.KnownTools.
var contextToolRules = map[string]contextRules{
	"claude": claudeCodeRules,
	"codex":  codexRules,
	"cursor": cursorRules,
}

// InspectContext lists, for each tool in tools, the instruction files
// that tool loads at session start in dir, and reports the issues ADR
// 0051 defines: files loaded from outside the repository (info),
// paragraphs repeated verbatim across loaded files (info), and Codex
// input over project_doc_max_bytes (warning). It only reads files and
// never reads user-level configuration such as ~/.claude or ~/.codex.
func InspectContext(dir string, tools []string) (ContextReport, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ContextReport{}, fmt.Errorf("doctor: context: resolve %s: %w", dir, err)
	}
	return inspectContext(abs, tools, filesystemRoot(abs))
}

// inspectContext is InspectContext with the outermost directory of
// every upward walk made explicit, so tests can fence the walk inside
// a temporary directory.
func inspectContext(dir string, tools []string, top string) (ContextReport, error) {
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return ContextReport{}, fmt.Errorf("doctor: context: resolve %s: %w", dir, err)
	}
	top, err = filepath.EvalSymlinks(top)
	if err != nil {
		return ContextReport{}, fmt.Errorf("doctor: context: resolve %s: %w", top, err)
	}
	repo, err := findRepo(resolved, top)
	if err != nil {
		return ContextReport{}, err
	}
	report := ContextReport{Dir: resolved, RepoRoot: repo.root, Tools: []ToolContext{}, contents: map[string][]byte{}}
	for _, tool := range tools {
		rules, ok := contextToolRules[tool]
		if !ok {
			return ContextReport{}, fmt.Errorf("%w: %q", ErrUnknownContextTool, tool)
		}
		loaded, err := rules.load(tool, resolved, repo, top, report.contents)
		if err != nil {
			return ContextReport{}, fmt.Errorf("doctor: context: %s: %w", tool, err)
		}
		report.Tools = append(report.Tools, loaded)
	}
	report.Issues = contextIssues(report)
	return report, nil
}

// filesystemRoot returns the root of the volume that holds dir.
func filesystemRoot(dir string) string {
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// repository is the git repository that holds the working directory.
type repository struct {
	// root is empty when no directory up to the walk's top holds `.git`.
	root string
	// linkedWorktree is true when root's `.git` is a file, as in a
	// linked worktree (or a submodule), rather than a directory.
	linkedWorktree bool
}

// findRepo returns the nearest directory from dir up to top that holds
// a `.git` entry.
func findRepo(dir, top string) (repository, error) {
	dirs := ancestorsTopDown(dir, top)
	for i := len(dirs) - 1; i >= 0; i-- {
		info, err := os.Lstat(filepath.Join(dirs[i], ".git"))
		if err == nil {
			return repository{root: dirs[i], linkedWorktree: !info.IsDir()}, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return repository{}, fmt.Errorf("doctor: context: stat %s: %w", filepath.Join(dirs[i], ".git"), err)
		}
	}
	return repository{}, nil
}

// ancestorsTopDown returns top, the directories between, and dir, in
// that order. When dir is not under top it returns dir alone.
func ancestorsTopDown(dir, top string) []string {
	var up []string
	for d := dir; ; d = filepath.Dir(d) {
		up = append(up, d)
		if d == top {
			break
		}
		if filepath.Dir(d) == d {
			return []string{dir}
		}
	}
	for i, j := 0, len(up)-1; i < j; i, j = i+1, j-1 {
		up[i], up[j] = up[j], up[i]
	}
	return up
}

// dirs returns the directories r searches, outermost first.
func (r contextRules) dirs(dir string, repo repository, top string) []string {
	switch r.walk {
	case walkFromFilesystemRoot:
		if r.stopAtLinkedWorktree && repo.linkedWorktree {
			return ancestorsTopDown(dir, repo.root)
		}
		return ancestorsTopDown(dir, top)
	case walkFromRepoRoot:
		if repo.root == "" {
			return []string{dir}
		}
		return ancestorsTopDown(dir, repo.root)
	default:
		return nil
	}
}

// contextLoader accumulates one tool's load list.
type contextLoader struct {
	rules    contextRules
	dir      string
	repoRoot string
	seen     map[string]bool
	contents map[string][]byte
	out      ToolContext
	// external counts imports that resolve outside dir.
	external int
}

// load builds tool's load list under r, recording each file's bytes
// in contents.
func (r contextRules) load(tool, dir string, repo repository, top string, contents map[string][]byte) (ToolContext, error) {
	l := &contextLoader{rules: r, dir: dir, repoRoot: repo.root, seen: map[string]bool{}, contents: contents}
	l.out = ToolContext{Tool: tool, Files: []ContextFile{}, Notes: append([]string{}, r.notes...)}
	for _, d := range r.dirs(dir, repo, top) {
		for _, name := range r.names {
			path := filepath.Join(d, name)
			data, ok, err := readInstructionFile(path)
			if err != nil {
				return ToolContext{}, err
			}
			if !ok || (r.skipEmpty && strings.TrimSpace(string(data)) == "") {
				continue
			}
			if err := l.add(path, data, "", 0); err != nil {
				return ToolContext{}, err
			}
			if r.firstPerDir {
				break
			}
		}
	}
	if l.external > 0 {
		// code.claude.com/docs/en/memory: "An import in a project-level
		// memory file is external when its path resolves outside your
		// working directory ... If you decline, the imports stay
		// disabled." The list assumes they were approved.
		l.out.Notes = append(l.out.Notes, fmt.Sprintf(
			"%d import(s) resolve outside this directory; %s loads them only after you approve external imports for the project",
			l.external, tool))
	}
	return l.out, nil
}

// add appends path (already read as data) and, within the hop limit,
// the files it imports. A file already in the list is not added again.
func (l *contextLoader) add(path string, data []byte, importedBy string, hop int) error {
	if l.seen[path] {
		return nil
	}
	l.seen[path] = true
	l.contents[path] = data
	l.out.Files = append(l.out.Files, ContextFile{
		Path:        path,
		Bytes:       len(data),
		OutsideRepo: l.repoRoot != "" && !isWithin(l.repoRoot, path),
		ImportedBy:  importedBy,
	})
	if hop >= l.rules.importHops {
		return nil
	}
	for _, imp := range importPaths(string(data)) {
		if strings.HasPrefix(imp.path, "~") {
			l.out.Notes = append(l.out.Notes, fmt.Sprintf(
				"@%s in %s:%d not followed: user-level files are out of scope",
				imp.path, displayPath(l.dir, path), imp.line))
			continue
		}
		target := imp.path
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
		}
		target = filepath.Clean(target)
		child, ok, err := readInstructionFile(target)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if !l.seen[target] && !isWithin(l.dir, target) {
			l.external++
		}
		if err := l.add(target, child, path, hop+1); err != nil {
			return err
		}
	}
	return nil
}

// readInstructionFile reads path when it is a regular file (following
// symlinks). A path that cannot name a file (missing, a directory, or
// malformed, as free-text `@` mentions often are) yields ok=false; any
// other failure is returned so the report never silently drops a file
// the tool would load.
func readInstructionFile(path string) (data []byte, ok bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrInvalid) ||
			errors.Is(err, syscall.ENOTDIR) || errors.Is(err, syscall.ENAMETOOLONG) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, false, nil
	}
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	return data, true, nil
}

// isWithin reports whether path is root or lies below it.
func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// displayPath renders path relative to dir in slash form, the same
// convention Issue.File uses. A path on another volume stays absolute.
func displayPath(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// contextIssues derives the ADR 0051 findings from the load lists.
func contextIssues(report ContextReport) []Issue {
	issues := []Issue{}
	for _, tc := range report.Tools {
		rules := contextToolRules[tc.Tool]
		issues = append(issues, outsideRepoIssues(report.Dir, tc, rules)...)
		if iss, ok := overLimitIssue(report.Dir, tc, rules, report.contents); ok {
			issues = append(issues, iss)
		}
	}
	issues = append(issues, duplicateParagraphIssues(report.Dir, report.Tools, report.contents)...)
	sort.SliceStable(issues, func(i, j int) bool { return issues[i].Level > issues[j].Level })
	return issues
}

// outsideRepoIssues reports each file the tool found by name outside
// the repository. Files reached by import are covered by their
// importer: excluding the importer drops them too, and an import
// written inside the repository is deliberate.
func outsideRepoIssues(dir string, tc ToolContext, rules contextRules) []Issue {
	var issues []Issue
	for i, f := range tc.Files {
		if !f.OutsideRepo || f.ImportedBy != "" {
			continue
		}
		imports, total := 0, f.Bytes
		for _, g := range tc.Files[i+1:] {
			if g.ImportedBy == "" {
				break
			}
			imports++
			total += g.Bytes
		}
		msg := fmt.Sprintf("%s loads this file from outside the repository", tc.Tool)
		if imports > 0 {
			msg += fmt.Sprintf(" with %d import(s), %d bytes in all", imports, total)
		} else {
			msg += fmt.Sprintf(", %d bytes", total)
		}
		if rules.excludeSetting != "" {
			msg += fmt.Sprintf(". To skip it, add %q to %s in .claude/settings.local.json."+
				" Exclusion drops the whole file and its imports, so first move any rules"+
				" this repository needs into a file it imports",
				excludeGlob(f.Path), rules.excludeSetting)
		}
		issues = append(issues, Issue{
			Level:   LevelInfo,
			File:    displayPath(dir, f.Path),
			Message: msg,
			Code:    CodeContextOutsideRepo,
		})
	}
	return issues
}

// excludeGlob turns an absolute path into a glob that matches only that
// path: slash-separated, with glob metacharacters escaped.
func excludeGlob(path string) string {
	var b strings.Builder
	for _, r := range filepath.ToSlash(path) {
		if strings.ContainsRune(`*?[]{}()!`, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// overLimitIssue reports a tool whose combined input passes its
// inputCap, pointing at the line where the cap falls.
func overLimitIssue(dir string, tc ToolContext, rules contextRules, contents map[string][]byte) (Issue, bool) {
	limit := rules.inputCap
	if limit == nil {
		return Issue{}, false
	}
	total := 0
	for _, f := range tc.Files {
		total += f.Bytes
	}
	remaining := limit.bytes
	for _, f := range tc.Files {
		if f.Bytes <= remaining {
			remaining -= f.Bytes
			continue
		}
		line := 1 + strings.Count(string(contents[f.Path][:remaining]), "\n")
		return Issue{
			Level: LevelWarning,
			File:  displayPath(dir, f.Path),
			Line:  line,
			Code:  limit.code,
			Message: fmt.Sprintf(
				"%s instruction files total %d bytes, over the %d-byte %s default;"+
					" %s stops at the limit, so at least the text from this line on is dropped without notice",
				tc.Tool, total, limit.bytes, limit.setting, tc.Tool),
		}, true
	}
	return Issue{}, false
}
