package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// brandArtifacts is the set of one-off authoring artifacts added in
// v0.9.2 (ADR 0031). They share the oneOffArtifact implementation, so
// the table-driven tests below exercise both with one code path.
var brandArtifacts = []struct {
	component  Component
	targetPath string
	enHeading  string
	jaHeading  string
}{
	{AppIcon, "docs/design/app-icon-concepts.md", "# App-icon Concepts", "# アプリアイコン案"},
	{Mascot, "docs/design/mascot-character-ideas.md", "# Mascot Character Ideas", "# マスコットキャラクター案"},
}

func TestBrandArtifacts_RenderEnAndJa(t *testing.T) {
	for _, a := range brandArtifacts {
		for _, tc := range []struct {
			lang    string
			heading string
		}{
			{"en", a.enHeading},
			{"ja", a.jaHeading},
		} {
			tmp := t.TempDir()
			var stdout bytes.Buffer
			ctx := AddContext{
				TargetDir:   tmp,
				ProjectName: "demo",
				Lang:        tc.lang,
				Clock:       fixedClock,
				Stdout:      &stdout,
			}
			if err := a.component.Add(ctx); err != nil {
				t.Fatalf("%s (%s): Add: %v", a.component.Name(), tc.lang, err)
			}
			full := filepath.Join(tmp, filepath.FromSlash(a.targetPath))
			body, err := os.ReadFile(full)
			if err != nil {
				t.Fatalf("%s (%s): expected %s: %v", a.component.Name(), tc.lang, a.targetPath, err)
			}
			bodyStr := string(body)
			if !strings.Contains(bodyStr, tc.heading) {
				t.Errorf("%s (%s): missing heading %q:\n%s", a.component.Name(), tc.lang, tc.heading, bodyStr)
			}
			if !strings.Contains(bodyStr, "project: demo") {
				t.Errorf("%s (%s): missing project frontmatter", a.component.Name(), tc.lang)
			}
			if !strings.Contains(bodyStr, "audience: [human, agent]") {
				t.Errorf("%s (%s): missing five-key frontmatter audience line", a.component.Name(), tc.lang)
			}
			if !strings.Contains(stdout.String(), a.targetPath) {
				t.Errorf("%s (%s): expected stdout to mention %s; got %q", a.component.Name(), tc.lang, a.targetPath, stdout.String())
			}
		}
	}
}

// TestBrandArtifacts_RefusesToClobber asserts a one-off artifact errors
// on a pre-existing file rather than overwriting it (ADR 0031 D2): the
// project owns the document after the first stamp.
func TestBrandArtifacts_RefusesToClobber(t *testing.T) {
	for _, a := range brandArtifacts {
		tmp := t.TempDir()
		ctx := AddContext{TargetDir: tmp, ProjectName: "demo", Lang: "en", Clock: fixedClock}
		if err := a.component.Add(ctx); err != nil {
			t.Fatalf("%s: first Add: %v", a.component.Name(), err)
		}
		err := a.component.Add(ctx)
		if err == nil {
			t.Fatalf("%s: expected an error on the second Add (collision), got nil", a.component.Name())
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("%s: collision error should mention the conflict; got %v", a.component.Name(), err)
		}
	}
}

// TestBrandArtifacts_DryRunWritesNothing asserts --dry-run prints a plan
// without touching the filesystem.
func TestBrandArtifacts_DryRunWritesNothing(t *testing.T) {
	for _, a := range brandArtifacts {
		tmp := t.TempDir()
		var stdout bytes.Buffer
		ctx := AddContext{TargetDir: tmp, ProjectName: "demo", Lang: "en", Clock: fixedClock, DryRun: true, Stdout: &stdout}
		if err := a.component.Add(ctx); err != nil {
			t.Fatalf("%s: dry-run Add: %v", a.component.Name(), err)
		}
		if !strings.Contains(stdout.String(), "Would write "+a.targetPath) {
			t.Errorf("%s: dry-run should announce the plan; got %q", a.component.Name(), stdout.String())
		}
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(a.targetPath))); !os.IsNotExist(err) {
			t.Errorf("%s: dry-run wrote a file; stat err = %v", a.component.Name(), err)
		}
	}
}

// TestBrandArtifacts_NoManifestNoConfig asserts a one-off artifact
// records nothing in .aikata/manifest.yaml and flips no schema flag, so
// the project owns the file and sync never restores it (ADR 0031 D2).
func TestBrandArtifacts_NoManifestNoConfig(t *testing.T) {
	for _, a := range brandArtifacts {
		tmp := t.TempDir()
		ctx := AddContext{TargetDir: tmp, ProjectName: "demo", Lang: "en", Clock: fixedClock}
		if err := a.component.Add(ctx); err != nil {
			t.Fatalf("%s: Add: %v", a.component.Name(), err)
		}
		if _, err := os.Stat(filepath.Join(tmp, ".aikata", "manifest.yaml")); !os.IsNotExist(err) {
			t.Errorf("%s: a one-off artifact must not create .aikata/manifest.yaml; stat err = %v", a.component.Name(), err)
		}
		if _, err := os.Stat(filepath.Join(tmp, ".aikata", "aikata.yaml")); !os.IsNotExist(err) {
			t.Errorf("%s: a one-off artifact must not create .aikata/aikata.yaml; stat err = %v", a.component.Name(), err)
		}
	}
}
