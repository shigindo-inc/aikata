package templates

import "testing"

func TestLangDir_KnownLang(t *testing.T) {
	got, fellBack, err := LangDir("memory", "en")
	if err != nil {
		t.Fatalf("LangDir: %v", err)
	}
	if got != "memory/en" {
		t.Errorf("LangDir = %q, want memory/en", got)
	}
	if fellBack {
		t.Error("expected fellBack=false for known lang")
	}
}

func TestLangDir_EmptyDefaultsToEn(t *testing.T) {
	got, fellBack, err := LangDir("memory", "")
	if err != nil {
		t.Fatalf("LangDir: %v", err)
	}
	if got != "memory/en" {
		t.Errorf("LangDir = %q, want memory/en", got)
	}
	if fellBack {
		t.Error("expected fellBack=false when lang is empty (default applied silently)")
	}
}

func TestLangDir_UnknownFallsBackToEn(t *testing.T) {
	got, fellBack, err := LangDir("memory", "xx")
	if err != nil {
		t.Fatalf("LangDir: %v", err)
	}
	if got != "memory/en" {
		t.Errorf("LangDir = %q, want memory/en", got)
	}
	if !fellBack {
		t.Error("expected fellBack=true for unknown lang")
	}
}

func TestLangDir_MissingBaseSurfacedAsError(t *testing.T) {
	if _, _, err := LangDir("does-not-exist", "en"); err == nil {
		t.Error("expected error when base directory is missing")
	}
}
