---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ROADMAP

> Direction, not promises. Versions are scoped by *what they unlock*, not by
> dates. The single ordering rule: **each milestone must leave aikata
> usable by its current user base** — never break Phase N users to ship
> Phase N+1.

For the why behind each item, see [SPEC.md](./SPEC.md). For the how, see
[ARCHITECTURE.md](./ARCHITECTURE.md). For open design questions affecting
sequencing, see [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md).

---

## Phase 1 — Documentation bootstrap (current)

**Status**: in progress.

**Goal**: aikata's repository is itself a coherent set of operational
markdown documents so that any AI agent and any new contributor can pick up
the project from `README.md` alone.

- [x] Split into operational documents: README, AGENTS, SPEC,
      ARCHITECTURE, ROADMAP, GLOSSARY, CHANGELOG, LICENSE, ADRs,
      open-questions. The original planning notes were collapsed into
      these and are preserved in git history.

**Exit criteria**:

- Every link in `AGENTS.md` resolves.
- `GLOSSARY.md` terms are used elsewhere.
- Top-level non-hidden file count ≤ 8.
- `.gitignore` follows the aikata-own-repo rule (commit generated
  artifacts).

---

## v0.1 — MVP

**Goal**: A working `aikata init` for two presets with Claude Code as the
only AI-tool target. Validates the core "scaffold from canonical
templates" loop end-to-end.

- [ ] Go project initialized (`go.mod`, `cmd/aikata`, `internal/`).
- [ ] cobra wired; `aikata --version`, `aikata --help` work.
- [ ] `embed.FS` template loader.
- [ ] `aikata init --preset minimal` produces README, AGENTS, SPEC.
- [ ] `aikata init --preset standard` produces the §3.1 default tree from
      [ARCHITECTURE.md](./ARCHITECTURE.md).
- [ ] `aikata init` interactive mode (default) via `huh`.
- [ ] Atomic file generation; `.aikata-proposed/` fallback for non-empty
      target dirs.
- [ ] `aikata generate` minimal Claude-Code output (`CLAUDE.md` from
      `AGENTS.md` + `SPEC.md` + `ARCHITECTURE.md` + `GLOSSARY.md`).
- [ ] Golden tests for both presets.
- [ ] CI green on macOS / Linux / Windows.

**Out of scope**: Cursor / Codex / Gemini / Copilot / Windsurf,
non-English templates, `aikata doctor`, `aikata add`, `aikata update`,
monorepo.

---

## v0.2 — Stack & Language

**Goal**: aikata becomes useful to a Japanese-speaking Flutter developer.

- [ ] `--preset flutter` (templates under `templates/presets/flutter/`,
      `docs/stacks/flutter.md`).
- [ ] `--lang ja` template set.
- [ ] `--with-memory` flag — provisions
      `docs/memory/{user,feedback,project,reference}.md` for opt-in
      long-term agent memory; see
      [ADR 0004](./docs/adr/0004-long-term-memory-slot.md).
- [ ] `aikata generate` targets for Cursor (`.cursor/rules/*.mdc`) and
      Codex (`AGENTS.md` is already the canonical; verify pass-through).
- [ ] First implementation of `aikata doctor` covering: frontmatter
      keys, broken links from `AGENTS.md`, ADR statuses, memory
      `memory_type` matches filename.

---

## v0.3 — Authoring ergonomics

**Goal**: Editing an aikata project is as ergonomic as creating one.

- [ ] `aikata add <component>` for `ui`, `api`, `tdd`, `changelog`, `adr`,
      `stack`, `ai-tool`, `memory`.
- [ ] ADR auto-numbering and template insertion.
- [ ] `--with-ui`, `--with-api`, `--with-tdd`, `--with-changelog` flags on
      `aikata init`.
- [ ] `--json` machine-readable output for `doctor`.
- [ ] Investigate memory generate-projection (ADR-0004 option δ): how
      to mirror `docs/memory/*` into tool-specific channels (Claude
      `.claude/memory/`, Cursor `.cursor/rules/long-term/`). Record
      findings in a new ADR; ship only if the cost is low.

---

## v0.4 — Operability

**Goal**: aikata projects can grow into monorepos and stay current.

- [ ] `--monorepo` initialization, nested `AGENTS.md` per app.
- [ ] `aikata update` interactive diff-merge.
- [ ] Migration framework for `.ai/aikata.yaml` schema versions.
- [ ] npm wrapper for `npx aikata` distribution.
- [ ] If v0.3 investigation justified it: ship
      `aikata generate --memory` for at least one AI-tool memory
      channel (ADR-0004 option δ).

---

## v1.0 — Stable surface

**Goal**: A surface that downstream tooling can depend on.

- [ ] Major AI tools all supported: Claude, Cursor, Codex, Gemini,
      Copilot, Windsurf.
- [ ] `--oss` adds `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`.
- [ ] Stable preset & template schema (semver guarantee).
- [ ] Official docs site (`aikata.dev`).
- [ ] External preset repositories (`aikata add stack github.com/foo/bar`).

---

## v1.x — Beyond bootstrap

Speculative. Order and inclusion depend on validating
[hypotheses H1–H4](./SPEC.md#7-hypotheses-to-validate).

- LLM-API-assisted document drafting (`aikata draft <topic>`).
- VS Code / JetBrains extensions.
- Reverse-analysis of existing projects to suggest an aikata layout
  (agentsmesh-like).
- Bilingual document mode (Japanese for humans, English for LLMs in a
  single canonical file).

---

## Dogfooding milestone

A standing goal across phases:

- The aikata repository itself should eventually be reproducible by
  `aikata init --preset standard --oss --ai-tools claude,cursor` plus a
  manual `git diff` review.
- This becomes a release gate from v0.3 onward.

---

## Out-of-scope, indefinitely

These are documented here so future scope-creep proposals can be
deflected.

- IDE GUIs for editing aikata config.
- Real-time rule enforcement (linting source files for style violations).
- Task / issue tracker integration.
- Direct write access to remote git providers (no `aikata push`).
