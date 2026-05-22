---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
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
| Update model | Init-only | `init` + `add` + `update` + `doctor` |
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
   (`aikata update`).
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
   `--with-memory`, planned v0.2); ephemeral **working memory** lives
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
- Read flags `--preset`, `--with-ui`, `--with-api`, `--with-tdd`,
  `--with-changelog`, `--with-memory`, `--oss`, `--monorepo`,
  `--lang ja|en`, `--ai-tools`, `--minimal`, `--no-interactive`,
  `--dry-run`, `--force`. (`--with-memory` enables the long-term
  agent memory slot defined in
  [ADR 0004](./docs/adr/0004-long-term-memory-slot.md); ships v0.2.)
- When run in an existing non-empty directory **without** `--force`:
  write proposed files under `.aikata-proposed/` instead of overwriting,
  and exit with a non-error message.
- In interactive mode (the default), ask: project name, language, AI tools,
  stack preset, optional features (UI / API / TDD / monorepo), and OSS
  intent.
- Default preset: `standard`. Default `--ai-tools`: `claude`. Default
  `--lang`: `en`.

**Acceptance**:

- `aikata init my-app --preset minimal --no-interactive` produces a
  predictable file set matching the `minimal` golden test.
- Re-running in the same directory without `--force` writes nothing outside
  `.aikata-proposed/`.

### 4.2 `aikata add <component>`

**Purpose**: Add a single component to an existing aikata project.

**Must be able to** add:

- A top-level file (`add ui` → `UI.md`, `add api` → `API.md`).
- A new ADR with a title (`add adr "use go modules"`).
- A stack (`add stack flutter`).
- An AI-tool target (`add ai-tool cursor`).

### 4.3 `aikata doctor`

**Purpose**: Check consistency of the document set.

**Must check**:

- Every file `AGENTS.md` links to exists.
- Every term defined in `GLOSSARY.md` is used at least once elsewhere
  (warning only).
- No ADR is in the `Deprecated` status without a replacement reference.
- Every variable in `.env.example` is mentioned in `AGENTS.md` or
  `ARCHITECTURE.md`.
- Document `updated:` fields are not more than 365 days old (warning).
- Every markdown file has the required frontmatter keys.

**Output**:

- `error` (blocking) / `warning` / `info` levels.
- Exit codes: 0 (no errors), 3 (errors found).
- `--fix` auto-fixes safe issues (e.g. stale `updated:` after a known edit).

### 4.4 `aikata generate`

**Purpose**: Produce per-AI-tool configuration files from canonical sources.

**Must be able to**:

- Read `.ai/aikata.yaml`, determine enabled AI tools.
- For each enabled tool, emit its expected file(s).
- Add generated artifact paths to `.gitignore` of the **target project** by
  default (see Do-No-Harm policy for the exception in aikata itself).

### 4.5 `aikata update`

**Purpose**: Merge upstream template updates into the user's project.

**Must be able to**:

- Detect template version drift between bundled templates and the user's
  files.
- Present diffs interactively with accept / reject per hunk.
- Preserve user edits (no silent overwrite).

Inspired by Copier's update flow.

### 4.6 `aikata list`

**Purpose**: List available presets, components, and AI-tool integrations.

**Must output**:

- Built-in presets (`minimal`, `standard`, plus stack presets).
- Available `add` components.
- AI tools the current aikata binary supports.

---

## 5. Non-Functional Requirements

### 5.1 Reliability

- File generation is **atomic**: a failed `init` must leave the target
  directory unchanged.
- No `panic`s in user-facing flows. Errors are returned and rendered
  with actionable messages.

### 5.2 Performance

- `aikata init --preset standard` completes in **< 500 ms** on a 2025-era
  laptop.
- Binary size **< 20 MB**.

### 5.3 Portability

- Supports macOS (Intel / ARM), Linux (x86_64 / ARM64), Windows (x86_64).
- All file paths use OS-neutral separators.

### 5.4 Compatibility

- The bundled template schema is versioned. Templates produced by aikata
  v0.1 remain readable by aikata v1.x.
- `.ai/aikata.yaml` has a top-level `version: 1` field for future migrations.

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
