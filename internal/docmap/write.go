package docmap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shigindo-inc/aikata/internal/config"
)

// Generate builds the document map for opts.TargetDir and writes the
// configured renderings under the aikata machine zone. The data layer
// (`docmap.yaml`) is always written; the readable view (`docmap.md`) and
// the optional `json` / `txt` / `mmd` renderers follow opts.Formats
// (default [yaml, md]). Each file is written atomically (temp file +
// rename) so a reader never sees a partial map; the `.aikata/` directory
// is created when absent.
//
// Generate is the single rebuild entry point shared by the explicit
// `aikata map` verb and the isolated lifecycle hooks; it never touches
// the manifest or sync state — freshness is guaranteed by the doctor
// check, not by merge (ADR 0044 D6).
func Generate(opts Options) error {
	m, err := Build(opts)
	if err != nil {
		return err
	}

	dir := filepath.Join(opts.TargetDir, config.PrimaryDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("docmap: mkdir %s: %w", dir, err)
	}

	want := resolveFormats(opts.Formats)
	for _, r := range renderers {
		if !contains(want, r.name) {
			continue
		}
		body, err := r.body(m)
		if err != nil {
			return err
		}
		if err := writeAtomic(pathFor(opts.TargetDir, r.file), body, 0o644); err != nil {
			return fmt.Errorf("docmap: write %s: %w", r.file, err)
		}
	}
	return nil
}

func contains(set []string, s string) bool {
	for _, v := range set {
		if v == s {
			return true
		}
	}
	return false
}

// writeAtomic writes body to path via a temp file + rename so the
// destination never sees a partial write. It mirrors
// internal/config.writeAtomic, kept local so docmap does not depend on
// an unexported config helper.
func writeAtomic(path string, body []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".aikata-docmap-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
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
