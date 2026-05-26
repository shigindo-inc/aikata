package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// AikataYamlMigrator transforms an aikata.yaml payload from one schema
// version into the next. Implementations operate on the raw YAML node
// graph rather than on the typed AikataYaml struct so they can read
// fields that no longer exist in the current Go type definitions.
type AikataYamlMigrator func(node *yaml.Node) error

// aikataYamlMigrators registers per-version forward migrations. A
// migrator at index N upgrades version N payloads in place to version
// N+1. v0.5 ships with the registry empty because v1 is the only
// schema in the wild; v2 will land here as a one-line addition.
//
// The registry sits at package scope so future migrations slot in
// without changing call-site code. Migrations run sequentially from
// the payload's current version up to Version.
var aikataYamlMigrators = map[int]AikataYamlMigrator{
	// Example shape, kept commented so the extension pattern is
	// obvious when v2 actually lands:
	//
	//   1: func(node *yaml.Node) error {
	//       // …mutate node so it satisfies the v2 schema…
	//       setVersionField(node, 2)
	//       return nil
	//   },
}

// ErrFutureSchema signals that the on-disk aikata.yaml carries a
// schema version newer than this binary knows how to read. The cobra
// layer maps this to an actionable "upgrade aikata" message.
var ErrFutureSchema = errors.New("config: aikata.yaml uses a newer schema than this aikata version supports")

// MigrateAikataYaml parses raw aikata.yaml bytes, runs any registered
// forward migrations, and returns the typed result. The returned
// bool is true when migrations were applied (callers can use it to
// decide whether to persist the upgraded form back to disk).
//
// The expected version progression is monotonic forward only:
//
//   - version == Version   → no-op, parse and return
//   - 0 < version < Version → run migrations[version], …, migrations[Version-1] in order
//   - version > Version    → ErrFutureSchema
//   - version == 0         → existing "missing version" error from
//     Unmarshal once the migrated payload is parsed
func MigrateAikataYaml(data []byte) (AikataYaml, bool, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return AikataYaml{}, false, fmt.Errorf("config: parse aikata.yaml for migration: %w", err)
	}
	currentVersion, err := readVersionField(&doc)
	if err != nil {
		return AikataYaml{}, false, err
	}
	if currentVersion > Version {
		return AikataYaml{}, false, fmt.Errorf("%w: payload version %d, supported up to version %d",
			ErrFutureSchema, currentVersion, Version)
	}
	migrated := false
	for v := currentVersion; v < Version; v++ {
		migrator, ok := aikataYamlMigrators[v]
		if !ok {
			return AikataYaml{}, false, fmt.Errorf("config: no migrator registered for aikata.yaml v%d -> v%d", v, v+1)
		}
		if err := migrator(&doc); err != nil {
			return AikataYaml{}, false, fmt.Errorf("config: migrate aikata.yaml v%d -> v%d: %w", v, v+1, err)
		}
		migrated = true
	}
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return AikataYaml{}, false, fmt.Errorf("config: re-marshal migrated aikata.yaml: %w", err)
	}
	cfg, err := Unmarshal(out)
	if err != nil {
		return AikataYaml{}, false, err
	}
	return cfg, migrated, nil
}

// LoadMigrated reads `.aikata/aikata.yaml` (preferring the primary
// path, falling back to the legacy `.ai/` directory) and runs
// MigrateAikataYaml. If migrations were applied, the upgraded YAML is
// written back to disk so subsequent commands see the modern schema.
//
// This is the entry point `aikata sync` uses; per ADR 0011 D3, schema
// migration is built into sync rather than living in a separate
// command.
func LoadMigrated(root string) (AikataYaml, bool, error) {
	path, _, err := Resolve(root)
	if err != nil {
		return AikataYaml{}, false, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return AikataYaml{}, false, fs.ErrNotExist
		}
		return AikataYaml{}, false, fmt.Errorf("config: read aikata.yaml: %w", err)
	}
	cfg, migrated, err := MigrateAikataYaml(body)
	if err != nil {
		return AikataYaml{}, false, err
	}
	if migrated {
		upgraded, err := Marshal(cfg)
		if err != nil {
			return cfg, true, err
		}
		if err := writeAtomic(path, upgraded, 0o644); err != nil {
			return cfg, true, fmt.Errorf("config: persist migrated aikata.yaml: %w", err)
		}
	}
	return cfg, migrated, nil
}

// readVersionField inspects the top-level `version:` scalar of a
// parsed YAML document and returns its integer value. Returns 0 when
// the field is absent so Unmarshal's "missing version" error path
// keeps owning that case.
func readVersionField(doc *yaml.Node) (int, error) {
	// `doc` is a document node wrapping a single mapping node. Walk
	// down to find the version key.
	if doc == nil || len(doc.Content) == 0 {
		return 0, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return 0, nil
	}
	for i := 0; i < len(root.Content)-1; i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		if k.Value == "version" {
			var n int
			if err := v.Decode(&n); err != nil {
				return 0, fmt.Errorf("config: parse version field: %w", err)
			}
			return n, nil
		}
	}
	return 0, nil
}
