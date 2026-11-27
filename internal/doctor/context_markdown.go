package doctor

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/shigindo-inc/aikata/internal/docmeta"
)

// contextImport is one `@path` import found in an instruction file.
type contextImport struct {
	path string
	line int
}

// importPattern follows code.claude.com/docs/en/memory: an `@` that
// starts the line or follows whitespace, then a path that ends at the
// first space not escaped with a backslash.
var importPattern = regexp.MustCompile(`(?:^|\s)@((?:\\ |\S)+)`)

// importPaths returns the `@path` imports in body in order of
// appearance. Fenced code blocks and code spans are skipped ("Import
// parsing skips Markdown code spans and fenced code blocks").
func importPaths(body string) []contextImport {
	var out []contextImport
	var fence string
	for i, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if fence != "" {
			if closesFence(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if f := openingFence(trimmed); f != "" {
			fence = f
			continue
		}
		for _, m := range importPattern.FindAllStringSubmatch(blankCodeSpans(line), -1) {
			// The path is echoed into notes and issues on the terminal; a
			// control character (an ANSI escape, say) is never a real
			// import and would let file content rewrite that output.
			if strings.IndexFunc(m[1], unicode.IsControl) >= 0 {
				continue
			}
			out = append(out, contextImport{path: strings.ReplaceAll(m[1], `\ `, " "), line: i + 1})
		}
	}
	return out
}

// openingFence returns the backtick or tilde run that opens a fenced
// code block on line, or "" when line opens none.
func openingFence(line string) string {
	for _, c := range []byte{'`', '~'} {
		n := 0
		for n < len(line) && line[n] == c {
			n++
		}
		if n >= 3 {
			return line[:n]
		}
	}
	return ""
}

// closesFence reports whether line closes a block opened by fence: a
// run of the same character at least as long, and nothing else.
func closesFence(line, fence string) bool {
	run := strings.TrimRight(line, " \t\r")
	if len(run) < len(fence) {
		return false
	}
	return strings.Trim(run, fence[:1]) == ""
}

// blankCodeSpans replaces each code span on line (a backtick run up to
// the next run of the same length) with spaces, so imports inside it
// are not seen. An unmatched backtick run stays literal text.
func blankCodeSpans(line string) string {
	b := []byte(line)
	for i := 0; i < len(b); {
		if b[i] != '`' {
			i++
			continue
		}
		n := 0
		for i+n < len(b) && b[i+n] == '`' {
			n++
		}
		end := -1
		for j := i + n; j < len(b); {
			if b[j] != '`' {
				j++
				continue
			}
			m := 0
			for j+m < len(b) && b[j+m] == '`' {
				m++
			}
			if m == n {
				end = j + m
				break
			}
			j += m
		}
		if end < 0 {
			i += n
			continue
		}
		for k := i; k < end; k++ {
			b[k] = ' '
		}
		i = end
	}
	return string(b)
}

// contextParagraph is one blank-line-separated block of an instruction
// file's body.
type contextParagraph struct {
	text string
	line int
}

// paragraphs splits body into the blocks the duplicate check compares.
// Frontmatter is metadata, not instructions, and is left out. So are
// blocks that are only a heading line or hold no letter or digit
// (rules such as `---`), which are structure, and blocks made only of
// `@path` imports, whose same text loads a different file in each
// directory. Trailing whitespace is ignored so CRLF and LF copies
// compare equal.
func paragraphs(body []byte) []contextParagraph {
	_, offset := docmeta.ParseFrontmatter(body)
	firstLine := 1 + strings.Count(string(body[:offset]), "\n")
	var out []contextParagraph
	var block []string
	start := 0
	flush := func() {
		if len(block) == 0 {
			return
		}
		text := strings.Join(block, "\n")
		block = nil
		if isHeadingOnly(text) || isImportOnly(text) || !strings.ContainsFunc(text, isWordRune) {
			return
		}
		out = append(out, contextParagraph{text: text, line: start})
	}
	for i, line := range strings.Split(string(body[offset:]), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			flush()
			continue
		}
		if len(block) == 0 {
			start = firstLine + i
		}
		block = append(block, line)
	}
	flush()
	return out
}

var headingLine = regexp.MustCompile(`^ {0,3}#{1,6}(\s|$)`)

func isHeadingOnly(text string) bool {
	return !strings.Contains(text, "\n") && headingLine.MatchString(text)
}

func isImportOnly(text string) bool {
	return len(importPaths(text)) == len(strings.Fields(text))
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// duplicateKey identifies one duplicate-paragraph finding: the later
// file, the earlier file it repeats, and what it repeats.
type duplicateKey struct {
	later, earlier   string
	count, bytes     int
	firstLaterLine   int
	firstEarlierLine int
}

// duplicateParagraphIssues reports, per loaded file, the paragraphs it
// repeats verbatim from a file loaded earlier by the same tool. Each
// repeat is charged to the first file that holds the paragraph. When
// several tools load the same pair with the same repeats, one issue
// names them all.
func duplicateParagraphIssues(dir string, tools []ToolContext, contents map[string][]byte) []Issue {
	var order []duplicateKey
	toolsByKey := map[duplicateKey][]string{}
	for _, tc := range tools {
		for _, k := range toolDuplicates(tc, contents) {
			if _, ok := toolsByKey[k]; !ok {
				order = append(order, k)
			}
			toolsByKey[k] = append(toolsByKey[k], tc.Tool)
		}
	}
	issues := make([]Issue, 0, len(order))
	for _, k := range order {
		issues = append(issues, Issue{
			Level: LevelInfo,
			File:  displayPath(dir, k.later),
			Line:  k.firstLaterLine,
			Code:  CodeContextDuplicateParagraph,
			Message: fmt.Sprintf(
				"%d paragraph(s), %d bytes, repeat verbatim from %s (from line %d); both files load in %s",
				k.count, k.bytes, displayPath(dir, k.earlier), k.firstEarlierLine,
				strings.Join(toolsByKey[k], ", ")),
		})
	}
	return issues
}

// toolDuplicates returns one duplicateKey per (later, earlier) file
// pair in tc's load list that shares at least one paragraph, in load
// order of the later file.
func toolDuplicates(tc ToolContext, contents map[string][]byte) []duplicateKey {
	type origin struct {
		path string
		line int
	}
	firstSeen := map[string]origin{}
	var keys []duplicateKey
	for _, f := range tc.Files {
		byEarlier := map[string]*duplicateKey{}
		var earlierOrder []string
		for _, p := range paragraphs(contents[f.Path]) {
			o, ok := firstSeen[p.text]
			if !ok {
				firstSeen[p.text] = origin{path: f.Path, line: p.line}
				continue
			}
			if o.path == f.Path {
				continue
			}
			k := byEarlier[o.path]
			if k == nil {
				k = &duplicateKey{later: f.Path, earlier: o.path, firstLaterLine: p.line, firstEarlierLine: o.line}
				byEarlier[o.path] = k
				earlierOrder = append(earlierOrder, o.path)
			}
			k.count++
			k.bytes += len(p.text)
		}
		for _, e := range earlierOrder {
			keys = append(keys, *byEarlier[e])
		}
	}
	return keys
}
