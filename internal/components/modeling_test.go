package components

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModeling_AddWritesBothDocuments(t *testing.T) {
	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := Modeling.Add(AddContext{
		TargetDir:   tmp,
		ProjectName: "demo",
		Lang:        "en",
		Clock:       fixedClock,
		Stdout:      &stdout,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	for _, rel := range []string{"docs/usecases.md", "docs/domain.md"} {
		b, rerr := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(rel)))
		if rerr != nil {
			t.Fatalf("read %s: %v", rel, rerr)
		}
		body := string(b)
		if !strings.HasPrefix(body, "---\n") {
			t.Errorf("%s: missing frontmatter opener", rel)
		}
		if !strings.Contains(body, "project: demo") {
			t.Errorf("%s: project name not interpolated", rel)
		}
	}
}

func TestModeling_DomainLinksToUsecasesAndBack(t *testing.T) {
	rendered, err := RenderModeling(ModelingParams{
		Lang: "en", ProjectName: "demo", Clock: fixedClock,
	})
	if err != nil {
		t.Fatalf("RenderModeling: %v", err)
	}
	if !strings.Contains(rendered["docs/domain.md"], "usecases.md") {
		t.Errorf("domain.md must link to usecases.md")
	}
	if !strings.Contains(rendered["docs/usecases.md"], "domain.md") {
		t.Errorf("usecases.md must link to domain.md")
	}
	if !strings.Contains(rendered["docs/domain.md"], "Related UC") {
		t.Errorf("domain.md must carry the field-granular Related UC column")
	}
}

func TestModeling_IsIdempotentAndNeverClobbers(t *testing.T) {
	tmp := t.TempDir()
	ctx := AddContext{
		TargetDir: tmp, ProjectName: "demo", Lang: "en",
		Clock: fixedClock, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{},
	}
	if err := Modeling.Add(ctx); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	target := filepath.Join(tmp, "docs", "usecases.md")
	if err := os.WriteFile(target, []byte("user edited\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := Modeling.Add(ctx); err != nil {
		t.Fatalf("second Add: %v", err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "user edited\n" {
		t.Errorf("re-run clobbered a user-edited file: %q", b)
	}
}

func TestModeling_IsRegisteredAsCapability(t *testing.T) {
	if _, ok := GetCapability("modeling"); !ok {
		t.Errorf("modeling is not registered in the capabilities registry")
	}
	if _, ok := GetArtifact("modeling"); ok {
		t.Errorf("modeling must be a capability, not an artifact (ADR 0017)")
	}
}

func TestModeling_NeverMentionsDomainDrivenDesign(t *testing.T) {
	for _, lang := range []string{"en", "ja"} {
		rendered, err := RenderModeling(ModelingParams{
			Lang: lang, ProjectName: "demo", Clock: fixedClock,
		})
		if err != nil {
			t.Fatalf("RenderModeling(%s): %v", lang, err)
		}
		for rel, body := range rendered {
			for _, banned := range []string{"DDD", "Domain-Driven Design", "ドメイン駆動設計"} {
				if strings.Contains(body, banned) {
					t.Errorf("%s (%s) contains banned term %q (spec D7)", rel, lang, banned)
				}
			}
		}
	}
}
