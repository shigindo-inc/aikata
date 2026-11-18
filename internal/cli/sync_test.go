package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func TestSync_AcceptOneResolvedConflict(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	if _, err := runInit(t, "samplekata", "--no-interactive"); err != nil {
		t.Fatal(err)
	}
	read := func(path string) []byte {
		t.Helper()
		body, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	write := func(path string, body []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, path), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(args ...string) (string, error) {
		t.Helper()
		cmd := newSyncCmd()
		cmd.SilenceUsage = true
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(append([]string{"--lang", "ja", "--json"}, args...))
		err := cmd.Execute()
		return buf.String(), err
	}
	for _, name := range []string{"README.md", "SPEC.md"} {
		write(name, append(read(name), []byte("\nLocal edit.\n")...))
	}
	if _, err := run(); err == nil {
		t.Fatal("expected conflicts after switching the upstream language")
	}
	marked := read("README.md")
	other := read("SPEC.md")
	base := read(config.PrimaryDir + "/" + config.ManifestFilename)
	if _, err := run("--accept-conflict", "README.md"); err == nil {
		t.Fatal("accepted a path with unresolved markers")
	}
	_, block, ok := strings.Cut(string(marked), "\n=======\n")
	upstream, _, end := strings.Cut(block, ">>>>>>> upstream")
	if !ok || !end {
		t.Fatal("expected upstream side in conflict")
	}
	resolved := []byte(upstream + "\nLocal edit.\n")
	write("README.md", resolved)
	if _, err := run("--dry-run", "--accept-conflict", "README.md"); err != nil {
		t.Fatalf("preview acceptance: %v", err)
	}
	if !bytes.Equal(read(config.PrimaryDir+"/"+config.ManifestFilename), base) {
		t.Fatal("preview modified the manifest")
	}
	if _, err := run("--accept-conflict", "README.md", "--accept-conflict", "SPEC.md"); err == nil {
		t.Fatal("accepted a batch containing an unresolved path")
	}
	if !bytes.Equal(read(config.PrimaryDir+"/"+config.ManifestFilename), base) {
		t.Fatal("failed acceptance partially modified the manifest")
	}
	if _, err := run("--accept-conflict", "README.md"); err != nil {
		t.Fatalf("accept resolution: %v", err)
	}
	out, err := run()
	var exit *ExitError
	if !errors.As(err, &exit) || exit.Code != 2 {
		t.Fatalf("remaining conflict: %v", err)
	}
	var report struct {
		Files []struct{ Path, Status string }
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatal(err)
	}
	status := ""
	for _, f := range report.Files {
		if f.Path == "README.md" {
			status = f.Status
		}
	}
	if status != "user-only-edit" {
		t.Fatalf("accepted file classified as %q", status)
	}
	if !bytes.Equal(read("README.md"), resolved) || !bytes.Equal(read("SPEC.md"), other) {
		t.Fatal("sync changed the accepted file or rewrapped the unresolved one")
	}
}
