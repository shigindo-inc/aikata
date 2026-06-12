package docmap

import "github.com/shigindo-inc/aikata/internal/config"

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
// A missing or unreadable config falls back to the built-in defaults.
func OptionsFor(targetDir string, managedGlobs []string) Options {
	opts := Options{TargetDir: targetDir, ManagedGlobs: managedGlobs}
	cfg, err := config.Load(targetDir)
	if err != nil {
		return opts
	}
	opts.Targets = cfg.Docmap.Targets
	opts.Exclude = cfg.Docmap.Exclude
	opts.Formats = cfg.Docmap.Formats
	return opts
}
