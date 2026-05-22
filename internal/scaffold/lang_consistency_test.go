package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// TestLangConsistency_PresetsEnJaFileSetMatches asserts that every
// preset directory has identical file-name sets under `en/` and `ja/`.
// This is the cheap pre-doctor guard for translation drift (Task 12).
// If a future translator adds an en file without the ja counterpart
// (or vice versa) this test fires loudly with the missing path.
func TestLangConsistency_PresetsEnJaFileSetMatches(t *testing.T) {
	for _, preset := range []string{"minimal", "standard", "flutter", "typescript"} {
		preset := preset
		t.Run(preset, func(t *testing.T) {
			enFiles := collectTemplateFiles(t, "presets/"+preset+"/en")
			jaFiles := collectTemplateFiles(t, "presets/"+preset+"/ja")
			compareFileSets(t, preset, enFiles, jaFiles)
		})
	}
}

// TestLangConsistency_MemoryEnJaFileSetMatches asserts the same for
// the memory templates that compose with every preset.
func TestLangConsistency_MemoryEnJaFileSetMatches(t *testing.T) {
	enFiles := collectTemplateFiles(t, "memory/en")
	jaFiles := collectTemplateFiles(t, "memory/ja")
	compareFileSets(t, "memory", enFiles, jaFiles)
}

// TestLangConsistency_AiToolsEnJaFileSetMatches asserts the same for
// generated AI-tool artifacts.
func TestLangConsistency_AiToolsEnJaFileSetMatches(t *testing.T) {
	for _, tool := range []string{"claude", "cursor"} {
		tool := tool
		t.Run(tool, func(t *testing.T) {
			enFiles := collectTemplateFiles(t, "ai_tools/"+tool+"/en")
			jaFiles := collectTemplateFiles(t, "ai_tools/"+tool+"/ja")
			compareFileSets(t, "ai_tools/"+tool, enFiles, jaFiles)
		})
	}
}

func collectTemplateFiles(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	fsys, err := templates.FS()
	if err != nil {
		t.Fatalf("templates.FS: %v", err)
	}
	out := make(map[string]struct{})
	err = fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".tmpl") {
			return nil
		}
		out[strings.TrimPrefix(p, root+"/")] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

func compareFileSets(t *testing.T, label string, en, ja map[string]struct{}) {
	t.Helper()
	for name := range en {
		if _, ok := ja[name]; !ok {
			t.Errorf("[%s] %s exists in en/ but is missing in ja/", label, name)
		}
	}
	for name := range ja {
		if _, ok := en[name]; !ok {
			t.Errorf("[%s] %s exists in ja/ but is missing in en/", label, name)
		}
	}
}

// TestLangFallback_UnknownLangFallsBackToEn confirms that requesting a
// language without templates falls back to en rather than failing.
// The fallback notice goes to opts.Stdout; we capture it to assert the
// message is surfaced.
func TestLangFallback_UnknownLangFallsBackToEn(t *testing.T) {
	tmp := t.TempDir()
	var buf strings.Builder
	opts := defaultOpts(tmp)
	opts.Lang = "fr"
	opts.Stdout = &buf
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), `language "fr" not available`) {
		t.Errorf("expected fallback notice in stdout, got: %q", buf.String())
	}
	// Output should still be en — verify README.md exists and contains
	// the English-only `Read first` header.
	body, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(body), "Read first") {
		t.Errorf("expected en README content; got:\n%s", body)
	}
}
