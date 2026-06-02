package repolint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// firstPartySkills is the v0.10.0 skill surface (ADR 0040): the
// CLI-wrapper responsibility and the in-repo context-maintenance loop,
// shipped from the single aikata plugin.
var firstPartySkills = []string{"aikata-cli", "aikata-context"}

// TestSkillCopiesMatchCanonical enforces the copy boundary of ADR 0040:
// `dist/universal-skill/<skill>/SKILL.md` is the single canonical source
// for each skill's content, and every per-platform copy (Codex plugin,
// Claude Code plugin, Claude Code standalone skill) is byte-identical to
// it. Copies exist only for per-platform discovery location/format, never
// for content. The Codex copy additionally carries a byte-identical
// `agents/openai.yaml`; Claude Code ignores that file, so its copies are
// the flat `.md` body alone.
func TestSkillCopiesMatchCanonical(t *testing.T) {
	root := repoRoot(t)

	for _, skill := range firstPartySkills {
		canonicalSkill := "dist/universal-skill/" + skill + "/SKILL.md"
		skillCopies := []string{
			"dist/codex/plugin/skills/" + skill + "/SKILL.md",
			"dist/claude-code/plugin/skills/" + skill + ".md",
			"dist/claude-code/skill/" + skill + ".md",
		}
		for _, copyPath := range skillCopies {
			assertFilesEqual(t, root, canonicalSkill, copyPath)
		}

		assertFilesEqual(t, root,
			"dist/universal-skill/"+skill+"/agents/openai.yaml",
			"dist/codex/plugin/skills/"+skill+"/agents/openai.yaml",
		)
	}
}

// TestClaudePluginListsBothSkills guards the Claude Code plugin manifest
// against drifting from the two-skill surface (ADR 0040).
func TestClaudePluginListsBothSkills(t *testing.T) {
	root := repoRoot(t)

	var plugin struct {
		Components struct {
			Skills []string `json:"skills"`
		} `json:"components"`
	}
	readJSON(t, filepath.Join(root, "dist", "claude-code", "plugin", "plugin.json"), &plugin)

	want := []string{"skills/aikata-cli.md", "skills/aikata-context.md"}
	got := plugin.Components.Skills
	if len(got) != len(want) {
		t.Fatalf("Claude plugin skills = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("Claude plugin skills[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestCodexMarketplacePointsAtTrackedPlugin(t *testing.T) {
	root := repoRoot(t)

	var marketplace struct {
		Plugins []struct {
			Name   string `json:"name"`
			Source struct {
				Source string `json:"source"`
				Path   string `json:"path"`
			} `json:"source"`
		} `json:"plugins"`
	}
	readJSON(t, filepath.Join(root, ".agents", "plugins", "marketplace.json"), &marketplace)

	if len(marketplace.Plugins) != 1 {
		t.Fatalf("Codex marketplace plugins = %d, want 1", len(marketplace.Plugins))
	}
	plugin := marketplace.Plugins[0]
	if plugin.Name != "aikata" {
		t.Fatalf("Codex marketplace plugin name = %q, want %q", plugin.Name, "aikata")
	}
	if plugin.Source.Source != "local" {
		t.Fatalf("Codex marketplace source = %q, want %q", plugin.Source.Source, "local")
	}
	if plugin.Source.Path != "./dist/codex/plugin" {
		t.Fatalf("Codex marketplace path = %q, want %q", plugin.Source.Path, "./dist/codex/plugin")
	}

	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(plugin.Source.Path)))
	if err != nil {
		t.Fatalf("stat Codex marketplace plugin directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("Codex marketplace plugin path is not a directory: %s", plugin.Source.Path)
	}
}

func TestPluginVersionsStayInLockstep(t *testing.T) {
	root := repoRoot(t)

	var claudeMarketplace struct {
		Version string `json:"version"`
		Plugins []struct {
			Version string `json:"version"`
		} `json:"plugins"`
	}
	readJSON(t, filepath.Join(root, ".claude-plugin", "marketplace.json"), &claudeMarketplace)
	if len(claudeMarketplace.Plugins) != 1 {
		t.Fatalf("Claude marketplace plugins = %d, want 1", len(claudeMarketplace.Plugins))
	}

	var claudePlugin struct {
		Version string `json:"version"`
	}
	readJSON(t, filepath.Join(root, "dist", "claude-code", "plugin", "plugin.json"), &claudePlugin)

	var codexPlugin struct {
		Version string `json:"version"`
	}
	readJSON(t, filepath.Join(root, "dist", "codex", "plugin", ".codex-plugin", "plugin.json"), &codexPlugin)

	want := claudeMarketplace.Version
	versions := map[string]string{
		".claude-plugin/marketplace.json plugins[0]":  claudeMarketplace.Plugins[0].Version,
		"dist/claude-code/plugin/plugin.json":         claudePlugin.Version,
		"dist/codex/plugin/.codex-plugin/plugin.json": codexPlugin.Version,
	}
	for path, got := range versions {
		if got != want {
			t.Errorf("%s version = %q, want lockstep version %q", path, got, want)
		}
	}
}

// TestSkillFrontmatterParsesAsYAML guards against malformed SKILL.md
// frontmatter shipping undetected. `aikata doctor` excludes dist/, so a
// description containing a YAML-breaking sequence (e.g. an unquoted
// `: ` colon-space, which the agent skill ecosystem and Claude Code /
// Codex parse as a mapping) would otherwise pass every existing check
// while leaving the skill unloadable on every platform. Each first-party
// skill file's frontmatter must parse as a YAML mapping with a non-empty
// name and description.
func TestSkillFrontmatterParsesAsYAML(t *testing.T) {
	root := repoRoot(t)

	var skillFiles []string
	for _, skill := range firstPartySkills {
		skillFiles = append(skillFiles,
			"dist/universal-skill/"+skill+"/SKILL.md",
			"dist/codex/plugin/skills/"+skill+"/SKILL.md",
			"dist/claude-code/plugin/skills/"+skill+".md",
			"dist/claude-code/skill/"+skill+".md",
		)
	}

	for _, rel := range skillFiles {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		fm, ok := frontmatterBytes(data)
		if !ok {
			t.Errorf("%s: missing `---` frontmatter block", rel)
			continue
		}
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal(fm, &meta); err != nil {
			t.Errorf("%s: frontmatter is not valid YAML: %v", rel, err)
			continue
		}
		if meta.Name == "" {
			t.Errorf("%s: frontmatter `name` is empty", rel)
		}
		if meta.Description == "" {
			t.Errorf("%s: frontmatter `description` is empty (likely a YAML parse truncation)", rel)
		}
	}
}

// frontmatterBytes returns the bytes between the first two `---` fence
// lines of a Markdown file, and whether such a block was found.
func frontmatterBytes(data []byte) ([]byte, bool) {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 || string(bytes.TrimRight(lines[0], "\r")) != "---" {
		return nil, false
	}
	var body [][]byte
	for _, line := range lines[1:] {
		if string(bytes.TrimRight(line, "\r")) == "---" {
			return bytes.Join(body, []byte("\n")), true
		}
		body = append(body, line)
	}
	return nil, false
}

func assertFilesEqual(t *testing.T, root, wantPath, gotPath string) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(wantPath)))
	if err != nil {
		t.Fatalf("read %s: %v", wantPath, err)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(gotPath)))
	if err != nil {
		t.Fatalf("read %s: %v", gotPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s must be byte-identical to %s", gotPath, wantPath)
	}
}

func readJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
