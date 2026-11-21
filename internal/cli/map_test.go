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

func TestMap_HookOffByDefault(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "demo", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, leaf := range []string{"docmap.yaml", "docmap.md"} {
		if _, err := os.Stat(filepath.Join(tmp, ".aikata", leaf)); !os.IsNotExist(err) {
			t.Fatalf("init wrote %s without docmap.enabled (err=%v)", leaf, err)
		}
	}
}

func TestMap_HookRunsWhenEnabled(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	if _, err := runInit(t, "demo", "--preset", "standard", "--no-interactive"); err != nil {
		t.Fatalf("init: %v", err)
	}
	cfgPath := filepath.Join(tmp, ".aikata", "aikata.yaml")
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, append(cfg, []byte("docmap:\n    enabled: true\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := runGenerate(t); err != nil {
		t.Fatalf("generate: %v (out: %s)", err, out)
	}
	for _, leaf := range []string{"docmap.yaml", "docmap.md"} {
		if _, err := os.Stat(filepath.Join(tmp, ".aikata", leaf)); err != nil {
			t.Fatalf("generate did not rebuild %s with docmap.enabled: %v", leaf, err)
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

func TestMap_InvalidConfigDoesNotWrite(t *testing.T) {
	for name, config := range map[string]string{
		"invalid YAML":   "docmap: [broken\n",
		"unknown format": "version: 2\ndocmap:\n  formats: [markdown]\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			chdir(t, root)
			writeMD(t, root, ".aikata/aikata.yaml", config)
			if _, err := runMap(t); err == nil {
				t.Fatal("expected an error from invalid config")
			}
			if _, err := os.Stat(filepath.Join(root, ".aikata", "docmap.yaml")); !os.IsNotExist(err) {
				t.Fatalf("map should be absent, stat returned %v", err)
			}
		})
	}
}
