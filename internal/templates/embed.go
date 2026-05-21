package templates

import (
	"embed"
	"fmt"
	"io/fs"
)

// fsys is the raw embedded filesystem. The on-disk layout is
//
//	internal/templates/data/...
//
// and is exposed externally via FS() rooted at "data". Keeping the embed
// root inside this package avoids `..`-relative embed paths, which the
// Go `embed` directive rejects.
//
//go:embed all:data
var fsys embed.FS

// FS returns the embedded template filesystem rooted at the `data`
// directory, so callers see paths like `presets/minimal/README.md.tmpl`
// rather than `data/presets/minimal/README.md.tmpl`.
//
// The sub-filesystem error is structurally impossible (the path is a
// compile-time constant the embed directive guarantees to exist), but we
// surface it as a returned error rather than a panic to honor
// AGENTS.md §4-9 ("No panic in non-test code paths").
func FS() (fs.FS, error) {
	sub, err := fs.Sub(fsys, "data")
	if err != nil {
		return nil, fmt.Errorf("templates: sub-fs %q: %w", "data", err)
	}
	return sub, nil
}
