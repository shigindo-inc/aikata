package templates

import (
	"errors"
	"fmt"
	"io/fs"
)

// LangDir picks the language-scoped subdirectory under base inside the
// embedded FS, falling back to "en" when lang is empty or the requested
// lang's directory is missing. It is the single source of truth for
// the per-locale layout under `data/<base>/<lang>/`.
//
// Returns (langDir, fellBack, err) where:
//   - langDir is the joined embed path (e.g. "memory/en") on success
//   - fellBack is true when lang was non-empty and non-"en" but no
//     directory existed for it, so "en" was used instead
//   - err is non-nil only when base/en is itself missing (a templates
//     packaging bug) or when fs.Stat surfaces a non-NotExist error
//
// The function is side-effect-free; callers that want to surface the
// fallback to the user emit their own notice using the returned
// fellBack flag.
func LangDir(base, lang string) (string, bool, error) {
	root, err := FS()
	if err != nil {
		return "", false, err
	}
	chosen := lang
	if chosen == "" {
		chosen = "en"
	}
	target := base + "/" + chosen
	if _, err := fs.Stat(root, target); err == nil {
		return target, false, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", false, fmt.Errorf("templates: stat %s: %w", target, err)
	}
	fallback := base + "/en"
	if _, err := fs.Stat(root, fallback); err != nil {
		return "", false, fmt.Errorf("templates: %s/en not found: %w", base, err)
	}
	return fallback, chosen != "en", nil
}
