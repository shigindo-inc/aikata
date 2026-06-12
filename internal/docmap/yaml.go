package docmap

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/shigindo-inc/aikata/internal/config"
)

// YAMLFilename is the leaf name of the doc map's data layer under the
// aikata machine zone (`.aikata/docmap.yaml`).
const YAMLFilename = "docmap.yaml"

// YAMLPath returns the path of the doc map data file under root.
func YAMLPath(root string) string {
	return filepath.Join(root, config.PrimaryDir, YAMLFilename)
}

// MarshalYAML serializes the Map to its `.aikata/docmap.yaml` byte form.
// Field and document order are deterministic (struct order; Docs sorted
// by Build), so identical inputs yield byte-identical output — a hard
// requirement for diff-stable commits and the doctor freshness check
// (ADR 0044 D6, docmap-design.md §5).
func MarshalYAML(m Map) ([]byte, error) {
	out, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("docmap: marshal docmap.yaml: %w", err)
	}
	return out, nil
}
