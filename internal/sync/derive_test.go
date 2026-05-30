package sync

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

// TestInferFlags_DetectsMonorepo asserts the v0.6.3 addition to
// `inferFlags`: when the manifest already contains `docs/monorepo.md`
// or any `apps/...` entry, sync infers the monorepo opt-in even
// without `features.monorepo: true` in aikata.yaml.
func TestInferFlags_DetectsMonorepo(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"explainer present", []string{"docs/monorepo.md"}, true},
		{"apps entry present", []string{"apps/_example/AGENTS.md"}, true},
		{"unrelated paths only", []string{"AGENTS.md", "README.md"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := config.Manifest{}
			for _, p := range tc.paths {
				m.Files = append(m.Files, config.ManifestFile{Path: p})
			}
			got := inferFlags(m).WithMonorepo
			if got != tc.want {
				t.Errorf("WithMonorepo for %+v = %v, want %v", tc.paths, got, tc.want)
			}
		})
	}
}

// TestDerivePlan_Rebaseline_HonoursAikataYaml is the regression test
// for the v0.6.1 footgun: `--rebaseline` used to ignore
// `aikata.yaml`'s `stacks` and `features.monorepo`. v0.6.3 reads them.
func TestDerivePlan_Rebaseline_HonoursAikataYaml(t *testing.T) {
	cfg := config.AikataYaml{
		Project: config.Project{Lang: "ja"},
		Stacks:  []string{"flutter"},
		Features: map[string]bool{
			"monorepo": true,
			"tdd":      true,
		},
	}
	preset, lang, flags, stacks, _ := derivePlan(config.Manifest{}, cfg, false, overrides{})

	if preset != "standard" {
		t.Errorf("preset = %q, want \"standard\" (rebaseline default)", preset)
	}
	if lang != "ja" {
		t.Errorf("lang = %q, want \"ja\" (from cfg.Project.Lang)", lang)
	}
	if !flags.WithMonorepo {
		t.Errorf("WithMonorepo not honoured: %+v", flags)
	}
	if !flags.WithTDD {
		t.Errorf("WithTDD not honoured: %+v", flags)
	}
	if !reflect.DeepEqual(stacks, []string{"flutter"}) {
		t.Errorf("stacks = %v, want [flutter]", stacks)
	}
}

// TestDerivePlan_Manifest_AikataYamlORs asserts that when a manifest
// is present, aikata.yaml's stacks and feature flags OR-merge with
// what `inferFlags` derives. The manifest is "what files exist";
// aikata.yaml is "what the user intends to enable"; both signals
// together cover post-init opt-ins the manifest may not have grown
// yet.
func TestDerivePlan_Manifest_AikataYamlORs(t *testing.T) {
	// Manifest knows about UI.md but nothing else opt-in.
	m := config.Manifest{
		Version: 1,
		Preset:  "standard",
		Lang:    "en",
		Files: []config.ManifestFile{
			{Path: "UI.md"},
		},
	}
	// aikata.yaml intends monorepo + tdd, plus a flutter stack.
	cfg := config.AikataYaml{
		Project: config.Project{Lang: "ja"}, // ignored when manifest present
		Stacks:  []string{"flutter"},
		Features: map[string]bool{
			"monorepo": true,
			"tdd":      true,
		},
	}
	preset, lang, flags, stacks, _ := derivePlan(m, cfg, true, overrides{})

	if preset != "standard" {
		t.Errorf("preset = %q, want manifest's \"standard\"", preset)
	}
	if lang != "en" {
		t.Errorf("lang = %q, want manifest's \"en\" (cfg.Project.Lang shouldn't win when manifest present)", lang)
	}
	if !flags.WithUI {
		t.Errorf("WithUI should be inferred from manifest UI.md entry: %+v", flags)
	}
	if !flags.WithMonorepo {
		t.Errorf("WithMonorepo should be OR'd in from cfg.Features: %+v", flags)
	}
	if !flags.WithTDD {
		t.Errorf("WithTDD should be OR'd in from cfg.Features: %+v", flags)
	}
	if !reflect.DeepEqual(stacks, []string{"flutter"}) {
		t.Errorf("stacks = %v, want [flutter] from cfg.Stacks", stacks)
	}
}

// TestDerivePlan_SchemaV2_ComponentsOR verifies the v0.7.0 addition:
// `cfg.Components.*` is OR-merged with whatever inferFlags or the
// legacy `features.*` keys produce. Schema v2 fields are the
// canonical signal; manifest path inference remains a safety net for
// projects that have not yet been re-touched by a writer.
func TestDerivePlan_SchemaV2_ComponentsOR(t *testing.T) {
	// Manifest knows about UI.md only.
	m := config.Manifest{
		Version: 1,
		Preset:  "standard",
		Lang:    "en",
		Files: []config.ManifestFile{
			{Path: "UI.md"},
		},
	}
	// Schema-v2 cfg declares memory + tdd + monorepo explicitly. UI is
	// false in the schema but still inferable from the manifest; the
	// OR-merge must keep UI true.
	cfg := config.AikataYaml{
		Version: 2,
		Components: config.Components{
			Memory:   true,
			TDD:      true,
			Monorepo: true,
		},
	}
	_, _, flags, _, _ := derivePlan(m, cfg, true, overrides{})

	if !flags.WithMemory {
		t.Errorf("Components.Memory not honoured: %+v", flags)
	}
	if !flags.WithTDD {
		t.Errorf("Components.TDD not honoured: %+v", flags)
	}
	if !flags.WithMonorepo {
		t.Errorf("Components.Monorepo not honoured: %+v", flags)
	}
	if !flags.WithUI {
		t.Errorf("UI must remain true via manifest OR-merge even when schema-v2 says false: %+v", flags)
	}
}

// TestDerivePlan_SchemaV2_LegacyFeaturesStillRead is the v0.7.0
// backwards-compatibility regression: a v1-shaped config (with
// `features.tdd` and `features.monorepo`, no `components` block) must
// keep producing the same scope, because the migrator runs on write
// paths only — read-only callers like `aikata doctor` still see the
// raw v1 payload.
func TestDerivePlan_SchemaV2_LegacyFeaturesStillRead(t *testing.T) {
	cfg := config.AikataYaml{
		Version: 1,
		Project: config.Project{Lang: "en"},
		Features: map[string]bool{
			"tdd":      true,
			"monorepo": true,
		},
	}
	_, _, flags, _, _ := derivePlan(config.Manifest{}, cfg, false, overrides{})
	if !flags.WithTDD || !flags.WithMonorepo {
		t.Errorf("legacy features.{tdd,monorepo} must still set flags pre-migration: %+v", flags)
	}
}

// TestDerivePlan_Overrides_TakePrecedence verifies the ADR 0013
// hierarchy: CLI override > manifest > aikata.yaml > defaults. Each
// field is asserted independently so a regression in one path doesn't
// hide regressions in the others.
func TestDerivePlan_Overrides_TakePrecedence(t *testing.T) {
	m := config.Manifest{
		Version: 1,
		Preset:  "standard",
		Lang:    "en",
		Files: []config.ManifestFile{
			{Path: "docs/monorepo.md"}, // would otherwise infer WithMonorepo=true
		},
	}
	cfg := config.AikataYaml{
		Project: config.Project{Lang: "ja"},
		Stacks:  []string{"flutter"},
	}

	wantPreset := "flutter"
	wantLang := "ja"
	wantStacks := []string{"typescript"}
	wantMono := false

	preset, lang, flags, stacks, _ := derivePlan(m, cfg, true, overrides{
		Preset:       &wantPreset,
		Lang:         &wantLang,
		Stacks:       &wantStacks,
		WithMonorepo: &wantMono,
	})

	if preset != wantPreset {
		t.Errorf("preset = %q, want %q (CLI override)", preset, wantPreset)
	}
	if lang != wantLang {
		t.Errorf("lang = %q, want %q (CLI override)", lang, wantLang)
	}
	if !reflect.DeepEqual(stacks, wantStacks) {
		t.Errorf("stacks = %v, want %v (CLI override replaces cfg.Stacks)", stacks, wantStacks)
	}
	if flags.WithMonorepo != wantMono {
		t.Errorf("WithMonorepo = %v, want %v (CLI override must beat inferFlags + cfg.Features)", flags.WithMonorepo, wantMono)
	}
}

// TestRun_OverridePreset_ChangesScope is the end-to-end version: a
// `aikata sync --preset minimal` against a manifest that records the
// `standard` preset narrows the merged file set to minimal's
// templates. Verifies the override path lands all the way through
// scaffold.Render.
func TestRun_OverridePreset_ChangesScope(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	minimal := "minimal"
	result, err := Run(Options{
		Root:           root,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
		OverridePreset: &minimal,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// minimal preset only renders AGENTS.md / README.md / SPEC.md.
	// Any standard-only file the manifest tracks (e.g. GLOSSARY.md)
	// should show up as upstream-removed in the result.
	var sawGlossaryRemoved bool
	for _, f := range result.Files {
		if f.Path == "GLOSSARY.md" && f.Status == StatusUpstreamRemoved {
			sawGlossaryRemoved = true
		}
	}
	if !sawGlossaryRemoved {
		t.Errorf("expected GLOSSARY.md status=upstream-removed under --preset minimal override; result.Files=%v", filePaths(result))
	}
}

// TestRun_OverridesAreTransient confirms that CLI overrides are not
// written back to the manifest or aikata.yaml. A subsequent sync
// without overrides should observe the original (un-overridden)
// scope.
func TestRun_OverridesAreTransient(t *testing.T) {
	root := t.TempDir()
	seedStandardProject(t, root)

	// Snapshot the original manifest hash so we can confirm it isn't
	// rewritten by the override-run (the override should only change
	// scaffold.Render's behaviour, not the persisted state).
	manifestPath := config.ManifestPath(root)
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest before: %v", err)
	}

	minimal := "minimal"
	if _, err := Run(Options{
		Root:           root,
		DryRun:         true, // keep the run side-effect-free
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
		OverridePreset: &minimal,
	}); err != nil {
		t.Fatalf("Run --dry-run with override: %v", err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest after: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("manifest changed across an override + dry-run; overrides must be transient\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

func filePaths(r RunResult) []string {
	out := make([]string, len(r.Files))
	for i, f := range r.Files {
		out[i] = f.Path + ":" + string(f.Status)
	}
	return out
}
