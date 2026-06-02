package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runDoctor(t *testing.T) (string, error) {
	t.Helper()
	return runDoctorArgs(t)
}

func runDoctorArgs(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestDoctor_HealthyAfterInitExitsZero(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	out, err := runDoctor(t)
	if err != nil {
		t.Fatalf("doctor on fresh standard preset should succeed, got %v\n%s", err, out)
	}
}

func TestDoctor_BrokenAGENTSLinkSetsExitCodeThree(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Append a known-bad relative link to AGENTS.md.
	agentsPath := filepath.Join(tmp, "AGENTS.md")
	body, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read agents: %v", err)
	}
	body = append(body, []byte("\n\n[broken](./no-such-file.md)\n")...)
	if err := os.WriteFile(agentsPath, body, 0o644); err != nil {
		t.Fatalf("write agents: %v", err)
	}

	out, err := runDoctor(t)
	if err == nil {
		t.Fatalf("doctor should return error for broken link, got nil\n%s", out)
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("expected ExitError code 3, got %v", err)
	}
	if !strings.Contains(out, "broken link") {
		t.Errorf("expected broken-link diagnostic in output:\n%s", out)
	}
}

// TestDoctor_StrictTreatsWarningsAsErrors covers the v0.5+ dogfood
// gate. A project with a stale `updated:` frontmatter triggers a
// warning-level finding; --strict must turn that into exit 3 while
// the plain run continues to exit 0.
func TestDoctor_StrictTreatsWarningsAsErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// Plant a stale `updated:` value on README.md so the doctor's
	// `updated.stale` warning fires (>365 days from any plausible
	// current date). The exact `updated:` value scaffold writes
	// depends on the test's wall clock, so rewrite the frontmatter
	// rather than pattern-match on the date.
	readmePath := filepath.Join(tmp, "README.md")
	stale := []byte("---\nproject: samplekata\nstatus: draft\nversion: 0.0.1\nupdated: 2000-01-01\naudience: [human, agent]\n---\n\n# README.md\n")
	if err := os.WriteFile(readmePath, stale, 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	// Without --strict, warnings do not affect exit code.
	if _, err := runDoctor(t); err != nil {
		t.Errorf("non-strict doctor should not fail on a warning, got: %v", err)
	}

	// With --strict, the warning bumps exit to 3.
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--strict"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("strict doctor should fail on a warning, got nil\n%s", buf.String())
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("strict doctor: expected ExitError code 3, got %v", err)
	}
	if !strings.Contains(err.Error(), "warning") {
		t.Errorf("strict error message should mention warnings: %v", err)
	}
}

// TestDoctor_ConfigExcludeSuppressesPluginErrors covers ADR 0021 under
// the v0.9.7 broad-audit mode. A SKILL.md under plugins/ (Claude Code
// plugin layout, frontmatter is `name`+`description` only) produces
// frontmatter errors under --all-markdown. Adding
// `doctor.exclude: ["plugins/**"]` must make even the broad audit
// exit 0. (Under the default managed-surface scope the plugin tree is
// already out of scope — see TestDoctor_DefaultScopeIgnoresThirdPartyMarkdown.)
func TestDoctor_ConfigExcludeSuppressesPluginErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	pluginDir := filepath.Join(tmp, "plugins", "job-search", "skills", "mock")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	skillBody := []byte("---\nname: mock\ndescription: practice interviews.\n---\n\n# Mock\n")
	if err := os.WriteFile(filepath.Join(pluginDir, "SKILL.md"), skillBody, 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// Without exclude, the broad audit → expect exit 3.
	if _, err := runDoctorArgs(t, "--all-markdown"); err == nil {
		t.Fatalf("doctor --all-markdown should fail on plugins/SKILL.md without exclude")
	}

	// Append doctor.exclude to .aikata/aikata.yaml and rerun.
	cfgPath := filepath.Join(tmp, ".aikata", "aikata.yaml")
	cfgBody, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read aikata.yaml: %v", err)
	}
	cfgBody = append(cfgBody, []byte("doctor:\n  exclude:\n    - plugins/**\n")...)
	if err := os.WriteFile(cfgPath, cfgBody, 0o644); err != nil {
		t.Fatalf("write aikata.yaml: %v", err)
	}

	out, err := runDoctorArgs(t, "--all-markdown")
	if err != nil {
		t.Fatalf("doctor --all-markdown should succeed with exclude, got %v\n%s", err, out)
	}
	if strings.Contains(out, "plugins/") {
		t.Errorf("excluded path should not appear in doctor output:\n%s", out)
	}
}

// seedThirdPartyMarkdown writes a Claude Code plugin SKILL.md (foreign
// frontmatter contract) and a project-owned CONTRIBUTING.md with no
// aikata frontmatter — the two file classes the v0.9.7 adoption report
// flagged. Returns their absolute paths.
func seedThirdPartyMarkdown(t *testing.T, root string) (skill, contributing string) {
	t.Helper()
	skillDir := filepath.Join(root, "skills", "foo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	skill = filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skill, []byte("---\nname: foo\ndescription: a skill.\n---\n\n# Foo\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	contributing = filepath.Join(root, "CONTRIBUTING.md")
	if err := os.WriteFile(contributing, []byte("# Contributing\n\nNo aikata frontmatter here.\n"), 0o644); err != nil {
		t.Fatalf("write CONTRIBUTING.md: %v", err)
	}
	return skill, contributing
}

// TestDoctor_DefaultScopeIgnoresThirdPartyMarkdown pins ADR 0037 D1:
// the default managed-surface walk does not validate third-party
// Markdown (skills/**) or project-owned prose (CONTRIBUTING.md), so a
// plain `doctor` exits 0 and `doctor --fix` leaves both byte-identical.
func TestDoctor_DefaultScopeIgnoresThirdPartyMarkdown(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	skill, contributing := seedThirdPartyMarkdown(t, tmp)
	skillBefore, _ := os.ReadFile(skill)
	contribBefore, _ := os.ReadFile(contributing)

	if out, err := runDoctor(t); err != nil {
		t.Fatalf("default doctor should exit 0 (third-party Markdown out of scope), got %v\n%s", err, out)
	}
	if out, err := runDoctorArgs(t, "--fix"); err != nil {
		t.Fatalf("default doctor --fix should exit 0, got %v\n%s", err, out)
	}

	skillAfter, _ := os.ReadFile(skill)
	contribAfter, _ := os.ReadFile(contributing)
	if string(skillAfter) != string(skillBefore) {
		t.Errorf("doctor --fix mutated third-party SKILL.md:\n%s", skillAfter)
	}
	if string(contribAfter) != string(contribBefore) {
		t.Errorf("doctor --fix mutated project-owned CONTRIBUTING.md:\n%s", contribAfter)
	}
}

// TestDoctor_AllMarkdownAuditsThirdPartyMarkdown pins the broad-audit
// opt-in: --all-markdown reports the same third-party files the default
// scope leaves alone (ADR 0037 D1).
func TestDoctor_AllMarkdownAuditsThirdPartyMarkdown(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "samplekata", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	seedThirdPartyMarkdown(t, tmp)

	out, err := runDoctorArgs(t, "--all-markdown")
	if err == nil {
		t.Fatalf("doctor --all-markdown should report third-party frontmatter errors, got nil\n%s", out)
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != 3 {
		t.Fatalf("expected ExitError code 3, got %v", err)
	}
	for _, needle := range []string{"skills/foo/SKILL.md", "CONTRIBUTING.md"} {
		if !strings.Contains(out, needle) {
			t.Errorf("broad audit should report %q:\n%s", needle, out)
		}
	}
}

func TestRootCmdShowsDoctorInHelp(t *testing.T) {
	cmd := newRootCmd("0.0.1-test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(buf.String(), "doctor") {
		t.Errorf("help does not mention `doctor`:\n%s", buf.String())
	}
}
