package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func runMap(t *testing.T) (string, error) {
	t.Helper()
	cmd := newMapCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	return buf.String(), err
}

func writeMD(t *testing.T, dir, rel, body string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMap_CreatesDocMapWithoutConfig(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	writeMD(t, tmp, "README.md", "# Readme\n\n> The project.\n[spec](SPEC.md)\n")
	writeMD(t, tmp, "SPEC.md", "# SPEC\n\nWhat and why.\n")

	out, err := runMap(t)
	if err != nil {
		t.Fatalf("map: %v (out: %s)", err, out)
	}
	yamlBody, err := os.ReadFile(filepath.Join(tmp, ".aikata", "docmap.yaml"))
	if err != nil {
		t.Fatalf("docmap.yaml not written: %v", err)
	}
	mdBody, err := os.ReadFile(filepath.Join(tmp, ".aikata", "docmap.md"))
	if err != nil {
		t.Fatalf("docmap.md not written: %v", err)
	}
	if !bytes.Contains(yamlBody, []byte("path: README.md")) {
		t.Errorf("docmap.yaml missing README entry:\n%s", yamlBody)
	}
	if !bytes.Contains(mdBody, []byte("## Index")) {
		t.Errorf("docmap.md missing Index section:\n%s", mdBody)
	}
}

func TestMap_HookRunsAfterInit(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "demo", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	// The init PostRunE hook should have produced the doc map.
	for _, leaf := range []string{"docmap.yaml", "docmap.md"} {
		if _, err := os.Stat(filepath.Join(tmp, ".aikata", leaf)); err != nil {
			t.Fatalf("init did not rebuild %s: %v", leaf, err)
		}
	}
}

func TestMap_Deterministic(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	writeMD(t, tmp, "A.md", "# A\n[b](B.md)\n")
	writeMD(t, tmp, "B.md", "# B\n[a](A.md)\n")

	if _, err := runMap(t); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(tmp, ".aikata", "docmap.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runMap(t); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(filepath.Join(tmp, ".aikata", "docmap.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("docmap.yaml not deterministic:\n%s\n---\n%s", first, second)
	}
}
