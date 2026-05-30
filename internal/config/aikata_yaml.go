package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Version is the schema version of `.aikata/aikata.yaml` that this
// binary writes and is expected to read. Bumped only when an
// incompatible schema change ships (see ARCHITECTURE.md §4.2).
//
// v2 (v0.7.0) introduces the typed `components:` block so optional
// template scope (`memory`, `ui`, `api`, `tdd`, `changelog`,
// `monorepo`) is recorded explicitly instead of being inferred from
// manifest paths. The v1 → v2 migrator in [schema_migrate.go] lifts
// the corresponding entries out of `features:` and into the new
// block; legacy projects continue to be readable via the migrator.
const Version = 2

// AikataYaml is the in-memory representation of `.aikata/aikata.yaml`.
// Field tags double as the YAML schema spec — keep them in sync with
// ARCHITECTURE.md §4.1.
type AikataYaml struct {
	Version    int             `yaml:"version"`
	Project    Project         `yaml:"project"`
	AITools    []string        `yaml:"ai_tools,omitempty"`
	Stacks     []string        `yaml:"stacks,omitempty"`
	Workflows  []string        `yaml:"workflows,omitempty"`
	Components Components      `yaml:"components"`
	Features   map[string]bool `yaml:"features,omitempty"`
	Docs       Docs            `yaml:"docs,omitempty"`
	Doctor     Doctor          `yaml:"doctor,omitempty"`
	Sync       Sync            `yaml:"sync,omitempty"`
}

// Project carries the human-facing identity of the project.
type Project struct {
	Name        string `yaml:"name"`
	Lang        string `yaml:"lang"`
	Description string `yaml:"description,omitempty"`
}

// Components is the schema-v2 explicit record of which optional
// template-scope components a project has enabled. Pre-v2 projects
// kept the same intent split across `features.tdd`,
// `features.monorepo`, and manifest-path inference; v2 makes the
// signal first-class so post-init commands and `aikata sync` can read
// a single declarative source. See ADR 0016.
type Components struct {
	Memory    bool `yaml:"memory"`
	UI        bool `yaml:"ui"`
	API       bool `yaml:"api"`
	TDD       bool `yaml:"tdd"`
	Changelog bool `yaml:"changelog"`
	Monorepo  bool `yaml:"monorepo"`
}

// Docs holds documentation-related preferences.
//
// The inert `generate_gitignore` field was removed in v0.8.3 (ADR 0025
// D3): it was defined, defaulted to true, and never read, so setting it
// had no effect. `.gitignore` remains managed by the ADR 0018 managed-
// append writer (non-destructive by default); a project that wants
// `aikata sync` to leave a file alone lists it under `sync.own`. Old
// configs still carrying `generate_gitignore:` parse without error —
// the unknown key is ignored.
type Docs struct {
	TaskFileLocation string `yaml:"task_file_location"`
}

// Sync holds optional `aikata sync` preferences. See ADR 0025.
type Sync struct {
	// Own is a list of glob patterns (same matcher and additive
	// semantics as `doctor.exclude`, ADR 0021) for files the user has
	// intentionally taken ownership of. A matching path is reported
	// with the `owned` status and is never rendered-compared,
	// conflict-markered, or overwritten by `aikata sync`. An absent
	// block is the empty list. Ownership lives here — in the user-
	// editable config — not in the machine-owned manifest (ADR 0011
	// forbids hand-editing `.aikata/manifest.yaml`).
	Own []string `yaml:"own,omitempty"`
}

// Doctor holds optional tuning for `aikata doctor`. The zero value
// preserves the pre-v0.7.3 behaviour exactly (no extra exclusions
// beyond the hardcoded `skippedDirs` / `skippedFiles` baselines in
// `internal/doctor`). See ADR 0021.
type Doctor struct {
	// Exclude is a list of glob patterns evaluated against the
	// slash-form path relative to the project root. Matching paths
	// are skipped at the markdown-walk layer, exempting them from
	// `checkFrontmatter`, `checkUpdated`, and `checkGlossary`
	// uniformly. Supported tokens: `*` (single segment), `**`
	// (recursive), and literals.
	Exclude []string `yaml:"exclude,omitempty"`
}

// Default returns an AikataYaml at the current schema version seeded
// for the given project name and language. The defaults match
// `aikata init --scope standard`:
//
//   - Only the `claude` AI tool is enabled by default.
//   - No stacks selected.
//   - All optional components off (schema-v2 `components` block).
//   - All non-scope feature flags off (`obsidian_hints`).
//   - `docs.task_file_location` points at the default task file.
func Default(name, lang string) AikataYaml {
	if lang == "" {
		lang = "en"
	}
	return AikataYaml{
		Version: Version,
		Project: Project{
			Name: name,
			Lang: lang,
		},
		AITools:    []string{"claude"},
		Stacks:     nil,
		Components: Components{},
		Features: map[string]bool{
			"obsidian_hints": false,
		},
		Docs: Docs{
			TaskFileLocation: "docs/tasks/current.md",
		},
	}
}

// Marshal serializes y to YAML using yaml.v3 defaults. Output is
// deterministic field order (driven by struct order) — important for
// golden-test stability.
func Marshal(y AikataYaml) ([]byte, error) {
	out, err := yaml.Marshal(y)
	if err != nil {
		return nil, fmt.Errorf("config: marshal aikata.yaml: %w", err)
	}
	return out, nil
}

// Unmarshal parses an aikata.yaml payload (either path). Returned
// errors wrap the underlying yaml parse error with the path-style
// context callers expect (ARCHITECTURE.md §7.1).
func Unmarshal(data []byte) (AikataYaml, error) {
	var y AikataYaml
	if err := yaml.Unmarshal(data, &y); err != nil {
		return AikataYaml{}, fmt.Errorf("config: parse aikata.yaml: %w", err)
	}
	if y.Version == 0 {
		return AikataYaml{}, fmt.Errorf("config: aikata.yaml is missing the required %q field", "version")
	}
	return y, nil
}
