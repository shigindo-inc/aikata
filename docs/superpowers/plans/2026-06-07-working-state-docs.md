---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-07
audience: [human, agent]
---

# Working State Documentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Clarify aikata's context model so generated projects understand `docs/memory/` as long-term memory, `docs/tasks/current.md` as short-lived working state, and `docs/adr/` as decision anchors.

**Architecture:** Keep the existing `docs/tasks/current.md` single-entry design for v0.11.1 and strengthen the documentation around it. Do not rename `docs/tasks/` or add task-detail files by default; capture the optional per-task-file model as a future design question because it changes how agents choose the entry point.

**Tech Stack:** Go scaffold templates, Markdown frontmatter docs, golden scaffold tests.

---

## Release Target

- Treat this as a patch release candidate for **v0.11.1**.
- Local evidence: `README.md` status is v0.11.0, `CHANGELOG.md` latest release is `0.11.0 - 2026-06-05`, local tag list contains `v0.11.0`, and `[Unreleased]` is empty.
- Do not bump plugin/marketplace versions unless the execution task explicitly cuts the release. For the implementation PR, add the changelog entry under `[Unreleased]`.

## File Structure

Modify generated source templates:

- `internal/templates/data/presets/standard/en/docs/tasks/current.md.tmpl`
- `internal/templates/data/presets/standard/ja/docs/tasks/current.md.tmpl`
- `internal/templates/data/presets/typescript/en/docs/tasks/current.md.tmpl`
- `internal/templates/data/presets/typescript/ja/docs/tasks/current.md.tmpl`
- `internal/templates/data/presets/flutter/en/docs/tasks/current.md.tmpl`
- `internal/templates/data/presets/flutter/ja/docs/tasks/current.md.tmpl`

Modify generated operating-rule templates:

- `internal/templates/data/presets/standard/en/AGENTS.md.tmpl`
- `internal/templates/data/presets/standard/ja/AGENTS.md.tmpl`
- `internal/templates/data/presets/typescript/en/AGENTS.md.tmpl`
- `internal/templates/data/presets/typescript/ja/AGENTS.md.tmpl`
- `internal/templates/data/presets/flutter/en/AGENTS.md.tmpl`
- `internal/templates/data/presets/flutter/ja/AGENTS.md.tmpl`

Modify memory README templates:

- `internal/templates/data/memory/en/README.md.tmpl`
- `internal/templates/data/memory/ja/README.md.tmpl`

Modify aikata's own planning docs:

- `docs/decisions/open-questions.md`
- `CHANGELOG.md`

Regenerate golden fixtures:

- `testdata/golden/**/AGENTS.md`
- `testdata/golden/**/docs/tasks/current.md`
- `testdata/golden/*-with-memory/docs/memory/README.md`

Optional distribution skill update if the implementation scope includes the first-party skills:

- `dist/universal-skill/aikata-context/SKILL.md`
- `dist/claude-code/skill/aikata-context/SKILL.md`
- `dist/claude-code/plugin/skills/aikata-context/SKILL.md`
- `dist/codex/plugin/skills/aikata-context/SKILL.md`

---

### Task 1: Strengthen `current.md` Templates

**Files:**
- Modify: `internal/templates/data/presets/standard/en/docs/tasks/current.md.tmpl`
- Modify: `internal/templates/data/presets/typescript/en/docs/tasks/current.md.tmpl`
- Modify: `internal/templates/data/presets/flutter/en/docs/tasks/current.md.tmpl`
- Modify: `internal/templates/data/presets/standard/ja/docs/tasks/current.md.tmpl`
- Modify: `internal/templates/data/presets/typescript/ja/docs/tasks/current.md.tmpl`
- Modify: `internal/templates/data/presets/flutter/ja/docs/tasks/current.md.tmpl`

- [ ] **Step 1: Replace the English guidance block**

Use this wording in all English `docs/tasks/current.md.tmpl` files:

```md
> This file is the agent's **short-term working state** for the current
> in-flight work. Rewrite it freely as work progresses.
>
> This directory is **not** a task archive, backlog, or project-management
> database. Keep only the state needed to resume the current work. Move
> durable facts or preferences to `docs/memory/`, design decisions to
> `docs/adr/`, requirements to `SPEC.md`, and invariant rules to
> [AGENTS.md](../../AGENTS.md).{{if .WithMemory}} Long-term context that should
> survive across sessions belongs in `docs/memory/`.{{end}}
```

- [ ] **Step 2: Replace the Japanese guidance block**

Use this wording in all Japanese `docs/tasks/current.md.tmpl` files:

```md
> 本ファイルは、現在進行中の作業のためのエージェントの**短期作業状態**です。
> 作業の進行に合わせて自由に書き換えて構いません。
>
> このディレクトリはタスク履歴、バックログ、プロジェクト管理 DB ではありません。
> 現在の作業を再開するために必要な状態だけを残します。永続化すべき事実や
> 好みは `docs/memory/`、設計判断は `docs/adr/`、要件は `SPEC.md`、
> 不変ルールは [AGENTS.md](../../AGENTS.md) に移します。{{if .WithMemory}}
> セッションをまたいで保持したい長期コンテキストは `docs/memory/` に置きます。{{end}}
```

- [ ] **Step 3: Keep the existing outline**

Leave `## Status`, `## Next`, and `## Notes` as the stable shape. Do not add `docs/tasks/README.md`, `next.md`, numbered files, or per-task slug files in this release.

---

### Task 2: Clarify Slot Separation in `AGENTS.md` Templates

**Files:**
- Modify: `internal/templates/data/presets/standard/en/AGENTS.md.tmpl`
- Modify: `internal/templates/data/presets/typescript/en/AGENTS.md.tmpl`
- Modify: `internal/templates/data/presets/flutter/en/AGENTS.md.tmpl`
- Modify: `internal/templates/data/presets/standard/ja/AGENTS.md.tmpl`
- Modify: `internal/templates/data/presets/typescript/ja/AGENTS.md.tmpl`
- Modify: `internal/templates/data/presets/flutter/ja/AGENTS.md.tmpl`

- [ ] **Step 1: Update the English before-start label**

Change the generated label from:

```md
[`docs/tasks/current.md`](./docs/tasks/current.md) — short-term working state
```

to:

```md
[`docs/tasks/current.md`](./docs/tasks/current.md) — short-term working state
```

- [ ] **Step 2: Add an English slot-separation section**

Insert after `## 4. Hard rules` or before `## 5. When stuck`, keeping numbering coherent:

```md
## 5. Context slots

Keep these slots separate:

| Slot | Lifetime | Use for |
|---|---|---|
| `AGENTS.md` | Long, invariant | Rules that must always hold |
| `docs/memory/` | Long, mutable | Durable facts, preferences, and references |
| `docs/tasks/current.md` | Short, rewriteable | Current in-flight work state |
| `docs/adr/` | Permanent | Design decisions and their rationale |

Do not use `docs/tasks/current.md` as a backlog or task archive. Move
durable information out of it before handoff.
```

If `.WithMemory` conditionals make `docs/memory/` absent, keep the row but phrase it as "when enabled" if needed:

```md
| `docs/memory/` (when enabled) | Long, mutable | Durable facts, preferences, and references |
```

- [ ] **Step 3: Add the Japanese equivalent**

Use:

```md
## 5. コンテキストの置き場所

次の置き場所を混同しないでください。

| 置き場所 | 寿命 | 用途 |
|---|---|---|
| `AGENTS.md` | 長期・不変 | 常に守るルール |
| `docs/memory/` | 長期・可変 | 永続的な事実、好み、参照情報 |
| `docs/tasks/current.md` | 短期・書き換え可 | 現在進行中の作業状態 |
| `docs/adr/` | 永続 | 設計判断とその理由 |

`docs/tasks/current.md` をバックログやタスク履歴として使わないでください。
引き継ぎ前に、永続化すべき情報は適切な置き場所へ移します。
```

For no-memory cases, use `docs/memory/` row with `有効な場合`.

- [ ] **Step 4: Renumber following headings**

If the new section becomes `## 5`, change existing `## 5. When stuck` to `## 6. When stuck` in English templates and `## 5. 困ったとき` to `## 6. 困ったとき` in Japanese templates.

---

### Task 3: Align Memory README Templates

**Files:**
- Modify: `internal/templates/data/memory/en/README.md.tmpl`
- Modify: `internal/templates/data/memory/ja/README.md.tmpl`

- [ ] **Step 1: Update English terminology**

Ensure `docs/tasks/current.md` is described as:

```md
**Working state** — current in-flight work state
```

Replace "Not a TODO list. Use `docs/tasks/current.md`." with:

```md
- **Not a TODO list or task archive.** Use `docs/tasks/current.md` only
  for short-lived in-flight work state.
```

- [ ] **Step 2: Update Japanese terminology**

Ensure `docs/tasks/current.md` is described as "作業状態", and replace the TODO-list warning with:

```md
- **TODO リストやタスク履歴ではありません。** `docs/tasks/current.md` は
  短期の進行中作業状態にだけ使います。
```

---

### Task 4: Record the Future Multi-File Question

**Files:**
- Modify: `docs/decisions/open-questions.md`

- [ ] **Step 1: Update Q-DESIGN-05**

Add the outcome of this discussion:

```md
- **2026-06-07 direction**: keep `docs/tasks/current.md` as the single
  mandatory entry point for v0.11.1. Clarify that `docs/tasks/` is
  short-lived working state, not a backlog or task archive.
- **Deferred option**: allow opt-in per-work files such as
  `docs/tasks/<slug>.md` for long-running or parallel work, with
  `current.md` remaining the required index / active pointer. Avoid
  `next.md` or numbered `next_XX.md` because they obscure which file an
  agent must read first.
```

Do not mark Q-DESIGN-05 closed unless an ADR is added.

---

### Task 5: Add Changelog Entry

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add an `[Unreleased]` entry**

Add:

```md
### Changed

- Clarified the generated context-slot model: `docs/memory/` is long-term
  memory, `docs/tasks/current.md` is short-lived working state rather
  than a backlog, and `docs/adr/` anchors design decisions. The default
  task slot remains a single `current.md` entry point for v0.11.1 while
  optional per-work task files remain a deferred design question.
```

---

### Task 6: Regenerate Golden Fixtures

**Files:**
- Modify: `testdata/golden/**`

- [ ] **Step 1: Run golden update**

Run:

```bash
go test ./internal/scaffold/... -update
```

Expected: PASS, and golden fixture files update to match the modified templates.

- [ ] **Step 2: Inspect changed golden files**

Run:

```bash
git diff -- testdata/golden
```

Expected: only generated `AGENTS.md`, `docs/tasks/current.md`, and memory README fixtures change. No unexpected scaffold files should be added.

---

### Task 7: Verify

**Files:**
- No direct file edits.

- [ ] **Step 1: Run scaffold tests**

Run:

```bash
go test ./internal/scaffold/...
```

Expected: PASS.

- [ ] **Step 2: Run full test suite**

Run:

```bash
make test
```

Expected: PASS.

- [ ] **Step 3: Run lint**

Run:

```bash
make lint
```

Expected: PASS.

- [ ] **Step 4: Run doctor**

Run:

```bash
go run ./cmd/aikata doctor
```

Expected: no errors. Warnings, if any, must be reviewed and reported.

---

### Task 8: Optional Skill Distribution Alignment

**Files:**
- Modify: `dist/universal-skill/aikata-context/SKILL.md`
- Modify: `dist/claude-code/skill/aikata-context/SKILL.md`
- Modify: `dist/claude-code/plugin/skills/aikata-context/SKILL.md`
- Modify: `dist/codex/plugin/skills/aikata-context/SKILL.md`

- [ ] **Step 1: Decide whether the release includes skill docs**

If the implementation is intended to update the published first-party skills, align their wording with the new slot model:

```md
| In-flight state for this task | `docs/tasks/current.md` | Short-lived working state; not a backlog or archive. |
```

- [ ] **Step 2: Keep version metadata unchanged**

Do not bump plugin metadata here unless this task becomes a release-cut task. Version lockstep belongs to release preparation, not this documentation clarification.

---

## Self-Review

- Spec coverage: The plan addresses the user's requested conceptual split, keeps `tasks/` naming stable, avoids standard multi-file task state, and captures the per-task-file idea as a future option.
- Placeholder scan: No implementation step uses "TBD" or asks for unspecified tests.
- Type consistency: All paths match the current repository layout; golden update command matches `internal/scaffold/golden_test.go`.
