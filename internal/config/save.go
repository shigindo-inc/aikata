package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Save persists cfg to the primary config path under root. The write
// is atomic (temp + rename) so a partial failure never leaves a
// half-written file.
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

// Load reads and parses `.aikata/aikata.yaml` for root via Resolve.
// Returns os.ErrNotExist when the path is missing, mirroring Resolve's
// contract.
func Load(root string) (AikataYaml, error) {
	path, err := Resolve(root)
	if err != nil {
		return AikataYaml{}, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return AikataYaml{}, fmt.Errorf("config: load %s: %w", path, err)
	}
	cfg, err := Unmarshal(body)
	if err != nil {
		return AikataYaml{}, err
	}
	return cfg, nil
}
