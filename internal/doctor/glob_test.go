package doctor

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		// literal
		{"literal exact", "plugins/x.md", "plugins/x.md", true},
		{"literal mismatch", "plugins/x.md", "plugins/y.md", false},

		// single-segment *
		{"star matches segment", "plugins/*", "plugins/foo", true},
		{"star is non-recursive", "plugins/*", "plugins/foo/bar", false},
		{"star prefix", "plugins/*.md", "plugins/foo.md", true},
		{"star prefix mismatch", "plugins/*.md", "plugins/foo.txt", false},
		{"star bare matches anything in segment", "*", "foo", true},
		{"star bare rejects slash", "*", "foo/bar", false},

		// recursive **
		{"double-star prefix matches all", "plugins/**", "plugins/foo/SKILL.md", true},
		{"double-star matches plugin root", "plugins/**", "plugins/foo", true},
		{"double-star prefix mismatch", "plugins/**", "other/foo.md", false},
		{"double-star suffix recurses", "**/SKILL.md", "plugins/foo/SKILL.md", true},
		{"double-star suffix top-level", "**/SKILL.md", "SKILL.md", true},
		{"double-star middle", "plugins/**/.claude-plugin/**", "plugins/a/.claude-plugin/marketplace.json", true},
		{"double-star alone matches anything", "**", "any/path/at/all.md", true},

		// mixed star + literal
		{"trailing star", "plugins/foo*", "plugins/foobar", true},
		{"trailing star empty match", "plugins/foo*", "plugins/foo", true},
		{"surrounding stars", "plugins/*foo*", "plugins/xfooy", true},
		{"surrounding stars no match", "plugins/*foo*", "plugins/xbary", false},

		// boundary
		{"empty pattern empty path", "", "", true},
		{"empty pattern nonempty path", "", "foo", false},
		{"path shorter than pattern", "a/b/c", "a/b", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Match(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
			}
		})
	}
}

func TestMatchAny(t *testing.T) {
	patterns := []string{"plugins/**", "**/SKILL.md", "vendor/legacy.md"}
	cases := []struct {
		path string
		want bool
	}{
		{"plugins/foo/SKILL.md", true},
		{"dist/claude-code/skill/SKILL.md", true}, // via **/SKILL.md
		{"vendor/legacy.md", true},
		{"docs/normal.md", false},
		{"plugins-of-mine/file.md", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := MatchAny(patterns, tc.path)
			if got != tc.want {
				t.Errorf("MatchAny(%v) for %q = %v, want %v", patterns, tc.path, got, tc.want)
			}
		})
	}
}

func TestMatchAny_EmptyPatternsNeverMatch(t *testing.T) {
	if MatchAny(nil, "anything") {
		t.Error("MatchAny(nil, ...) should return false")
	}
	if MatchAny([]string{}, "anything") {
		t.Error("MatchAny([], ...) should return false")
	}
}
