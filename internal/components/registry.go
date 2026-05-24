package components

import "sort"

// registry holds every component the binary ships with. Order here is
// irrelevant; All() returns a fresh slice sorted by Name so callers
// can mutate freely.
var registry = []Component{
	Adr,
	Memory,
}

// All returns the full component list, sorted by Name. The slice is
// freshly built on every call.
func All() []Component {
	out := append([]Component(nil), registry...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Get returns the Component registered under name, or (nil, false) if
// no such component exists. Lookups are case-sensitive; component
// names are lowercase by convention.
func Get(name string) (Component, bool) {
	for _, c := range registry {
		if c.Name() == name {
			return c, true
		}
	}
	return nil, false
}

// ActiveNames returns the names of components whose Status is
// StatusActive, sorted alphabetically. Used by cli layers that only
// want add-able identifiers (e.g. unknown-name error messages).
func ActiveNames() []string {
	var out []string
	for _, c := range registry {
		if c.Status() == StatusActive {
			out = append(out, c.Name())
		}
	}
	sort.Strings(out)
	return out
}
