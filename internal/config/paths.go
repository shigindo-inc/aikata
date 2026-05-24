package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Filename is the leaf name of the aikata project config file. Same
// across the primary and legacy directories.
const Filename = "aikata.yaml"

// PrimaryDir is the v0.3.2+ aikata-owned configuration directory
// (per ADR 0008). New `aikata init` writes here.
const PrimaryDir = ".aikata"

// LegacyDir is the v0.2 / v0.3.0 / v0.3.1 directory that aikata
// continues to read for backwards compatibility. New writes never
// target it; aikata generate and aikata doctor surface a deprecation
// notice when this path is the only one present.
const LegacyDir = ".ai"

// PrimaryPath returns the absolute (or root-relative) path to the
// v0.3.2+ config file under root.
func PrimaryPath(root string) string {
	return filepath.Join(root, PrimaryDir, Filename)
}

// LegacyPath returns the path to the pre-v0.3.2 config file under
// root. Read-only for new code paths.
func LegacyPath(root string) string {
	return filepath.Join(root, LegacyDir, Filename)
}

// Resolve picks the config file aikata should read from a project at
// root. It prefers the primary path (.aikata/aikata.yaml) when both
// exist, falls back to the legacy path (.ai/aikata.yaml) otherwise,
// and returns os.ErrNotExist when neither is present.
//
// The isLegacy return is true only when the legacy path was selected
// because the primary was absent. Callers use it to emit deprecation
// warnings or trigger auto-migration.
func Resolve(root string) (path string, isLegacy bool, err error) {
	primary := PrimaryPath(root)
	if exists, statErr := fileExists(primary); statErr != nil {
		return "", false, statErr
	} else if exists {
		return primary, false, nil
	}

	legacy := LegacyPath(root)
	if exists, statErr := fileExists(legacy); statErr != nil {
		return "", false, statErr
	} else if exists {
		return legacy, true, nil
	}

	return "", false, fs.ErrNotExist
}

// fileExists reports whether path is an existing regular file. Errors
// other than ENOENT are propagated so the caller can distinguish
// "missing" from "unreadable".
func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.Mode().IsRegular(), nil
}
