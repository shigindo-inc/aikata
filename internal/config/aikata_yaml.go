package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Version is the schema version of `.aikata/aikata.yaml` (and the
// pre-v0.3.2 `.ai/aikata.yaml` fallback) that this binary writes and
// is expected to read. Bumped only when an incompatible schema
// change ships (see ARCHITECTURE.md §4.2).
const Version = 1

// AikataYaml is the in-memory representation of the aikata config
// file (`.aikata/aikata.yaml`, or the legacy `.ai/aikata.yaml`).
// Field tags double as the YAML schema spec — keep them in sync with
// ARCHITECTURE.md §4.1.
type AikataYaml struct {
	Version  int             `yaml:"version"`
	Project  Project         `yaml:"project"`
	AITools  []string        `yaml:"ai_tools,omitempty"`
	Stacks   []string        `yaml:"stacks,omitempty"`
	Features map[string]bool `yaml:"features,omitempty"`
	Docs     Docs            `yaml:"docs,omitempty"`
}

// Project carries the human-facing identity of the project.
type Project struct {
	Name        string `yaml:"name"`
	Lang        string `yaml:"lang"`
	Description string `yaml:"description,omitempty"`
}

// Docs holds documentation-related preferences.
type Docs struct {
	GenerateGitignore bool   `yaml:"generate_gitignore"`
	TaskFileLocation  string `yaml:"task_file_location"`
}

// Default returns a v1 AikataYaml seeded for the given project name and
// language. The defaults match `aikata init --preset standard`:
//
//   - Only the `claude` AI tool is enabled by default.
//   - No stacks selected.
//   - All feature flags off.
//   - Generated AI-tool artifacts are gitignored by default
//     (target-project default — aikata's own repo overrides this per
//     ADR 0003).
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
		AITools: []string{"claude"},
		Stacks:  nil,
		Features: map[string]bool{
			"tdd":            false,
			"obsidian_hints": false,
			"monorepo":       false,
		},
		Docs: Docs{
			GenerateGitignore: true,
			TaskFileLocation:  "docs/tasks/current.md",
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
