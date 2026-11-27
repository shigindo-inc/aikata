package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListADRs_JSON(t *testing.T) {
	root := t.TempDir()
	adrDir := filepath.Join(root, "docs", "adr")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"0001-first.md":  "# First\n\n- **Status**: Superseded by [ADR 0002](./0002-second.md)\n",
		"0002-second.md": "# Second\n\n> **状態**: active（一部失効。ADR 0003 が決定 2 を置き換えた）\n",
		"0003-third.md":  "---\nstatus: accepted\n---\n\n# Third\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(adrDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, root)

	cmd := newRootCmd("0.0.1-test")
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"list", "adrs", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list adrs: %v\nstderr: %s", err, errBuf.String())
	}

	var got adrReport
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v\n%s", err, out.String())
	}
	if got.Kind != "adrs" || got.Summary.Total != 3 {
		t.Fatalf("kind=%q total=%d, want adrs/3", got.Kind, got.Summary.Total)
	}
	wantClass := []string{"superseded", "effective", "missing"}
	wantRefs := []int{0, 1, 1}
	for i, it := range got.Items {
		if string(it.Class) != wantClass[i] || it.DocRefs != wantRefs[i] {
			t.Errorf("item %d: class=%s refs=%d, want %s/%d", it.Number, it.Class, it.DocRefs, wantClass[i], wantRefs[i])
		}
		if it.Links == nil {
			t.Errorf("item %d: links is null, want an array", it.Number)
		}
	}
	if got.Summary.ByClass["missing"] != 1 {
		t.Errorf("by_class = %v, want missing=1", got.Summary.ByClass)
	}
}
