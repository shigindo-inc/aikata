package scaffold

import "sort"

// PresetInfo carries the user-facing metadata for one preset. It is
// curated by hand (not derived from the template tree) so descriptions
// and reserved-name semantics are explicit rather than inferred.
type PresetInfo struct {
	// Name is the identifier accepted by `aikata init --preset`.
	Name string
	// Description is a one-line English summary (no trailing period in
	// the canonical form; callers add punctuation when rendering).
	Description string
	// Status is "active" for presets that init can scaffold today, or
	// "reserved" for names held for a future release (no templates yet).
	Status string
	// Languages lists the locale codes a preset ships templates for.
	// Empty for "reserved" presets.
	Languages []string
}

// Status values used by PresetInfo.Status. StatusActive marks a preset
// that init can scaffold today; StatusReserved holds a name for a
// future release.
const (
	StatusActive   = "active"
	StatusReserved = "reserved"
)

// Presets returns the curated preset list, sorted alphabetically by
// name. The slice is rebuilt on every call so callers can mutate it
// freely without affecting the registry.
func Presets() []PresetInfo {
	out := []PresetInfo{
		{
			Name:        "minimal",
			Description: "Minimal preset: AGENTS.md plus the smallest README/SPEC/ARCH/GLOSSARY scaffold to bootstrap an aikata project.",
			Status:      StatusActive,
			Languages:   []string{"en", "ja"},
		},
		{
			Name:        "standard",
			Description: "Standard preset: the full canonical document set including ROADMAP, CHANGELOG, ADR templates, and docs/decisions/.",
			Status:      StatusActive,
			Languages:   []string{"en", "ja"},
		},
		{
			Name:        "flutter",
			Description: "Standard preset + Flutter stack brief at docs/stacks/flutter.md.",
			Status:      StatusActive,
			Languages:   []string{"en", "ja"},
		},
		{
			Name:        "typescript",
			Description: "Standard preset + TypeScript stack brief at docs/stacks/typescript.md.",
			Status:      StatusActive,
			Languages:   []string{"en", "ja"},
		},
		{
			Name:        "extended",
			Description: "Reserved for the v1.0 operational-readiness pack (CONTRIBUTING.md, SECURITY.md, CODE_OF_CONDUCT.md, GitHub issue/PR templates). Not selectable in v0.3.x.",
			Status:      StatusReserved,
		},
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// FindPreset returns the PresetInfo for name, or (zero, false) if no
// such preset is registered. Lookups are case-sensitive — preset names
// are lowercase by convention.
func FindPreset(name string) (PresetInfo, bool) {
	for _, p := range Presets() {
		if p.Name == name {
			return p, true
		}
	}
	return PresetInfo{}, false
}

// Stacks returns the stack identifiers currently bundled with the
// canonical templates. v0.3.x ships flutter and typescript via the
// matching presets; v0.4 introduces `aikata add stack` which will
// extend this list.
func Stacks() []string {
	return []string{"flutter", "typescript"}
}

// ActivePresetNames returns the names of presets whose Status is
// StatusActive, in registry order. Used by code paths that only want
// to enumerate scaffold-capable presets (e.g. init flag validation).
func ActivePresetNames() []string {
	infos := Presets()
	out := make([]string, 0, len(infos))
	for _, p := range infos {
		if p.Status == StatusActive {
			out = append(out, p.Name)
		}
	}
	return out
}
