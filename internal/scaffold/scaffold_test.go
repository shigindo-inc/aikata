package scaffold

import (
	"bytes"
	"errors"
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
	// No extra files for the minimal preset.
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if got, want := len(entries), 3; got != want {
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
	if !strings.Contains(out, "Would write 3 file(s)") {
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
