package components

import "sort"

// sortedKeys returns m's keys in lexicographic order. Shared between
// Component implementations that emit deterministic listings (dry-run,
// notice text).
func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
