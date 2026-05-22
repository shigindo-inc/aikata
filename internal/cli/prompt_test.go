package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunPrompt_HappyPath(t *testing.T) {
	in := strings.NewReader("myproj\nminimal\ny\n")
	var out bytes.Buffer
	got, err := runPrompt(in, &out, promptResult{Preset: "standard"})
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Name != "myproj" {
		t.Errorf("Name = %q, want %q", got.Name, "myproj")
	}
	if got.Preset != "minimal" {
		t.Errorf("Preset = %q, want %q", got.Preset, "minimal")
	}
	if !got.WithMemory {
		t.Errorf("WithMemory = false, want true")
	}
	// Prompts should mention the things they ask about.
	prompts := out.String()
	for _, needle := range []string{"Project name", "Preset", "memory"} {
		if !strings.Contains(prompts, needle) {
			t.Errorf("prompt output missing %q:\n%s", needle, prompts)
		}
	}
}

func TestRunPrompt_KeepsDefaultsOnBlankInput(t *testing.T) {
	in := strings.NewReader("myproj\n\n\n")
	var out bytes.Buffer
	got, err := runPrompt(in, &out, promptResult{Preset: "standard", WithMemory: true})
	if err != nil {
		t.Fatalf("runPrompt: %v", err)
	}
	if got.Preset != "standard" {
		t.Errorf("blank preset input should keep default: got %q", got.Preset)
	}
	if !got.WithMemory {
		t.Errorf("blank memory input should keep WithMemory=true default")
	}
}

func TestRunPrompt_SkipsNameWhenAlreadySet(t *testing.T) {
	// Only two answers: preset and memory. Name is supplied via the
	// defaults, so the prompt must not ask for it.
	in := strings.NewReader("standard\nn\n")
	var out bytes.Buffer
	got, err := runPrompt(in, &out, promptResult{Name: "preset-name", Preset: "minimal"})
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

func TestRunPrompt_EmptyNameIsError(t *testing.T) {
	in := strings.NewReader("\nstandard\nn\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{})
	if err == nil {
		t.Fatalf("expected error for empty project name")
	}
}

func TestRunPrompt_UnknownPresetIsError(t *testing.T) {
	in := strings.NewReader("myproj\nflutter\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{Preset: "standard"})
	if err == nil {
		t.Fatalf("expected error for unknown preset")
	}
}

func TestRunPrompt_NumericPresetChoices(t *testing.T) {
	cases := map[string]string{
		"1": "standard",
		"2": "minimal",
	}
	for input, expected := range cases {
		in := strings.NewReader("myproj\n" + input + "\nn\n")
		var out bytes.Buffer
		got, err := runPrompt(in, &out, promptResult{Preset: "standard"})
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
		in := strings.NewReader("myproj\nstandard\n" + input + "\n")
		var out bytes.Buffer
		got, err := runPrompt(in, &out, promptResult{})
		if err != nil {
			t.Fatalf("runPrompt(%q): %v", input, err)
		}
		if got.WithMemory != expected {
			t.Errorf("memory %q → got %v, want %v", input, got.WithMemory, expected)
		}
	}
}

func TestRunPrompt_UnknownYesNoIsError(t *testing.T) {
	in := strings.NewReader("myproj\nstandard\nmaybe\n")
	var out bytes.Buffer
	_, err := runPrompt(in, &out, promptResult{})
	if err == nil {
		t.Fatalf("expected error for unknown yes/no value")
	}
}
