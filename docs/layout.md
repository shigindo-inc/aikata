---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-11
audience: [human, agent]
---

# Recommended Layout — the aikata structure on one page

> The single, cross-referenced index of **where each kind of document
> belongs** in an aikata project, what governs it, and whether
> `aikata doctor` validates it by default. This is the **prescriptive**
> counterpart to the **descriptive** [doc map](./decisions/docmap-design.md)
> (`.aikata/docmap.{yaml,md}`, [ADR 0044](./adr/0044-doc-map-derived-artifact.md)):
> `layout.md` says *what should exist*; the doc map records *what does
> exist*; reconciling the two is a future migration assistant (ROADMAP
> v0.14.0).

This page is an **index, not a second source of truth**. The authorities it
points at are: [ARCHITECTURE.md §3](../ARCHITECTURE.md#3-generated-project-structure)
(the narrative of what `aikata init` produces), `internal/doctor/scope.go`
(`canonicalDocNames` / `managedDocGlobs`, the surface `aikata doctor`
validates by default, [ADR 0033](./adr/0033-doctor-default-scope-direction.md)
/ [ADR 0037](./adr/0037-tighten-adoption-mutation-boundaries.md)), and the
per-component ADRs cited below. When they disagree, **they win** and this
page is corrected.

---

## 1. Top-level documents

Top-level minimalism caps the project root at **8 non-hidden files** after
`aikata init --scope standard` ([GLOSSARY](../GLOSSARY.md#top-level-minimalism)).

| Path | Role | Scope / capability | Governed by | doctor-managed? |
|---|---|---|---|---|
| `README.md` | Thin nav for humans + LLMs | standard | — | ✅ canonical |
| `AGENTS.md` | Canonical agent instructions (rules) | standard | [ADR 0002](./adr/0002-agents-md-as-canonical.md) | ✅ canonical |
| `SPEC.md` | What / why | standard | — | ✅ canonical |
| `ARCHITECTURE.md` | How | standard | — | ✅ canonical |
| `GLOSSARY.md` | Terminology pin (ja/en) | standard | — | ✅ canonical |
| `ROADMAP.md` | Direction | `--oss` / extended | — | ✅ canonical |
| `CHANGELOG.md` | Release notes | `--with-changelog` | — | ✅ canonical |
| `UI.md` | UI / UX guidelines | `--with-ui` | — | ✅ canonical |
| `API.md` | API interface spec | `--with-api` | — | ✅ canonical |
| `CONTRIBUTING.md` / `SECURITY.md` | Contributor / security policy | `--oss` (v1.0) | — | prose (not frontmatter-validated) |

"✅ canonical" = present in `canonicalDocNames`; validated for frontmatter,
links, and freshness under the default `doctor` scope.

## 2. `docs/` subtrees (scaffolded for user projects)

The surface `aikata doctor` manages by default (`managedDocGlobs`).

| Path | Role | Scope / capability | Governed by |
|---|---|---|---|
| `docs/adr/` | One ADR per decision; immutable once Accepted | standard | [ADR 0001](./adr/0001-record-architecture-decisions.md) |
| `docs/tasks/current.md` | Short-term working state (not a backlog) | standard | [SPEC §4.1](../SPEC.md) |
| `docs/troubleshooting.md` | Common failures & fixes | standard | — |
| `docs/memory/` | Long-term agent memory (`user`/`feedback`/`project`/`reference`) | `--with-memory` | [ADR 0004](./adr/0004-long-term-memory-slot.md) |
| `docs/stacks/<name>.md` | Per-stack brief (code-free) | `--stack` | [ADR 0029](./adr/0029-stack-brief-layout-convention.md) |
| `docs/workflows/<domain>.md` | Opt-in collaboration policy | `enable workflow` | [ADR 0026](./adr/0026-workflow-guides-as-opt-in-collaboration-docs.md) |
| `docs/design/` | One-off brand-exploration artifacts | `new app-icon` / `new mascot` | [ADR 0031](./adr/0031-brand-exploration-documents-as-one-off-artifacts.md) |
| `docs/prompts.md` | Reusable-prompt library | `enable prompts` | [ADR 0034](./adr/0034-reusable-prompts-opt-in-capability.md) |
| `docs/monorepo.md` + `apps/**/AGENTS.md` | Monorepo per-app instructions | `enable monorepo` | — |

## 3. Machine zone (`.aikata/`) and transient paths

Hidden, aikata-owned state — does **not** count against top-level
minimalism.

| Path | Role | Governed by |
|---|---|---|
| `.aikata/aikata.yaml` | Project config (schema v2) | [ADR 0008](./adr/0008-aikata-owned-config-directory.md) / [ADR 0016](./adr/0016-aikata-yaml-schema-v2.md) |
| `.aikata/manifest.yaml` | Rendered-file hashes for `sync` 3-way merge | [ADR 0014](./adr/0014-manifest-living-record.md) |
| `.aikata/docmap.yaml` · `.aikata/docmap.md` | Doc map (derived; not manifest-tracked) | [ADR 0044](./adr/0044-doc-map-derived-artifact.md) |
| `.aikata-proposed/` | Proposal tree from `init` in a non-empty dir (transient) | [ADR 0037](./adr/0037-tighten-adoption-mutation-boundaries.md) |
| Generated AI-tool artifacts (`CLAUDE.md`, `.cursor/rules/`, …) | Disposable; overwritten by `generate` | [ADR 0002](./adr/0002-agents-md-as-canonical.md) |

## 4. aikata-repo-internal only (NOT scaffolded for users)

These exist in **this repository** to develop aikata; `aikata init` does
**not** create them, and they are **not** in the default `doctor`
managed surface:

- `docs/decisions/` — open questions and design notes (this file lives here
  too).
- `docs/origin/` — historical planning docs (do not edit).
- `docs/roadmap-archive.md` — archived released-milestone detail
  ([ADR 0039](./adr/0039-documentation-hygiene-and-context-budget.md)).

---

## 5. The prescriptive / descriptive / reconcile triad

- **Prescriptive** — *this page* defines the recommended structure.
- **Descriptive** — the [doc map](./decisions/docmap-design.md) records the
  documents that actually exist, tagging each
  [managed or external](../GLOSSARY.md#managed--external-document).
- **Reconcile** — a future, opt-in **structure-migration assistant** (a
  skill, user-confirmed; ROADMAP v0.14.0) proposes moving off-structure
  (`external`) documents into the layout above. The aikata CLI stays
  **observation-only**; the agent performs moves under the skill, only with
  explicit user approval. The mutation boundary is to be fixed in its own
  ADR before implementation.
