package scaffold

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden, when true, overwrites the golden directory with the
// current output rather than comparing. Run via:
//
//	go test ./internal/scaffold/... -update
//
// The flag is shared across all golden tests in this package.
var updateGolden = flag.Bool("update", false, "rewrite testdata/golden/* with current scaffold output")

// goldenRoot returns the path of testdata/golden/<presetCase> relative to
// the scaffold package directory. Resolved at test time because that
// keeps the path readable in failure messages.
func goldenRoot(t *testing.T, presetCase string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "golden", presetCase)
}

func TestGolden_Minimal(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(defaultOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	compareWithGolden(t, goldenRoot(t, "minimal"), tmp)
}

func TestGolden_Standard(t *testing.T) {
	tmp := t.TempDir()
	if err := Run(standardOpts(tmp)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	compareWithGolden(t, goldenRoot(t, "standard"), tmp)
}

// compareWithGolden either updates the golden tree from gotDir (when
// -update is set) or asserts that the two trees are byte-identical.
func compareWithGolden(t *testing.T, goldenDir, gotDir string) {
	t.Helper()
	gotFiles, err := snapshotDir(gotDir)
	if err != nil {
		t.Fatalf("snapshot got: %v", err)
	}

	if *updateGolden {
		writeGolden(t, goldenDir, gotFiles)
		return
	}

	goldenFiles, err := snapshotDir(goldenDir)
	if err != nil {
		t.Fatalf("snapshot golden (run `go test ./internal/scaffold/... -update` to seed it): %v", err)
	}

	for rel, want := range goldenFiles {
		got, ok := gotFiles[rel]
		if !ok {
			t.Errorf("missing in output: %s", rel)
			continue
		}
		if string(want) != string(got) {
			t.Errorf("file %s differs from golden\n--- want (%s) ---\n%s\n--- got (%s) ---\n%s",
				rel, goldenDir, want, gotDir, got)
		}
	}
	for rel := range gotFiles {
		if _, ok := goldenFiles[rel]; !ok {
			t.Errorf("extra in output (not in golden %s): %s", goldenDir, rel)
		}
	}
}

// snapshotDir reads every regular file under root and keys it by its
// slash-separated path relative to root.
func snapshotDir(root string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = body
		return nil
	})
	return out, err
}

func writeGolden(t *testing.T, goldenDir string, files map[string][]byte) {
	t.Helper()
	if err := os.RemoveAll(goldenDir); err != nil {
		t.Fatalf("clear golden %s: %v", goldenDir, err)
	}
	for rel, body := range files {
		full := filepath.Join(goldenDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, body, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", full, err)
		}
	}
}
