// Package docmeta extracts mechanically-derivable facts from Markdown
// documents — YAML front-matter scalars and relative inter-document
// links. It is the single shared parser used by both `aikata doctor`
// (link/frontmatter validation) and `aikata map` (the doc-cartography
// artifact, ADR 0044), so link and front-matter parsing cannot drift
// between the two features (docmap-design.md §6).
//
// The package reads document text only: front-matter, headings, the
// first lines, and a link regex. It never reads source code, which keeps
// the stack-agnostic core intact (ADR 0044 D4).
package docmeta

import (
	"regexp"
	"strings"
)

// DefaultSkipDirs is the shared baseline of directory names the document
// scan never descends into: external areas (vendored dependencies, build
// outputs, VCS), generated AI-tool rule trees, and aikata-internal
// machine/scratch areas. Both `aikata doctor` (validation walk) and
// `aikata map` (cartography scan) use this one set so the surface they
// consider cannot drift apart.
//
// `.aikata` is skipped so the doc map never catalogs its own output
// (docmap.{yaml,md}) and doctor never flags the generated, frontmatter-
// free docmap.md. `dist` holds first-party distribution payloads
// (skill/plugin Markdown) validated by the repolint distribution tests,
// not by doctor; `testdata` holds golden fixtures.
var DefaultSkipDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "build": {}, "dist": {},
	".dart_tool": {}, ".next": {}, "vendor": {}, ".turbo": {},
	".cursor": {}, ".github": {}, "testdata": {}, ".remember": {},
	".serena": {}, ".aikata-proposed": {}, ".aikata": {},
}

// DefaultSkipFiles names individual *.md leaves the scan always skips:
// generated AI-tool artifacts that carry no front-matter by design.
var DefaultSkipFiles = map[string]struct{}{
	"CLAUDE.md": {}, "GEMINI.md": {},
}

// ParseFrontmatter extracts the YAML front-matter block delimited by
// `---` lines at the very start of body. Returns the lines (without the
// fences) and the byte offset where the body proper begins. If the file
// has no front-matter, lines is nil and offset is 0.
func ParseFrontmatter(body []byte) (lines []string, offset int) {
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

// FrontmatterValue returns the value associated with key in the parsed
// front-matter lines. Only top-level scalar keys are supported (which is
// all aikata uses for required keys). The returned line is 1-based and
// already offset by the opening `---` fence, matching the source line in
// the original document.
func FrontmatterValue(lines []string, key string) (value string, line int, ok bool) {
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

// linkRE matches markdown links of the form [text](path). The captured
// group is the link target, which may be an absolute URL, a mailto:
// address, or a relative path; ExtractLinks filters to the relative
// paths.
var linkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// Link is one relative Markdown link found in a document.
type Link struct {
	// Raw is the target text exactly as written inside the parentheses,
	// suitable for echoing in a diagnostic ("broken link: ...").
	Raw string
	// Target is the cleaned, slash-form relative path: URL fragment and
	// query stripped, a leading "./" removed, surrounding space trimmed.
	// It is resolved by the caller relative to the linking document's own
	// directory (doctor resolves AGENTS.md links against the root; docmap
	// resolves each document's links against that document's directory).
	Target string
	// Line is the 1-based line number of the link in the document body.
	Line int
}

// ExtractLinks returns every relative inter-document link in body, in
// source order. Absolute URLs (http/https), mailto: addresses, and empty
// targets are excluded — only links that can resolve to a file on disk
// are returned. Resolution (joining Target with a base directory and
// checking existence) is left to the caller so the same parse serves
// both the root-relative doctor check and the per-document docmap graph.
func ExtractLinks(body []byte) []Link {
	var links []Link
	lineNum := 0
	for _, line := range strings.Split(string(body), "\n") {
		lineNum++
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			raw := m[1]
			target := raw
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
			target = strings.TrimPrefix(target, "./")
			links = append(links, Link{Raw: raw, Target: target, Line: lineNum})
		}
	}
	return links
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
