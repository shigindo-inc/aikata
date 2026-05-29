package repolint

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// selfPath is the repo-relative path of this scanner. It is skipped during the
// tree scan because its self-test fixtures (below) deliberately contain
// known-bad samples that would otherwise trip the very patterns under test.
const selfPath = "internal/repolint/secretscan_test.go"

// pattern is one named secret / privacy signature.
//
// The patterns are deliberately tighter than a bare keyword grep: each
// requires the *shape of a real leak* (a path segment, an address, a PEM
// header, a value after the assignment) rather than the bare token. This is
// what lets the project's own prose document the patterns — e.g. "do not
// commit `/Users/...` paths or `api_key=` values" — without the scanner
// flagging its own documentation. TestPatternsAreLive proves, on every run,
// that this precision did not silently render a pattern dead.
type pattern struct {
	name string
	re   *regexp.Regexp
}

var patterns = []pattern{
	// A real absolute macOS home path has a username segment after /Users/.
	// Placeholder mentions in docs ("/Users/...", "/Users/" in a string
	// literal) are followed by a dot or quote, so they do not match.
	{"local-user-path", regexp.MustCompile(`/Users/[A-Za-z0-9]`)},

	// A real local dev path under the home workspace has a segment after it.
	// Doc mentions ("~/Workspace" in backticks) are not followed by "/x".
	{"workspace-path", regexp.MustCompile(`~/Workspace/[A-Za-z0-9]`)},

	// Personal mailbox providers. The maintainer address lives in git
	// authorship, never in tracked file content.
	{"private-email", regexp.MustCompile(`[A-Za-z0-9._%+-]+@(?:gmail|icloud|me)\.com`)},

	// A real PEM private-key header carries the full dashed envelope and the
	// literal "PRIVATE KEY". Docs that mention "BEGIN (RSA|OPENSSH|PRIVATE)"
	// as a shorthand lack the dashes and the "KEY" suffix, so they miss.
	{"private-key", regexp.MustCompile(`-----BEGIN (?:RSA |DSA |EC |OPENSSH |PGP )?PRIVATE KEY-----`)},

	// A real credential assignment has a value after the separator. Doc
	// mentions ("`api_key=`,") are followed by a backtick, not a value char,
	// so they miss; a quoted value ("api_key=\"AKIA…\"") still matches.
	{"credential-assignment", regexp.MustCompile(`(?i)(?:api[_-]?key|client[_-]?secret|refresh[_-]?token)\s*[:=]\s*["']?[A-Za-z0-9]`)},
}

// skipExt lists extensions whose contents are binary and not worth scanning.
var skipExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".ico": true, ".pdf": true, ".zip": true,
	".gz": true, ".tar": true, ".woff": true, ".woff2": true,
}

// forbiddenTracked are paths that must never be tracked, even though they are
// gitignored — a belt-and-braces check against a stray `git add -f`.
var forbiddenTracked = []string{".env", ".env.local"}

// forbiddenSuffix are tracked-path suffixes reserved for local-only files.
var forbiddenSuffix = []string{".local", ".local.yaml", ".local.yml"}

// TestPatternsAreLive is the load-bearing guard: it proves every pattern still
// detects a known-bad sample and ignores a known-good one. A pattern that was
// silently broken (e.g. by an escaping mistake) would match nothing and let
// the tree scan pass vacuously — this catches that, red.
func TestPatternsAreLive(t *testing.T) {
	cases := []struct {
		sample string
		match  bool // must a known-bad HIT, must a known-good MISS
	}{
		{"/Users/satoshi/Workspace/x", true},
		{"~/Workspace/develop/aikata", true},
		{"someone@gmail.com", true},
		{"dev@icloud.com", true},
		{"-----BEGIN RSA PRIVATE KEY-----", true},
		{"-----BEGIN OPENSSH PRIVATE KEY-----", true},
		{"-----BEGIN PRIVATE KEY-----", true},
		{"api_key=AKIAIOSFODNN7EXAMPLE", true},
		{`client_secret: "s3cr3tValue"`, true},
		{"refresh_token = abc123", true},
		// Known-good: the placeholder / documentation forms that the project
		// legitimately commits must NOT match.
		{"local path (`/Users/...`)", false},
		{`strings.Contains(body, "/Users/")`, false},
		{"local user paths (`/Users/...`, `~/Workspace`)", false},
		{"grep for `BEGIN (RSA|OPENSSH|PRIVATE)`", false},
		{"key material (`api_key=`, `client_secret=`)", false},
		{"contact the maintainers", false},
		{"a description of the api surface", false},
	}
	for _, c := range cases {
		hit := false
		for _, p := range patterns {
			if p.re.MatchString(c.sample) {
				hit = true
				break
			}
		}
		if hit != c.match {
			t.Errorf("self-test: %q expected match=%v, got match=%v", c.sample, c.match, hit)
		}
	}

	// Each pattern must independently match at least one known-bad sample,
	// so an entirely dead pattern cannot hide behind the others above.
	for _, p := range patterns {
		live := false
		for _, c := range cases {
			if c.match && p.re.MatchString(c.sample) {
				live = true
				break
			}
		}
		if !live {
			t.Errorf("self-test: pattern %q matched no known-bad sample — it may be dead", p.name)
		}
	}
}

// TestNoSecretsInTrackedFiles scans every tracked file and fails if any
// pattern hits, or if a local-only path is tracked at all.
func TestNoSecretsInTrackedFiles(t *testing.T) {
	root := repoRoot(t)

	out, err := exec.Command("git", "-C", root, "ls-files").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")

	var failures []string
	for _, rel := range files {
		if rel == "" || rel == selfPath {
			continue
		}
		for _, forbidden := range forbiddenTracked {
			if rel == forbidden {
				failures = append(failures, rel+": local-only file must not be tracked")
			}
		}
		for _, suf := range forbiddenSuffix {
			if strings.HasSuffix(rel, suf) {
				failures = append(failures, rel+": local-only suffix must not be tracked")
			}
		}
		if skipExt[strings.ToLower(filepath.Ext(rel))] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if !utf8.Valid(data) {
			continue // skip binary blobs without a known extension
		}
		text := string(data)
		for _, p := range patterns {
			if loc := p.re.FindStringIndex(text); loc != nil {
				line := 1 + strings.Count(text[:loc[0]], "\n")
				failures = append(failures, formatHit(rel, line, p.name))
			}
		}
	}

	if len(failures) > 0 {
		t.Fatalf("secret/privacy scan found %d issue(s):\n%s",
			len(failures), strings.Join(failures, "\n"))
	}
}

func formatHit(rel string, line int, name string) string {
	return rel + ":" + itoa(line) + ": " + name
}

// itoa avoids pulling strconv just for line numbers in failure output.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// repoRoot resolves the repository top level so the scan covers the whole
// tree regardless of which directory `go test` runs the package from.
func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse --show-toplevel: %v", err)
	}
	return strings.TrimSpace(string(out))
}
