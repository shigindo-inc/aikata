---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# Glossary

Terminology used across aikata documentation and source code. Each entry
contains the English term, an optional Japanese reading (`yomi`), and a short
definition. Terms marked **(domain)** are aikata-specific concepts; others are
industry-standard but redefined here to fix the project's interpretation.

> **Why this file matters**: aikata targets bilingual (Japanese / English)
> projects. Pinning terminology in one place reduces translation drift in LLM
> output and helps `aikata doctor` flag mismatches.

---

## A

### ADR — Architecture Decision Record

A short markdown document that captures a single architectural decision,
its context, and its consequences. Stored under `docs/adr/` with the file
name pattern `NNNN-title.md`. Format follows
[`0001-record-architecture-decisions.md`](./docs/adr/0001-record-architecture-decisions.md).

### agent

An LLM-driven coding assistant (Claude Code, Cursor, Codex, Gemini CLI,
Copilot, Windsurf, …) that reads project documentation and produces or edits
code. aikata treats agents and humans as **first-class co-readers** of every
document.

### `AGENTS.md`

The single, hand-written, human-and-agent-readable instruction document at
the project root. In aikata it is the **canonical source** for agent
behavior. See
[ADR 0002](./docs/adr/0002-agents-md-as-canonical.md) for why it is canonical
and how tool-specific files (e.g. `CLAUDE.md`) relate to it.

### `.ai/`

Legacy v0.1-v0.3.1 directory at the project root used to hold aikata's
config file (`.ai/aikata.yaml`). ADR 0008 moved current writes to the
aikata-owned `.aikata/` namespace because `.ai/` is too generic for
durable project state. ADR 0020 removes the read fallback in v0.7.4;
the term remains here as historical context for old repositories and
release notes.

### `.aikata/`

Current v0.3.2+ directory at the project root for **aikata-owned durable
configuration**, primarily `.aikata/aikata.yaml`. It replaces `.ai/` as the
preferred config namespace while generated tool-facing artifacts continue to
live in their native locations (`CLAUDE.md`, `.cursor/rules/`, etc.). See
[ADR 0008](./docs/adr/0008-aikata-owned-config-directory.md).

---

## C

### canonical source

The **single source of truth** for a piece of information. aikata enforces
canonical sources to keep lossy generation safe: if a generated artifact
diverges from its canonical source, the canonical source wins. Example:
`AGENTS.md` is canonical for agent instructions; `CLAUDE.md` (when present)
is generated and disposable.

### `CLAUDE.md`

A Claude Code-specific instruction file. In aikata, it is **not hand-written**
and **not present in Phase 1**. It will be produced later by
`aikata generate` from `AGENTS.md` plus optional Claude-only extensions.

### Conventional Commits

Commit message convention (`<type>(<scope>): <subject>`). aikata mandates it
for all commits; see [AGENTS.md](./AGENTS.md) for the allowed `type` values
and the **no-AI-signature** rule.

---

## D

### dogfooding (ドッグフーディング)

Using one's own product internally. aikata is dogfooded: the aikata repo
itself is structured the way `aikata init --preset standard --oss` would
produce. The migration of aikata's own layout to a fully generated form is
tracked in [ROADMAP.md](./ROADMAP.md).

---

## F

### frontmatter

The YAML block at the top of a markdown file delimited by `---`. aikata uses
frontmatter for cross-document metadata: `project`, `status`, `version`,
`updated`, `audience`. Designed to be readable as plain text **and** to work
as Obsidian Properties without harming non-Obsidian users (ADR 0003).

---

## G

### generate (verb) — `aikata generate`

The command that produces AI-tool-facing artifacts (`CLAUDE.md`,
`.cursor/rules/*.mdc`, `.github/copilot-instructions.md`, …) from the
canonical documents. The output of `generate` is **lossy** and disposable.

### golden test

A test that compares produced output against a checked-in expected output
under `testdata/golden/`. Used to validate `aikata init` and `aikata
generate`.

---

## H

### Human-LLM dual readable

Design principle: every document must be useful to a human and an LLM
without rewriting. Implications include: short paragraphs, explicit
navigation, no diagrams-only documents, no cleverness that requires a UI
viewer.

---

## I

### init (verb) — `aikata init`

The command that scaffolds a new project. The MVP target is `--preset
minimal` and `--preset standard`.

---

## L

### long-term memory — `docs/memory/`

The third class of agent-facing artifact, complementary to **rules**
(`AGENTS.md`) and **working memory** (`docs/tasks/current.md`).
Captures mutable facts that should survive across sessions: user
profile and preferences, validated continuing instructions, ongoing
project context, and external references. Subdivided into four
**memory types** — `user`, `feedback`, `project`, `reference` — one
file per type under `docs/memory/`. Opt-in via `--with-memory`
(planned v0.2). Defined by
[ADR 0004](./docs/adr/0004-long-term-memory-slot.md).

### lossy generation

The acceptance that **derived files may lose information** vs. their
canonical source. This is acceptable as long as the canonical source remains
authoritative and regeneration is cheap (Design Principle 7).

---

## M

### memory type — `user` / `feedback` / `project` / `reference`

The four canonical buckets inside [long-term memory](#long-term-memory--docsmemory):

- `user` — profile, role, knowledge, preferences of the user.
- `feedback` — continuing instructions (corrections + validated approaches).
- `project` — ongoing project context not derivable from code or git log.
- `reference` — pointers to external systems.

These names are deliberately aligned with the Claude Code auto-memory
taxonomy and the superpowers plugin so agents can transfer convention.

### memory slot

A single file (`docs/memory/<type>.md`) corresponding to one memory
type. Always one file per type; subdirectory layouts are deferred to
v1.x.

### MADR — Markdown Architecture Decision Record

A lightweight ADR format ([adr.github.io](https://adr.github.io/madr/)).
aikata's ADR template is inspired by MADR but simplified;
see [ADR 0001](./docs/adr/0001-record-architecture-decisions.md).

---

## O

### owned (sync status) — (所有)

The `aikata sync` status for a path listed under `sync.own` in
`.aikata/aikata.yaml` (a glob list using the same matcher as
`doctor.exclude`). An `owned` file is one the user has intentionally
taken over: `aikata sync` never rendered-compares, conflict-markers,
overwrites, or manifest-tracks it. Introduced in v0.8.3
([ADR 0025](./docs/adr/0025-sync-divergent-file-preservation.md) D2) as
the declarative replacement for a manual `git restore` workaround.

### opinionated, but small

aikata's positioning: it makes a few strong choices (file names, frontmatter
schema, canonical source rules) but does **not** lock the user into a
particular tool chain. Compared to Vite vs. ai-rulez's Terraform-like
breadth, aikata is closer to Vite/Astro.

---

## P

### preset — (プリセット)

**Deprecated alias** (removed in v1.0) for the orthogonal `scope` and
`stack` axes — see
[ADR 0024](./docs/adr/0024-scope-stack-axes-split.md). `aikata init
--preset` still works and prints a deprecation notice: `minimal` /
`standard` map to `--scope`, and `flutter` / `typescript` map to
`--scope standard --stack <name>`. New work should use `--scope` and
`--stack` directly. Internally, "preset" survives only as the name of a
template tree under `internal/templates/data/presets/<tree>/`.

---

## S

### scaffold (verb / noun) — 雛形 (hinagata)

To generate the initial structure (directories + template files) of a
project. The noun form refers to the generated structure itself.

### scope — (スコープ)

The documentation-breadth axis of `aikata init`, selected with
`--scope`: `minimal` (AGENTS.md + the smallest README/SPEC/ARCH/GLOSSARY
scaffold) or `standard` (the full canonical document set). `extended` is
reserved for the v1.0 operational-readiness pack. Orthogonal to `stack`
and `--lang` (ADR 0024). As of v0.8.2 only the combinations that have a
template tree are buildable — `minimal`, `standard`, `standard + flutter`,
`standard + typescript`; other pairings error explicitly until the
template refactor that unlocks them.

### stack — (スタック)

The target-technology axis of `aikata init`, selected with `--stack`
(`flutter` | `typescript`; repeatable / comma-separated in syntax,
empty = stack-agnostic). A stack adds a `docs/stacks/<name>.md` brief and
records itself in `aikata.yaml`'s `stacks:` list. Orthogonal to `scope`
and `--lang` (ADR 0024). v0.8.2 accepts a single stack paired with
`--scope standard`.

### stack-agnostic core

Design principle: the aikata CLI core knows nothing about specific
technology stacks. Stack knowledge lives entirely in templates under
`templates/presets/<tree>/`. Adding a stack must not require modifying
core code (this guides the v1.x plugin design).

---

## T

### top-level minimalism

Design rule: at most **8 non-hidden files** at the project root after
`aikata init --scope standard`. Dot-files (`.gitignore`, `.env.example`,
`.ai/`, `.aikata/`) do not count. Enforced by `aikata doctor`.

---

## Y

### yomi (読み)

The Japanese phonetic reading of a kanji term. Recorded in this glossary
for terms that have a non-obvious reading, so LLMs and non-Japanese
contributors can pronounce them correctly. Example: `相方 (aikata)`,
`雛形 (hinagata)`.
