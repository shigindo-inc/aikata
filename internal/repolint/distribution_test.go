package repolint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCodexPluginSkillCopiesMatchUniversal(t *testing.T) {
	root := repoRoot(t)

	assertFilesEqual(t, root,
		"dist/universal-skill/SKILL.md",
		"dist/codex/plugin/skills/aikata/SKILL.md",
	)
	assertFilesEqual(t, root,
		"dist/universal-skill/agents/openai.yaml",
		"dist/codex/plugin/skills/aikata/agents/openai.yaml",
	)
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
