package components

import (
	"io"
	"os"
	"sort"
)

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

// stdout returns the AddContext's stdout sink or os.Stdout when nil.
// Shared by every Component.Add implementation that needs to emit
// user-facing progress messages.
func stdout(ctx AddContext) io.Writer {
	if ctx.Stdout != nil {
		return ctx.Stdout
	}
	return os.Stdout
}

// stderr returns the AddContext's stderr sink or os.Stderr when nil.
func stderr(ctx AddContext) io.Writer {
	if ctx.Stderr != nil {
		return ctx.Stderr
	}
	return os.Stderr
}
