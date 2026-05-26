package config

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// ManifestFilename is the leaf name of the per-project sync manifest
// (`.aikata/manifest.yaml`). Recorded at `aikata init` time and read
// by `aikata sync` (v0.5) to perform a 3-way diff-merge against
// upstream templates. Hand editing is not supported; the file is
// regenerated on every successful sync.
const ManifestFilename = "manifest.yaml"

// ManifestVersion is the schema version of `.aikata/manifest.yaml`.
// Bumped only on incompatible shape changes; v0.5 ships with v1.
const ManifestVersion = 1

// Manifest records which template-derived files aikata wrote into a
// project at init time, along with a content hash for each. `aikata
// sync` uses these hashes as the common ancestor in its 3-way merge
// (ADR 0011 D4). The file lives at `.aikata/manifest.yaml` and is
// committed to the project's source repo so other contributors can
// reproduce the same merge.
type Manifest struct {
	Version int            `yaml:"version"`
	Preset  string         `yaml:"preset"`
	Lang    string         `yaml:"lang"`
	Files   []ManifestFile `yaml:"files"`
}

// ManifestFile is one entry in the manifest's file list. Path is the
// project-root-relative slash-separated path of the rendered artifact;
// SHA256 is the lowercase hex digest of the rendered content.
type ManifestFile struct {
	Path   string `yaml:"path"`
	SHA256 string `yaml:"sha256"`
}

// ManifestPath returns the absolute (or root-relative) path of the
// per-project manifest file under root.
func ManifestPath(root string) string {
	return filepath.Join(root, PrimaryDir, ManifestFilename)
}

// HashContent returns the canonical hex-encoded SHA-256 of body. This
// is the only hashing helper aikata exposes for manifest entries;
// keeping it here makes the hash algorithm one decision instead of one
// per call site.
func HashContent(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// BuildManifest constructs a Manifest from preset / lang / a map of
// rendered files keyed by target-relative path. The returned manifest
// has Files sorted alphabetically by Path so the YAML output is
// deterministic across runs.
//
// excludePaths lists target-relative paths that must NOT appear in the
// manifest — typically `.aikata/aikata.yaml` (mutable config) and the
// manifest itself (would be circular). Unknown paths in excludePaths
// are silently ignored.
func BuildManifest(preset, lang string, rendered map[string]string, excludePaths []string) Manifest {
	excluded := make(map[string]struct{}, len(excludePaths))
	for _, p := range excludePaths {
		excluded[p] = struct{}{}
	}
	files := make([]ManifestFile, 0, len(rendered))
	for path, body := range rendered {
		if _, skip := excluded[path]; skip {
			continue
		}
		files = append(files, ManifestFile{
			Path:   path,
			SHA256: HashContent([]byte(body)),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Manifest{
		Version: ManifestVersion,
		Preset:  preset,
		Lang:    lang,
		Files:   files,
	}
}

// MarshalManifest serializes m to YAML. Field order is driven by the
// struct definition; the Files slice is expected to be pre-sorted by
// the caller (BuildManifest does this).
func MarshalManifest(m Manifest) ([]byte, error) {
	out, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("config: marshal manifest.yaml: %w", err)
	}
	return out, nil
}

// UnmarshalManifest parses a manifest YAML payload. The Version field
// is required: a zero value is treated as a parse error so callers can
// distinguish "no manifest yet" (fs.ErrNotExist on Load) from "garbled
// manifest on disk" (Unmarshal error).
func UnmarshalManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("config: parse manifest.yaml: %w", err)
	}
	if m.Version == 0 {
		return Manifest{}, fmt.Errorf("config: manifest.yaml is missing the required %q field", "version")
	}
	return m, nil
}

// LoadManifest reads .aikata/manifest.yaml under root and returns the
// parsed manifest. If the file is missing, returns (Manifest{},
// fs.ErrNotExist) so callers can distinguish "no manifest" from an
// I/O failure.
func LoadManifest(root string) (Manifest, error) {
	path := ManifestPath(root)
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Manifest{}, fs.ErrNotExist
		}
		return Manifest{}, fmt.Errorf("config: read manifest.yaml: %w", err)
	}
	return UnmarshalManifest(body)
}

// SaveManifest writes m to .aikata/manifest.yaml under root via an
// atomic temp-file rename so the destination never sees a partial
// write. Creates the .aikata/ directory if it does not exist.
func SaveManifest(root string, m Manifest) error {
	path := ManifestPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: mkdir for manifest.yaml: %w", err)
	}
	body, err := MarshalManifest(m)
	if err != nil {
		return err
	}
	if err := writeAtomic(path, body, 0o644); err != nil {
		return fmt.Errorf("config: write manifest.yaml: %w", err)
	}
	return nil
}
