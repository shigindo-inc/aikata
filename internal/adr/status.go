package adr

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/shigindo-inc/aikata/internal/docmeta"
)

// Class groups a decision status by what it means for the reader. It is
// computed from the first word of the status; prose after that word
// (dates, partial-supersession notes) never changes the class (ADR 0052).
type Class string

// Decision status classes.
const (
	ClassPending    Class = "pending"
	ClassEffective  Class = "effective"
	ClassSuperseded Class = "superseded"
	ClassRetired    Class = "retired"
	ClassUnknown    Class = "unknown"
)

// decisionWords maps each decision word to its class. Frontmatter
// `status` must not hold any of these: it means the document lifecycle
// on every Markdown file, ADRs included (ADR 0052).
var decisionWords = map[string]Class{
	"proposed":   ClassPending,
	"accepted":   ClassEffective,
	"active":     ClassEffective,
	"superseded": ClassSuperseded,
	"deprecated": ClassRetired,
	"rejected":   ClassRetired,
}

// statusLineRE captures the value after `**Status**:` (or the Japanese
// `**状態**:`, with an ASCII or a full-width colon). The ja templates keep
// the English key, but ADRs written by hand in Japanese projects use the
// localized one.
var statusLineRE = regexp.MustCompile(`(?i)\*\*(?:Status|状態)\*\*\s*[:：]\s*([^\n]+)`)

var (
	leadingWordRE = regexp.MustCompile(`^[\s*_` + "`" + `]*([A-Za-z]+)`)
	adrMentionRE  = regexp.MustCompile(`\bADR[\s-]?(\d{4})\b`)
	adrFileRE     = regexp.MustCompile(`(?:^|/)(\d{4})-[a-z0-9][a-z0-9-]*\.md$`)
)

// Status is the decision status declared on an ADR's body status line.
type Status struct {
	// Raw is the text after the colon, trimmed.
	Raw string
	// Class is derived from the first word of Raw.
	Class Class
	// Line is the 1-based line of the status line in the file.
	Line int
	// Links are the ADR numbers the status line names, in order, without
	// duplicates (for example the superseding ADR).
	Links []int
}

// ParseStatus reads the body status line of an ADR. It returns false
// when the ADR has none. Frontmatter is never consulted: frontmatter
// `status` is the document lifecycle (ADR 0052).
func ParseStatus(body []byte) (Status, bool) {
	loc := statusLineRE.FindSubmatchIndex(body)
	if loc == nil {
		return Status{}, false
	}
	raw := strings.TrimSpace(string(body[loc[2]:loc[3]]))
	return Status{
		Raw:   raw,
		Class: ClassOf(raw),
		Line:  strings.Count(string(body[:loc[0]]), "\n") + 1,
		Links: referencedNumbers([]byte(raw)),
	}, true
}

// ClassOf classifies a status value by its first word. Values whose first
// word is not a decision word (for example 「一部 active」) are
// ClassUnknown, so a person or the inventory skill reads them.
func ClassOf(value string) Class {
	m := leadingWordRE.FindStringSubmatch(value)
	if m == nil {
		return ClassUnknown
	}
	if c, ok := decisionWords[strings.ToLower(m[1])]; ok {
		return c
	}
	return ClassUnknown
}

// DecisionWord reports whether a frontmatter `status` value is a decision
// word rather than a lifecycle value, and returns its class.
func DecisionWord(value string) (Class, bool) {
	c, ok := decisionWords[strings.ToLower(strings.TrimSpace(value))]
	return c, ok
}

// referencedNumbers returns the ADR numbers text names, either as
// `ADR NNNN` or as a link to an `NNNN-slug.md` file, in order of first
// appearance.
func referencedNumbers(text []byte) []int {
	type hit struct{ pos, n int }
	var hits []hit
	for _, m := range adrMentionRE.FindAllSubmatchIndex(text, -1) {
		n, _ := strconv.Atoi(string(text[m[2]:m[3]]))
		hits = append(hits, hit{m[0], n})
	}
	for _, l := range docmeta.ExtractLinks(text) {
		if m := adrFileRE.FindStringSubmatch(l.Target); m != nil {
			n, _ := strconv.Atoi(m[1])
			hits = append(hits, hit{strings.Index(string(text), l.Raw), n})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].pos < hits[j].pos })
	var out []int
	seen := map[int]struct{}{}
	for _, h := range hits {
		if _, dup := seen[h.n]; dup {
			continue
		}
		seen[h.n] = struct{}{}
		out = append(out, h.n)
	}
	return out
}
