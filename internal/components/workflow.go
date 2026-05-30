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

// workflowComponent adds an opt-in collaboration workflow guide under
// docs/workflows/<domain>.md and registers the domain in the
// `workflows:` list of .aikata/aikata.yaml (ADR 0026). Git is the first
// built-in domain; the command shape is the broader `enable workflow
// <domain>` so future release / deployment / incident / review guides
// can share the same category and the same list-shaped config axis
// (mirroring stacks: / ai_tools: rather than one boolean per workflow
// under components:).
//
// Unlike the other enable-tier capabilities, `enable workflow <domain>`
// also injects a short pointer section into AGENTS.md so agents discover
// the guide from the canonical instruction file (ADR 0026 D2). AGENTS.md
// is intentionally NOT re-recorded in the manifest: its ancestor stays at
// the upstream rendering (no pointer), so the injected pointer is
// preserved across syncs as a `user-only-edit` divergence (ADR 0025 D1).
type workflowComponent struct{}

// Workflow is the singleton registered in the capabilities registry.
var Workflow Component = workflowComponent{}

// bundledWorkflows names the workflow domains shipped with the binary.
// The list is the single source of truth for `enable workflow <domain>`
// validation; future domains (release, deploy, incident, review) extend
// it alongside their templates.
func bundledWorkflows() []string {
	return []string{"git"}
}

// Workflows returns the bundled workflow-domain identifiers in sorted
// order.
func Workflows() []string {
	out := append([]string(nil), bundledWorkflows()...)
	sort.Strings(out)
	return out
}

func (workflowComponent) Name() string { return "workflow" }
func (workflowComponent) Description() string {
	return "Collaboration workflow guide under docs/workflows/<domain>.md plus an aikata.yaml workflows: entry (e.g. `enable workflow git`)."
}
func (workflowComponent) Status() string { return StatusActive }

// Add registers the workflow domain named in ctx.Args[0]. Behavior:
//
//   - The domain must be one of bundledWorkflows(); unknown names exit
//     non-zero with the known set in the error message.
//   - If docs/workflows/<domain>.md is absent, render the built-in
//     guide. Existing on-disk files are never overwritten.
//   - The rendered guide is recorded in the manifest unconditionally
//     (ADR 0014) so a user-customised guide registers as user-only-edit
//     on the next sync.
//   - If aikata.yaml workflows: does not already include the domain, add
//     it via config.Save (atomic).
//   - A short `## Workflow` pointer is appended to AGENTS.md when it does
//     not already reference the guide (idempotent).
//
// The result is idempotent: re-running on a fully-registered project is
// a no-op + notice.
func (workflowComponent) Add(ctx AddContext) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("components: workflow: domain is required (usage: aikata enable workflow <domain>; known: %s)", strings.Join(Workflows(), ", "))
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: workflow: target directory is required")
	}
	domain := strings.ToLower(strings.TrimSpace(ctx.Args[0]))
	if domain == "" {
		return fmt.Errorf("components: workflow: domain cannot be blank")
	}
	if !workflowKnown(domain) {
		return fmt.Errorf("components: workflow: unknown workflow %q (known: %s)", domain, strings.Join(Workflows(), ", "))
	}

	// LoadMigrated picks up any pending v1 -> v2 schema migration (ADR
	// 0016) so the subsequent config.Save lands at the current schema.
	cfg, _, err := config.LoadMigrated(ctx.TargetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("components: workflow: aikata config not found at %s (run `aikata init` first)", config.PrimaryPath(ctx.TargetDir))
		}
		return fmt.Errorf("components: workflow: load config: %w", err)
	}

	guideRel := "docs/workflows/" + domain + ".md"
	guideFile := filepath.Join(ctx.TargetDir, "docs", "workflows", domain+".md")
	fileExists := false
	if _, statErr := os.Stat(guideFile); statErr == nil {
		fileExists = true
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("components: workflow: stat %s: %w", guideFile, statErr)
	}

	cfgHasWorkflow := contains(cfg.Workflows, domain)

	// Always render — the rendered bytes are the manifest ancestor hash
	// whether or not the on-disk file already exists (same invariant as
	// the stack / singleFile components). RenderWorkflowGuide is shared
	// with scaffold's sync upstream rendering so both produce a
	// byte-identical guide.
	guideMap, rerr := RenderWorkflowGuide(domain, SingleFileParams{
		Lang:        ctx.Lang,
		ProjectName: ctx.ProjectName,
		Clock:       ctx.Clock,
	})
	if rerr != nil {
		return rerr
	}
	rendered := guideMap[guideRel]

	pointerNeeded, perr := agentsNeedsWorkflowPointer(ctx.TargetDir, guideRel)
	if perr != nil {
		return perr
	}

	if ctx.DryRun {
		w := stdout(ctx)
		if !fileExists {
			if _, werr := fmt.Fprintf(w, "Would write %s\n", guideRel); werr != nil {
				return werr
			}
		}
		if !cfgHasWorkflow {
			if _, werr := fmt.Fprintf(w, "Would add %q to .aikata/aikata.yaml workflows:\n", domain); werr != nil {
				return werr
			}
		}
		if pointerNeeded {
			if _, werr := fmt.Fprintf(w, "Would add a Workflow pointer to AGENTS.md\n"); werr != nil {
				return werr
			}
		}
		return nil
	}

	if !fileExists {
		if err := os.MkdirAll(filepath.Dir(guideFile), 0o755); err != nil {
			return fmt.Errorf("components: workflow: mkdir %s: %w", filepath.Dir(guideFile), err)
		}
		if err := os.WriteFile(guideFile, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("components: workflow: write %s: %w", guideFile, err)
		}
		if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", guideRel); werr != nil {
			return werr
		}
	}

	// Record the guide in the manifest unconditionally (ADR 0014).
	if rerr := RecordInManifest(ctx.TargetDir, map[string]string{guideRel: rendered}); rerr != nil {
		return rerr
	}

	if !cfgHasWorkflow {
		cfg.Workflows = append(cfg.Workflows, domain)
		sort.Strings(cfg.Workflows)
		if err := config.Save(ctx.TargetDir, cfg); err != nil {
			return fmt.Errorf("components: workflow: save config: %w", err)
		}
		if _, werr := fmt.Fprintf(stdout(ctx),
			"updated .aikata/aikata.yaml workflows: += %q\n", domain); werr != nil {
			return werr
		}
	}

	// Inject the AGENTS.md pointer (ADR 0026 D2). AGENTS.md is the
	// canonical instruction file and is NOT re-recorded in the manifest,
	// so the pointer survives `aikata sync` as a user-only-edit (ADR 0025).
	if pointerNeeded {
		if err := injectWorkflowPointer(ctx.TargetDir, guideRel, ctx.Lang); err != nil {
			return err
		}
		if _, werr := fmt.Fprintf(stdout(ctx), "updated AGENTS.md (added Workflow pointer)\n"); werr != nil {
			return werr
		}
	}

	if fileExists && cfgHasWorkflow && !pointerNeeded {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: workflow %q already present (%s + aikata.yaml + AGENTS.md pointer)\n", domain, guideRel); werr != nil {
			return werr
		}
	}
	return nil
}

// RenderWorkflowGuide renders the built-in guide for one workflow
// domain, keyed by its target path docs/workflows/<domain>.md. Shared by
// the enable-tier Add path and scaffold's sync upstream rendering so the
// guide a project syncs against is byte-identical to the one `enable`
// wrote. Unknown domains error rather than rendering nothing.
func RenderWorkflowGuide(domain string, p SingleFileParams) (map[string]string, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if !workflowKnown(domain) {
		return nil, fmt.Errorf("components: workflow: unknown workflow %q (known: %s)", domain, strings.Join(Workflows(), ", "))
	}
	if p.ProjectName == "" {
		return nil, fmt.Errorf("components: workflow: project name is required")
	}
	langDir, _, err := templates.LangDir("workflows/"+domain, p.Lang)
	if err != nil {
		return nil, fmt.Errorf("components: workflow: %w", err)
	}
	tmplPath := langDir + "/" + domain + ".md.tmpl"
	data := map[string]any{
		"ProjectName": p.ProjectName,
		"Lang":        p.Lang,
	}
	content, err := templates.Render(tmplPath, data, p.Clock)
	if err != nil {
		return nil, fmt.Errorf("components: workflow: render %s: %w", tmplPath, err)
	}
	return map[string]string{"docs/workflows/" + domain + ".md": content}, nil
}

func workflowKnown(domain string) bool {
	for _, w := range bundledWorkflows() {
		if w == domain {
			return true
		}
	}
	return false
}

// agentsNeedsWorkflowPointer reports whether AGENTS.md should receive a
// pointer to guideRel. It returns false (no-op) when AGENTS.md is absent
// or already references the guide, keeping `enable workflow` idempotent
// and safe on projects whose canonical file the user removed.
func agentsNeedsWorkflowPointer(targetDir, guideRel string) (bool, error) {
	body, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("components: workflow: read AGENTS.md: %w", err)
	}
	return !strings.Contains(string(body), guideRel), nil
}

// workflowPointerSection returns the localized `## Workflow` pointer
// section linking guideRel. The injected text must match the language of
// the surrounding canonical AGENTS.md (which is rendered from the
// `--lang` preset template) so a ja project does not gain an English
// section. Unknown langs fall back to English, mirroring templates.LangDir.
func workflowPointerSection(guideRel, lang string) string {
	if lang == "ja" {
		return fmt.Sprintf(
			"## ワークフロー (Workflow)\n\n存在する場合は [`%s`](./%s) に従い、ブランチ・コミット・PR・\nレビュー・マージ・リリース・CI ゲートの規則を守ること。\n",
			guideRel, guideRel)
	}
	return fmt.Sprintf(
		"## Workflow\n\nIf present, follow [`%s`](./%s) for branch, commit, PR,\nreview, merge, release, and CI gate rules.\n",
		guideRel, guideRel)
}

// injectWorkflowPointer appends a short, language-matched `## Workflow`
// section to AGENTS.md referencing guideRel. The caller must have
// verified via agentsNeedsWorkflowPointer that the pointer is absent. The
// link form is used so `aikata doctor` validates the reference resolves.
func injectWorkflowPointer(targetDir, guideRel, lang string) error {
	path := filepath.Join(targetDir, "AGENTS.md")
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("components: workflow: read AGENTS.md: %w", err)
	}
	section := workflowPointerSection(guideRel, lang)
	out := string(body)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	out += "\n" + section
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		return fmt.Errorf("components: workflow: write AGENTS.md: %w", err)
	}
	return nil
}
