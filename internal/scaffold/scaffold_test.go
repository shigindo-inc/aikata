package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixedClock returns a deterministic time for golden-stable rendering.
func fixedClock() func() time.Time {
	t := time.Date(2026, time.May, 21, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func defaultOpts(target string) Options {
	return Options{
		ProjectName: "samplekata",
		Preset:      "minimal",
		TargetDir:   target,
		Clock:       fixedClock(),
	}
}

func TestRun_GeneratesAllMinimalFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(defaultOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, name := range []string{"README.md", "AGENTS.md", "SPEC.md"} {
		path := filepath.Join(tmp, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", name)
		}
	}
	// Minimal preset writes its three markdown files plus the
	// `.aikata/manifest.yaml` sync ancestor (ADR 0011 D4).
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if got, want := len(entries), 4; got != want {
		t.Errorf("entry count = %d, want %d (%v)", got, want, entries)
	}
}

func TestRun_ProjectNameReflectedInOutput(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(defaultOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(tmp, "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	if !strings.Contains(string(readme), "samplekata") {
		t.Errorf("README does not contain ProjectName:\n%s", readme)
	}
}

func TestRun_NonEmptyDirWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "preexisting.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	err := Run(defaultOpts(tmp))
	if !errors.Is(err, ErrTargetDirNotEmpty) {
		t.Fatalf("expected ErrTargetDirNotEmpty, got %v", err)
	}
	// Confirm scaffold did not write anything.
	if _, err := os.Stat(filepath.Join(tmp, "README.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("README.md should not exist after failed run: %v", err)
	}
}

func TestRun_NonEmptyDirWithForce(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "preexisting.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	opts := defaultOpts(tmp)
	opts.Force = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run with --force: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "README.md")); err != nil {
		t.Errorf("README.md not written: %v", err)
	}
	// Pre-existing file is preserved.
	if _, err := os.Stat(filepath.Join(tmp, "preexisting.txt")); err != nil {
		t.Errorf("preexisting.txt should remain: %v", err)
	}
}

func TestRun_DryRunWritesNothing(t *testing.T) {
	tmp := t.TempDir()
	opts := defaultOpts(tmp)
	opts.DryRun = true
	var buf bytes.Buffer
	opts.Stdout = &buf
	if err := Run(opts); err != nil {
		t.Fatalf("Run dry-run: %v", err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run should not write any files, found %v", entries)
	}
	out := buf.String()
	if !strings.Contains(out, "Would write 4 file(s)") {
		t.Errorf("dry-run output missing summary line: %q", out)
	}
	for _, name := range []string{"README.md", "AGENTS.md", "SPEC.md"} {
		if !strings.Contains(out, name) {
			t.Errorf("dry-run output missing %s: %q", name, out)
		}
	}
}

func TestRun_UnknownPreset(t *testing.T) {
	tmp := t.TempDir()
	opts := defaultOpts(tmp)
	opts.Preset = "does-not-exist"
	err := Run(opts)
	if err == nil {
		t.Fatalf("expected error for unknown preset")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error should mention the preset name, got: %v", err)
	}
}

func TestRun_ValidateRequiredFields(t *testing.T) {
	tmp := t.TempDir()
	cases := []struct {
		name string
		opts Options
	}{
		{
			name: "missing project name",
			opts: Options{Preset: "minimal", TargetDir: tmp},
		},
		{
			name: "missing preset",
			opts: Options{ProjectName: "x", TargetDir: tmp},
		},
		{
			name: "missing target dir",
			opts: Options{ProjectName: "x", Preset: "minimal"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := Run(c.opts); err == nil {
				t.Fatalf("expected validation error for case %q", c.name)
			}
		})
	}
}

func TestRun_NoLocalPathsInOutput(t *testing.T) {
	// OSS readiness: generated files must not contain /Users/ or other
	// host-specific absolute paths leaked from templates.
	tmp := t.TempDir()
	if err := Run(defaultOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, name := range []string{"README.md", "AGENTS.md", "SPEC.md"} {
		body, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(body), "/Users/") {
			t.Errorf("%s contains a /Users/ path (OSS leak)", name)
		}
		for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
			if strings.Contains(string(body), sec) {
				t.Errorf("%s contains secret-like pattern %q", name, sec)
			}
		}
	}
}

// --- standard preset ---

func standardOpts(target string) Options {
	o := defaultOpts(target)
	o.Preset = "standard"
	return o
}

// expectedStandardFiles lists every file the standard preset must
// produce, expressed as forward-slash paths relative to TargetDir.
var expectedStandardFiles = []string{
	"AGENTS.md",
	"ARCHITECTURE.md",
	"GLOSSARY.md",
	"README.md",
	"ROADMAP.md",
	"SPEC.md",
	".env.example",
	".gitignore",
	".aikata/aikata.yaml",
	"docs/adr/0001-record-architecture-decisions.md",
	"docs/prompts.md",
	"docs/tasks/current.md",
	"docs/troubleshooting.md",
}

func TestRun_GeneratesAllStandardFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(standardOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedStandardFiles {
		full := filepath.Join(tmp, filepath.FromSlash(rel))
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", rel)
		}
	}
}

func TestRun_StandardAikataYamlContent(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(standardOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read .aikata/aikata.yaml: %v", err)
	}
	out := string(body)
	for _, needle := range []string{
		"version: 2",
		"name: samplekata",
		"lang: en",
		"- claude",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("aikata.yaml missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_StandardGitignoreMentionsGeneratedArtifacts(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(standardOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	out := string(body)
	for _, needle := range []string{"CLAUDE.md", ".cursor/rules/", ".aikata-proposed/"} {
		if !strings.Contains(out, needle) {
			t.Errorf(".gitignore missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_StandardOSSReadiness(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(standardOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedStandardFiles {
		body, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if strings.Contains(string(body), "/Users/") {
			t.Errorf("%s contains a /Users/ path (OSS leak)", rel)
		}
		for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
			if strings.Contains(string(body), sec) {
				t.Errorf("%s contains secret-like pattern %q", rel, sec)
			}
		}
	}
}

// --- flutter preset (Task 10) ---

func flutterOpts(target string) Options {
	o := defaultOpts(target)
	o.Preset = "flutter"
	o.Stacks = []string{"flutter"}
	return o
}

var expectedFlutterFiles = []string{
	"AGENTS.md",
	"ARCHITECTURE.md",
	"GLOSSARY.md",
	"README.md",
	"ROADMAP.md",
	"SPEC.md",
	".env.example",
	".gitignore",
	".aikata/aikata.yaml",
	"docs/adr/0001-record-architecture-decisions.md",
	"docs/prompts.md",
	"docs/stacks/flutter.md",
	"docs/tasks/current.md",
	"docs/troubleshooting.md",
}

func TestRun_GeneratesAllFlutterFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(flutterOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedFlutterFiles {
		full := filepath.Join(tmp, filepath.FromSlash(rel))
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", rel)
		}
	}
}

func TestRun_FlutterAikataYamlIncludesStack(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(flutterOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read .aikata/aikata.yaml: %v", err)
	}
	out := string(body)
	for _, needle := range []string{"version: 2", "name: samplekata", "stacks:", "- flutter"} {
		if !strings.Contains(out, needle) {
			t.Errorf("aikata.yaml missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_FlutterAGENTSReferencesStackDoc(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(flutterOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(body), "docs/stacks/flutter.md") {
		t.Errorf("flutter AGENTS.md should reference docs/stacks/flutter.md:\n%s", body)
	}
}

func TestRun_FlutterGitignoreCoversFlutterArtifacts(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(flutterOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	out := string(body)
	for _, needle := range []string{".dart_tool/", "build/", "ios/Pods/", "pubspec.lock"} {
		if !strings.Contains(out, needle) {
			t.Errorf(".gitignore missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_FlutterOSSReadiness(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(flutterOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedFlutterFiles {
		body, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if strings.Contains(string(body), "/Users/") {
			t.Errorf("%s contains a /Users/ path (OSS leak)", rel)
		}
		for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
			if strings.Contains(string(body), sec) {
				t.Errorf("%s contains secret-like pattern %q", rel, sec)
			}
		}
	}
}

// TestRun_NonStack_NoStackFootprint is the Do-No-Harm regression gate
// (ADR 0003): minimal / standard presets must not contain any
// stack-specific identifier (`flutter`, `typescript`, or
// `docs/stacks/` references).
func TestRun_NonStack_NoStackFootprint(t *testing.T) {
	for _, preset := range []string{"minimal", "standard"} {
		preset := preset
		t.Run(preset, func(t *testing.T) {
			tmp := t.TempDir()
			opts := defaultOpts(tmp)
			opts.Preset = preset
			if err := Run(opts); err != nil {
				t.Fatalf("Run: %v", err)
			}
			err := filepath.WalkDir(tmp, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				lower := strings.ToLower(string(body))
				for _, needle := range []string{"flutter", "typescript", "docs/stacks/"} {
					if strings.Contains(lower, needle) {
						rel, _ := filepath.Rel(tmp, path)
						t.Errorf("%s mentions %q in non-stack preset:\n%s",
							rel, needle, body)
					}
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}
		})
	}
}

// --- typescript preset (Task 11) ---

func typescriptOpts(target string) Options {
	o := defaultOpts(target)
	o.Preset = "typescript"
	o.Stacks = []string{"typescript"}
	return o
}

var expectedTypescriptFiles = []string{
	"AGENTS.md",
	"ARCHITECTURE.md",
	"GLOSSARY.md",
	"README.md",
	"ROADMAP.md",
	"SPEC.md",
	".env.example",
	".gitignore",
	".aikata/aikata.yaml",
	"docs/adr/0001-record-architecture-decisions.md",
	"docs/prompts.md",
	"docs/stacks/typescript.md",
	"docs/tasks/current.md",
	"docs/troubleshooting.md",
}

func TestRun_GeneratesAllTypescriptFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(typescriptOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedTypescriptFiles {
		full := filepath.Join(tmp, filepath.FromSlash(rel))
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", rel)
		}
	}
}

func TestRun_TypescriptAikataYamlIncludesStack(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(typescriptOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read .aikata/aikata.yaml: %v", err)
	}
	out := string(body)
	for _, needle := range []string{"version: 2", "name: samplekata", "stacks:", "- typescript"} {
		if !strings.Contains(out, needle) {
			t.Errorf("aikata.yaml missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_TypescriptAGENTSReferencesStackDoc(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(typescriptOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(body), "docs/stacks/typescript.md") {
		t.Errorf("typescript AGENTS.md should reference docs/stacks/typescript.md:\n%s", body)
	}
}

func TestRun_TypescriptGitignoreCoversNodeArtifacts(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(typescriptOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	out := string(body)
	for _, needle := range []string{"node_modules/", "dist/", "*.tsbuildinfo", "coverage/"} {
		if !strings.Contains(out, needle) {
			t.Errorf(".gitignore missing %q:\n%s", needle, out)
		}
	}
}

func TestRun_TypescriptOSSReadiness(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(typescriptOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedTypescriptFiles {
		body, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if strings.Contains(string(body), "/Users/") {
			t.Errorf("%s contains a /Users/ path (OSS leak)", rel)
		}
		for _, sec := range []string{"AKIA", "ghp_", "sk-", "xoxb-"} {
			if strings.Contains(string(body), sec) {
				t.Errorf("%s contains secret-like pattern %q", rel, sec)
			}
		}
	}
}

// --- --with-memory (Task 5A) ---

var expectedMemoryFiles = []string{
	"docs/memory/README.md",
	"docs/memory/user.md",
	"docs/memory/feedback.md",
	"docs/memory/project.md",
	"docs/memory/reference.md",
}

func TestRun_WithMemory_GeneratesMemoryFilesUnderMinimal(t *testing.T) {
	tmp := t.TempDir()
	opts := defaultOpts(tmp)
	opts.WithMemory = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range expectedMemoryFiles {
		full := filepath.Join(tmp, filepath.FromSlash(rel))
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", rel)
		}
	}
}

func TestRun_WithMemory_MemoryTypeMatchesFilename(t *testing.T) {
	tmp := t.TempDir()
	opts := defaultOpts(tmp)
	opts.WithMemory = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	cases := map[string]string{
		"docs/memory/user.md":      "memory_type: user",
		"docs/memory/feedback.md":  "memory_type: feedback",
		"docs/memory/project.md":   "memory_type: project",
		"docs/memory/reference.md": "memory_type: reference",
	}
	for rel, want := range cases {
		body, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if !strings.Contains(string(body), want) {
			t.Errorf("%s missing %q in frontmatter", rel, want)
		}
	}
}

func TestRun_WithMemory_StandardAGENTSReferencesMemory(t *testing.T) {
	tmp := t.TempDir()
	opts := standardOpts(tmp)
	opts.WithMemory = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	agents, err := os.ReadFile(filepath.Join(tmp, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	body := string(agents)
	if !strings.Contains(body, "docs/memory/") {
		t.Errorf("standard AGENTS.md should reference docs/memory/ when WithMemory:\n%s", body)
	}
	if !strings.Contains(body, "Recall user / project context") {
		t.Errorf("standard AGENTS.md should include the Recall navigation row")
	}
}

// TestRun_WithoutMemory_NoMemoryFootprint is the Do-No-Harm regression
// gate (ADR 0003 / 0004): with WithMemory=false the generated tree must
// not contain `docs/memory/` and no rendered file may mention the
// directory or the slot.
func TestRun_WithoutMemory_NoMemoryFootprint(t *testing.T) {
	for _, preset := range []string{"minimal", "standard"} {
		preset := preset
		t.Run(preset, func(t *testing.T) {
			tmp := t.TempDir()
			opts := defaultOpts(tmp)
			opts.Preset = preset
			opts.WithMemory = false
			if err := Run(opts); err != nil {
				t.Fatalf("Run: %v", err)
			}
			if _, err := os.Stat(filepath.Join(tmp, "docs", "memory")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("docs/memory/ should not exist when WithMemory=false: %v", err)
			}
			err := filepath.WalkDir(tmp, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(string(body), "docs/memory/") {
					rel, _ := filepath.Rel(tmp, path)
					t.Errorf("%s mentions docs/memory/ even though WithMemory=false:\n%s",
						rel, body)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}
		})
	}
}

// TestRun_ForcedInit_MergesExistingGitignore pins the v0.7.2
// managed-append integration (ADR 0018): when --force is set and
// the target already contains a `.gitignore`, the scaffold writes
// the aikata-owned block in place via internal/managed.ApplyBlock,
// preserving every user-owned line outside the markers.
func TestRun_ForcedInit_MergesExistingGitignore(t *testing.T) {
	tmp := t.TempDir()
	userBefore := []byte("# user content\nuser.local\n.cache/\n")
	if err := os.WriteFile(filepath.Join(tmp, ".gitignore"), userBefore, 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	opts := standardOpts(tmp)
	opts.Force = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(tmp, ".gitignore"))
	if err != nil {
		t.Fatalf("read merged .gitignore: %v", err)
	}
	got := string(body)
	if !strings.Contains(got, "user.local") {
		t.Errorf("user content was dropped by managed-append merge:\n%s", got)
	}
	if !strings.Contains(got, "# >>> aikata managed >>>") {
		t.Errorf("aikata block marker missing from merged .gitignore:\n%s", got)
	}
	if !strings.Contains(got, "/.aikata-proposed/") {
		t.Errorf("aikata-owned entry missing from merged .gitignore:\n%s", got)
	}
}
