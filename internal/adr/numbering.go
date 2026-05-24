// Package adr offers a single source of truth for the ADR filename
// numbering convention used in docs/adr/. The naming rule (matched by
// Pattern below) is `NNNN-slug.md`, with NNNN being a 4-digit, zero-
// padded decimal counter starting at 0001. Scaffolding code, doctor
// checks, and the future `aikata add adr` command all consume this
// package so the rule stays in one place.
package adr

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// Pattern matches the canonical ADR filename. The slug part is
// restricted to lowercase alphanumerics plus hyphens to keep generated
// names portable across filesystems.
var Pattern = regexp.MustCompile(`^(\d{4})-[a-z0-9][a-z0-9-]*\.md$`)

// Entry pairs a discovered ADR number with its on-disk filename.
type Entry struct {
	Number   int
	Filename string
}

// Scan returns the ADR entries discovered in dir, sorted by Number
// ascending. Files whose name does not match Pattern are silently
// skipped — doctor surfaces those separately. A nonexistent directory
// returns (nil, nil) rather than an error so callers can treat the
// "no ADRs yet" case the same as "directory missing".
func Scan(dir string) ([]Entry, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("adr: stat %s: %w", dir, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("adr: %s is not a directory", dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("adr: read %s: %w", dir, err)
	}
	var entries []Entry
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		m := Pattern.FindStringSubmatch(f.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			// Pattern already guarantees \d{4}; this is unreachable
			// in practice but keeps the parser honest.
			continue
		}
		entries = append(entries, Entry{Number: n, Filename: f.Name()})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Number != entries[j].Number {
			return entries[i].Number < entries[j].Number
		}
		return entries[i].Filename < entries[j].Filename
	})
	return entries, nil
}

// Next returns the smallest unused ADR number, equal to one greater
// than the maximum existing number. Returns 1 when dir is missing or
// holds no canonical ADRs. Duplicates among the existing files are
// tolerated (Scan would surface both via Entry, doctor reports the
// duplicate); Next still advances past the maximum to avoid handing
// out a number already in use.
func Next(dir string) (int, error) {
	entries, err := Scan(dir)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 1, nil
	}
	return entries[len(entries)-1].Number + 1, nil
}

// Filename builds the canonical filename for the given number and
// slug. The slug is taken as-is — callers are responsible for
// normalizing it to match Pattern.
func Filename(number int, slug string) string {
	return fmt.Sprintf("%04d-%s.md", number, slug)
}

// Path joins adrDir with Filename for convenience at call sites that
// already hold the parent directory.
func Path(adrDir string, number int, slug string) string {
	return filepath.Join(adrDir, Filename(number, slug))
}
