package doctor

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shigindo-inc/aikata/internal/adr"
)

// frontmatterKeys are the keys every markdown file must declare. memory
// files declare an extra `memory_type` key which is verified by
// checkMemory rather than here.
var frontmatterKeys = []string{"project", "status", "version", "updated", "audience"}

// skippedDirs lists directories doctor will not descend into. They are
// either external (vendored deps, build outputs) or aikata-internal
// scratch areas (testdata fixtures, ephemeral memory).
var skippedDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "build": {}, "dist": {},
	".dart_tool": {}, ".next": {}, "vendor": {}, ".turbo": {},
	".cursor": {}, ".github": {}, "testdata": {}, ".remember": {},
	".serena": {}, ".aikata-proposed": {},
}

// skippedFiles names individual *.md files doctor should not inspect.
// These are generated AI-tool artifacts (no frontmatter by design) or
// project-level boilerplate that lives outside the doctor contract.
var skippedFiles = map[string]struct{}{
	"CLAUDE.md": {}, "GEMINI.md": {},
}

// walkMarkdown invokes fn for every regular *.md file under
// opts.TargetDir, returning the slash-separated path relative to
// TargetDir. Skipped directories and generated artifact files are
// excluded, as are paths matching any user-configured
// opts.Excludes glob (see ADR 0021).
func walkMarkdown(opts Options, fn func(rel string, body []byte) error) error {
	return filepath.WalkDir(opts.TargetDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if _, skip := skippedDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		if _, skip := skippedFiles[d.Name()]; skip {
			return nil
		}
		rel, err := filepath.Rel(opts.TargetDir, p)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if MatchAny(opts.Excludes, relSlash) {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return fn(relSlash, body)
	})
}

// parseFrontmatter extracts the YAML front-matter block delimited by
// `---` lines at the very start of body. Returns the lines (without
// the fences) and the byte offset where the body proper begins. If
// the file has no front-matter, lines is nil and offset is 0.
func parseFrontmatter(body []byte) (lines []string, offset int) {
	text := string(body)
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return nil, 0
	}
	// Find the closing `---` line.
	rest := text[len("---\n"):]
	if strings.HasPrefix(text, "---\r\n") {
		rest = text[len("---\r\n"):]
	}
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		// Also accept `\n---` at EOF.
		if i := strings.Index(rest, "\n---"); i >= 0 && i+len("\n---") == len(rest) {
			end = i
		} else {
			return nil, 0
		}
	}
	fm := rest[:end]
	for _, ln := range strings.Split(fm, "\n") {
		lines = append(lines, strings.TrimRight(ln, "\r"))
	}
	// Offset points just past the closing fence.
	closer := "\n---\n"
	if !strings.Contains(rest, closer) {
		closer = "\n---"
	}
	offset = len("---\n") + end + len(closer)
	return lines, offset
}

// frontmatterValue returns the value associated with key in the parsed
// front-matter lines. Only top-level scalar keys are supported (which
// is all aikata uses for required keys).
func frontmatterValue(lines []string, key string) (value string, line int, ok bool) {
	prefix := key + ":"
	for i, ln := range lines {
		trimmed := strings.TrimLeft(ln, " \t")
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		v := strings.TrimSpace(trimmed[len(prefix):])
		v = strings.Trim(v, `"'`)
		return v, i + 2, true // +1 frontmatter offset, +1 1-based
	}
	return "", 0, false
}

// checkFrontmatter asserts every .md file contains the required
// frontmatter keys.
func checkFrontmatter(opts Options) ([]Issue, error) {
	var issues []Issue
	err := walkMarkdown(opts, func(rel string, body []byte) error {
		lines, _ := parseFrontmatter(body)
		if lines == nil {
			issues = append(issues, Issue{
				Level: LevelError, File: rel, Line: 1,
				Code: "frontmatter.missing",
				Message: "missing YAML frontmatter (required keys: " +
					strings.Join(frontmatterKeys, ", ") + ")",
			})
			return nil
		}
		for _, key := range frontmatterKeys {
			if _, _, ok := frontmatterValue(lines, key); !ok {
				issues = append(issues, Issue{
					Level: LevelError, File: rel,
					Code:    "frontmatter.missing-key." + key,
					Message: fmt.Sprintf("frontmatter missing required key %q", key),
				})
			}
		}
		return nil
	})
	return issues, err
}

// linkRE matches markdown links of the form [text](path) where path is
// not an absolute URL. The relative paths it captures are subject to
// the link-existence check.
var linkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// checkLinks verifies that every relative markdown link in AGENTS.md
// resolves to an existing file under TargetDir.
func checkLinks(opts Options) ([]Issue, error) {
	agentsPath := filepath.Join(opts.TargetDir, "AGENTS.md")
	body, err := os.ReadFile(agentsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return []Issue{{
			Level: LevelError, File: "AGENTS.md",
			Message: "AGENTS.md not found in project root",
		}}, nil
	}
	if err != nil {
		return nil, err
	}

	var issues []Issue
	lineNum := 0
	for _, line := range strings.Split(string(body), "\n") {
		lineNum++
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			target := m[1]
			// Strip URL fragment and query.
			if i := strings.IndexAny(target, "#?"); i >= 0 {
				target = target[:i]
			}
			target = strings.TrimSpace(target)
			if target == "" {
				continue
			}
			if isAbsoluteURL(target) || strings.HasPrefix(target, "mailto:") {
				continue
			}
			// Resolve relative to AGENTS.md's directory.
			rel := target
			if strings.HasPrefix(target, "./") {
				rel = target[2:]
			}
			full := filepath.Join(opts.TargetDir, filepath.FromSlash(rel))
			if _, err := os.Stat(full); err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					issues = append(issues, Issue{
						Level: LevelError, File: "AGENTS.md", Line: lineNum,
						Message: fmt.Sprintf("broken link: %s", m[1]),
					})
				} else {
					return nil, err
				}
			}
		}
	}
	return issues, nil
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// adrStatusRE captures the value after `**Status**:` on a single ADR
// metadata line.
var adrStatusRE = regexp.MustCompile(`(?i)\*\*Status\*\*:\s*([^\n]+)`)

// checkADR verifies Deprecated ADRs reference their replacement.
func checkADR(opts Options) ([]Issue, error) {
	adrDir := filepath.Join(opts.TargetDir, "docs", "adr")
	if _, err := os.Stat(adrDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var issues []Issue
	entries, err := os.ReadDir(adrDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rel := filepath.ToSlash(filepath.Join("docs", "adr", e.Name()))
		body, err := os.ReadFile(filepath.Join(adrDir, e.Name()))
		if err != nil {
			return nil, err
		}
		m := adrStatusRE.FindStringSubmatch(string(body))
		if m == nil {
			issues = append(issues, Issue{
				Level: LevelWarning, File: rel,
				Message: "ADR missing **Status** line",
			})
			continue
		}
		status := strings.TrimSpace(m[1])
		if strings.HasPrefix(strings.ToLower(status), "deprecated") {
			if !strings.Contains(string(body), "Replaced by") &&
				!strings.Contains(string(body), "Superseded by") {
				issues = append(issues, Issue{
					Level: LevelError, File: rel,
					Message: "Deprecated ADR must reference Replaced by / Superseded by ADR-NNNN",
				})
			}
		}
	}
	return issues, nil
}

// checkADRNumbering reports duplicate numbers and missing slots in the
// 0001..max range. Findings are LevelInfo because the project may
// intentionally retire a number; the advisory only flags them so a
// human (or `aikata add adr` later on) can decide what to do.
func checkADRNumbering(opts Options) ([]Issue, error) {
	adrDir := filepath.Join(opts.TargetDir, "docs", "adr")
	entries, err := adr.Scan(adrDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	byNumber := make(map[int][]string, len(entries))
	for _, e := range entries {
		byNumber[e.Number] = append(byNumber[e.Number], e.Filename)
	}
	var issues []Issue
	for _, e := range entries {
		if len(byNumber[e.Number]) <= 1 {
			continue
		}
		rel := filepath.ToSlash(filepath.Join("docs", "adr", e.Filename))
		issues = append(issues, Issue{
			Level: LevelInfo, File: rel,
			Message: fmt.Sprintf("duplicate ADR number %04d", e.Number),
		})
	}
	maxNum := entries[len(entries)-1].Number
	dirRel := filepath.ToSlash(filepath.Join("docs", "adr"))
	for n := 1; n <= maxNum; n++ {
		if _, ok := byNumber[n]; ok {
			continue
		}
		issues = append(issues, Issue{
			Level: LevelInfo, File: dirRel,
			Message: fmt.Sprintf("ADR number %04d is unused (gap below %04d)", n, maxNum),
		})
	}
	return issues, nil
}

// checkMemory verifies docs/memory/<type>.md frontmatter's
// memory_type matches the file name.
func checkMemory(opts Options) ([]Issue, error) {
	memDir := filepath.Join(opts.TargetDir, "docs", "memory")
	if _, err := os.Stat(memDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	validTypes := map[string]struct{}{
		"user": {}, "feedback": {}, "project": {}, "reference": {},
	}
	var issues []Issue
	entries, err := os.ReadDir(memDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), ".md")
		if stem == "README" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join("docs", "memory", e.Name()))
		body, err := os.ReadFile(filepath.Join(memDir, e.Name()))
		if err != nil {
			return nil, err
		}
		lines, _ := parseFrontmatter(body)
		mt, _, ok := frontmatterValue(lines, "memory_type")
		if !ok {
			issues = append(issues, Issue{
				Level: LevelError, File: rel,
				Message: "memory file missing frontmatter key 'memory_type'",
			})
			continue
		}
		if _, valid := validTypes[mt]; !valid {
			issues = append(issues, Issue{
				Level: LevelError, File: rel,
				Message: fmt.Sprintf("memory_type %q must be one of user/feedback/project/reference", mt),
			})
			continue
		}
		if mt != stem {
			issues = append(issues, Issue{
				Level: LevelError, File: rel,
				Message: fmt.Sprintf("memory_type %q does not match filename %q", mt, stem),
			})
		}
	}
	return issues, nil
}

// checkUpdated warns when a frontmatter `updated:` field is older than
// 365 days. Only ISO 8601 (YYYY-MM-DD) values are recognised; other
// formats are skipped without an error.
func checkUpdated(opts Options) ([]Issue, error) {
	cutoff := opts.Now.AddDate(-1, 0, 0)
	var issues []Issue
	err := walkMarkdown(opts, func(rel string, body []byte) error {
		lines, _ := parseFrontmatter(body)
		if lines == nil {
			return nil
		}
		raw, lineNum, ok := frontmatterValue(lines, "updated")
		if !ok {
			return nil
		}
		t, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
		if err != nil {
			return nil
		}
		if t.Before(cutoff) {
			issues = append(issues, Issue{
				Level: LevelWarning, File: rel, Line: lineNum,
				Code:    "updated.stale",
				Message: fmt.Sprintf("updated %s is more than 365 days old", raw),
			})
		}
		return nil
	})
	return issues, err
}

// envVarRE matches a top-level VAR=value entry in a .env-style file
// (allowing optional `export ` prefix).
var envVarRE = regexp.MustCompile(`^(?:export\s+)?([A-Z][A-Z0-9_]*)=`)

// checkEnvExample warns when a variable declared in .env.example is
// not referenced by AGENTS.md or ARCHITECTURE.md.
func checkEnvExample(opts Options) ([]Issue, error) {
	envPath := filepath.Join(opts.TargetDir, ".env.example")
	body, err := os.ReadFile(envPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var docs strings.Builder
	for _, name := range []string{"AGENTS.md", "ARCHITECTURE.md"} {
		b, err := os.ReadFile(filepath.Join(opts.TargetDir, name))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		docs.Write(b)
		docs.WriteByte('\n')
	}
	corpus := docs.String()

	var issues []Issue
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := envVarRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name := m[1]
		if !strings.Contains(corpus, name) {
			issues = append(issues, Issue{
				Level: LevelWarning, File: ".env.example", Line: lineNum,
				Message: fmt.Sprintf("variable %q is not mentioned in AGENTS.md or ARCHITECTURE.md", name),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return issues, nil
}

// glossaryHeadingRE matches the `### Term` headings under each letter
// section of a GLOSSARY.md file.
var glossaryHeadingRE = regexp.MustCompile(`(?m)^###\s+(.+?)\s*$`)

// checkGlossary warns when a glossary term is not used by any other
// markdown file in the project.
func checkGlossary(opts Options) ([]Issue, error) {
	glossaryPath := filepath.Join(opts.TargetDir, "GLOSSARY.md")
	body, err := os.ReadFile(glossaryPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	matches := glossaryHeadingRE.FindAllStringSubmatchIndex(string(body), -1)
	type entry struct {
		Term string
		Line int
	}
	var terms []entry
	for _, m := range matches {
		text := string(body)[m[2]:m[3]]
		// Drop trailing dash-followed-by-translation if present
		// ("Term — translation" → "Term").
		if i := strings.Index(text, " —"); i > 0 {
			text = text[:i]
		}
		text = strings.TrimSpace(text)
		if text == "" || strings.HasPrefix(text, "(") {
			continue
		}
		line := 1 + strings.Count(string(body)[:m[2]], "\n")
		terms = append(terms, entry{Term: text, Line: line})
	}

	// Build the corpus of non-glossary markdown contents.
	var corpus strings.Builder
	err = walkMarkdown(opts, func(rel string, body []byte) error {
		if rel == "GLOSSARY.md" {
			return nil
		}
		corpus.Write(body)
		corpus.WriteByte('\n')
		return nil
	})
	if err != nil {
		return nil, err
	}
	text := corpus.String()

	var issues []Issue
	for _, t := range terms {
		if !strings.Contains(text, t.Term) {
			issues = append(issues, Issue{
				Level: LevelInfo, File: "GLOSSARY.md", Line: t.Line,
				Message: fmt.Sprintf("term %q is not referenced from any other markdown file", t.Term),
			})
		}
	}
	return issues, nil
}
