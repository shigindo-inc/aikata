---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
audience: [human, agent]
---

# SPEC — What & Why

> This document describes **what** aikata does and **why**. For technical
> design (the **how**), read [ARCHITECTURE.md](./ARCHITECTURE.md). For the
> release timeline, read [ROADMAP.md](./ROADMAP.md). For terminology, read
> [GLOSSARY.md](./GLOSSARY.md).

---

## 1. Purpose

**aikata** is a lightweight CLI tool that, in a single command, scaffolds a
project with markdown documents and per-AI-tool configuration files designed
for the AI-coding era.

The name `aikata` (相方) means "partner" — a companion that helps human and
LLM collaborate as equals during development.

### 1.1 Problem statement

Modern projects must be readable by **both humans and multiple AI coding
agents** (Claude Code, Cursor, Codex, Gemini CLI, Copilot, Windsurf, …) at
the same time. Today this requires:

- Hand-writing several near-duplicate instruction files (`CLAUDE.md`,
  `.cursor/rules/*.mdc`, `.github/copilot-instructions.md`, `AGENTS.md`, …).
- Keeping them in sync as the project evolves.
- Inventing project-by-project conventions for **what** to document and
  **where** to put it.

Existing scaffolding tools (ai-rulez, Ruler, block/ai-rules, agentsmesh)
focus on **rules**, target **English-speaking enterprises**, and tend to be
heavy, all-in-one configurations.

### 1.2 Solution

aikata is an opinionated, **document-centered** scaffold. The unit of truth
is the markdown document that both humans and LLMs read, not the rules
artifact. Tool-specific files are **generated** from those canonical
documents and are disposable.

### 1.3 Differentiation

| Axis | Existing tools | aikata |
|---|---|---|
| Unit of truth | Rules | Documents |
| Audience | English / enterprise | Bilingual (ja / en) first, English default |
| Scope | All-in-one config | Minimal core + presets |
| Update model | Init-only | `init` + `add` + `sync` + `doctor` |
| Optional features | Tightly coupled | Strictly opt-in (Do-No-Harm Policy) |

---

## 2. Goals & Non-Goals

### 2.1 Goals (In Scope)

1. Generate a coherent set of project-root markdown documents
   (`README.md`, `AGENTS.md`, `SPEC.md`, `ARCHITECTURE.md`, `GLOSSARY.md`,
   …) for a new project.
2. Generate per-AI-tool configuration files (`CLAUDE.md`,
   `.cursor/rules/`, `.github/copilot-instructions.md`, …) from canonical
   documents.
3. Verify consistency of the document set (`aikata doctor`).
4. Merge upstream template updates without overwriting user edits
   (`aikata sync`).
5. Provide stack-specific presets (initial: `minimal`, `standard`; later
   `flutter`, `typescript`, …).
6. Support monorepo layouts (nested `AGENTS.md` per app).

### 2.2 Non-Goals (Out of Scope)

- IDE plugins (CLI only).
- Real-time rule enforcement (that is ai-rulez's domain).
- Automatic LLM-driven document generation (no API calls in v1).
- A GUI for editing AI-tool settings.
- Project-management features (issue tracking, task boards, etc.).

### 2.3 Users

| Tier | Persona | Need |
|---|---|---|
| Primary | Solo developers and small teams using **multiple AI tools** simultaneously | Coherent, bilingual scaffolding without per-tool maintenance |
| Secondary | OSS maintainers providing AI-friendly structure to contributors | Generate `CONTRIBUTING.md`, `AGENTS.md`, etc. consistently |
| Tertiary | Japanese-speaking teams with internal documentation conventions | Glossary-driven, low-translation-drift scaffolding |

---

## 3. Design Principles

aikata follows eight principles. They are the test against which every
feature request is judged.

1. **Human-LLM dual readable** — every document is equally useful to both.
2. **Convention over configuration** — minimal config to start, full
   override available.
3. **Do no harm** — opt-out features (Obsidian, TDD, Flutter, monorepo,
   long-term memory) never penalize users who don't adopt them.
   Codified in [ADR 0003](./docs/adr/0003-do-no-harm-policy.md).
4. **Top-level minimalism** — at most 8 non-hidden files at project root.
5. **Composable, not monolithic** — features ship as files, addable and
   removable individually.
6. **Stack-agnostic core, opinionated presets** — core has no stack
   knowledge; presets are where opinions live.
7. **Lossy generation is OK** — generated files may lose information; the
   canonical source remains authoritative.
8. **Rules ≠ memory ≠ working state** — three distinct agent-facing
   slots with distinct lifetimes: invariant **rules** live in
   `AGENTS.md`; mutable **long-term memory** (user preferences,
   project context, references) lives under `docs/memory/` (opt-in via
   `--with-memory`, planned v0.2); ephemeral **working state** lives
   in `docs/tasks/current.md` (created with the standard preset in
   v0.1). Codified in
   [ADR 0004](./docs/adr/0004-long-term-memory-slot.md).

---

## 4. Functional Requirements (CLI)

The CLI exposes six commands. Each is described in terms of **what** it
must do and the user-visible behavior. Implementation details live in
[ARCHITECTURE.md](./ARCHITECTURE.md).

### 4.1 `aikata init [name]`

**Purpose**: Scaffold a new project.

**Must be able to**:

- Create a default file set in an empty or new directory.
- Read flags `--scope`, `--stack`, `--with-ui`, `--with-api`,
  `--with-tdd`, `--with-changelog`, `--with-prompts`, `--with-memory`,
  `--oss`,
  `--monorepo`, `--lang ja|en`, `--ai-tools`, `--no-interactive`,
  `--dry-run`, `--force`. `--scope` (documentation breadth) and
  `--stack` (target technology) are orthogonal axes
  ([ADR 0024](./docs/adr/0024-scope-stack-axes-split.md), v0.8.2);
  `--preset` survives as a deprecated alias for them, removed in v1.0.
  (`--with-memory` enables the long-term agent memory slot defined in
  [ADR 0004](./docs/adr/0004-long-term-memory-slot.md); ships v0.2.)
- When run in an existing non-empty directory **without** `--force`:
  write proposed files under `.aikata-proposed/` instead of overwriting,
  and exit with a non-error message.
- In interactive mode (the default), ask: project name, scope, stack,
  language, AI tools, and the optional-component questions (long-term
  memory, UI, API, TDD, changelog, prompts). Each optional-component
  question maps 1:1 to its `--with-*` flag and defaults to N. Questions whose
  flag was explicitly set on the command line are silently skipped. The
  `extended` / OSS intent question remains scheduled for v1.0; the
  `--monorepo` flag is scheduled for v0.6.
- Default scope: `standard`. Default stack: none (stack-agnostic).
  Default `--ai-tools`: `claude`. Default `--lang`: `en`.

**Acceptance**:

- `aikata init my-app --scope minimal --no-interactive` produces a
  predictable file set matching the `minimal` golden test.
- Re-running in the same directory without `--force` writes nothing outside
  `.aikata-proposed/`.

### 4.2 `aikata enable <capability>` / `aikata new <artifact>`

**Purpose**: Extend an existing aikata project after `init`. The verb is
split by lifetime ([ADR 0017](./docs/adr/0017-post-init-command-taxonomy.md)):
`enable` records a durable capability; `new` stamps a one-off artifact.
The pre-v0.7.1 `aikata add` parent was removed without an alias.

**`aikata enable <capability>`** must be able to enable:

- A single-file component (`enable ui` → `UI.md`, `enable api` → `API.md`,
  plus `tdd`, `changelog`, and `prompts` → `docs/prompts.md`, the opt-in
  reusable-prompt library per
  [ADR 0034](./docs/adr/0034-reusable-prompts-opt-in-capability.md)).
- The long-term memory slot (`enable memory`).
- A stack (`enable stack flutter`) and an AI-tool target
  (`enable ai-tool cursor`).
- The monorepo layout (`enable monorepo`) and a workflow guide
  (`enable workflow git`).

Each leaf renders its files, records them in `.aikata/manifest.yaml`, and
flips the matching `components.*` flag or appends to the
`stacks` / `ai_tools` / `workflows` list in `.aikata/aikata.yaml`.

**`aikata new <artifact>`** must be able to stamp one-off authoring
scaffolds with no durable schema change:

- A new ADR with a title (`new adr "use go modules"`).
- Brand-exploration documents (`new app-icon` →
  `docs/design/app-icon-concepts.md`, `new mascot` →
  `docs/design/mascot-character-ideas.md`) per
  [ADR 0031](./docs/adr/0031-brand-exploration-documents-as-one-off-artifacts.md).
  These are not recorded in the manifest and `aikata sync` does not
  restore them; a second `new` on an existing file refuses rather than
  overwriting.

### 4.3 `aikata doctor`

**Purpose**: Check consistency of the document set.

**Must check**:

- Every file `AGENTS.md` links to exists.
- Every term defined in `GLOSSARY.md` is used at least once elsewhere
  (warning only).
- No ADR is in the `Deprecated` status without a replacement reference.
- Every variable in `.env.example` is mentioned in `AGENTS.md` or
  `ARCHITECTURE.md` (when `.env.example` is present; it is opt-in via the
  `env` capability from v0.9.7).
- Document `updated:` fields are not more than 365 days old (warning).
- Every markdown file has the required frontmatter keys.

**Scope**: The Markdown-walking checks (frontmatter / updated / glossary)
validate **only the document surface aikata manages** by default — the
canonical top-level documents, the known `docs/` subtrees (`adr`,
`memory`, `stacks`, `tasks`, `workflows`, `design`), and any
manifest-tracked Markdown. Third-party Markdown governed by another
contract (Claude Code `skills/**`, vendored docs, project-owned prose) is
left alone. Pass `--all-markdown` to audit every Markdown file in the
tree; `doctor.exclude` ([ADR 0021](./docs/adr/0021-doctor-scope-and-exclusion.md))
subtracts paths under either mode. The default flip and broad-audit opt-in
are recorded in
[ADR 0033](./docs/adr/0033-doctor-default-scope-direction.md) and
[ADR 0037](./docs/adr/0037-tighten-adoption-mutation-boundaries.md).

**Output**:

- `error` (blocking) / `warning` / `info` levels.
- Exit codes: 0 (no errors), 3 (errors found).
- `--fix` auto-fixes safe issues (e.g. stale `updated:` after a known edit).
- `--json` emits a machine-readable report on stdout. The schema is
  versioned for forward compatibility:

  ```json
  {
    "version": 1,
    "issues": [
      {
        "level": "error",
        "file": "SPEC.md",
        "line": 12,
        "code": "frontmatter.missing-key.version",
        "message": "frontmatter missing required key \"version\""
      }
    ],
    "summary": { "errors": 1, "warnings": 0, "info": 0 }
  }
  ```

  `line` and `code` are omitted when empty so consumers do not need
  to distinguish "missing field" from "zero value". When `--json` is
  combined with `--fix`, the post-fix report is emitted (the text
  pre-print is suppressed so the JSON stream stays clean). Errors
  from cobra itself are written to stderr and never mixed into the
  JSON stream.

### 4.4 `aikata generate`

**Purpose**: Produce per-AI-tool configuration files from canonical sources.

**Must be able to**:

- Read `.aikata/aikata.yaml` and determine enabled AI tools.
- For each enabled tool, emit its expected file(s).
- Add generated artifact paths to `.gitignore` of the **target project** by
  default (see Do-No-Harm policy for the exception in aikata itself).

### 4.5 `aikata update`

**Purpose**: Update the aikata CLI itself.

**Must be able to**:

- Check whether a newer aikata release is available when explicitly
  requested.
- Update native / installer-managed aikata binaries without requiring
  Go.
- Delegate package-manager installs to their package manager (`brew`,
  npm, etc.) instead of overwriting managed files directly.
- Print an actionable manual command when the install source is
  unknown or cannot be updated safely.

`aikata update` follows the user expectation established by tools such
as Claude Code: `update` means "update this CLI". Network access is
explicit and limited to release metadata and release assets needed for
the update operation. See
[ADR 0009](./docs/adr/0009-update-command-owns-cli-version-updates.md).

### 4.6 `aikata sync`

**Purpose**: Merge upstream template updates into the user's project.

**Must be able to**:

- Detect template version drift between bundled templates and the user's
  files.
- Present diffs interactively with accept / reject per hunk.
- Preserve user edits (no silent overwrite) — **durably**, across
  repeated syncs. A user-edited / upstream-unchanged file
  (`user-only-edit`) keeps its divergence on every run; the manifest
  ancestor is regenerated from the upstream rendering, never from the
  post-merge on-disk bytes (ADR 0025 D1).
- Honor a per-file ownership opt-out: paths listed under
  `sync.own` in `.aikata/aikata.yaml` report the `owned` status and are
  never rendered-compared, conflict-markered, overwritten, or
  manifest-tracked (ADR 0025 D2). Same glob matcher as `doctor.exclude`.
- Re-anchor an existing manifest on demand with `--reseed` (manifest-
  only write; no source file touched). `--rebaseline` seeds a *missing*
  manifest and emits a notice pointing at `--reseed` if one already
  exists (ADR 0025 D4).

`.gitignore` is managed by the ADR 0018 managed-append writer
(non-destructive by default); the removed `docs.generate_gitignore`
flag (ADR 0025 D3) was inert, and a project that wants sync to leave any
file alone lists it under `sync.own`.

Inspired by Copier's update flow. `aikata generate` may warn that a
project is behind bundled templates, but it must not silently perform
this sync because canonical project documents may contain user edits.

### 4.7 `aikata list`

**Purpose**: List available scopes, stacks, capabilities, artifacts, and
AI-tool integrations.

**Must output**:

- Built-in scopes (`minimal`, `standard`) and stacks (`flutter`,
  `typescript`).
- Available capabilities (`list capabilities`: memory, ui, api, tdd,
  changelog, prompts, monorepo, stack, ai-tool, workflow) and one-off
  artifacts (`list artifacts`: adr, app-icon, mascot).
- AI tools the current aikata binary supports.

### 4.8 `aikata fill`

**Purpose**: Bring a repository to a complete canonical document set by
writing only the documents that are **missing**, never overwriting an
existing file. Adoption / completion verb (ADR 0042); resolves
Q-INTEROP-03.

**Behaviour**:

- Option-free and idempotent. Scope is inferred: from `.aikata/`
  (manifest preset/lang + `aikata.yaml` components) when the project is
  already managed; otherwise the `standard` scope with project name = the
  working-directory basename, adopting the repo (writes `aikata.yaml` +
  `manifest.yaml`).
- Existing files — including hand-edited ones — are left byte-for-byte
  untouched (`components.WriteIfMissing`).
- The manifest is rebuilt from the rendered (upstream) hashes
  (`components.RecordInManifest`), so a hand-edited file's ancestor is the
  upstream rendering and a subsequent `aikata sync` classifies it as
  `user-only-edit` and preserves it.
- Distinct from `init` (scaffolds new / proposes in a non-empty dir),
  `sync` (pulls upstream changes, respects deletions), and `enable` (one
  capability, requires config).

---

## 5. Non-Functional Requirements

### 5.1 Reliability

- File generation is **atomic**: a failed `init` must leave the target
  directory unchanged.
- No `panic`s in user-facing flows. Errors are returned and rendered
  with actionable messages.

### 5.2 Performance

- `aikata init --scope standard` completes in **< 500 ms** on a 2025-era
  laptop.
- Binary size **< 20 MB**.

### 5.3 Portability

- Supports macOS (Intel / ARM), Linux (x86_64 / ARM64), Windows (x86_64).
- All file paths use OS-neutral separators.

### 5.4 Compatibility

- The bundled template schema is versioned. Templates produced by aikata
  v0.1 remain readable by aikata v1.x.
- `.aikata/aikata.yaml` has a top-level `version` field for future
  migrations. v0.7.4 removes the pre-v0.3.2 `.ai/aikata.yaml` read
  fallback; projects must use `.aikata/aikata.yaml`.

### 5.5 Internationalization

- Default document language: **English** (`--lang en`).
- Japanese (`--lang ja`) is a first-class target, not an after-thought,
  but is **planned for v0.2** (see [ROADMAP.md](./ROADMAP.md)).
- A bilingual mode (Japanese for humans + English for LLMs in the same file)
  is under design — see
  [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md).

---

## 6. The Do-No-Harm Policy (Summary)

A summary; the full ADR is
[ADR 0003](./docs/adr/0003-do-no-harm-policy.md).

aikata's optional features must not impose cost on users who don't adopt
them. Concretely:

- **Obsidian**: standard markdown links only; no `[[wikilinks]]`; no
  `.obsidian/`; no Dataview queries in canonical files.
- **TDD**: testing documentation only when `--with-tdd`; no testing rules
  leak into `AGENTS.md` otherwise.
- **Flutter and other stacks**: isolated to `docs/stacks/`, included into
  `AGENTS.md` only when enabled.
- **Workflow guides**: collaboration policy isolated to `docs/workflows/`,
  enabled only via `aikata enable workflow <domain>`; `AGENTS.md` gains a
  short pointer (never the full policy) only when a guide is enabled
  ([ADR 0026](./docs/adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)).
- **Monorepo**: no nested layout unless `--monorepo`.
- **AI tools**: only `claude` by default; others opt-in via `--ai-tools`.
- **Language**: English is the default to maximize reach.

---

## 7. Hypotheses to Validate

aikata is built on four hypotheses; each has an explicit success criterion
in [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md).

- **H1** — Users of multiple AI tools prefer aikata over heavier
  ai-rulez-style configuration.
- **H2** — A document-centered scaffold is more adoptable than a
  rules-centered one.
- **H3** — Flutter developers actively want an AI-coding scaffold for their
  stack.
- **H4** — Being identifiably a Japanese OSS strengthens adoption rather
  than weakens it.

Dogfooding by the author plus a small community circle is the initial
validation channel.

---

## 8. Out-of-Scope Examples (For Clarity)

To prevent scope creep:

- aikata does **not** call any LLM API.
- aikata does **not** modify source code files of the target project.
- aikata does **not** install dependencies for the target project.
- aikata does **not** ship VS Code / JetBrains extensions in v1.

These boundaries may be revisited per [ROADMAP.md](./ROADMAP.md) §"v1.x".

---

## 9. References

- Open questions:
  [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md)
- [agents.md open spec](https://agents.md/)
