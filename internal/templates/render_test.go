package templates

import (
	"io/fs"
	"strings"
	"testing"
	"time"
)

func fixedClock(year int, month time.Month, day int) Clock {
	t := time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestHelpersNow(t *testing.T) {
	got := Helpers(fixedClock(2026, time.May, 21))["now"].(func() string)()
	want := "2026-05-21"
	if got != want {
		t.Fatalf("now() = %q, want %q", got, want)
	}
}

func TestHelpersJoinSlash(t *testing.T) {
	join := Helpers(nil)["joinSlash"].(func(...string) string)
	got := join("docs", "adr", "0001-foo.md")
	want := "docs/adr/0001-foo.md"
	if got != want {
		t.Fatalf("joinSlash = %q, want %q", got, want)
	}
}

func TestHelpersKebab(t *testing.T) {
	kebab := Helpers(nil)["kebab"].(func(string) string)
	cases := []struct {
		in, want string
	}{
		{"My App", "my-app"},
		{"  spaced  ", "spaced"},
		{"Mixed_Case-123", "mixed-case-123"},
		{"!!!", ""},
		{"Foo  Bar  Baz", "foo-bar-baz"},
		{"日本語Mixed", "mixed"}, // non-ASCII is dropped (documented limitation)
		{"", ""},
	}
	for _, c := range cases {
		got := kebab(c.in)
		if got != c.want {
			t.Errorf("kebab(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFSContainsMinimalPreset(t *testing.T) {
	root, err := FS()
	if err != nil {
		t.Fatalf("FS: %v", err)
	}
	// Presets are now lang-scoped (Task 12); read the en subdirectory.
	entries, err := fs.ReadDir(root, "presets/minimal/en")
	if err != nil {
		t.Fatalf("ReadDir presets/minimal/en: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("presets/minimal/en has no entries")
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md.tmpl") {
			t.Errorf("unexpected non-template file: %s", e.Name())
		}
	}
}

func TestRenderMinimalReadme(t *testing.T) {
	got, err := Render("presets/minimal/en/README.md.tmpl",
		map[string]any{"ProjectName": "samplekata", "Lang": "en"},
		fixedClock(2026, time.May, 21))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(got, "samplekata") {
		t.Errorf("rendered README does not contain ProjectName:\n%s", got)
	}
}

func TestRenderMissingKeyIsAnError(t *testing.T) {
	// Pass empty data so {{.ProjectName}} resolves to <no value>; the
	// missingkey=error option turns this into a real error rather than a
	// silent template output of "<no value>".
	_, err := Render("presets/minimal/en/README.md.tmpl", map[string]any{},
		fixedClock(2026, time.May, 21))
	if err == nil {
		t.Fatalf("expected error for missing template key; got nil")
	}
}

func TestRenderUnknownTemplate(t *testing.T) {
	_, err := Render("presets/does-not-exist/file.md.tmpl", nil, nil)
	if err == nil {
		t.Fatalf("expected error for missing template; got nil")
	}
}
