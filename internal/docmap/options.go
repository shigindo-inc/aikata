package docmap

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
)

// OptionsFor assembles the Build/Generate options for targetDir, reading
// the optional `docmap:` block from `.aikata/aikata.yaml` for targets,
// exclude, and formats. managedGlobs is injected by the caller (the CLI
// and doctor both pass doctor.ManagedIncludeGlobs) so docmap never needs
// to import doctor — keeping the dependency one-directional so doctor's
// freshness check can import docmap.
//
// It is the single place that maps config to Options, so the explicit
// `aikata map` verb, the lifecycle hooks, and the doctor freshness check
// all build the map identically (no drift between writer and checker).
// An absent config uses the built-in defaults.
func OptionsFor(targetDir string, managedGlobs []string) (Options, error) {
	opts := Options{TargetDir: targetDir, ManagedGlobs: managedGlobs}
	cfg, err := config.Load(targetDir)
	if errors.Is(err, fs.ErrNotExist) {
		return opts, nil
	}
	if err != nil {
		return Options{}, fmt.Errorf("load doc map config: %w", err)
	}
	for _, format := range cfg.Docmap.Formats {
		name := strings.ToLower(strings.TrimSpace(format))
		if !slices.ContainsFunc(renderers, func(r renderer) bool { return r.name == name }) {
			return Options{}, fmt.Errorf("unknown docmap format %q", format)
		}
	}
	opts.Targets = cfg.Docmap.Targets
	opts.Exclude = cfg.Docmap.Exclude
	opts.Formats = cfg.Docmap.Formats
	opts.Enabled = cfg.Docmap.Enabled
	return opts, nil
}
