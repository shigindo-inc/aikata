package doctor

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// Level grades the severity of an Issue. Doctor sets exit code 3 if any
// LevelError is reported; warnings and infos are advisory.
type Level int

// Severity levels for doctor Issues, in ascending order of seriousness.
const (
	LevelInfo Level = iota
	LevelWarning
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	case LevelInfo:
		return "info"
	}
	return "unknown"
}

// Issue is a single finding emitted by a check function. File is a
// path relative to Options.TargetDir; Line is 1-based (0 = no specific
// line). Code, when set, is a stable machine-readable discriminator
// consumed by `doctor --fix`. The zero value (empty Code) marks the
// issue as advisory-only — `--fix` will skip it.
type Issue struct {
	Level   Level
	File    string
	Line    int
	Message string
	Code    string
}

// Options carries the inputs every check needs.
type Options struct {
	// TargetDir is the project root to inspect.
	TargetDir string
	// Now overrides time.Now for the stale-updated check; zero value
	// resolves to time.Now() at Run time.
	Now time.Time
	// Excludes is an optional list of glob patterns evaluated against
	// the slash-form path relative to TargetDir. Markdown files whose
	// relative path matches any pattern are skipped by walkMarkdown
	// and therefore by checkFrontmatter, checkUpdated, and
	// checkGlossary. Supported tokens: `*`, `**`, and literals; see
	// internal/doctor/glob.go. The zero value (nil slice) preserves
	// the pre-v0.7.3 behaviour. See ADR 0021.
	Excludes []string
	// Includes scopes the Markdown walk to the document surface aikata
	// manages (ADR 0033 / ADR 0037 D1). When non-empty, walkMarkdown
	// only visits Markdown files whose relative path matches one of
	// these globs; Excludes still subtracts on top. An empty/nil slice
	// means "no include filter" — the broad-audit walk over every
	// Markdown file (`aikata doctor --all-markdown`). The cli layer
	// populates this from ManagedIncludeGlobs by default.
	Includes []string
}

// Check is the contract every individual check function implements.
// Returning an error means the check could not run (I/O failure); a
// successfully-running check that found problems returns issues with
// nil error.
type Check func(opts Options) ([]Issue, error)

// builtinChecks lists the checks Run executes, in stable order.
var builtinChecks = []struct {
	Name string
	Fn   Check
}{
	{"frontmatter", checkFrontmatter},
	{"links", checkLinks},
	{"adr", checkADR},
	{"adr-numbering", checkADRNumbering},
	{"memory", checkMemory},
	{"updated", checkUpdated},
	{"env", checkEnvExample},
	{"glossary", checkGlossary},
}

// Run walks the target directory and produces a sorted slice of
// Issues. The order is: Level descending (error first), then file,
// then line. The error return is reserved for I/O failures that
// prevent doctor from running at all.
func Run(opts Options) ([]Issue, error) {
	if opts.TargetDir == "" {
		return nil, fmt.Errorf("doctor: target directory is required")
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}

	var all []Issue
	for _, c := range builtinChecks {
		got, err := c.Fn(opts)
		if err != nil {
			return nil, fmt.Errorf("doctor: %s: %w", c.Name, err)
		}
		all = append(all, got...)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Level != all[j].Level {
			return all[i].Level > all[j].Level
		}
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		return all[i].Line < all[j].Line
	})
	return all, nil
}

// Format writes the human-readable rendering of issues to w. Each
// issue takes one line in `file:line: [level] message` shape; lines
// without a specific source line render as `file: [level] message`.
func Format(w io.Writer, issues []Issue) error {
	for _, iss := range issues {
		var err error
		if iss.Line > 0 {
			_, err = fmt.Fprintf(w, "%s:%d: [%s] %s\n", iss.File, iss.Line, iss.Level, iss.Message)
		} else {
			_, err = fmt.Fprintf(w, "%s: [%s] %s\n", iss.File, iss.Level, iss.Message)
		}
		if err != nil {
			return fmt.Errorf("doctor: format: %w", err)
		}
	}
	return nil
}

// HasErrors reports whether the issue set contains any LevelError
// items. The cli layer uses this to choose between exit code 0 and 3.
func HasErrors(issues []Issue) bool {
	for _, iss := range issues {
		if iss.Level == LevelError {
			return true
		}
	}
	return false
}
