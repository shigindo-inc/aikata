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
	if got.Docs.TaskFileLocation != "docs/tasks/current.md" {
		t.Errorf("Docs.TaskFileLocation = %q, want %q", got.Docs.TaskFileLocation, "docs/tasks/current.md")
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
		"version: 2",
		"project:",
		"name: samplekata",
		"lang: en",
		"ai_tools:",
		"- claude",
		"components:",
		"memory: false",
		"ui: false",
		"api: false",
		"tdd: false",
		"changelog: false",
		"monorepo: false",
		"features:",
		"docs:",
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

func TestMarshal_OmitsDoctorWhenEmpty(t *testing.T) {
	y := Default("samplekata", "en")
	body, err := Marshal(y)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(body), "doctor:") {
		t.Errorf("default marshal should omit empty doctor block; got:\n%s", body)
	}
}

func TestMarshalRoundtrip_PreservesDoctorExclude(t *testing.T) {
	original := Default("roundtrip", "en")
	original.Doctor.Exclude = []string{"plugins/**", "**/SKILL.md"}
	body, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(body), "doctor:") {
		t.Errorf("doctor block should be marshalled when Exclude is set; got:\n%s", body)
	}
	parsed, err := Unmarshal(body)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got, want := parsed.Doctor.Exclude, original.Doctor.Exclude; len(got) != len(want) {
		t.Fatalf("Doctor.Exclude length = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i, p := range original.Doctor.Exclude {
		if parsed.Doctor.Exclude[i] != p {
			t.Errorf("Doctor.Exclude[%d] = %q, want %q", i, parsed.Doctor.Exclude[i], p)
		}
	}
}

// TestUnmarshal_ToleratesRemovedGenerateGitignore covers ADR 0025 D3:
// the inert `docs.generate_gitignore` field was removed, but old configs
// that still carry the key must keep parsing (the unknown key is
// ignored, not rejected).
func TestUnmarshal_ToleratesRemovedGenerateGitignore(t *testing.T) {
	legacy := "version: 2\n" +
		"project:\n  name: x\n  lang: en\n" +
		"docs:\n  generate_gitignore: false\n  task_file_location: docs/tasks/current.md\n"
	got, err := Unmarshal([]byte(legacy))
	if err != nil {
		t.Fatalf("legacy config with generate_gitignore should still parse: %v", err)
	}
	if got.Docs.TaskFileLocation != "docs/tasks/current.md" {
		t.Errorf("task_file_location lost: %q", got.Docs.TaskFileLocation)
	}
}

// TestMarshalRoundtrip_PreservesSyncOwn covers ADR 0025 D2: the optional
// `sync.own` glob list round-trips and is omitted when empty.
func TestMarshalRoundtrip_PreservesSyncOwn(t *testing.T) {
	empty := Default("x", "en")
	body, err := Marshal(empty)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(body), "sync:") {
		t.Errorf("empty sync.own should be omitted; got:\n%s", body)
	}

	owned := Default("x", "en")
	owned.Sync.Own = []string{"README.md", ".gitignore"}
	body, err = Marshal(owned)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	parsed, err := Unmarshal(body)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(parsed.Sync.Own) != 2 || parsed.Sync.Own[0] != "README.md" || parsed.Sync.Own[1] != ".gitignore" {
		t.Errorf("sync.own lost in roundtrip: %v", parsed.Sync.Own)
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
