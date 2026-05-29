package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Filename is the leaf name of the aikata project config file.
const Filename = "aikata.yaml"

// PrimaryDir is the aikata-owned configuration directory (per ADR
// 0008). `aikata init` writes here.
const PrimaryDir = ".aikata"

// PrimaryPath returns the absolute (or root-relative) path to the
// config file under root.
func PrimaryPath(root string) string {
	return filepath.Join(root, PrimaryDir, Filename)
}

// Resolve picks the config file aikata should read from a project at
// root. v0.7.4 removed the pre-v0.3.2 `.ai/aikata.yaml` fallback, so
// only `.aikata/aikata.yaml` is accepted.
func Resolve(root string) (path string, err error) {
	primary := PrimaryPath(root)
	if exists, statErr := fileExists(primary); statErr != nil {
		return "", statErr
	} else if exists {
		return primary, nil
	}

	return "", fs.ErrNotExist
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
