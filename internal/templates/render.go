package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
	"time"
)

// Clock is the time source used by template helpers. Tests pass a fixed
// time so golden output is reproducible.
type Clock func() time.Time

// Helpers returns the template FuncMap used by all aikata templates.
// Pass clock = nil to use time.Now.
//
// Available helpers (kept intentionally small — sprig is out of scope
// per ARCHITECTURE.md §10 / SPEC D6):
//
//   - now        → today's date in YYYY-MM-DD (UTC)
//   - joinSlash  → slash-join a variadic list of path segments
//   - kebab      → lowercase + non-alphanumeric runs collapsed to '-'
func Helpers(clock Clock) template.FuncMap {
	if clock == nil {
		clock = time.Now
	}
	return template.FuncMap{
		"now": func() string {
			return clock().UTC().Format("2006-01-02")
		},
		"joinSlash": func(parts ...string) string {
			return strings.Join(parts, "/")
		},
		"kebab": Kebab,
	}
}

// Render reads the named template from the embedded FS, parses it, and
// returns the rendered string. The path is relative to the FS root
// returned by FS() (e.g. "presets/minimal/README.md.tmpl").
func Render(path string, data any, clock Clock) (string, error) {
	root, err := FS()
	if err != nil {
		return "", err
	}
	raw, err := fs.ReadFile(root, path)
	if err != nil {
		return "", fmt.Errorf("templates: read %q: %w", path, err)
	}
	tmpl, err := template.New(path).Funcs(Helpers(clock)).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("templates: parse %q: %w", path, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("templates: execute %q: %w", path, err)
	}
	return buf.String(), nil
}

// Kebab converts s to kebab-case: ASCII lowercase, with runs of
// non-alphanumeric characters collapsed into a single '-'. Leading and
// trailing hyphens are trimmed. Intended for slug derivation (project
// names, ADR titles, stack identifiers); not Unicode-aware on purpose
// — those normalizations belong in a different helper, if ever needed.
func Kebab(s string) string {
	var b strings.Builder
	pendingHyphen := false
	wroteAny := false
	for _, r := range strings.ToLower(s) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			if pendingHyphen && wroteAny {
				b.WriteRune('-')
			}
			b.WriteRune(r)
			pendingHyphen = false
			wroteAny = true
		default:
			pendingHyphen = true
		}
	}
	return b.String()
}
