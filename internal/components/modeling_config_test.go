package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func TestEnableComponentInConfig_FlipsModeling(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("demo", "en")
	body, err := config.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	dir := filepath.Join(tmp, ".aikata")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "aikata.yaml"), body, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := EnableComponentInConfig(tmp, "modeling"); err != nil {
		t.Fatalf("EnableComponentInConfig: %v", err)
	}

	got, _, err := config.LoadMigrated(tmp)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !got.Components.Modeling {
		t.Errorf("components.modeling = false, want true")
	}
}
