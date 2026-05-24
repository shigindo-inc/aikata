package scaffold

import (
	"sort"
	"testing"
)

func TestPresets_SortedAndIncludesReservedExtended(t *testing.T) {
	got := Presets()
	if len(got) < 5 {
		t.Fatalf("Presets(): expected at least 5 entries, got %d", len(got))
	}

	names := make([]string, len(got))
	for i, p := range got {
		names[i] = p.Name
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("Presets() not alphabetically sorted: %v", names)
	}

	var extended *PresetInfo
	for i := range got {
		if got[i].Name == "extended" {
			extended = &got[i]
			break
		}
	}
	if extended == nil {
		t.Fatalf("Presets(): expected reserved 'extended' entry, got %v", names)
	}
	if extended.Status != StatusReserved {
		t.Fatalf("Presets(): 'extended' status = %q, want %q", extended.Status, StatusReserved)
	}
	if len(extended.Languages) != 0 {
		t.Fatalf("Presets(): reserved 'extended' should declare no languages, got %v", extended.Languages)
	}
}

func TestPresets_ActiveSetMatchesTemplateTree(t *testing.T) {
	wantActive := map[string]bool{"minimal": true, "standard": true, "flutter": true, "typescript": true}
	for _, p := range Presets() {
		if p.Status != StatusActive {
			continue
		}
		if !wantActive[p.Name] {
			t.Errorf("Presets(): unexpected active preset %q (not in templates/data/presets)", p.Name)
		}
		delete(wantActive, p.Name)
	}
	for missing := range wantActive {
		t.Errorf("Presets(): missing active preset %q", missing)
	}
}

func TestFindPreset(t *testing.T) {
	if p, ok := FindPreset("standard"); !ok || p.Name != "standard" {
		t.Fatalf("FindPreset(standard) = %v, %v; want active standard", p, ok)
	}
	if p, ok := FindPreset("extended"); !ok || p.Status != StatusReserved {
		t.Fatalf("FindPreset(extended) = %v, %v; want reserved entry", p, ok)
	}
	if _, ok := FindPreset("does-not-exist"); ok {
		t.Fatalf("FindPreset(does-not-exist): expected ok=false")
	}
}

func TestActivePresetNames(t *testing.T) {
	got := ActivePresetNames()
	want := []string{"flutter", "minimal", "standard", "typescript"}
	if len(got) != len(want) {
		t.Fatalf("ActivePresetNames() = %v, want %v", got, want)
	}
	gotSorted := append([]string(nil), got...)
	sort.Strings(gotSorted)
	for i := range want {
		if gotSorted[i] != want[i] {
			t.Fatalf("ActivePresetNames() sorted = %v, want %v", gotSorted, want)
		}
	}
}

func TestStacks(t *testing.T) {
	got := Stacks()
	if len(got) != 2 {
		t.Fatalf("Stacks() len = %d, want 2; got %v", len(got), got)
	}
	want := map[string]bool{"flutter": true, "typescript": true}
	for _, s := range got {
		if !want[s] {
			t.Errorf("Stacks(): unexpected entry %q", s)
		}
	}
}
