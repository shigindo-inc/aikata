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
	"github.com/shigindo-inc/aikata/internal/docmeta"
)

// frontmatterKeys are the keys every markdown file must declare. memory
// files declare an extra `memory_type` key which is verified by
// checkMemory rather than here.
var frontmatterKeys = []string{"project", "status", "version", "updated", "audience"}

// walkMarkdown invokes fn for every regular *.md file under
// opts.TargetDir, returning the slash-separated path relative to
// TargetDir. Skipped directories (docmeta.SkipDir) and generated
// artifact files are excluded, as are paths matching any user-configured
// opts.Excludes glob (see ADR 0021).
func walkMarkdown(opts Options, fn func(rel string, body []byte) error) error {
	return filepath.WalkDir(opts.TargetDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if docmeta.SkipDir(d.Name(), p, opts.TargetDir) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		if _, skip := docmeta.DefaultSkipFiles[d.Name()]; skip {
			return nil
		}
		rel, err := filepath.Rel(opts.TargetDir, p)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		// Managed-surface scope (ADR 0037 D1): when an include set is
		// configured, only validate Markdown the include globs match.
		// An empty include set is the broad-audit walk (--all-markdown).
		if len(opts.Includes) > 0 && !MatchAny(opts.Includes, relSlash) {
			return nil
		}
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

// checkFrontmatter asserts every .md file contains the required
// frontmatter keys.
func checkFrontmatter(opts Options) ([]Issue, error) {
	var issues []Issue
	err := walkMarkdown(opts, func(rel string, body []byte) error {
		lines, _ := docmeta.ParseFrontmatter(body)
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
			if _, _, ok := docmeta.FrontmatterValue(lines, key); !ok {
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

// checkLinks verifies that every relative markdown link in AGENTS.md
// resolves to an existing file under TargetDir. Link extraction is
// shared with `aikata map` via internal/docmeta so the two cannot drift.
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
	for _, link := range docmeta.ExtractLinks(body) {
		// AGENTS.md lives in the project root, so links resolve relative
		// to TargetDir.
		full := filepath.Join(opts.TargetDir, filepath.FromSlash(link.Target))
		if _, err := os.Stat(full); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				issues = append(issues, Issue{
					Level: LevelError, File: "AGENTS.md", Line: link.Line,
					Message: fmt.Sprintf("broken link: %s", link.Raw),
				})
			} else {
				return nil, err
			}
		}
	}
	return issues, nil
}

// adrStatusRE captures the value after `**Status**:` (or the Japanese
// `**状態**:`, with either an ASCII or a full-width colon) on a single ADR
// metadata line. The ja templates keep the English key, but ADRs written
// by hand in Japanese projects use the localized one.
var adrStatusRE = regexp.MustCompile(`(?i)\*\*(?:Status|状態)\*\*\s*[:：]\s*([^\n]+)`)

// adrStatus returns the ADR's decision status. The body metadata line wins
// because aikata's own templates use the frontmatter `status` for the
// document lifecycle (draft / stable) and the body line for the decision;
// projects that keep only the frontmatter key fall back to it.
func adrStatus(body []byte) (string, bool) {
	if m := adrStatusRE.FindSubmatch(body); m != nil {
		return strings.TrimSpace(string(m[1])), true
	}
	lines, _ := docmeta.ParseFrontmatter(body)
	if v, _, ok := docmeta.FrontmatterValue(lines, "status"); ok && v != "" {
		return v, true
	}
	return "", false
}

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
		status, ok := adrStatus(body)
		if !ok {
			issues = append(issues, Issue{
				Level: LevelWarning, File: rel,
				Code:    "adr.missing-status",
				Message: "ADR has no status (expected a **Status** line or frontmatter status)",
			})
			continue
		}
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

// checkADRNumbering reports duplicate numbers as errors and missing slots
// in the 0001..max range as info. A duplicate makes "ADR NNNN" ambiguous in
// every reference to it, so it must be renumbered; a gap is usually a
// retired number and needs no action (ADR 0049).
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
			Level: LevelError, File: rel,
			Code:    "adr.duplicate-number",
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
		lines, _ := docmeta.ParseFrontmatter(body)
		mt, _, ok := docmeta.FrontmatterValue(lines, "memory_type")
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
// one year and rejects dates outside the YYYY-MM-DD format.
func checkUpdated(opts Options) ([]Issue, error) {
	cutoff := opts.Now.AddDate(-1, 0, 0)
	var issues []Issue
	err := walkMarkdown(opts, func(rel string, body []byte) error {
		lines, _ := docmeta.ParseFrontmatter(body)
		if lines == nil {
			return nil
		}
		raw, lineNum, ok := docmeta.FrontmatterValue(lines, "updated")
		if !ok {
			return nil
		}
		t, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
		if err != nil {
			issues = append(issues, Issue{
				Level: LevelError, File: rel, Line: lineNum,
				Code:    "updated.invalid",
				Message: "updated must be a calendar date in YYYY-MM-DD format",
			})
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

// checkGlossary reports glossary terms that no other markdown file uses.
// It runs only with Options.Verbose: an unreferenced term is usually core
// vocabulary, and reporting each one by default buried the findings that
// need action (ADR 0049).
func checkGlossary(opts Options) ([]Issue, error) {
	if !opts.Verbose {
		return nil, nil
	}
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
