package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runGenerate(t *testing.T) (string, error) {
	t.Helper()
	cmd := newGenerateCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	return buf.String(), err
}

func TestGenerate_RequiresAikataYaml(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runGenerate(t)
	if err == nil {
		t.Fatalf("expected error when .aikata/aikata.yaml is missing")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got: %v", err)
	}
}

func TestGenerate_AfterInitProducesCLAUDE(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	// Scaffold a standard preset (which seeds .aikata/aikata.yaml).
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Re-create the generate command in the same cwd; runGenerate
	// builds a fresh cobra command per call to avoid state bleed.
	out, err := runGenerate(t)
	if err != nil {
		t.Fatalf("generate: %v (out: %s)", err, out)
	}
	body, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not produced: %v", err)
	}
	for _, needle := range []string{
		"AGENTS.md",
		"canonical",
		"aikata generate",
		"samplekata",
		"SPEC.md",         // standard preset has SPEC.md
		"ARCHITECTURE.md", // standard preset has ARCHITECTURE.md
		"GLOSSARY.md",     // standard preset has GLOSSARY.md
	} {
		if !strings.Contains(string(body), needle) {
			t.Errorf("CLAUDE.md missing %q:\n%s", needle, body)
		}
	}
}

func TestGenerate_MinimalAfterInitProducesCLAUDE(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "minimal", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// minimal preset doesn't include .aikata/aikata.yaml, so generate
	// should fail with our exit-2 message. This documents that
	// `aikata generate` currently requires the standard preset (or a
	// hand-written .aikata/aikata.yaml) — refining this is a v0.2 concern.
	_, err := runGenerate(t)
	if err == nil {
		t.Fatalf("expected error: minimal does not include .aikata/aikata.yaml")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got: %v", err)
	}
}

func TestGenerate_CursorAndCodex(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Rewrite ai_tools to enable claude + cursor + codex.
	yamlPath := filepath.Join(tmp, ".aikata", "aikata.yaml")
	orig, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read aikata.yaml: %v", err)
	}
	rewritten := strings.Replace(string(orig),
		"ai_tools:\n    - claude\n",
		"ai_tools:\n    - claude\n    - cursor\n    - codex\n", 1)
	if rewritten == string(orig) {
		t.Fatalf("aikata.yaml format unexpected, could not enable cursor/codex:\n%s", orig)
	}
	if err := os.WriteFile(yamlPath, []byte(rewritten), 0o644); err != nil {
		t.Fatalf("rewrite aikata.yaml: %v", err)
	}

	out, err := runGenerate(t)
	if err != nil {
		t.Fatalf("generate: %v (out: %s)", err, out)
	}
	if !strings.Contains(out, "[codex] no files generated") {
		t.Errorf("expected codex no-op notice in output, got:\n%s", out)
	}
	for _, rel := range []string{"CLAUDE.md", filepath.Join(".cursor", "rules", "main.mdc")} {
		if _, err := os.Stat(filepath.Join(tmp, rel)); err != nil {
			t.Errorf("expected %s to exist after generate: %v", rel, err)
		}
	}
}

func TestGenerate_DoesNotReadAIConfig(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	newPath := filepath.Join(tmp, ".aikata", "aikata.yaml")
	oldDir := filepath.Join(tmp, ".ai")
	oldPath := filepath.Join(oldDir, "aikata.yaml")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	if err := os.Rename(newPath, oldPath); err != nil {
		t.Fatalf("rename to .ai: %v", err)
	}
	manifestPath := filepath.Join(tmp, ".aikata", "manifest.yaml")
	if err := os.Remove(manifestPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rm manifest.yaml: %v", err)
	}
	if err := os.Remove(filepath.Dir(newPath)); err != nil {
		t.Fatalf("rm empty .aikata dir: %v", err)
	}

	out, err := runGenerate(t)
	if err == nil {
		t.Fatalf("expected generate to fail without .aikata/aikata.yaml; out: %s", out)
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 {
		t.Fatalf("expected ExitError code 2, got: %v", err)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Errorf(".ai/aikata.yaml should be left untouched: %v", err)
	}
}

func TestRootCmdShowsGenerateInHelp(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(buf.String(), "generate") {
		t.Errorf("help does not mention `generate`:\n%s", buf.String())
	}
}
