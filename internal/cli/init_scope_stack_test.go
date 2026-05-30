package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveTemplateTree(t *testing.T) {
	cases := []struct {
		name       string
		scope      string
		stacks     []string
		wantTree   string
		wantStacks []string
		wantErr    bool
	}{
		{name: "standard no stack", scope: "standard", wantTree: "standard"},
		{name: "minimal no stack", scope: "minimal", wantTree: "minimal"},
		{name: "standard flutter", scope: "standard", stacks: []string{"flutter"}, wantTree: "flutter", wantStacks: []string{"flutter"}},
		{name: "standard typescript", scope: "standard", stacks: []string{"typescript"}, wantTree: "typescript", wantStacks: []string{"typescript"}},
		{name: "stack case-normalized", scope: "standard", stacks: []string{"Flutter"}, wantTree: "flutter", wantStacks: []string{"flutter"}},
		{name: "minimal with stack errors", scope: "minimal", stacks: []string{"flutter"}, wantErr: true},
		{name: "multi-stack errors", scope: "standard", stacks: []string{"flutter", "typescript"}, wantErr: true},
		{name: "unknown stack errors", scope: "standard", stacks: []string{"rust"}, wantErr: true},
		{name: "extended reserved errors", scope: "extended", wantErr: true},
		{name: "unknown scope errors", scope: "rust", wantErr: true},
		{name: "empty scope errors", scope: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, normStacks, err := resolveTemplateTree(tc.scope, tc.stacks)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got tree=%q stacks=%v", tree, normStacks)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tree != tc.wantTree {
				t.Errorf("tree = %q, want %q", tree, tc.wantTree)
			}
			if len(normStacks) != 0 || len(tc.wantStacks) != 0 {
				if !reflect.DeepEqual(normStacks, tc.wantStacks) {
					t.Errorf("stacks = %v, want %v", normStacks, tc.wantStacks)
				}
			}
		})
	}
}

func TestPresetToScopeStack(t *testing.T) {
	cases := []struct {
		preset     string
		wantScope  string
		wantStacks []string
		wantErr    bool
	}{
		{preset: "minimal", wantScope: "minimal"},
		{preset: "standard", wantScope: "standard"},
		{preset: "flutter", wantScope: "standard", wantStacks: []string{"flutter"}},
		{preset: "typescript", wantScope: "standard", wantStacks: []string{"typescript"}},
		{preset: "rust", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.preset, func(t *testing.T) {
			scope, stacks, err := presetToScopeStack(tc.preset)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.preset)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if scope != tc.wantScope {
				t.Errorf("scope = %q, want %q", scope, tc.wantScope)
			}
			if len(stacks) != 0 || len(tc.wantStacks) != 0 {
				if !reflect.DeepEqual(stacks, tc.wantStacks) {
					t.Errorf("stacks = %v, want %v", stacks, tc.wantStacks)
				}
			}
		})
	}
}

func TestNormalizeStacks(t *testing.T) {
	got := normalizeStacks([]string{" Flutter ", "flutter", "", "TypeScript"})
	want := []string{"flutter", "typescript"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("normalizeStacks = %v, want %v (lowercased, trimmed, deduped, order-preserving)", got, want)
	}
}

// TestInit_ScopeStandardStackFlutter is the backward-compat anchor: the
// new axes must reproduce what `--preset flutter` produced — the flutter
// template tree plus stacks: [flutter] in aikata.yaml.
func TestInit_ScopeStandardStackFlutter(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--scope", "standard", "--stack", "flutter", "--no-interactive")
	if err != nil {
		t.Fatalf("init --scope standard --stack flutter: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs", "stacks", "flutter.md")); err != nil {
		t.Errorf("expected docs/stacks/flutter.md from the flutter tree: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read aikata.yaml: %v", err)
	}
	if !strings.Contains(string(body), "flutter") {
		t.Errorf("aikata.yaml stacks: must list flutter:\n%s", body)
	}
}

func TestInit_MinimalWithStackErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	out, err := runInit(t, "samplekata", "--scope", "minimal", "--stack", "flutter", "--no-interactive")
	if err == nil {
		t.Fatalf("expected error for --scope minimal --stack flutter; out: %s", out)
	}
	if !strings.Contains(err.Error(), "not yet supported") {
		t.Errorf("error should explain the bounded combination, got: %v", err)
	}
}

func TestInit_MultiStackErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--stack", "flutter,typescript", "--no-interactive")
	if err == nil {
		t.Fatalf("expected error for multi-stack")
	}
}

// TestInit_DeprecatedPresetFlutterStillWorks asserts the alias keeps
// working and prints a deprecation warning (runInit folds stderr into
// the returned output buffer).
func TestInit_DeprecatedPresetFlutterStillWorks(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	out, err := runInit(t, "samplekata", "--preset", "flutter", "--no-interactive")
	if err != nil {
		t.Fatalf("init --preset flutter: %v (out: %s)", err, out)
	}
	if !strings.Contains(out, "deprecated") {
		t.Errorf("expected a deprecation warning for --preset, got: %s", out)
	}
	if _, err := os.Stat(filepath.Join(tmp, "docs", "stacks", "flutter.md")); err != nil {
		t.Errorf("--preset flutter must still scaffold the flutter tree: %v", err)
	}
}

func TestInit_PresetCombinedWithScopeErrors(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)
	_, err := runInit(t, "samplekata", "--preset", "flutter", "--scope", "standard", "--no-interactive")
	if err == nil {
		t.Fatalf("expected error combining --preset with --scope")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("error should explain the conflict, got: %v", err)
	}
}
