package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	intadr "github.com/shigindo-inc/aikata/internal/adr"
	"github.com/shigindo-inc/aikata/internal/templates"
)

// adrComponent scaffolds a new Architecture Decision Record under
// docs/adr/. The filename / numbering rule is owned by internal/adr;
// this component is the user-facing shape (template + slug + title).
type adrComponent struct{}

// Adr is the singleton registered in the components registry.
var Adr Component = adrComponent{}

func (adrComponent) Name() string        { return "adr" }
func (adrComponent) Description() string { return "Architecture Decision Record under docs/adr/." }
func (adrComponent) Status() string      { return StatusActive }

// Add creates docs/adr/NNNN-<slug>.md from the title supplied as
// positional arguments. The slug is derived via templates.Kebab so a
// title like `aikata add adr "Use Go modules"` produces
// `docs/adr/NNNN-use-go-modules.md`. Refuses to clobber an existing
// ADR with the same slug at any number; the user must pick a
// different title or rename the prior file.
func (adrComponent) Add(ctx AddContext) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("components: adr: title is required (usage: aikata add adr \"<title>\")")
	}
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: adr: project name is required")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: adr: target directory is required")
	}

	title := strings.TrimSpace(strings.Join(ctx.Args, " "))
	if title == "" {
		return fmt.Errorf("components: adr: title cannot be blank")
	}
	slug := templates.Kebab(title)
	if slug == "" {
		return fmt.Errorf("components: adr: title %q produced an empty slug; pick a title with at least one ASCII alphanumeric", title)
	}

	adrDir := filepath.Join(ctx.TargetDir, "docs", "adr")
	entries, err := intadr.Scan(adrDir)
	if err != nil {
		return fmt.Errorf("components: adr: %w", err)
	}
	slugSuffix := "-" + slug + ".md"
	for _, e := range entries {
		if strings.HasSuffix(e.Filename, slugSuffix) {
			return fmt.Errorf("components: adr: docs/adr/%s already covers slug %q (pick a different title or rename the existing ADR)", e.Filename, slug)
		}
	}
	number, err := intadr.Next(adrDir)
	if err != nil {
		return fmt.Errorf("components: adr: %w", err)
	}
	filename := intadr.Filename(number, slug)
	full := filepath.Join(adrDir, filename)

	langDir, _, err := templates.LangDir("components/adr", ctx.Lang)
	if err != nil {
		return fmt.Errorf("components: adr: %w", err)
	}
	tmplPath := langDir + "/adr.md.tmpl"
	data := map[string]any{
		"ProjectName": ctx.ProjectName,
		"Lang":        ctx.Lang,
		"Number":      number,
		"NumberStr":   fmt.Sprintf("%04d", number),
		"Title":       title,
		"Slug":        slug,
	}
	content, err := templates.Render(tmplPath, data, ctx.Clock)
	if err != nil {
		return fmt.Errorf("components: adr: render %s: %w", tmplPath, err)
	}

	if ctx.DryRun {
		if _, werr := fmt.Fprintf(stdout(ctx), "Would write %s\n", relFromTarget(ctx.TargetDir, full)); werr != nil {
			return werr
		}
		return nil
	}

	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		return fmt.Errorf("components: adr: mkdir %s: %w", adrDir, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		return fmt.Errorf("components: adr: write %s: %w", full, err)
	}
	if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", relFromTarget(ctx.TargetDir, full)); werr != nil {
		return werr
	}
	return nil
}

// relFromTarget converts an absolute path under target into a path
// relative to target, falling back to the absolute path if rel fails.
func relFromTarget(target, abs string) string {
	if r, err := filepath.Rel(target, abs); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(abs)
}
