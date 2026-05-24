package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Save persists cfg to the primary config path under root. The write
// is atomic (temp + rename) so a partial failure never leaves a
// half-written file. Save always writes to the primary path
// (.aikata/aikata.yaml) even when a legacy file (.ai/aikata.yaml)
// exists at the same root — callers that want to migrate the legacy
// file first should run MoveLegacyToPrimary explicitly.
//
// Save creates the .aikata/ directory if needed (0755). On success
// the new file inherits 0644.
func Save(root string, cfg AikataYaml) error {
	body, err := Marshal(cfg)
	if err != nil {
		return err
	}
	path := PrimaryPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: save: mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := writeAtomic(path, body, 0o644); err != nil {
		return fmt.Errorf("config: save: write %s: %w", path, err)
	}
	return nil
}

// Load reads and parses the aikata config for root via Resolve, so
// both `.aikata/aikata.yaml` and the legacy `.ai/aikata.yaml` are
// accepted. The isLegacy return is true when the legacy path was
// used; callers may surface a deprecation notice or schedule a
// migration based on it.
//
// Returns os.ErrNotExist when neither path is present, mirroring
// Resolve's contract.
func Load(root string) (cfg AikataYaml, isLegacy bool, err error) {
	path, isLegacy, err := Resolve(root)
	if err != nil {
		return AikataYaml{}, false, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return AikataYaml{}, false, fmt.Errorf("config: load %s: %w", path, err)
	}
	cfg, err = Unmarshal(body)
	if err != nil {
		return AikataYaml{}, false, err
	}
	return cfg, isLegacy, nil
}
