package components

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/templates"
)

// stackComponent adds a docs/stacks/<name>.md guide to an existing
// project and registers the stack in .aikata/aikata.yaml. Bundled
// stacks (flutter, typescript) reuse the preset template tree — there
// is no parallel docs/stacks/ template until external stack
// registries land in v1.0.
type stackComponent struct{}

// Stack is the singleton registered in the components registry.
var Stack Component = stackComponent{}

// bundledStacks names the stack guides shipped with the binary. The
// list is the single source of truth for `aikata list stacks` and
// the `aikata add stack` validation; future external stacks would
// extend this through a registry.
func bundledStacks() []string {
	return []string{"flutter", "typescript"}
}

// Stacks returns the bundled stack identifiers in sorted order. Used
// by cli/list.go so the listing and the addable set are one slice.
func Stacks() []string {
	out := append([]string(nil), bundledStacks()...)
	sort.Strings(out)
	return out
}

func (stackComponent) Name() string { return "stack" }
func (stackComponent) Description() string {
	return "Stack-specific guide under docs/stacks/<name>.md plus an aikata.yaml stacks: entry."
}
func (stackComponent) Status() string { return StatusActive }

// Add registers the stack named in ctx.Args[0]. Behavior:
//
//   - The stack name must be one of bundledStacks(); unknown names
//     exit non-zero with the known set in the error message.
//   - If docs/stacks/<name>.md is absent, render the bundled stack
//     brief (reused from the matching preset template tree).
//   - If aikata.yaml stacks: does not already include the name, add
//     it via config.Save (atomic).
//   - Existing on-disk files are never overwritten (the user may
//     have customized them); cfg updates still apply if needed.
//
// The result is idempotent: re-running on a project where the stack
// is fully registered is a no-op + notice.
func (stackComponent) Add(ctx AddContext) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("components: stack: stack name is required (usage: aikata add stack <name>)")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: stack: target directory is required")
	}
	name := strings.ToLower(strings.TrimSpace(ctx.Args[0]))
	if name == "" {
		return fmt.Errorf("components: stack: stack name cannot be blank")
	}
	if !stackKnown(name) {
		return fmt.Errorf("components: stack: unknown stack %q (known: %s)", name, strings.Join(Stacks(), ", "))
	}

	cfg, _, err := config.Load(ctx.TargetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("components: stack: aikata config not found at %s (run `aikata init` first)", config.PrimaryPath(ctx.TargetDir))
		}
		return fmt.Errorf("components: stack: load config: %w", err)
	}

	stackFile := filepath.Join(ctx.TargetDir, "docs", "stacks", name+".md")
	fileExists := false
	if _, statErr := os.Stat(stackFile); statErr == nil {
		fileExists = true
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("components: stack: stat %s: %w", stackFile, statErr)
	}

	cfgHasStack := contains(cfg.Stacks, name)

	// Always render the template — the rendered bytes are the
	// manifest ancestor hash, regardless of whether the on-disk file
	// already exists. This matches the singleFile / memory pattern
	// (record template hash even for skipped writes) so user-customised
	// stack guides register as `user-only-edit` on the next sync
	// rather than upstream-added or conflict-markered.
	langDir, _, lerr := templates.LangDir("presets/"+name, ctx.Lang)
	if lerr != nil {
		return fmt.Errorf("components: stack: %w", lerr)
	}
	tmplPath := langDir + "/docs/stacks/" + name + ".md.tmpl"
	data := map[string]any{
		"ProjectName": ctx.ProjectName,
		"Lang":        ctx.Lang,
		"Stacks":      []string{name},
	}
	rendered, rerr := templates.Render(tmplPath, data, ctx.Clock)
	if rerr != nil {
		return fmt.Errorf("components: stack: render %s: %w", tmplPath, rerr)
	}

	if ctx.DryRun {
		w := stdout(ctx)
		if !fileExists {
			if _, werr := fmt.Fprintf(w, "Would write %s\n", relFromTarget(ctx.TargetDir, stackFile)); werr != nil {
				return werr
			}
		}
		if !cfgHasStack {
			if _, werr := fmt.Fprintf(w, "Would add %q to .aikata/aikata.yaml stacks:\n", name); werr != nil {
				return werr
			}
		}
		return nil
	}

	if !fileExists {
		if err := os.MkdirAll(filepath.Dir(stackFile), 0o755); err != nil {
			return fmt.Errorf("components: stack: mkdir %s: %w", filepath.Dir(stackFile), err)
		}
		if err := os.WriteFile(stackFile, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("components: stack: write %s: %w", stackFile, err)
		}
		if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", relFromTarget(ctx.TargetDir, stackFile)); werr != nil {
			return werr
		}
	}

	// Record the stack guide in the manifest unconditionally (ADR 0014).
	// Even when the on-disk file pre-existed and we did not write,
	// the manifest must track the template hash so the next sync
	// classifies any user customisation as user-only-edit.
	stackRel := "docs/stacks/" + name + ".md"
	if rerr := RecordInManifest(ctx.TargetDir, map[string]string{stackRel: rendered}); rerr != nil {
		return rerr
	}

	if !cfgHasStack {
		cfg.Stacks = append(cfg.Stacks, name)
		sort.Strings(cfg.Stacks)
		if err := config.Save(ctx.TargetDir, cfg); err != nil {
			return fmt.Errorf("components: stack: save config: %w", err)
		}
		if _, werr := fmt.Fprintf(stdout(ctx),
			"updated .aikata/aikata.yaml stacks: += %q\n", name); werr != nil {
			return werr
		}
	}

	if fileExists && cfgHasStack {
		// Fully idempotent: file present, cfg has the stack, manifest
		// just refreshed. Mirror the original "nothing to do" notice so
		// scripts that grep for it keep working.
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: stack %q already present (docs/stacks/%s.md + aikata.yaml)\n", name, name); werr != nil {
			return werr
		}
	}
	return nil
}

func stackKnown(name string) bool {
	for _, s := range bundledStacks() {
		if s == name {
			return true
		}
	}
	return false
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
