package docmap

import (
	"path"
	"strings"
)

// title returns the document's display title. In priority order: a
// `title:` front-matter scalar, the first H1 heading, then the file's
// base name. Always non-empty.
func title(lines []string, body []byte, offset int, rel string) string {
	if v := frontmatterScalar(lines, "title"); v != "" {
		return v
	}
	if h := firstH1(body, offset); h != "" {
		return h
	}
	return path.Base(rel)
}

// summary returns a best-effort one-line gist, degrading gracefully so a
// document with no aikata front-matter still maps (design §3). In
// priority order, first non-empty wins:
//
//  1. optional `summary:` front-matter key (Q-DOCMAP-02);
//  2. a leading `>` blockquote;
//  3. the first non-heading paragraph after the H1;
//  4. the H1 text;
//  5. the file's base name.
func summary(lines []string, body []byte, offset int, rel string) string {
	if v := frontmatterScalar(lines, "summary"); v != "" {
		return v
	}
	bodyLines := strings.Split(string(body[offset:]), "\n")
	if bq := leadingBlockquote(bodyLines); bq != "" {
		return bq
	}
	if p := firstParagraphAfterH1(bodyLines); p != "" {
		return p
	}
	if h := firstH1(body, offset); h != "" {
		return h
	}
	return path.Base(rel)
}

// firstH1 returns the text of the first `# ` heading in the body proper
// (after the front-matter offset), or "" when none is present.
func firstH1(body []byte, offset int) string {
	for _, ln := range strings.Split(string(body[offset:]), "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "# ") {
			return strings.TrimSpace(t[2:])
		}
	}
	return ""
}

// leadingBlockquote joins the first contiguous run of `>` lines found
// before any non-heading prose, with the markers stripped. Headings and
// blank lines above the blockquote are skipped (documents commonly open
// `# Title` then `> one-line description`). Returns "" if the first prose
// encountered is not a blockquote.
func leadingBlockquote(bodyLines []string) string {
	var quote []string
	for _, ln := range bodyLines {
		t := strings.TrimSpace(ln)
		switch {
		case t == "":
			if len(quote) > 0 {
				return joinProse(quote)
			}
			continue
		case strings.HasPrefix(t, "#"):
			if len(quote) > 0 {
				return joinProse(quote)
			}
			continue
		case strings.HasPrefix(t, ">"):
			quote = append(quote, strings.TrimSpace(strings.TrimPrefix(t, ">")))
		default:
			// First prose is not a blockquote.
			return joinProse(quote)
		}
	}
	return joinProse(quote)
}

// firstParagraphAfterH1 returns the first contiguous run of non-heading,
// non-blockquote prose lines following the first H1, joined to one line.
func firstParagraphAfterH1(bodyLines []string) string {
	seenH1 := false
	var para []string
	for _, ln := range bodyLines {
		t := strings.TrimSpace(ln)
		if !seenH1 {
			if strings.HasPrefix(t, "# ") {
				seenH1 = true
			}
			continue
		}
		switch {
		case t == "":
			if len(para) > 0 {
				return joinProse(para)
			}
		case strings.HasPrefix(t, "#") || strings.HasPrefix(t, ">"):
			if len(para) > 0 {
				return joinProse(para)
			}
		default:
			para = append(para, t)
		}
	}
	return joinProse(para)
}

// joinProse collapses prose lines into a single trimmed line.
func joinProse(lines []string) string {
	return strings.TrimSpace(strings.Join(lines, " "))
}
