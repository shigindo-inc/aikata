package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRun_WithModelingWritesPairAndFlagsConfig pins the init-time wiring
// for the modeling capability (Task 3): --with-modeling must render the
// docs/usecases.md + docs/domain.md pair and flip `modeling: true` in
// .aikata/aikata.yaml, mirroring how --with-prompts is wired.
func TestRun_WithModelingWritesPairAndFlagsConfig(t *testing.T) {
	tmp := t.TempDir()
	opts := standardOpts(tmp)
	opts.WithModeling = true
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range []string{"docs/usecases.md", "docs/domain.md"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected %s: %v", rel, err)
		}
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), "modeling: true") {
		t.Errorf("config must record modeling: true:\n%s", body)
	}
}
