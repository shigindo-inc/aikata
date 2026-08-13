# `modeling` Capability & `model-feature` Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship an opt-in `modeling` capability that scaffolds `docs/usecases.md` + `docs/domain.md`, and a first-party `model-feature` skill that fills that pair one feature at a time.

**Architecture:** `modeling` is a multi-file capability (like `memory`, not like `singleFile`) rendering two fixed paths from `internal/templates/data/components/modeling/<lang>/`. It flips `components.modeling` in schema v2 and joins the `doctor` managed surface. The skill is authored in `dist/universal-skill/model-feature/SKILL.md` and mirrored byte-identically into the three other distribution channels, which `internal/repolint/distribution_test.go` enforces. No new CLI verb, no new `doctor` check.

**Tech Stack:** Go 1.x (stdlib + `gopkg.in/yaml.v3` + cobra), Go table-driven tests, embedded `text/template` templates, Markdown.

**Spec:** [`docs/decisions/2026-08-14-modeling-capability-and-model-feature-skill-design.md`](../../decisions/2026-08-14-modeling-capability-and-model-feature-skill-design.md)

## Global Constraints

- **Never write "DDD" or "Domain-Driven Design"** in any template, skill body, ADR, or user-facing string (spec D7). LLM priors would import aggregates/repositories unbidden.
- **Capability name is exactly `modeling`; skill name is exactly `model-feature`.** ADR 0017 forbids compatibility aliases, so a rename after release is a breaking change.
- **`docs/design/` is taken** by ADR 0031 (brand artifacts). Never write there.
- **Zero new CLI machinery** (spec D8): no `aikata modeling …` verb, no new `doctor` check. The skill calls only existing `fill` / `doctor` / `map`.
- **Do not create `docs/workflows/design.md`** (spec D9).
- **Target release: v0.15.0** (current released version is 0.14.0). Per the project's version-lockstep rule, `dist/claude-code/plugin/.claude-plugin/plugin.json`, `dist/codex/plugin/.codex-plugin/plugin.json`, and `.claude-plugin/marketplace.json` (both version fields) bump together in Task 8.
- **Every new/edited Markdown doc carries the frontmatter contract**: `project`, `status`, `version`, `updated`, `audience` — `aikata doctor` validates these on the managed surface.
- **Commit messages are English** (project convention), Conventional Commits style. Do not add `Co-Authored-By` or agent trailers.
- **Do not run `git push`.** Commits only.

---

### Task 1: Schema — `components.modeling`

Adds the durable flag before anything can flip it. No renderer yet.

**Files:**
- Modify: `internal/config/aikata_yaml.go:71-87` (the `Components` struct)
- Modify: `internal/config/schema_migrate.go:230-257` (`upsertComponentsBlock`)
- Modify: `internal/components/component.go:182-230` (`EnableComponentInConfig` switch)
- Test: `internal/config/schema_migrate_test.go`
- Test: `internal/components/modeling_config_test.go` (create)

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `config.Components.Modeling bool` (yaml key `modeling`); `components.EnableComponentInConfig(targetDir, "modeling") error` accepts `"modeling"`.

- [ ] **Step 1: Write the failing migration test**

Append to `internal/config/schema_migrate_test.go`:

```go
func TestUpsertComponentsBlock_SeedsModelingKey(t *testing.T) {
	body := []byte("version: 1\nproject:\n  name: legacy\n  lang: en\n")
	got, migrated, err := MigrateAikataYaml(body)
	if err != nil {
		t.Fatalf("MigrateAikataYaml: %v", err)
	}
	if !migrated {
		t.Fatalf("v1 payload should report migrated=true")
	}
	if got.Components.Modeling {
		t.Errorf("modeling should default to false on migration, got true")
	}
	out, err := Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), "modeling:") {
		t.Errorf("migrated config must carry an explicit modeling key:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestUpsertComponentsBlock_SeedsModelingKey -v`
Expected: FAIL — compile error `got.Components.Modeling undefined`.

- [ ] **Step 3: Add the struct field**

In `internal/config/aikata_yaml.go`, inside `type Components struct`, after the `Env` field:

```go
	// Modeling records the opt-in use-case ledger and domain model at
	// docs/usecases.md + docs/domain.md. New in v0.15.0; pre-v0.15.0 v2
	// configs omit the key and read as false.
	Modeling bool `yaml:"modeling"`
```

- [ ] **Step 4: Seed the key in both migration branches**

In `internal/config/schema_migrate.go`, `upsertComponentsBlock`:

In the "block already exists" branch, after `ensureBoolKey(existing, "env", false)`:

```go
		ensureBoolKey(existing, "modeling", false)
```

In the fresh-block branch, after `scalarString("env"), scalarBool(false),`:

```go
		scalarString("modeling"), scalarBool(false),
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestUpsertComponentsBlock_SeedsModelingKey -v`
Expected: PASS

- [ ] **Step 6: Write the failing enable-flag test**

Create `internal/components/modeling_config_test.go`:

```go
package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shigindo-inc/aikata/internal/config"
)

func TestEnableComponentInConfig_FlipsModeling(t *testing.T) {
	tmp := t.TempDir()
	cfg := config.Default("demo", "en")
	body, err := config.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	dir := filepath.Join(tmp, ".aikata")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "aikata.yaml"), body, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := EnableComponentInConfig(tmp, "modeling"); err != nil {
		t.Fatalf("EnableComponentInConfig: %v", err)
	}

	got, _, err := config.LoadMigrated(tmp)
	if err != nil {
		t.Fatalf("LoadMigrated: %v", err)
	}
	if !got.Components.Modeling {
		t.Errorf("components.modeling = false, want true")
	}
}
```

- [ ] **Step 7: Run test to verify it fails**

Run: `go test ./internal/components/ -run TestEnableComponentInConfig_FlipsModeling -v`
Expected: FAIL — `components.modeling = false, want true` (the switch has no `"modeling"` case, so it falls through without setting anything).

- [ ] **Step 8: Add the switch case**

In `internal/components/component.go`, in the `EnableComponentInConfig` switch, add a case alongside the existing ones (keep the existing early-return-if-already-true shape):

```go
	case "modeling":
		if cfg.Components.Modeling {
			return nil
		}
		cfg.Components.Modeling = true
```

- [ ] **Step 9: Run tests to verify they pass**

Run: `go test ./internal/config/ ./internal/components/ -v`
Expected: PASS (all tests in both packages)

- [ ] **Step 10: Commit**

```bash
git add internal/config/aikata_yaml.go internal/config/schema_migrate.go internal/config/schema_migrate_test.go internal/components/component.go internal/components/modeling_config_test.go
git commit -m "feat(config): add components.modeling schema flag"
```

---

### Task 2: `modeling` capability — templates, renderer, registration

Delivers a working `aikata enable modeling`.

**Files:**
- Create: `internal/templates/data/components/modeling/en/usecases.md.tmpl`
- Create: `internal/templates/data/components/modeling/en/domain.md.tmpl`
- Create: `internal/templates/data/components/modeling/ja/usecases.md.tmpl`
- Create: `internal/templates/data/components/modeling/ja/domain.md.tmpl`
- Create: `internal/components/modeling.go`
- Modify: `internal/components/registry.go:10-22` (the `capabilities` slice)
- Test: `internal/components/modeling_test.go` (create)

**Interfaces:**
- Consumes: `components.EnableComponentInConfig(targetDir, "modeling")` (Task 1); existing `components.WriteIfMissing`, `components.RecordInManifest`, `templates.LangDir`, `templates.Render`, `templates.Clock`.
- Produces: `components.Modeling Component` (registry singleton); `components.RenderModeling(p ModelingParams) (map[string]string, error)` returning keys `"docs/usecases.md"` and `"docs/domain.md"`; `components.ModelingParams{Lang string; ProjectName string; Clock templates.Clock}`.

- [ ] **Step 1: Write the failing renderer test**

Create `internal/components/modeling_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/components/ -run TestModeling -v`
Expected: FAIL — compile error `undefined: Modeling`, `undefined: RenderModeling`, `undefined: ModelingParams`.

- [ ] **Step 3: Write the English use-case template**

Create `internal/templates/data/components/modeling/en/usecases.md.tmpl`:

```
---
project: {{.ProjectName}}
status: draft
version: 0.0.1
updated: {{now}}
audience: [human, agent]
---

# Use cases — {{.ProjectName}}

> One entry per externally observable behaviour: who acts, what triggers
> it, what success looks like, and how it fails. The structural
> counterpart is [`docs/domain.md`](./domain.md), whose field table links
> back to the IDs here. For what and why, read
> [`SPEC.md`](../SPEC.md).

Each use case owns an ID (`UC-NN`). `docs/domain.md` references these IDs
per field, so **never reuse an ID** after retiring a use case.

Four fields are required. Add `Precondition` only where it carries
weight — an optional field that is always empty teaches nothing.

---

## UC-01 — _TODO: the outcome an actor achieves, phrased as a goal_

- **Actor**: _TODO: who initiates this_
- **Trigger**: _TODO: what starts it_
- **Success state**: _TODO: what is true afterwards_
- **Main exception**: _TODO: the most likely way this fails, and what happens then_
```

- [ ] **Step 4: Write the English domain template**

Create `internal/templates/data/components/modeling/en/domain.md.tmpl`:

```
---
project: {{.ProjectName}}
status: draft
version: 0.0.1
updated: {{now}}
audience: [human, agent]
---

# Domain model — {{.ProjectName}}

> The entities this project reasons about: what they hold, what must
> always be true of them, and how they change state. The behavioural
> counterpart is [`docs/usecases.md`](./usecases.md). Terms introduced
> here belong in [`GLOSSARY.md`](../GLOSSARY.md); structure and
> implementation approach belong in
> [`ARCHITECTURE.md`](../ARCHITECTURE.md).

Every principal field carries a **Related UC**. This runs the check in
both directions: a use case whose success state needs data absent from
this file, and a field here that no use case reaches, are both visible.
The second direction is the point — it is what catches a speculative
field.

List only fields that carry meaning. Mechanical ones (`id`,
`created_at`) are omitted deliberately.

---

## Entities

### _TODO: EntityName_

- **Description**: _TODO: what one of these represents_
- **Invariant**: _TODO: what must always hold, however it is reached_

| Field | Type | Meaning | Related UC |
|---|---|---|---|
| _TODO_ | _TODO_ | _TODO_ | UC-01 |

- **Transitions**: _TODO: stateA → stateB (UC-01)_
```

- [ ] **Step 5: Write the Japanese templates**

Create `internal/templates/data/components/modeling/ja/usecases.md.tmpl`:

```
---
project: {{.ProjectName}}
status: draft
version: 0.0.1
updated: {{now}}
audience: [human, agent]
---

# ユースケース — {{.ProjectName}}

> 外から観測できるふるまい 1 つにつき 1 項目。誰が・何をきっかけに・
> 何を達成し・どう失敗するか。構造側の対になる文書は
> [`docs/domain.md`](./domain.md) で、そのフィールド表がここの ID を
> 参照する。何を・なぜは [`SPEC.md`](../SPEC.md) を参照。

各ユースケースは ID（`UC-NN`）を持つ。`docs/domain.md` がフィールド単位で
この ID を参照するため、**廃止した ID を再利用しない**こと。

必須は 4 項目。`事前条件` は意味があるときだけ足す — 常に空の任意項目は
何も伝えない。

---

## UC-01 — _TODO: アクターが達成する結果をゴールとして書く_

- **アクター**: _TODO: 誰が始めるか_
- **トリガー**: _TODO: 何をきっかけに始まるか_
- **成功状態**: _TODO: 完了後に何が真になっているか_
- **主な例外**: _TODO: 最も起こりやすい失敗と、そのときどうなるか_
```

Create `internal/templates/data/components/modeling/ja/domain.md.tmpl`:

```
---
project: {{.ProjectName}}
status: draft
version: 0.0.1
updated: {{now}}
audience: [human, agent]
---

# ドメインモデル — {{.ProjectName}}

> このプロジェクトが扱うエンティティ。何を持ち、何が常に真でなければ
> ならず、どう状態が変わるか。ふるまい側の対になる文書は
> [`docs/usecases.md`](./usecases.md)。ここで導入した用語は
> [`GLOSSARY.md`](../GLOSSARY.md) に、構造と実装方針は
> [`ARCHITECTURE.md`](../ARCHITECTURE.md) に置く。

主要フィールドはすべて **関連UC** を持つ。これにより検査が双方向になる
— ユースケースの成功状態に必要なのにここに無いデータと、ここにあるのに
どのユースケースからも辿れないフィールドの両方が見える。後者が要点で、
投機的なフィールドを捕まえるのはこちら。

意味を持つフィールドだけを書く。機械的なもの（`id` / `created_at`）は
意図的に省く。

---

## エンティティ

### _TODO: エンティティ名_

- **説明**: _TODO: これ 1 件が何を表すか_
- **不変条件**: _TODO: どの経路を通っても常に成り立つべきこと_

| フィールド | 型 | 意味 | 関連UC |
|---|---|---|---|
| _TODO_ | _TODO_ | _TODO_ | UC-01 |

- **状態遷移**: _TODO: 状態A → 状態B (UC-01)_
```

- [ ] **Step 6: Write the component**

Create `internal/components/modeling.go`:

```go
package components

import (
	"fmt"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// modelingComponent provides the opt-in document pair docs/usecases.md
// (behaviour) + docs/domain.md (structure). The two are one capability
// because they are read and edited as a pair: docs/domain.md links back
// to use-case IDs per field, so half the pair cannot discharge either
// side of that check.
//
// It cannot use singleFile (single targetPath by construction), so it
// follows memoryComponent's multi-file shape with fixed paths rather
// than a template-tree walk.
type modelingComponent struct{}

// Modeling is the singleton registered in the capabilities registry.
var Modeling Component = modelingComponent{}

func (modelingComponent) Name() string { return "modeling" }
func (modelingComponent) Description() string {
	return "Use-case ledger and domain model at docs/usecases.md + docs/domain.md."
}
func (modelingComponent) Status() string { return StatusActive }

// modelingFiles maps the target path to its template file name under
// components/modeling/<lang>/.
var modelingFiles = map[string]string{
	"docs/usecases.md": "usecases.md.tmpl",
	"docs/domain.md":   "domain.md.tmpl",
}

// ModelingParams carries the inputs RenderModeling needs. Reduced form
// of scaffold.Options so the init-time and enable-time paths share one
// renderer.
type ModelingParams struct {
	Lang        string
	ProjectName string
	Clock       templates.Clock
}

// RenderModeling returns the pair keyed by target path relative to the
// project root ("docs/usecases.md", "docs/domain.md").
func RenderModeling(p ModelingParams) (map[string]string, error) {
	if p.ProjectName == "" {
		return nil, fmt.Errorf("components: modeling: project name is required")
	}
	langDir, _, err := templates.LangDir("components/modeling", p.Lang)
	if err != nil {
		return nil, fmt.Errorf("components: modeling: %w", err)
	}
	data := map[string]any{
		"ProjectName": p.ProjectName,
		"Lang":        p.Lang,
	}
	out := make(map[string]string, len(modelingFiles))
	for rel, tmplName := range modelingFiles {
		tmplPath := langDir + "/" + tmplName
		content, rerr := templates.Render(tmplPath, data, p.Clock)
		if rerr != nil {
			return nil, fmt.Errorf("components: modeling: render %s: %w", tmplPath, rerr)
		}
		out[rel] = content
	}
	return out, nil
}

// Add renders the pair under ctx.TargetDir. Existing files are
// preserved, so re-running is idempotent and never clobbers a
// hand-edited document.
func (m modelingComponent) Add(ctx AddContext) error {
	if ctx.ProjectName == "" {
		return fmt.Errorf("components: modeling: project name is required")
	}
	if ctx.TargetDir == "" {
		return fmt.Errorf("components: modeling: target directory is required")
	}
	rendered, err := RenderModeling(ModelingParams{
		Lang:        ctx.Lang,
		ProjectName: ctx.ProjectName,
		Clock:       ctx.Clock,
	})
	if err != nil {
		return err
	}

	if ctx.DryRun {
		for _, rel := range sortedKeys(rendered) {
			if _, werr := fmt.Fprintf(stdout(ctx), "Would write %s\n", rel); werr != nil {
				return werr
			}
		}
		return nil
	}

	written, skipped, err := WriteIfMissing(ctx.TargetDir, rendered)
	if err != nil {
		return err
	}
	// ADR 0014: register both paths so the next `aikata sync` treats
	// them as aikata-managed templates.
	if err := RecordInManifest(ctx.TargetDir, rendered); err != nil {
		return err
	}
	if err := EnableComponentInConfig(ctx.TargetDir, "modeling"); err != nil {
		return err
	}
	if written == 0 {
		if _, werr := fmt.Fprintf(stderr(ctx),
			"notice: modeling already present (%d file(s)); nothing to do\n", skipped); werr != nil {
			return werr
		}
		return nil
	}
	for _, rel := range sortedKeys(rendered) {
		if _, werr := fmt.Fprintf(stdout(ctx), "wrote %s\n", rel); werr != nil {
			return werr
		}
	}
	return nil
}
```

- [ ] **Step 7: Register the capability**

In `internal/components/registry.go`, add `Modeling,` to the `capabilities` slice, keeping alphabetical order (between `Memory` and `Monorepo`):

```go
	Memory,
	Modeling,
	Monorepo,
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test ./internal/components/ -run TestModeling -v`
Expected: PASS (all five `TestModeling*` tests)

- [ ] **Step 9: Verify end-to-end against the real CLI**

```bash
go build -o /tmp/aikata-dev ./cmd/aikata
TMP=$(mktemp -d) && /tmp/aikata-dev init demo --dir "$TMP" && /tmp/aikata-dev enable modeling --dir "$TMP" && ls "$TMP"/docs/ && grep -n "modeling" "$TMP"/.aikata/aikata.yaml
```

Expected: `docs/usecases.md` and `docs/domain.md` listed; `modeling: true` in the config. If `--dir` is not the flag name, run `/tmp/aikata-dev init --help` and use the equivalent; do not skip this verification.

- [ ] **Step 10: Commit**

```bash
git add internal/templates/data/components/modeling internal/components/modeling.go internal/components/modeling_test.go internal/components/registry.go
git commit -m "feat(components): add modeling capability (docs/usecases.md + docs/domain.md)"
```

---

### Task 3: init-time wiring — `--with-modeling`

Makes the capability reachable at `aikata init` and re-derivable by `aikata fill`.

**Files:**
- Modify: `internal/scaffold/scaffold.go:52-56` (Options), `:244` (render table), `:338` (config write-back)
- Modify: `internal/cli/init.go:99-106`, `:124`, `:143`, `:183`, `:208-215`
- Modify: `internal/cli/prompt.go:26`, `:46`, `:179`
- Modify: `internal/fill/fill.go:295`
- Test: `internal/scaffold/` existing test file for options coverage; `internal/cli/init_test.go` if present

**Interfaces:**
- Consumes: `components.RenderModeling(ModelingParams)` and `config.Components.Modeling` (Tasks 1–2).
- Produces: `scaffold.Options.WithModeling bool`; CLI flag `--with-modeling`.

- [ ] **Step 1: Write the failing scaffold test**

Create `internal/scaffold/modeling_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_WithModelingWritesPairAndFlagsConfig(t *testing.T) {
	tmp := t.TempDir()
	opts := Options{
		ProjectName:  "demo",
		TargetDir:    tmp,
		Lang:         "en",
		WithModeling: true,
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, rel := range []string{"docs/usecases.md", "docs/domain.md"} {
		if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected %s: %v", rel, err)
		}
	}
	body, err := os.ReadFile(filepath.Join(tmp, ".aikata", "aikata.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), "modeling: true") {
		t.Errorf("config must record modeling: true:\n%s", body)
	}
}
```

**Before running this:** read `internal/scaffold/scaffold_test.go` and mirror its happy-path `Options` literal, adding only `WithModeling: true`. `Options` may require additional mandatory fields (`Preset`, `Scope`, `Clock`) that this snippet omits; copy them rather than guessing. The same applies to `Run`'s exact signature.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/scaffold/ -run TestRun_WithModelingWritesPairAndFlagsConfig -v`
Expected: FAIL — compile error `unknown field WithModeling in struct literal`.

- [ ] **Step 3: Add the scaffold option**

In `internal/scaffold/scaffold.go`, in `Options`, after the `WithPrompts` field:

```go
	// WithModeling provisions the opt-in use-case ledger and domain
	// model at docs/usecases.md + docs/domain.md.
	WithModeling bool
```

- [ ] **Step 4: Add the render-table entry**

In `internal/scaffold/scaffold.go` near line 244, alongside the `WithPrompts` entry (match the surrounding entries' exact shape — `sfp` is the shared params value already in scope):

```go
		{opts.WithModeling, func() (map[string]string, error) {
			return components.RenderModeling(components.ModelingParams{
				Lang: sfp.Lang, ProjectName: sfp.ProjectName, Clock: sfp.Clock,
			})
		}},
```

If `components.ModelingParams` field names do not line up with `sfp`'s, construct it from the same values the neighbouring `RenderPrompts(sfp)` call receives.

- [ ] **Step 5: Add the config write-back**

In `internal/scaffold/scaffold.go` near line 338, after `cfg.Components.Prompts = opts.WithPrompts`:

```go
	cfg.Components.Modeling = opts.WithModeling
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -run TestRun_WithModelingWritesPairAndFlagsConfig -v`
Expected: PASS

- [ ] **Step 7: Add the CLI flag**

In `internal/cli/init.go`:

Declare the variable next to `withPrompts` / `withEnv`, then register the flag after line 215:

```go
	cmd.Flags().BoolVar(&withModeling, "with-modeling", false, "include a use-case ledger and domain model at docs/usecases.md + docs/domain.md")
```

Add the four wiring sites, mirroring `WithPrompts` exactly:

- near line 105: `WithModeling: cmd.Flags().Changed("with-modeling"),`
- near line 124: `WithModeling: withModeling,`
- near line 143: `withModeling = result.WithModeling`
- near line 183: `WithModeling: withModeling,`

Update the doc comment at `internal/cli/init.go:18-19` to include `--with-modeling`.

- [ ] **Step 8: Add the interactive prompt entry**

In `internal/cli/prompt.go`, add `WithModeling bool` to both structs (near line 26 and line 46), then add to the prompt table near line 179:

```go
		{skip.WithModeling, &result.WithModeling, "Include docs/usecases.md + docs/domain.md (use cases and domain model)?"},
```

- [ ] **Step 9: Wire the fill path**

In `internal/fill/fill.go`, after line 295:

```go
		opts.WithModeling = cfg.Components.Modeling
```

- [ ] **Step 10: Run the full suite**

Run: `go build ./... && go test ./...`
Expected: PASS

- [ ] **Step 11: Verify the flag end-to-end**

```bash
go build -o /tmp/aikata-dev ./cmd/aikata
TMP=$(mktemp -d) && /tmp/aikata-dev init demo --dir "$TMP" --with-modeling && ls "$TMP"/docs/
```

Expected: `usecases.md` and `domain.md` present after a single `init`.

- [ ] **Step 12: Commit**

```bash
git add internal/scaffold internal/cli/init.go internal/cli/prompt.go internal/fill/fill.go
git commit -m "feat(cli): wire --with-modeling through init, prompt, and fill"
```

---

### Task 4: `doctor` managed surface

Puts the pair under default validation. No new check is added (spec D8) — only the existing frontmatter/link/freshness walk gains two paths.

**Files:**
- Modify: `internal/doctor/scope.go:26-41` (`managedDocGlobs`)
- Test: `internal/doctor/scope_test.go` (create if absent)

**Interfaces:**
- Consumes: nothing new.
- Produces: `doctor.ManagedIncludeGlobs(dir)` includes `docs/usecases.md` and `docs/domain.md`.

- [ ] **Step 1: Write the failing test**

Append to `internal/doctor/scope_test.go` (create the file with `package doctor` if it does not exist):

```go
func TestManagedIncludeGlobs_CoversModelingPair(t *testing.T) {
	globs := ManagedIncludeGlobs(t.TempDir())
	want := []string{"docs/usecases.md", "docs/domain.md"}
	for _, w := range want {
		found := false
		for _, g := range globs {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("managed globs missing %q; got %v", w, globs)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/doctor/ -run TestManagedIncludeGlobs_CoversModelingPair -v`
Expected: FAIL — `managed globs missing "docs/usecases.md"`

- [ ] **Step 3: Add the globs**

In `internal/doctor/scope.go`, in `managedDocGlobs`, after `"docs/prompts.md",`:

```go
	// The opt-in modeling pair. Listed individually (not as a subtree)
	// because they are two fixed files, and validated whenever a
	// project has enabled the capability.
	"docs/usecases.md",
	"docs/domain.md",
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/doctor/ -v`
Expected: PASS

- [ ] **Step 5: Verify doctor accepts the generated templates**

```bash
go build -o /tmp/aikata-dev ./cmd/aikata
TMP=$(mktemp -d) && /tmp/aikata-dev init demo --dir "$TMP" --with-modeling && /tmp/aikata-dev doctor --dir "$TMP"
```

Expected: 0 errors. If the freshly generated templates produce doctor errors, fix the **templates** (Task 2 files) until doctor is clean — a scaffold that fails its own validator is a defect.

- [ ] **Step 6: Commit**

```bash
git add internal/doctor/
git commit -m "feat(doctor): include the modeling document pair in the managed surface"
```

---

### Task 5: `model-feature` skill

Four distribution channels, byte-identical bodies, enforced by `internal/repolint/distribution_test.go`. Skill-only — **no** command wrapper (ADR 0043 D4 precedent: `track-context` ships skill-only).

**Files:**
- Create: `dist/universal-skill/model-feature/SKILL.md` (canonical)
- Create: `dist/universal-skill/model-feature/agents/openai.yaml`
- Create: `dist/codex/plugin/skills/model-feature/SKILL.md`
- Create: `dist/codex/plugin/skills/model-feature/agents/openai.yaml`
- Create: `dist/claude-code/plugin/skills/model-feature/SKILL.md`
- Create: `dist/claude-code/skill/model-feature/SKILL.md`
- Modify: `internal/repolint/distribution_test.go:19` (`firstPartySkills`)

**Interfaces:**
- Consumes: `docs/usecases.md` / `docs/domain.md` (Task 2) as the destinations it writes into.
- Produces: a skill named `model-feature`, referenced by name from `track-context` in Task 6.

- [ ] **Step 1: Add the skill to the parity list (failing test first)**

In `internal/repolint/distribution_test.go` line 19:

```go
var firstPartySkills = []string{"manage-docs", "track-context", "refresh-docs", "migrate-structure", "model-feature"}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/repolint/ -v`
Expected: FAIL — missing `dist/universal-skill/model-feature/SKILL.md` and the three mirrors.

- [ ] **Step 3: Write the canonical SKILL.md**

Create `dist/universal-skill/model-feature/SKILL.md`:

````markdown
---
name: model-feature
user-invocable: false
description: Use when designing one feature whose externally observable behaviour changes, in a repository that has aikata's modeling capability enabled (`docs/usecases.md` and `docs/domain.md` present). Walks a single feature from use case to domain model to a hand-off point just before implementation — writing the use case into docs/usecases.md, propagating required data into docs/domain.md, fixing new terms in GLOSSARY.md, and recording a decision in docs/adr/ when alternatives existed. Do not use for refactors, bug fixes, copy changes, or UI position adjustments (no observable behaviour change), in repositories without docs/usecases.md, or to carry out the implementation itself — this skill stops at the hand-off. For the daily context loop use `track-context`; for CLI invocation use `manage-docs`.
---

# model-feature

Design one feature at a time by writing it into the canonical documents,
in an order that keeps the data model honest: **behaviour first,
structure second**. Modelling data before behaviour is what admits
fields nothing actually needs.

This skill owns only the mapping from content to canonical destination.
It does not replace whatever thinking or planning process the project
already uses, and it stops before implementation.

## When this applies

Run this loop when **externally observable behaviour changes**. That is
the whole firing criterion — do not run it for refactors, bug fixes,
copy tweaks, or UI position adjustments.

It requires `docs/usecases.md` and `docs/domain.md`. If they are absent,
the repository has not enabled the capability; hand off to `manage-docs`
to run `aikata enable modeling` and confirm with the user before
proceeding.

Run it once per feature. A project kickoff is just this loop run several
times.

## The loop

### Step 1 — Write the use case

Add or update an entry in `docs/usecases.md`. Give a new use case the
next free `UC-NN` id; never reuse a retired id.

**Done when** the entry states an actor, a trigger, and a success state,
**and names at least one exception path**. The exception is required:
it is what distinguishes a use case from a feature list, and it is the
seed of a test scenario later.

### Step 2 — Propagate to the domain model

Update `docs/domain.md` with whatever entities, fields, invariants, and
state transitions this use case implies. Put every new term into
`GLOSSARY.md`.

**Done when both directions hold:**

- **Forward** — every piece of data the use case's success state needs
  exists in `docs/domain.md`.
- **Reverse** — no field or state you just added is unreachable from any
  use case.

The reverse direction is the point of this step. Forward-only catches
omissions but not drift; only the reverse direction catches a field that
nothing needs. When it fails you have two honest options — delete the
field, or write the use case that justifies it. Say which one you are
taking.

Expect to go **back to Step 1** here. Sketching the model routinely
invalidates a use case; that round trip is the loop working, not a
mistake.

### Step 3 — Record a decision, if there was one

If the shape you chose had real alternatives, write an ADR under
`docs/adr/` (hand off to `manage-docs` for `aikata new adr "<title>"`).

Skip this step when there was no genuine fork. Most features have none.

### Step 4 — Check and hand off

- Run `aikata doctor` (via `manage-docs`) and report the result.
- Check `GLOSSARY.md` has no unused terms.
- Note the in-flight feature in `docs/tasks/current.md` — **one line**.
  Step-by-step progress is not tracked anywhere; that is working state,
  not knowledge.

Then stop and hand off to the project's normal implementation flow.
Implementation is explicitly outside this skill.

## Reporting, not gating

Unmet conditions are **reported, not enforced**. State plainly what is
unsatisfied and let the human decide whether to proceed. Never block,
and never silently fill a gap with a guess.

## Document shapes

`docs/usecases.md` — one section per use case, four required fields:

```markdown
## UC-01 — Cancel an order
- Actor: buyer
- Trigger: chooses "cancel" on the order detail screen
- Success state: order becomes `cancelled`; a refund is scheduled
- Main exception: already shipped ⇒ cancellation refused, routed to returns
```

`docs/domain.md` — per entity, with `Related UC` at **field**
granularity. That granularity is deliberate: entity-level linkage is too
coarse to reveal an unreachable field, which is the failure this exists
to catch.

```markdown
### Order
- Description: one confirmed purchase by a buyer
- Invariant: an order that reached `cancelled` cannot leave it

| Field | Type | Meaning | Related UC |
|---|---|---|---|
| status | enum | progress state of the order | UC-01, UC-03 |
| cancel_reason | text | why it was cancelled | UC-01 |

- Transitions: placed → shipped (UC-02) / placed → cancelled (UC-01)
```

List only fields that carry meaning. Mechanical ones (`id`,
`created_at`) are omitted.

## Boundaries

- **Traceability stays inside `docs/`.** Do not push use-case ids into
  test names or source comments. aikata does not read code, so such
  links cannot be verified and will rot.
- **No new files.** Everything lands in `docs/usecases.md`,
  `docs/domain.md`, `GLOSSARY.md`, `docs/adr/`, or the one line in
  `docs/tasks/current.md`.
- **No progress file.** "Which step are we on" is working state.
- Sibling skills: `track-context` (daily context loop), `manage-docs`
  (raw CLI surface).
````

- [ ] **Step 4: Write the Codex agent descriptor**

Create `dist/universal-skill/model-feature/agents/openai.yaml`:

```yaml
interface:
  display_name: "model-feature (aikata)"
  short_description: "Design one feature from use case to domain model before implementing"
  default_prompt: "Use $model-feature when a change alters externally observable behaviour: write the use case, propagate it into the domain model, then hand off."
policy:
  allow_implicit_invocation: true
```

- [ ] **Step 5: Mirror into the three other channels**

```bash
mkdir -p dist/codex/plugin/skills/model-feature/agents \
         dist/claude-code/plugin/skills/model-feature \
         dist/claude-code/skill/model-feature
cp dist/universal-skill/model-feature/SKILL.md dist/codex/plugin/skills/model-feature/SKILL.md
cp dist/universal-skill/model-feature/agents/openai.yaml dist/codex/plugin/skills/model-feature/agents/openai.yaml
cp dist/universal-skill/model-feature/SKILL.md dist/claude-code/plugin/skills/model-feature/SKILL.md
cp dist/universal-skill/model-feature/SKILL.md dist/claude-code/skill/model-feature/SKILL.md
```

Do **not** create `dist/claude-code/plugin/commands/model-feature.md` — this skill ships skill-only.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/repolint/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add dist/universal-skill/model-feature dist/codex/plugin/skills/model-feature dist/claude-code/plugin/skills/model-feature dist/claude-code/skill/model-feature internal/repolint/distribution_test.go
git commit -m "feat(skills): add model-feature skill across distribution channels"
```

---

### Task 6: Point `track-context` at `model-feature`

One line, four byte-identical copies.

**Files:**
- Modify: `dist/universal-skill/track-context/SKILL.md`
- Modify: `dist/codex/plugin/skills/track-context/SKILL.md`
- Modify: `dist/claude-code/plugin/skills/track-context/SKILL.md`
- Modify: `dist/claude-code/skill/track-context/SKILL.md`

**Interfaces:**
- Consumes: the skill name `model-feature` (Task 5).
- Produces: nothing consumed by later tasks.

- [ ] **Step 1: Edit the canonical copy**

In `dist/universal-skill/track-context/SKILL.md`, in the section listing sibling hand-offs (around the existing `## When to hand off to \`manage-docs\`` section), add:

```markdown
## When to hand off to `model-feature`

If the work about to start changes **externally observable behaviour**
(a new or altered feature, not a refactor or bug fix), use the
`model-feature` skill first: it writes the use case into
`docs/usecases.md`, propagates the required data into `docs/domain.md`,
and hands back just before implementation. Skip it for refactors, bug
fixes, copy changes, and UI position adjustments.
```

Also append to the sibling-skills line near the end of the file:

```markdown
- Sibling skill: `model-feature` (per-feature design loop into `docs/usecases.md` + `docs/domain.md`).
```

- [ ] **Step 2: Mirror to the other three channels**

```bash
cp dist/universal-skill/track-context/SKILL.md dist/codex/plugin/skills/track-context/SKILL.md
cp dist/universal-skill/track-context/SKILL.md dist/claude-code/plugin/skills/track-context/SKILL.md
cp dist/universal-skill/track-context/SKILL.md dist/claude-code/skill/track-context/SKILL.md
```

- [ ] **Step 3: Run tests to verify parity holds**

Run: `go test ./internal/repolint/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add dist/universal-skill/track-context dist/codex/plugin/skills/track-context dist/claude-code/plugin/skills/track-context dist/claude-code/skill/track-context
git commit -m "docs(skills): hand off from track-context to model-feature"
```

---

### Task 7: Boundary ADR and canonical document updates

The durable WHY, plus every index that must know the pair exists.

**Files:**
- Create: `docs/adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md`
- Modify: `GLOSSARY.md`
- Modify: `docs/layout.md` (§2 table)
- Modify: `docs/decisions/open-questions.md`
- Modify: `SPEC.md` (§4.2 capability list), `ARCHITECTURE.md` (§3 generated structure)

**Interfaces:**
- Consumes: everything shipped in Tasks 1–6.
- Produces: ADR 0047, cited by the CHANGELOG entry in Task 8.

- [ ] **Step 1: Write ADR 0047**

Create `docs/adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md`. Follow the exact structure of `docs/adr/0046-structure-migration-assistant-boundary.md` (read it first for the house style). Required content:

- **Status**: Accepted; **Date**: 2026-08-14; **Deciders**: aikata maintainers
- **Related**: ADR 0007, 0017, 0026, 0028, 0031, 0043, 0045, 0046
- **Context**: the observed gap — nothing owns the band between personas and functional requirements, and nothing owns the domain model. Cite the design note as the rationale source.
- **Decision D1**: `modeling` is an opt-in capability rendering `docs/usecases.md` + `docs/domain.md` as a pair.
- **Decision D2**: the use-case ledger lives under `docs/`, not in `SPEC.md` — discharging ADR 0007 via update-frequency asymmetry.
- **Decision D3**: `model-feature` is a per-feature design loop that **ends before implementation**, is advisory not gating, and fires on one criterion (externally observable behaviour changes).
- **Decision D4**: traceability closes inside `docs/` at field granularity; it is never extended into code or tests (SPEC §8).
- **Decision D5**: zero new CLI machinery and no new `doctor` check.
- **Consequences**, including this negative stated plainly: **the per-feature granularity is a pre-measurement design decision and may be superseded** once the loop has been run on real features.
- **Alternatives considered**: use cases as a `SPEC.md` section; a paired `docs/workflows/design.md` convention file (deferred per spec D9); extending `track-context` instead of a new skill; entity-granular rather than field-granular traceability.

- [ ] **Step 2: Add the GLOSSARY entry**

In `GLOSSARY.md`, add an entry that pins the two senses apart (spec D7). Match the file's existing entry format:

- **document-centered** — aikata's existing positioning: the unit of truth is the markdown document rather than a rules artifact (SPEC §1.2). Describes **structure**.
- **per-feature design loop** — the ordering `use case → domain model → decision → hand-off` walked by the `model-feature` skill. Describes **process**, and sits on top of the document-centered structure. Deliberately not abbreviated, to avoid collision with the unrelated established meaning of "DDD".

- [ ] **Step 3: Add the layout.md rows**

In `docs/layout.md` §2, add to the table:

```markdown
| `docs/usecases.md` | Use-case ledger (behaviour) | `enable modeling` | [ADR 0047](./adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md) |
| `docs/domain.md` | Domain model (structure) | `enable modeling` | [ADR 0047](./adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md) |
```

- [ ] **Step 4: Record the deferred items**

In `docs/decisions/open-questions.md`, add two entries matching the file's existing question format:

- **Test strategy** — extend the `tdd` component and/or add `docs/workflows/testing.md`. Deferred. Re-entry criterion: a concrete instance of agent/human disagreement about test layers or scope. Note that use-case exception paths are already the natural input.
- **Data handling & privacy** — collected fields, permissions, store declarations. Deferred. Re-entry criterion: work actually blocked on a store declaration or permission design. Open sub-question: generic capability vs. the `flutter` stack brief, since iOS and Android declaration formats differ. Note that `docs/domain.md`'s field table is the intended carrier.

Do **not** add either to `ROADMAP.md`.

- [ ] **Step 5: Update SPEC.md and ARCHITECTURE.md**

- `SPEC.md` §4.2 — add `modeling` to the enumerated `aikata enable <capability>` targets.
- `ARCHITECTURE.md` §3 — add `docs/usecases.md` and `docs/domain.md` to the generated-structure coverage, marked opt-in.

- [ ] **Step 6: Verify the docs pass their own validator**

Run: `go build -o /tmp/aikata-dev ./cmd/aikata && /tmp/aikata-dev doctor`
Expected: 0 errors on this repository. Fix any frontmatter or broken-link errors introduced by the new/edited docs before committing.

- [ ] **Step 7: Commit**

```bash
git add docs/adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md GLOSSARY.md docs/layout.md docs/decisions/open-questions.md SPEC.md ARCHITECTURE.md
git commit -m "docs(adr): record the modeling capability and per-feature design loop boundary"
```

---

### Task 8: Release wiring — v0.15.0

**Files:**
- Modify: `CHANGELOG.md` (the `## [Unreleased]` section)
- Modify: `dist/claude-code/plugin/.claude-plugin/plugin.json`
- Modify: `dist/codex/plugin/.codex-plugin/plugin.json`
- Modify: `.claude-plugin/marketplace.json` (**both** version fields — line 4 and line 14)
- Modify: `ROADMAP.md`
- Regenerate: `.aikata/docmap.yaml`, `.aikata/docmap.md`

**Interfaces:**
- Consumes: all prior tasks.
- Produces: a releasable v0.15.0 tree.

- [ ] **Step 1: Write the CHANGELOG entry**

Under `## [Unreleased]` in `CHANGELOG.md`, add (matching the file's existing subsection style):

```markdown
### Added

- `aikata enable modeling` / `aikata init --with-modeling` scaffolds an
  opt-in document pair: `docs/usecases.md` (use-case ledger) and
  `docs/domain.md` (domain model, with per-field use-case links).
  Both join the default `aikata doctor` managed surface. (ADR 0047)
- `model-feature` first-party skill — a per-feature design loop that
  writes a use case, propagates it into the domain model, fixes new
  terms in `GLOSSARY.md`, and hands off before implementation. Fires
  only when externally observable behaviour changes; advisory, never
  gating. (ADR 0047)

### Changed

- `track-context` now hands off to `model-feature` when the work about
  to start changes externally observable behaviour.
```

- [ ] **Step 2: Bump the version in lockstep**

Set `0.15.0` in all three files (four fields total):

```bash
grep -n '"version"' dist/claude-code/plugin/.claude-plugin/plugin.json dist/codex/plugin/.codex-plugin/plugin.json .claude-plugin/marketplace.json
```

Edit each to `0.15.0`, then re-run the grep to confirm no `0.14.0` remains.

- [ ] **Step 3: Update ROADMAP.md**

Add a v0.15.0 milestone entry describing the shipped capability and skill. Do not list the deferred items from Task 7 Step 4.

- [ ] **Step 4: Regenerate the doc map**

```bash
go build -o /tmp/aikata-dev ./cmd/aikata && /tmp/aikata-dev map
```

- [ ] **Step 5: Run the full verification suite**

```bash
go build ./... && go test ./... && /tmp/aikata-dev doctor
```

Expected: build clean, all tests PASS, doctor 0 errors. **Report the actual output.** If anything fails, fix it before the commit — do not commit a red tree.

- [ ] **Step 6: Confirm the working tree holds only intended changes**

```bash
git status --porcelain
git branch --show-current
```

Expected: only the files this plan touched. If anything unexpected appears, stop and report rather than committing it.

- [ ] **Step 7: Commit**

```bash
git add CHANGELOG.md ROADMAP.md dist/claude-code/plugin/.claude-plugin/plugin.json dist/codex/plugin/.codex-plugin/plugin.json .claude-plugin/marketplace.json .aikata/docmap.yaml .aikata/docmap.md
git commit -m "chore(release): prepare v0.15.0"
```

---

## Post-implementation obligation

The spec ships this without the recommended local-validation-first
sequence, so §7 of the design note commits to one follow-up:

**Run `model-feature` on one real feature in an actual app repository
immediately after release**, and record what the per-feature granularity
actually felt like. If it is wrong, ADR 0047 is written to be superseded
rather than amended. Release is the start of validation, not the end of
the work.
