package repolint

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// adrFileRE matches a numbered ADR filename (NNNN-slug.md). Non-ADR files in
// docs/adr/ (for example a future README or template) are ignored.
var adrFileRE = regexp.MustCompile(`^\d{4}-.*\.md$`)

// TestReadmeAdrIndexCoversAllAdrs guards against the recurring drift where a
// new ADR lands under docs/adr/ but is never linked from the README "Decisions
// & design" index. It is a narrow mechanical check — not a general Markdown
// linter — that pins exactly one invariant: every numbered ADR file is
// referenced by path from README.md.
func TestReadmeAdrIndexCoversAllAdrs(t *testing.T) {
	root := repoRoot(t)

	entries, err := os.ReadDir(filepath.Join(root, "docs", "adr"))
	if err != nil {
		t.Fatalf("read docs/adr: %v", err)
	}

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readmeText := string(readme)

	var missing []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !adrFileRE.MatchString(name) {
			continue
		}
		// The README links each ADR by its docs/adr/<file> path.
		if !strings.Contains(readmeText, "docs/adr/"+name) {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("README.md ADR index is missing %d ADR(s); add a link under "+
			"\"Decisions & design\" for each:\n%s",
			len(missing), strings.Join(missing, "\n"))
	}
}
