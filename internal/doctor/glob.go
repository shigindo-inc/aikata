package doctor

import "github.com/shigindo-inc/aikata/internal/glob"

// MatchAny reports whether path matches any pattern in patterns.
// Thin wrapper over internal/glob so `doctor.exclude` (ADR 0021) and
// `sync.own` (ADR 0025) share one matcher. Paths are expected in
// slash-form relative to the project root (as produced by
// walkMarkdown).
func MatchAny(patterns []string, path string) bool {
	return glob.MatchAny(patterns, path)
}

// Match reports whether the slash-form path satisfies pattern. See
// internal/glob for the supported token set.
func Match(pattern, path string) bool {
	return glob.Match(pattern, path)
}
