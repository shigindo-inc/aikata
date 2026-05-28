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
// N+1.
//
// The registry sits at package scope so future migrations slot in
// without changing call-site code. Migrations run sequentially from
// the payload's current version up to Version.
var aikataYamlMigrators = map[int]AikataYamlMigrator{
	1: migrateV1ToV2,
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

// migrateV1ToV2 lifts the schema-v1 `features.tdd` and
// `features.monorepo` keys into the schema-v2 `components` block and
// stamps `version: 2`. Components that v1 had no schema slot for
// (memory, ui, api, changelog) start as `false`; `aikata sync` keeps
// inferring them from manifest paths in the meantime so legacy
// projects do not lose their scope at the migration boundary. See
// ADR 0016.
func migrateV1ToV2(doc *yaml.Node) error {
	root := documentRoot(doc)
	if root == nil {
		return fmt.Errorf("config: migrate v1 -> v2: document has no root mapping")
	}

	liftedTDD := liftMapKeyBool(root, "features", "tdd")
	liftedMono := liftMapKeyBool(root, "features", "monorepo")
	if isMapEmpty(root, "features") {
		removeMapKey(root, "features")
	}

	upsertComponentsBlock(root, liftedTDD, liftedMono)
	setMapKey(root, "version", scalarInt(2))
	return nil
}

// documentRoot returns the top-level mapping node of a parsed YAML
// document, or nil when the document is empty or non-mapping.
func documentRoot(doc *yaml.Node) *yaml.Node {
	if doc == nil || len(doc.Content) == 0 {
		return nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}
	return root
}

// findMapValue returns the value node paired with key on a mapping
// node, or nil when absent. Caller must own the mapping node.
func findMapValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// removeMapKey deletes the key/value pair for key from a mapping node.
// Reports whether a pair was removed.
func removeMapKey(m *yaml.Node, key string) bool {
	if m == nil || m.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return true
		}
	}
	return false
}

// setMapKey inserts or replaces a key/value pair on a mapping node.
// New keys are appended at the end; existing keys keep their position.
func setMapKey(m *yaml.Node, key string, value *yaml.Node) {
	if m == nil || m.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return
		}
	}
	m.Content = append(m.Content, scalarString(key), value)
}

// liftMapKeyBool reads a nested `<parent>.<key>` scalar as a bool,
// removes the inner key, and returns the value. Returns false when
// either level is missing or the scalar is not parseable as bool.
func liftMapKeyBool(root *yaml.Node, parent, key string) bool {
	p := findMapValue(root, parent)
	if p == nil || p.Kind != yaml.MappingNode {
		return false
	}
	v := findMapValue(p, key)
	if v == nil {
		return false
	}
	var b bool
	if err := v.Decode(&b); err != nil {
		removeMapKey(p, key)
		return false
	}
	removeMapKey(p, key)
	return b
}

// isMapEmpty reports whether the mapping at key has no remaining
// child pairs. Returns false when the key is absent or not a mapping.
func isMapEmpty(root *yaml.Node, key string) bool {
	v := findMapValue(root, key)
	if v == nil || v.Kind != yaml.MappingNode {
		return false
	}
	return len(v.Content) == 0
}

// upsertComponentsBlock ensures a `components:` mapping exists with
// all six v2 fields, seeded from the lifted v1 features.
func upsertComponentsBlock(root *yaml.Node, tdd, monorepo bool) {
	existing := findMapValue(root, "components")
	if existing != nil && existing.Kind == yaml.MappingNode {
		// Honour any v2-shaped block already present; only fill the
		// fields the user hasn't set yet.
		ensureBoolKey(existing, "memory", false)
		ensureBoolKey(existing, "ui", false)
		ensureBoolKey(existing, "api", false)
		ensureBoolKey(existing, "tdd", tdd)
		ensureBoolKey(existing, "changelog", false)
		ensureBoolKey(existing, "monorepo", monorepo)
		return
	}
	block := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	block.Content = append(block.Content,
		scalarString("memory"), scalarBool(false),
		scalarString("ui"), scalarBool(false),
		scalarString("api"), scalarBool(false),
		scalarString("tdd"), scalarBool(tdd),
		scalarString("changelog"), scalarBool(false),
		scalarString("monorepo"), scalarBool(monorepo),
	)
	setMapKey(root, "components", block)
}

// ensureBoolKey sets key=value on a mapping when key is missing.
// Existing keys are left untouched.
func ensureBoolKey(m *yaml.Node, key string, value bool) {
	if findMapValue(m, key) != nil {
		return
	}
	m.Content = append(m.Content, scalarString(key), scalarBool(value))
}

func scalarString(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func scalarBool(v bool) *yaml.Node {
	s := "false"
	if v {
		s = "true"
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: s}
}

func scalarInt(v int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", v)}
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
