package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const contextTestConfig = "version: 2\nproject:\n  name: ctx\n  lang: en\nai_tools:\n  - claude\n  - codex\n"

// seedContextRepo writes a git repository whose CLAUDE.md imports an
// AGENTS.md larger than Codex's 32 KiB default.
func seedContextRepo(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		filepath.Join(".aikata", "aikata.yaml"): contextTestConfig,
		filepath.Join(".git", "HEAD"):           "ref: refs/heads/main\n",
		"CLAUDE.md":                             "@AGENTS.md\n",
		"AGENTS.md":                             strings.Repeat("Rule line.\n", 3500),
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

type contextJSON struct {
	Version int    `json:"version"`
	Kind    string `json:"kind"`
	Tools   []struct {
		Tool  string `json:"tool"`
		Files []struct {
			Path       string `json:"path"`
			ImportedBy string `json:"imported_by"`
		} `json:"files"`
	} `json:"tools"`
	Issues []struct {
		Level string `json:"level"`
		File  string `json:"file"`
		Code  string `json:"code"`
	} `json:"issues"`
	Summary struct {
		Warnings int `json:"warnings"`
	} `json:"summary"`
}

func TestDoctorContext_JSON(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	seedContextRepo(t, tmp)

	out, err := runDoctorArgs(t, "--context", "--json")
	if err != nil {
		t.Fatalf("doctor --context --json: %v\n%s", err, out)
	}
	var got contextJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not one JSON document: %v\n%s", err, out)
	}
	if got.Version != 1 || got.Kind != "context" {
		t.Errorf("envelope version=%d kind=%q", got.Version, got.Kind)
	}
	if len(got.Tools) != 2 || got.Tools[0].Tool != "claude" || got.Tools[1].Tool != "codex" {
		t.Fatalf("tools = %+v, want claude, codex from ai_tools", got.Tools)
	}
	// Rows above the temp dir depend on the machine; keep the repo's own.
	var claude []string
	for _, f := range got.Tools[0].Files {
		if !strings.HasPrefix(f.Path, "../") {
			claude = append(claude, f.Path+"<"+f.ImportedBy)
		}
	}
	if strings.Join(claude, ",") != "CLAUDE.md<,AGENTS.md<CLAUDE.md" {
		t.Errorf("claude rows in the repo = %q", claude)
	}
	if got.Summary.Warnings != 1 || len(got.Issues) == 0 || got.Issues[0].Code != "context.codex-over-limit" {
		t.Errorf("want the codex over-limit warning first; issues = %+v", got.Issues)
	}
}

func TestDoctorContext_StrictFailsOnWarning(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	seedContextRepo(t, tmp)

	_, err := runDoctorArgs(t, "--context", "--strict")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("--context --strict with a warning: err = %v, want exit 3", err)
	}
}

// TestDoctorContext_RejectsCheckFlags: --context reads files only, so
// flags that configure or fix the checks are refused before anything
// runs, and --fix leaves a fixable file untouched.
func TestDoctorContext_RejectsCheckFlags(t *testing.T) {
	for _, flag := range []string{"--fix", "--dry-run", "--all-markdown", "--verbose"} {
		t.Run(flag, func(t *testing.T) {
			tmp := t.TempDir()
			chdir(t, tmp)
			seedContextRepo(t, tmp)
			fixable := filepath.Join(tmp, "SPEC.md")
			const body = "# SPEC without frontmatter\n"
			if err := os.WriteFile(fixable, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}

			out, err := runDoctorArgs(t, "--context", flag)
			var ee *ExitError
			if !errors.As(err, &ee) || ee.Code != 2 {
				t.Fatalf("--context %s: err = %v, want exit 2\n%s", flag, err, out)
			}
			if !strings.Contains(err.Error(), "--context") || !strings.Contains(err.Error(), flag) {
				t.Errorf("error should name both flags: %v", err)
			}
			if strings.Contains(out, "claude:") {
				t.Errorf("nothing should be reported after the rejection:\n%s", out)
			}
			after, rerr := os.ReadFile(fixable)
			if rerr != nil || string(after) != body {
				t.Errorf("SPEC.md changed under --context %s: %q (%v)", flag, after, rerr)
			}
		})
	}
}

func TestDoctorContext_RequiresConfig(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	_, err := runDoctorArgs(t, "--context")
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 2 || !strings.Contains(err.Error(), "aikata.yaml") {
		t.Fatalf("--context without .aikata/aikata.yaml: err = %v, want exit 2 naming the file", err)
	}
}

func TestDoctorContext_TextListsTools(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	seedContextRepo(t, tmp)

	var buf bytes.Buffer
	cmd := newDoctorCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--context"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor --context: %v\n%s", err, buf.String())
	}
	for _, want := range []string{"claude: ", "codex: 1 file(s)", "AGENTS.md:", "[warning]"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("text output missing %q:\n%s", want, buf.String())
		}
	}
}
