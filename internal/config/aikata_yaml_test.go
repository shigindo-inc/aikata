package config

import (
	"strings"
	"testing"
)

func TestDefault_PopulatesRequiredFields(t *testing.T) {
	got := Default("samplekata", "en")
	if got.Version != Version {
		t.Errorf("Version = %d, want %d", got.Version, Version)
	}
	if got.Project.Name != "samplekata" {
		t.Errorf("Project.Name = %q, want %q", got.Project.Name, "samplekata")
	}
	if got.Project.Lang != "en" {
		t.Errorf("Project.Lang = %q, want %q", got.Project.Lang, "en")
	}
	if len(got.AITools) != 1 || got.AITools[0] != "claude" {
		t.Errorf("AITools = %v, want [claude]", got.AITools)
	}
	if !got.Docs.GenerateGitignore {
		t.Errorf("Docs.GenerateGitignore should default to true")
	}
}

func TestDefault_EmptyLangFallsBackToEn(t *testing.T) {
	got := Default("x", "")
	if got.Project.Lang != "en" {
		t.Errorf("Project.Lang = %q, want %q (empty → en)", got.Project.Lang, "en")
	}
}

func TestMarshal_StableShape(t *testing.T) {
	y := Default("samplekata", "en")
	body, err := Marshal(y)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	out := string(body)
	for _, needle := range []string{
		"version: 1",
		"project:",
		"name: samplekata",
		"lang: en",
		"ai_tools:",
		"- claude",
		"features:",
		"docs:",
		"generate_gitignore: true",
		"task_file_location: docs/tasks/current.md",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("marshal output missing %q:\n%s", needle, out)
		}
	}
}

func TestMarshalRoundtrip(t *testing.T) {
	original := Default("roundtrip", "ja")
	body, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	parsed, err := Unmarshal(body)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if parsed.Project.Name != original.Project.Name {
		t.Errorf("Name lost in roundtrip: got %q, want %q", parsed.Project.Name, original.Project.Name)
	}
	if parsed.Project.Lang != original.Project.Lang {
		t.Errorf("Lang lost in roundtrip: got %q, want %q", parsed.Project.Lang, original.Project.Lang)
	}
	if parsed.Version != Version {
		t.Errorf("Version lost in roundtrip: got %d, want %d", parsed.Version, Version)
	}
}

func TestUnmarshal_MissingVersionIsError(t *testing.T) {
	_, err := Unmarshal([]byte("project:\n  name: x\n  lang: en\n"))
	if err == nil {
		t.Fatalf("expected error for missing version")
	}
}

func TestUnmarshal_MalformedYamlIsError(t *testing.T) {
	_, err := Unmarshal([]byte("this: is: not: yaml"))
	if err == nil {
		t.Fatalf("expected error for malformed yaml")
	}
}
