package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// Default skip — every question is asked.
func noSkip() promptSkip { return promptSkip{} }

func TestRunPrompt_HappyPath(t *testing.T) {
	// Order: name, preset, lang, ai-tools, memory, ui, api, tdd, changelog, monorepo.
	in := strings.NewReader("myproj\nminimal\nja\nclaude,cursor\ny\ny\nn\ny\nn\nn\n")
	var out bytes.Buffer
	got, err := runPrompt(in, &out, promptResult{Preset: "standard", Lang: "en"}, noSkip())
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Name != "myproj" {
		t.Errorf("Name = %q, want %q", got.Name, "myproj")
	}
	if got.Preset != "minimal" {
		t.Errorf("Preset = %q, want %q", got.Preset, "minimal")
	}
	if got.Lang != "ja" {
		t.Errorf("Lang = %q, want %q", got.Lang, "ja")
	}
	if !reflect.DeepEqual(got.AITools, []string{"claude", "cursor"}) {
		t.Errorf("AITools = %v, want [claude cursor]", got.AITools)
	}
	if !got.WithMemory {
		t.Errorf("WithMemory = false, want true")
	}
	if !got.WithUI || got.WithAPI || !got.WithTDD || got.WithChangelog {
		t.Errorf("optional flags = %+v, want UI=true API=false TDD=true Changelog=false",
			[4]bool{got.WithUI, got.WithAPI, got.WithTDD, got.WithChangelog})
	}
	prompts := out.String()
	for _, needle := range []string{"Project name", "Preset", "Document language", "AI tools", "memory", "UI.md", "API.md", "testing.md", "CHANGELOG.md", "monorepo"} {
		if !strings.Contains(prompts, needle) {
			t.Errorf("prompt output missing %q:\n%s", needle, prompts)
		}
	}
}

func TestRunPrompt_KeepsDefaultsOnBlankInput(t *testing.T) {
	// Ten blanks (name + 4 existing + 4 v0.4 single-file + 1 v0.6
	// monorepo question) accept every default.
	in := strings.NewReader("myproj\n\n\n\n\n\n\n\n\n\n")
	var out bytes.Buffer
	defaults := promptResult{
		Preset:     "standard",
		Lang:       "en",
		AITools:    []string{"claude"},
		WithMemory: true,
	}
	got, err := runPrompt(in, &out, defaults, noSkip())
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Preset != "standard" {
		t.Errorf("Preset default lost: %q", got.Preset)
	}
	if got.Lang != "en" {
		t.Errorf("Lang default lost: %q", got.Lang)
	}
	if !reflect.DeepEqual(got.AITools, []string{"claude"}) {
		t.Errorf("AITools default lost: %v", got.AITools)
	}
	if !got.WithMemory {
		t.Errorf("WithMemory default lost")
	}
	if got.WithUI || got.WithAPI || got.WithTDD || got.WithChangelog {
		t.Errorf("new optional-component defaults clobbered to true: %+v",
			[4]bool{got.WithUI, got.WithAPI, got.WithTDD, got.WithChangelog})
	}
}

func TestRunPrompt_SkipsNameWhenAlreadySet(t *testing.T) {
	in := strings.NewReader("standard\nen\nclaude\nn\nn\nn\nn\nn\nn\n")
	var out bytes.Buffer
	got, err := runPrompt(in, &out, promptResult{Name: "preset-name", Preset: "minimal"}, noSkip())
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Name != "preset-name" {
		t.Errorf("expected supplied name to be kept, got %q", got.Name)
	}
	if strings.Contains(out.String(), "Project name") {
		t.Errorf("prompt asked for name even though one was supplied:\n%s", out.String())
	}
}

func TestRunPrompt_SkipsFieldsExplicitlySet(t *testing.T) {
	// Lang and AITools were pinned via flags; the prompt should ask
	// only about preset and memory.
	in := strings.NewReader("myproj\nstandard\nn\nn\nn\nn\nn\nn\n")
	var out bytes.Buffer
	defaults := promptResult{
		Lang:    "ja",
		AITools: []string{"claude", "codex"},
	}
	got, err := runPrompt(in, &out, defaults, promptSkip{Lang: true, AITools: true})
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Lang != "ja" {
		t.Errorf("Lang clobbered by skipped prompt: %q", got.Lang)
	}
	if !reflect.DeepEqual(got.AITools, []string{"claude", "codex"}) {
		t.Errorf("AITools clobbered by skipped prompt: %v", got.AITools)
	}
	if strings.Contains(out.String(), "Document language") {
		t.Errorf("Lang question rendered despite skip:\n%s", out.String())
	}
	if strings.Contains(out.String(), "AI tools") {
		t.Errorf("AI tools question rendered despite skip:\n%s", out.String())
	}
}

func TestRunPrompt_EmptyNameIsError(t *testing.T) {
	in := strings.NewReader("\nstandard\nen\nclaude\nn\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{}, noSkip())
	if err == nil {
		t.Fatalf("expected error for empty project name")
	}
}

func TestRunPrompt_UnknownPresetIsError(t *testing.T) {
	in := strings.NewReader("myproj\nrust\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{Preset: "standard"}, noSkip())
	if err == nil {
		t.Fatalf("expected error for unknown preset")
	}
}

func TestRunPrompt_UnknownLanguageIsError(t *testing.T) {
	in := strings.NewReader("myproj\nstandard\nfr\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{Preset: "standard"}, noSkip())
	if err == nil {
		t.Fatalf("expected error for unknown language")
	}
}

func TestRunPrompt_UnknownAIToolIsError(t *testing.T) {
	in := strings.NewReader("myproj\nstandard\nen\nclaude,jetbrains\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{Preset: "standard"}, noSkip())
	if err == nil {
		t.Fatalf("expected error for unknown ai-tool")
	}
}

func TestRunPrompt_NumericPresetChoices(t *testing.T) {
	cases := map[string]string{
		"1": "standard",
		"2": "minimal",
		"3": "flutter",
		"4": "typescript",
	}
	for input, expected := range cases {
		in := strings.NewReader("myproj\n" + input + "\nen\nclaude\nn\nn\nn\nn\nn\nn\n")
		var out bytes.Buffer
		got, err := runPrompt(in, &out, promptResult{Preset: "standard"}, noSkip())
		if err != nil {
			t.Fatalf("runPrompt(%q): %v", input, err)
		}
		if got.Preset != expected {
			t.Errorf("preset %q → got %q, want %q", input, got.Preset, expected)
		}
	}
}

func TestRunPrompt_MemoryYesNoVariants(t *testing.T) {
	cases := map[string]bool{
		"y":   true,
		"yes": true,
		"Y":   true,
		"YES": true,
		"n":   false,
		"no":  false,
		"N":   false,
	}
	for input, expected := range cases {
		in := strings.NewReader("myproj\nstandard\nen\nclaude\n" + input + "\nn\nn\nn\nn\nn\n")
		var out bytes.Buffer
		got, err := runPrompt(in, &out, promptResult{}, noSkip())
		if err != nil {
			t.Fatalf("runPrompt(%q): %v", input, err)
		}
		if got.WithMemory != expected {
			t.Errorf("memory %q → got %v, want %v", input, got.WithMemory, expected)
		}
	}
}

func TestRunPrompt_OptionalComponentSkipsHonored(t *testing.T) {
	// Pin every optional question via skip and provide no Y/N input
	// for them. runPrompt must traverse without asking, keeping the
	// supplied defaults intact.
	in := strings.NewReader("myproj\nstandard\nen\nclaude\n")
	var out bytes.Buffer
	defaults := promptResult{
		WithMemory:    true,
		WithUI:        true,
		WithAPI:       false,
		WithTDD:       true,
		WithChangelog: false,
	}
	skip := promptSkip{
		WithMemory:    true,
		WithUI:        true,
		WithAPI:       true,
		WithTDD:       true,
		WithChangelog: true,
		WithMonorepo:  true,
	}
	got, err := runPrompt(in, &out, defaults, skip)
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.WithMemory != true || got.WithUI != true || got.WithAPI != false ||
		got.WithTDD != true || got.WithChangelog != false {
		t.Errorf("skipped optional defaults clobbered: %+v", got)
	}
	for _, needle := range []string{"memory", "UI.md", "API.md", "testing.md", "CHANGELOG.md", "monorepo"} {
		if strings.Contains(out.String(), needle) {
			t.Errorf("expected %q question to be skipped; got:\n%s", needle, out.String())
		}
	}
}

func TestRunPrompt_UnknownYesNoOnNewOptionalIsError(t *testing.T) {
	in := strings.NewReader("myproj\nstandard\nen\nclaude\nn\nmaybe\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{}, noSkip())
	if err == nil {
		t.Fatalf("expected error for unknown yes/no on optional question")
	}
}

func TestRunPrompt_UnknownYesNoIsError(t *testing.T) {
	in := strings.NewReader("myproj\nstandard\nen\nclaude\nmaybe\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{}, noSkip())
	if err == nil {
		t.Fatalf("expected error for unknown yes/no value")
	}
}

func TestParseAITools_DedupAndSort(t *testing.T) {
	got, err := parseAITools("cursor, claude ,Cursor,codex")
	if err != nil {
		t.Fatalf("parseAITools: %v", err)
	}
	want := []string{"cursor", "claude", "codex"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v (insertion-order de-duplication)", got, want)
	}
}

func TestParseAITools_RejectsUnknown(t *testing.T) {
	if _, err := parseAITools("claude,foo"); err == nil {
		t.Fatalf("expected error for unknown tool")
	}
}

func TestParseAITools_RejectsEmpty(t *testing.T) {
	if _, err := parseAITools(""); err == nil {
		t.Fatalf("expected error for empty list")
	}
	if _, err := parseAITools("  ,  "); err == nil {
		t.Fatalf("expected error for whitespace-only list")
	}
}
