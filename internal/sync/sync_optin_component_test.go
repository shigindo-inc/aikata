package sync

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/components"
	"github.com/shigindo-inc/aikata/internal/config"
)

// enableComponent runs an enable-tier component against a scaffolded
// project, mirroring `aikata enable <name>`. Add() also flips the
// schema-v2 `components.*` bool, so this covers the whole enable path.
func enableComponent(t *testing.T, root string, c components.Component) {
	t.Helper()
	if err := c.Add(components.AddContext{
		TargetDir:   root,
		ProjectName: "samplekata",
		Lang:        "en",
		// nil Clock -> time.Now, matching sync's own upstream
		// rendering, so same-day renders classify as unchanged rather
		// than as a date-only re-render.
		Clock:  nil,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("enable component: %v", err)
	}
}

// TestInferFlags_DetectsOptInComponentPairs covers the manifest-side
// half of Q-MODELING-03: `inferFlags` knew nothing about the modeling,
// prompts, and env components, so a project that enabled one of them
// re-entered sync with the flag off and its files absent from the
// upstream render.
func TestInferFlags_DetectsOptInComponentPairs(t *testing.T) {
	cases := []struct {
		name string
		path string
		get  func(withFlags) bool
	}{
		{"usecases ledger", "docs/usecases.md", func(f withFlags) bool { return f.WithModeling }},
		{"domain model", "docs/domain.md", func(f withFlags) bool { return f.WithModeling }},
		{"prompt library", "docs/prompts.md", func(f withFlags) bool { return f.WithPrompts }},
		{"env template", ".env.example", func(f withFlags) bool { return f.WithEnv }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := config.Manifest{Files: []config.ManifestFile{{Path: tc.path}}}
			if !tc.get(inferFlags(m)) {
				t.Errorf("inferFlags did not detect the opt-in from %q", tc.path)
			}
		})
	}

	// Narrow detection: an unrelated manifest must not light any of them up.
	f := inferFlags(config.Manifest{Files: []config.ManifestFile{{Path: "README.md"}}})
	if f.WithModeling || f.WithPrompts || f.WithEnv {
		t.Errorf("unrelated manifest lit up an opt-in flag: %+v", f)
	}
}

// TestDerivePlan_SchemaV2_OptInComponentsOR covers the aikata.yaml-side
// half: `components.{modeling,prompts,env}: true` must reach the
// upstream render even when the manifest has not grown the paths yet.
func TestDerivePlan_SchemaV2_OptInComponentsOR(t *testing.T) {
	cfg := config.AikataYaml{}
	cfg.Components.Modeling = true
	cfg.Components.Prompts = true
	cfg.Components.Env = true

	_, _, flags, _, _ := derivePlan(config.Manifest{}, cfg, false, overrides{})

	if !flags.WithModeling {
		t.Errorf("components.modeling not honoured: %+v", flags)
	}
	if !flags.WithPrompts {
		t.Errorf("components.prompts not honoured: %+v", flags)
	}
	if !flags.WithEnv {
		t.Errorf("components.env not honoured: %+v", flags)
	}
}

// TestRun_OptInComponents_ParticipateAsManagedDocuments is the
// end-to-end regression for Q-MODELING-03: a file enabled post-init
// must classify as `unchanged` rather than `upstream-removed`, must
// keep its manifest entry across repeated syncs, and must stay on
// disk.
func TestRun_OptInComponents_ParticipateAsManagedDocuments(t *testing.T) {
	cases := []struct {
		name      string
		component components.Component
		paths     []string
	}{
		{"modeling", components.Modeling, []string{"docs/usecases.md", "docs/domain.md"}},
		{"prompts", components.Prompts, []string{"docs/prompts.md"}},
		{"env", components.Env, []string{".env.example"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			seedStandardProject(t, root)
			enableComponent(t, root, tc.component)

			for _, rel := range tc.paths {
				if manifestHash(t, root, rel) == "" {
					t.Fatalf("%s missing from manifest after enable", rel)
				}
			}

			for pass, res := range []RunResult{runSync(t, root, Options{}), runSync(t, root, Options{})} {
				for _, rel := range tc.paths {
					if got := statusOf(res, rel); got != StatusUnchanged {
						t.Errorf("sync %d: %s = %q, want %q", pass+1, rel, got, StatusUnchanged)
					}
					if manifestHash(t, root, rel) == "" {
						t.Errorf("sync %d: %s must keep its manifest entry", pass+1, rel)
					}
					if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
						t.Errorf("sync %d: %s must remain on disk: %v", pass+1, rel, err)
					}
				}
			}
		})
	}
}
