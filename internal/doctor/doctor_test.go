package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixedNow is the deterministic "now" used by tests so the
// stale-updated check fires (or doesn't) predictably.
func fixedNow() time.Time {
	return time.Date(2026, time.May, 23, 12, 0, 0, 0, time.UTC)
}

// scaffoldHealthyProject writes a minimal but doctor-clean tree to dir.
// Tests build on top of this and mutate one file to assert a single
// check.
func scaffoldHealthyProject(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"AGENTS.md": `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: agent
---

# Agent rules

See [SPEC.md](./SPEC.md) and [docs/adr/0001-foo.md](./docs/adr/0001-foo.md).
`,
		"SPEC.md": `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# Spec
`,
		"ARCHITECTURE.md": `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# Architecture

DATABASE_URL is required.
`,
		"GLOSSARY.md": `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# Glossary

## A

### agent

An LLM-driven coding assistant. AGENTS.md captures the rules.

### orphan-term

This term is only defined here.
`,
		".env.example": "DATABASE_URL=postgres://localhost/sample\n",
		"docs/adr/0001-foo.md": `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR 0001 — Foo

- **Status**: Accepted
`,
	}
	for rel, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", full, err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

func runDoctor(t *testing.T, dir string) []Issue {
	t.Helper()
	issues, err := Run(Options{TargetDir: dir, Now: fixedNow()})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return issues
}

func TestRun_RequiresTargetDir(t *testing.T) {
	_, err := Run(Options{})
	if err == nil {
		t.Fatalf("expected error for empty TargetDir")
	}
}

func TestRun_HealthyProjectHasNoErrors(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	issues := runDoctor(t, tmp)
	if HasErrors(issues) {
		t.Errorf("healthy project should have no error-level issues, got:\n%+v", issues)
	}
}

func TestCheckFrontmatter_MissingKeyIsError(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// Rewrite SPEC.md without `version:`.
	bad := `---
project: sample
status: draft
updated: 2026-05-20
audience: [human, agent]
---

# Spec
`
	if err := os.WriteFile(filepath.Join(tmp, "SPEC.md"), []byte(bad), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.File == "SPEC.md" && iss.Level == LevelError && strings.Contains(iss.Message, "version") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an error about missing version in SPEC.md, got:\n%+v", issues)
	}
}

func TestCheckFrontmatter_SkipsSerenaLocalState(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	serenaMemory := filepath.Join(tmp, ".serena", "memories", "core.md")
	if err := os.MkdirAll(filepath.Dir(serenaMemory), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(serenaMemory, []byte("# Local Serena memory\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	issues := runDoctor(t, tmp)
	for _, iss := range issues {
		if strings.HasPrefix(iss.File, ".serena/") {
			t.Fatalf("expected .serena files to be skipped, got issue: %+v", iss)
		}
	}
}

func TestCheckLinks_BrokenLinkIsError(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// Remove docs/adr/0001-foo.md so AGENTS.md's link to it dangles.
	if err := os.Remove(filepath.Join(tmp, "docs", "adr", "0001-foo.md")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.Level == LevelError && strings.Contains(iss.Message, "broken link") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected broken-link error, got:\n%+v", issues)
	}
}

func TestCheckADR_DeprecatedWithoutReplacementIsError(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	bad := `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR 0001 — Foo

- **Status**: Deprecated
`
	if err := os.WriteFile(filepath.Join(tmp, "docs", "adr", "0001-foo.md"), []byte(bad), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.Level == LevelError && strings.Contains(iss.Message, "Deprecated ADR") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected deprecated-ADR error, got:\n%+v", issues)
	}
}

func TestCheckMemory_TypeMismatchIsError(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	memDir := filepath.Join(tmp, "docs", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Name says user.md but memory_type is "feedback" — mismatch.
	bad := `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
memory_type: feedback
---

# Memory — user
`
	if err := os.WriteFile(filepath.Join(memDir, "user.md"), []byte(bad), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.Level == LevelError && strings.Contains(iss.Message, "does not match filename") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected memory_type mismatch error, got:\n%+v", issues)
	}
}

func TestCheckUpdated_StaleIsWarning(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// SPEC.md updated more than 365 days before fixedNow.
	stale := `---
project: sample
status: draft
version: 0.0.1
updated: 2024-01-01
audience: [human, agent]
---

# Spec
`
	if err := os.WriteFile(filepath.Join(tmp, "SPEC.md"), []byte(stale), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.File == "SPEC.md" && iss.Level == LevelWarning && strings.Contains(iss.Message, "365 days old") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected stale-updated warning on SPEC.md, got:\n%+v", issues)
	}
}

func TestCheckEnvExample_UnreferencedIsWarning(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// Append a variable that AGENTS.md / ARCHITECTURE.md don't mention.
	env := "DATABASE_URL=postgres://localhost/sample\nUNUSED_TOKEN=secret\n"
	if err := os.WriteFile(filepath.Join(tmp, ".env.example"), []byte(env), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.File == ".env.example" && iss.Level == LevelWarning && strings.Contains(iss.Message, "UNUSED_TOKEN") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected env unreferenced warning, got:\n%+v", issues)
	}
}

func TestCheckGlossary_UnusedTermIsInfo(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	issues := runDoctor(t, tmp)
	found := false
	for _, iss := range issues {
		if iss.File == "GLOSSARY.md" && iss.Level == LevelInfo && strings.Contains(iss.Message, "orphan-term") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected glossary unused-term info, got:\n%+v", issues)
	}
}

func TestCheckADRNumbering_DuplicateAndGapAreInfo(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)
	// Healthy scaffold already has 0001-foo.md. Add a duplicate of
	// number 0001 and a gap by jumping to 0003.
	adrBody := `---
project: sample
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR

- **Status**: Accepted
`
	if err := os.WriteFile(filepath.Join(tmp, "docs", "adr", "0001-bar.md"), []byte(adrBody), 0o644); err != nil {
		t.Fatalf("write dup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "docs", "adr", "0003-baz.md"), []byte(adrBody), 0o644); err != nil {
		t.Fatalf("write gap: %v", err)
	}

	issues := runDoctor(t, tmp)

	var foundDup, foundGap bool
	for _, iss := range issues {
		if iss.Level != LevelInfo {
			continue
		}
		if strings.Contains(iss.Message, "duplicate ADR number 0001") {
			foundDup = true
		}
		if strings.Contains(iss.Message, "ADR number 0002 is unused") {
			foundGap = true
		}
	}
	if !foundDup {
		t.Errorf("expected duplicate-number info issue, got:\n%+v", issues)
	}
	if !foundGap {
		t.Errorf("expected gap info issue for 0002, got:\n%+v", issues)
	}
	if HasErrors(issues) {
		t.Errorf("ADR numbering findings must stay at info level, got errors:\n%+v", issues)
	}
}

func TestExcludes_SuppressFrontmatterUpdatedAndGlossary(t *testing.T) {
	tmp := t.TempDir()
	scaffoldHealthyProject(t, tmp)

	// A SKILL.md without aikata frontmatter (Claude Code plugin
	// layout). Without exclusion this triggers
	// `frontmatter.missing` errors. It also contains an obvious
	// `orphan-term` reference; if the glossary check still sees
	// it, the existing TestCheckGlossary_UnusedTermIsInfo would
	// no longer find the info issue.
	pluginDir := filepath.Join(tmp, "plugins", "job-search", "skills", "mock")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	skillBody := `---
name: mock-interview
description: Use when the user practices interviews.
---

# Mock interview

orphan-term is mentioned here.
`
	if err := os.WriteFile(filepath.Join(pluginDir, "SKILL.md"), []byte(skillBody), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	// A stale-dated file under plugins/ would trip checkUpdated
	// without exclusion.
	staleBody := `---
name: legacy
description: legacy reference.
updated: 2024-01-01
---

# legacy
`
	if err := os.WriteFile(filepath.Join(pluginDir, "references.md"), []byte(staleBody), 0o644); err != nil {
		t.Fatalf("write references.md: %v", err)
	}

	// Sanity: without exclude, the SKILL.md error is reported.
	t.Run("without exclude reports errors", func(t *testing.T) {
		issues, err := Run(Options{TargetDir: tmp, Now: fixedNow()})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		var sawFrontmatter bool
		for _, iss := range issues {
			if strings.HasPrefix(iss.File, "plugins/") && iss.Level == LevelError {
				sawFrontmatter = true
			}
		}
		if !sawFrontmatter {
			t.Errorf("expected at least one error under plugins/ without exclude, got:\n%+v", issues)
		}
	})

	// With exclude, no plugins/ issue should remain at any level,
	// and the glossary check's orphan-term info should still fire
	// (the SKILL.md mention is excluded from the corpus).
	t.Run("with exclude suppresses all checks", func(t *testing.T) {
		issues, err := Run(Options{
			TargetDir: tmp,
			Now:       fixedNow(),
			Excludes:  []string{"plugins/**"},
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		var sawGlossaryOrphan bool
		for _, iss := range issues {
			if strings.HasPrefix(iss.File, "plugins/") {
				t.Errorf("excluded path still reported: %+v", iss)
			}
			if iss.File == "GLOSSARY.md" && iss.Level == LevelInfo && strings.Contains(iss.Message, "orphan-term") {
				sawGlossaryOrphan = true
			}
		}
		if !sawGlossaryOrphan {
			t.Errorf("excluded corpus should not satisfy glossary reference; expected orphan-term info to still fire, got:\n%+v", issues)
		}
	})
}

func TestFormat_RendersAllLevels(t *testing.T) {
	issues := []Issue{
		{Level: LevelError, File: "a.md", Line: 5, Message: "boom"},
		{Level: LevelWarning, File: "b.md", Message: "hmm"},
		{Level: LevelInfo, File: "c.md", Line: 1, Message: "fyi"},
	}
	var buf bytes.Buffer
	if err := Format(&buf, issues); err != nil {
		t.Fatalf("Format: %v", err)
	}
	want := []string{"a.md:5: [error] boom", "b.md: [warning] hmm", "c.md:1: [info] fyi"}
	for _, w := range want {
		if !strings.Contains(buf.String(), w) {
			t.Errorf("Format output missing %q:\n%s", w, buf.String())
		}
	}
}

func TestHasErrors(t *testing.T) {
	if HasErrors(nil) {
		t.Error("HasErrors(nil) should be false")
	}
	warnings := []Issue{{Level: LevelWarning, File: "x"}}
	if HasErrors(warnings) {
		t.Error("HasErrors on warnings-only should be false")
	}
	mixed := []Issue{{Level: LevelWarning, File: "x"}, {Level: LevelError, File: "y"}}
	if !HasErrors(mixed) {
		t.Error("HasErrors on mixed should be true")
	}
}
