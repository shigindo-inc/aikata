package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// MoveLegacyToPrimary moves .ai/aikata.yaml to .aikata/aikata.yaml
// under root. The move is atomic: bytes are written to the new path
// first, fsynced, then the legacy file is unlinked. The empty .ai/
// directory is removed afterwards so it does not linger in `ls -a`.
//
// Returns moved=true only when bytes were actually moved on disk.
// The contract is intentionally conservative:
//
//   - If primary already exists, returns (false, nil) without touching
//     anything (caller can read the primary and ignore the legacy).
//   - If legacy is missing, returns (false, fs.ErrNotExist).
//   - On I/O error the function aborts and the legacy file remains
//     untouched; the caller can re-run safely.
func MoveLegacyToPrimary(root string) (bool, error) {
	primary := PrimaryPath(root)
	legacy := LegacyPath(root)

	if exists, err := fileExists(primary); err != nil {
		return false, fmt.Errorf("config: migrate: stat primary: %w", err)
	} else if exists {
		return false, nil
	}

	body, err := os.ReadFile(legacy)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fs.ErrNotExist
		}
		return false, fmt.Errorf("config: migrate: read legacy: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(primary), 0o755); err != nil {
		return false, fmt.Errorf("config: migrate: mkdir .aikata: %w", err)
	}

	if err := writeAtomic(primary, body, 0o644); err != nil {
		return false, fmt.Errorf("config: migrate: write primary: %w", err)
	}

	if err := os.Remove(legacy); err != nil {
		return true, fmt.Errorf("config: migrate: unlink legacy: %w", err)
	}

	// Best-effort cleanup of the legacy directory; ignore any error
	// (the user may have other files under .ai/ we should not touch).
	legacyDir := filepath.Dir(legacy)
	if entries, statErr := os.ReadDir(legacyDir); statErr == nil && len(entries) == 0 {
		_ = os.Remove(legacyDir)
	}
	return true, nil
}

// writeAtomic writes body to path via a temp file + rename so the
// destination never sees a partial write. Mode mirrors any existing
// file at path or defaults to perm.
func writeAtomic(path string, body []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".aikata-migrate-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		// Rename consumes the temp file; ignore the not-exist case.
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
