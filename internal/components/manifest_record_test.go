package components

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// seedAikataConfig writes a minimal `.aikata/aikata.yaml` so the
// RecordInManifest helpers can establish project identity. Mirrors
// what `aikata init` writes for the standard preset.
func seedAikataConfig(t *testing.T, root, name, lang string) {
	t.Helper()
	cfg := config.Default(name, lang)
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("seed aikata.yaml: %v", err)
	}
}

func TestRecordInManifest_SingleFile_AddsEntry(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	if err := UI.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add UI: %v", err)
	}
	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("LoadManifest after Add: %v", err)
	}
	if !manifestHasPath(m, "UI.md") {
		t.Errorf("manifest should record UI.md after `add ui`; got files=%v", manifestPaths(m))
	}
}

func TestRecordInManifest_Memory_AddsAllRenderedPaths(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	if err := Memory.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add memory: %v", err)
	}
	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("LoadManifest after Add: %v", err)
	}
	for _, want := range []string{"docs/memory/user.md", "docs/memory/project.md"} {
		if !manifestHasPath(m, want) {
			t.Errorf("manifest should record %s after `add memory`; got files=%v", want, manifestPaths(m))
		}
	}
}

func TestRecordInManifest_PreservesPreviousEntries(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	// Hand-author an existing manifest (simulates a prior `aikata init`
	// that recorded AGENTS.md). Subsequent `aikata add ui` must keep
	// AGENTS.md and add UI.md, not clobber the existing slate.
	initial := config.Manifest{
		Version: config.ManifestVersion,
		Preset:  "standard",
		Lang:    "en",
		Files: []config.ManifestFile{
			{Path: "AGENTS.md", SHA256: config.HashContent([]byte("agents body"))},
		},
	}
	if err := config.SaveManifest(tmp, initial); err != nil {
		t.Fatalf("seed manifest: %v", err)
	}

	if err := UI.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add UI: %v", err)
	}
	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if !manifestHasPath(m, "AGENTS.md") {
		t.Errorf("AGENTS.md should survive the Add operation; got %v", manifestPaths(m))
	}
	if !manifestHasPath(m, "UI.md") {
		t.Errorf("UI.md missing from manifest after Add; got %v", manifestPaths(m))
	}
}

func TestRecordInManifest_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	ctx := AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
		Stderr:      &bytes.Buffer{},
	}
	if err := UI.Add(ctx); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	body1, err := os.ReadFile(filepath.Join(tmp, ".aikata", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest after first Add: %v", err)
	}

	if err := UI.Add(ctx); err != nil {
		t.Fatalf("second Add: %v", err)
	}
	body2, err := os.ReadFile(filepath.Join(tmp, ".aikata", "manifest.yaml"))
	if err != nil {
		t.Fatalf("read manifest after second Add: %v", err)
	}
	if !bytes.Equal(body1, body2) {
		t.Errorf("repeated Add should produce byte-identical manifests:\n--- first ---\n%s\n--- second ---\n%s", body1, body2)
	}
}

func TestRecordInManifest_NoConfig_NoOpSilent(t *testing.T) {
	// Verifies that running `aikata add ui` against a directory that
	// was never init'd (no .aikata/aikata.yaml) does NOT mint a
	// manifest. The file write still happens (legacy behaviour), but
	// no .aikata/manifest.yaml is created.
	tmp := t.TempDir()

	if err := UI.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := config.LoadManifest(tmp); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("manifest should not be minted when aikata.yaml is absent; LoadManifest err=%v", err)
	}
	// And the file is still written, so the legacy behaviour is preserved.
	if _, err := os.Stat(filepath.Join(tmp, "UI.md")); err != nil {
		t.Errorf("UI.md should still be written even without aikata.yaml: %v", err)
	}
}

func TestRecordInManifest_Stack_RegistersGuide(t *testing.T) {
	tmp := t.TempDir()
	seedAikataConfig(t, tmp, "demo", "en")

	if err := Stack.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Args:        []string{"flutter"},
		Stdout:      &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("Add stack flutter: %v", err)
	}
	m, err := config.LoadManifest(tmp)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if !manifestHasPath(m, "docs/stacks/flutter.md") {
		t.Errorf("docs/stacks/flutter.md missing from manifest after `add stack flutter`; got %v", manifestPaths(m))
	}
}

func manifestHasPath(m config.Manifest, path string) bool {
	for _, f := range m.Files {
		if f.Path == path {
			return true
		}
	}
	return false
}

func manifestPaths(m config.Manifest) []string {
	out := make([]string, len(m.Files))
	for i, f := range m.Files {
		out[i] = f.Path
	}
	return out
}
