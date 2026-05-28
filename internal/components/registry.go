package components

import "sort"

// capabilities is the set of components exposed by
// `aikata enable <capability>` (v0.7.1, ADR 0017). Each capability
// represents a persistent project feature: schema-v2 component flags
// (memory, ui, api, tdd, changelog, monorepo), stack registrations,
// and AI-tool enrolments.
var capabilities = []Component{
	AITool,
	API,
	Changelog,
	Memory,
	Monorepo,
	Stack,
	TDD,
	UI,
}

// artifacts is the set of components exposed by
// `aikata new <artifact>` (v0.7.1, ADR 0017). Each artifact is an
// authoring scaffold (ADRs, future templates) that adds a single
// file without flipping any durable project flag.
var artifacts = []Component{
	Adr,
}

// Capabilities returns the enable-tier components sorted by Name.
// The slice is freshly built so callers can mutate freely.
func Capabilities() []Component {
	out := append([]Component(nil), capabilities...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Artifacts returns the new-tier components sorted by Name.
func Artifacts() []Component {
	out := append([]Component(nil), artifacts...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// All returns every registered component (capabilities + artifacts)
// sorted by Name. Retained as a transitional convenience for
// internals that don't yet care about the enable/new split.
func All() []Component {
	combined := make([]Component, 0, len(capabilities)+len(artifacts))
	combined = append(combined, capabilities...)
	combined = append(combined, artifacts...)
	sort.Slice(combined, func(i, j int) bool { return combined[i].Name() < combined[j].Name() })
	return combined
}

// GetCapability returns the capability registered under name, or
// (nil, false) if no such capability exists. Lookups are
// case-sensitive; capability names are lowercase by convention.
func GetCapability(name string) (Component, bool) {
	for _, c := range capabilities {
		if c.Name() == name {
			return c, true
		}
	}
	return nil, false
}

// GetArtifact returns the artifact registered under name, or
// (nil, false) if no such artifact exists.
func GetArtifact(name string) (Component, bool) {
	for _, c := range artifacts {
		if c.Name() == name {
			return c, true
		}
	}
	return nil, false
}

// ActiveCapabilityNames returns the names of active capabilities,
// sorted. Used by the cli layer's unknown-name error messages.
func ActiveCapabilityNames() []string {
	var out []string
	for _, c := range capabilities {
		if c.Status() == StatusActive {
			out = append(out, c.Name())
		}
	}
	sort.Strings(out)
	return out
}

// ActiveArtifactNames returns the names of active artifacts, sorted.
func ActiveArtifactNames() []string {
	var out []string
	for _, c := range artifacts {
		if c.Status() == StatusActive {
			out = append(out, c.Name())
		}
	}
	sort.Strings(out)
	return out
}
