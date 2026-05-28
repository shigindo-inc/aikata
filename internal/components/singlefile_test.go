package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// singleFileCases enumerates every Component that uses the singleFile
// helper. Adding a new optional file-emitting component is one line
// here plus its component file; the test body never changes.
var singleFileCases = []struct {
	name       string
	component  Component
	targetPath string
}{
	{"ui", UI, "UI.md"},
	{"api", API, "API.md"},
	{"tdd", TDD, "docs/testing.md"},
	{"changelog", Changelog, "CHANGELOG.md"},
}

func TestSingleFile_HappyPathWritesTargetWithFrontmatter(t *testing.T) {
	for _, tc := range singleFileCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			var stdout bytes.Buffer
			err := tc.component.Add(AddContext{
				TargetDir:   tmp,
				ProjectName: "demo",
				Lang:        "en",
				Clock:       fixedClock,
				Stdout:      &stdout,
			})
			if err != nil {
				t.Fatalf("Add: %v", err)
			}
			full := filepath.Join(tmp, filepath.FromSlash(tc.targetPath))
			body, err := os.ReadFile(full)
			if err != nil {
				t.Fatalf("read %s: %v", tc.targetPath, err)
			}
			s := string(body)
			for _, key := range []string{"project: demo", "status: draft", "version: 0.0.1", "updated: 2026-05-24", "audience: [human, agent]"} {
				if !strings.Contains(s, key) {
					t.Errorf("rendered file missing %q\n--- body ---\n%s", key, s)
				}
			}
			if !strings.Contains(stdout.String(), "wrote "+tc.targetPath) {
				t.Errorf("expected progress message; got %q", stdout.String())
			}
		})
	}
}

func TestSingleFile_JaRendersJaTemplate(t *testing.T) {
	for _, tc := range singleFileCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			err := tc.component.Add(AddContext{
				TargetDir:   tmp,
				ProjectName: "demo",
				Lang:        "ja",
				Clock:       fixedClock,
				Stdout:      &bytes.Buffer{},
			})
			if err != nil {
				t.Fatalf("Add: %v", err)
			}
			body, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(tc.targetPath)))
			if err != nil {
				t.Fatalf("read %s: %v", tc.targetPath, err)
			}
			s := string(body)
			// Each ja template carries at least one Japanese-script
			// section heading; spot-check on hiragana/katakana/kanji
			// presence is enough to distinguish from the en variant.
			if !containsAny(s, "ガイドライン", "エンドポイント", "戦略", "重要な変更点") {
				t.Errorf("expected ja-language content in rendered %s; got:\n%s", tc.targetPath, s)
			}
		})
	}
}

func TestSingleFile_AlreadyPresentIsNoOp(t *testing.T) {
	for _, tc := range singleFileCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			full := filepath.Join(tmp, filepath.FromSlash(tc.targetPath))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(full, []byte("custom"), 0o644); err != nil {
				t.Fatalf("seed: %v", err)
			}
			var stderr bytes.Buffer
			err := tc.component.Add(AddContext{
				TargetDir:   tmp,
				ProjectName: "demo",
				Lang:        "en",
				Clock:       fixedClock,
				Stdout:      &bytes.Buffer{},
				Stderr:      &stderr,
			})
			if err != nil {
				t.Fatalf("Add: %v", err)
			}
			body, _ := os.ReadFile(full)
			if string(body) != "custom" {
				t.Errorf("user file was overwritten; got %q", string(body))
			}
			if !strings.Contains(stderr.String(), "already present") {
				t.Errorf("expected idempotent notice; got %q", stderr.String())
			}
		})
	}
}

func TestSingleFile_DryRunDoesNotWrite(t *testing.T) {
	for _, tc := range singleFileCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			var stdout bytes.Buffer
			err := tc.component.Add(AddContext{
				TargetDir:   tmp,
				ProjectName: "demo",
				Lang:        "en",
				DryRun:      true,
				Clock:       fixedClock,
				Stdout:      &stdout,
			})
			if err != nil {
				t.Fatalf("Add: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(tc.targetPath))); !os.IsNotExist(err) {
				t.Errorf("expected %s not to exist after --dry-run; got err=%v", tc.targetPath, err)
			}
			if !strings.Contains(stdout.String(), tc.targetPath) {
				t.Errorf("expected dry-run plan to name the target; got %q", stdout.String())
			}
		})
	}
}

func TestSingleFile_RequiresProjectName(t *testing.T) {
	for _, tc := range singleFileCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			err := tc.component.Add(AddContext{
				TargetDir: tmp,
				Lang:      "en",
				Stdout:    &bytes.Buffer{},
			})
			if err == nil {
				t.Fatal("expected error when project name is missing")
			}
		})
	}
}

func TestSingleFile_RegistryListsAllSingleFileNames(t *testing.T) {
	want := []string{"api", "changelog", "tdd", "ui"}
	got := ActiveCapabilityNames()
	for _, w := range want {
		if !containsString(got, w) {
			t.Errorf("ActiveCapabilityNames() missing %q; got %v", w, got)
		}
	}
}

func TestSingleFile_RenderShimsMatchAddOutput(t *testing.T) {
	cases := []struct {
		name   string
		render func(SingleFileParams) (map[string]string, error)
		target string
	}{
		{"ui", RenderUI, "UI.md"},
		{"api", RenderAPI, "API.md"},
		{"tdd", RenderTDD, "docs/testing.md"},
		{"changelog", RenderChangelog, "CHANGELOG.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.render(SingleFileParams{
				Lang:        "en",
				ProjectName: "demo",
				Clock:       fixedClock,
			})
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if len(out) != 1 {
				t.Fatalf("expected 1 file; got %d (%v)", len(out), keysOf(out))
			}
			if _, ok := out[tc.target]; !ok {
				t.Errorf("expected key %q; got %v", tc.target, keysOf(out))
			}
		})
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func containsString(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
