package adr

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shigindo-inc/aikata/internal/docmeta"
)

// Dir is the ADR directory relative to a project root.
var Dir = filepath.Join("docs", "adr")

// Record is one ADR as `aikata list adrs` reports it. Every field is
// computed from document text; judgments such as "partially superseded"
// or "not yet implemented" are left to the inventory skill (ADR 0052).
type Record struct {
	Number int
	Title  string
	// Path is slash-separated and relative to the project root.
	Path string
	// HasStatus is false when the ADR has no body status line.
	HasStatus bool
	Status    Status
	// DocRefs counts Markdown files other than this ADR that link to it
	// or name it as `ADR NNNN`. Source code is not read (SPEC §8).
	DocRefs int
	Bytes   int
}

// Inventory lists every ADR under root/docs/adr in number order.
func Inventory(root string) ([]Record, error) {
	adrDir := filepath.Join(root, Dir)
	entries, err := Scan(adrDir)
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(entries))
	byNumber := map[int][]int{}
	byPath := map[string]int{}
	for _, e := range entries {
		body, err := os.ReadFile(filepath.Join(adrDir, e.Filename))
		if err != nil {
			return nil, err
		}
		rel := filepath.ToSlash(filepath.Join(Dir, e.Filename))
		st, ok := ParseStatus(body)
		byNumber[e.Number] = append(byNumber[e.Number], len(records))
		byPath[rel] = len(records)
		records = append(records, Record{
			Number:    e.Number,
			Title:     firstHeading(body),
			Path:      rel,
			HasStatus: ok,
			Status:    st,
			Bytes:     len(body),
		})
	}
	if len(records) == 0 {
		return records, nil
	}
	err = walkMarkdown(root, func(rel string, body []byte) {
		hit := map[int]struct{}{}
		for _, m := range adrMentionRE.FindAllSubmatch(body, -1) {
			n, _ := strconv.Atoi(string(m[1]))
			for _, i := range byNumber[n] {
				hit[i] = struct{}{}
			}
		}
		base := filepath.ToSlash(filepath.Dir(rel))
		for _, l := range docmeta.ExtractLinks(body) {
			target := filepath.ToSlash(filepath.Clean(filepath.Join(base, filepath.FromSlash(l.Target))))
			if i, ok := byPath[target]; ok {
				hit[i] = struct{}{}
			}
		}
		for i := range hit {
			if records[i].Path != rel {
				records[i].DocRefs++
			}
		}
	})
	return records, err
}

// walkMarkdown calls fn for every Markdown file under root that the
// shared document scan covers (docmeta.SkipDir, docmeta.DefaultSkipFiles).
func walkMarkdown(root string, fn func(rel string, body []byte)) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if docmeta.SkipDir(d.Name(), p, root) {
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
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fn(filepath.ToSlash(rel), body)
		return nil
	})
}

// firstHeading returns the text of the first `# ` heading, or "".
func firstHeading(body []byte) string {
	_, offset := docmeta.ParseFrontmatter(body)
	for _, line := range strings.Split(string(body[offset:]), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}
