package doctor

import "strings"

// MatchAny reports whether path matches any pattern in patterns.
// Empty patterns slice returns false. Paths are expected in
// slash-form relative to the project root (as produced by
// walkMarkdown). See ADR 0021 for the supported glob shape.
func MatchAny(patterns []string, path string) bool {
	for _, p := range patterns {
		if Match(p, path) {
			return true
		}
	}
	return false
}

// Match reports whether the slash-form path satisfies pattern.
// Supported tokens:
//
//   - `*`  — zero or more characters within a single path segment.
//   - `**` — zero or more complete path segments (recursive). `**`
//     must appear as an entire segment (e.g. `plugins/**`, `**/x`);
//     a bare `**` inside other characters is treated as `*`.
//   - literal — exact byte match (including `/`).
//
// `?` and character classes are intentionally not supported.
func Match(pattern, path string) bool {
	return matchSegments(splitSegments(pattern), splitSegments(path))
}

// splitSegments returns s split on `/`. An empty string yields a
// single empty segment so callers can distinguish "empty pattern"
// (matches only "") from "single segment".
func splitSegments(s string) []string {
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "/")
}

// matchSegments greedily resolves `**` segments while delegating
// per-segment matches to matchSegment.
func matchSegments(pat, path []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			if len(rest) == 0 {
				return true
			}
			for i := 0; i <= len(path); i++ {
				if matchSegments(rest, path[i:]) {
					return true
				}
			}
			return false
		}
		if len(path) == 0 {
			return false
		}
		if !matchSegment(pat[0], path[0]) {
			return false
		}
		pat = pat[1:]
		path = path[1:]
	}
	return len(path) == 0
}

// matchSegment handles `*` within a single segment. The algorithm:
// split pat on `*` into literal chunks, then verify the segment
// starts with the first chunk, ends with the last chunk, and the
// intermediate chunks appear in order.
func matchSegment(pat, seg string) bool {
	if !strings.Contains(pat, "*") {
		return pat == seg
	}
	chunks := strings.Split(pat, "*")
	// pat = "a*b*c" → chunks = ["a", "b", "c"]
	// pat = "*x"   → chunks = ["", "x"]
	// pat = "x*"   → chunks = ["x", ""]
	// pat = "*"    → chunks = ["", ""]
	if !strings.HasPrefix(seg, chunks[0]) {
		return false
	}
	last := chunks[len(chunks)-1]
	if !strings.HasSuffix(seg, last) {
		return false
	}
	// Trim already-matched prefix/suffix and look for middle chunks
	// in order, left to right.
	seg = seg[len(chunks[0]):]
	if len(chunks) == 1 {
		return true
	}
	// Reserve the suffix for the last chunk.
	if len(seg) < len(last) {
		return false
	}
	middle := seg[:len(seg)-len(last)]
	for _, c := range chunks[1 : len(chunks)-1] {
		if c == "" {
			continue
		}
		idx := strings.Index(middle, c)
		if idx < 0 {
			return false
		}
		middle = middle[idx+len(c):]
	}
	return true
}
