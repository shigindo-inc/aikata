---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR 0002 — `AGENTS.md` is the Canonical Source for Agent Instructions

- **Status**: Accepted
- **Date**: 2026-05-20
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy);
  [`docs/origin/initial-design.md`](../origin/initial-design.md) §9.1.1

## Context

aikata targets users who run multiple AI coding agents (Claude Code,
Cursor, Codex, Gemini CLI, Copilot, Windsurf, …). Each tool expects
instructions in a tool-specific file:

| Tool | Expected file |
|---|---|
| Claude Code | `CLAUDE.md` |
| Cursor | `.cursor/rules/*.mdc` |
| Codex / OpenAI | `AGENTS.md` (open spec) |
| Copilot | `.github/copilot-instructions.md` |
| Gemini CLI | `GEMINI.md` |
| Windsurf | `.windsurfrules` |

Hand-maintaining all of these in lock-step is the exact pain aikata
exists to solve. There are two reasonable strategies:

- **(A) Canonical `AGENTS.md`** — author one file, generate the rest.
  Bets on the [`agents.md` open spec](https://agents.md/) gaining
  traction. Codex already reads it natively.
- **(B) Tool-specific authoring** — author each tool's file directly.
  Plays to per-tool features (Claude's CLAUDE.md supports skills /
  hooks; Cursor's MDC supports glob-scoped rules) at the cost of
  duplication.

The pure-(B) path makes aikata's value proposition collapse into "a
folder of starter files," which is what existing scaffolders already do.

## Decision

We adopt **(A) `AGENTS.md` is canonical**.

Concretely:

- aikata's own repository hand-writes only `AGENTS.md`.
- `CLAUDE.md`, `.cursor/rules/`, `.github/copilot-instructions.md`,
  `GEMINI.md`, etc. are **generated** by `aikata generate` from
  `AGENTS.md` (with optional per-tool extension blocks under
  `templates/ai_tools/<tool>/`).
- Generated artifacts are **disposable**. The canonical source wins
  every conflict.

**Phase 1 deviation** (this milestone): `aikata generate` is not yet
implemented (it lands in v0.1 for Claude only). Until then:

- aikata's own repo ships **`AGENTS.md` only**. No `CLAUDE.md` exists.
- Claude Code reads `AGENTS.md` directly (it does, per the
  [`agents.md` spec](https://agents.md/)).
- When `aikata generate` is shipped (v0.1), the maintainers will
  re-evaluate whether to commit a generated `CLAUDE.md` in this repo;
  the default for **user projects** is documented in
  [ARCHITECTURE.md §6](../../ARCHITECTURE.md#6-distribution--generated-artifacts).

## Consequences

**Positive**:

- One file to keep current.
- aikata is aligned with an open spec rather than any one vendor.
- `aikata generate` becomes the natural place to host all
  tool-specific knowledge — the locus of "lossy generation is OK"
  (SPEC §3 principle 7).

**Negative**:

- Bets on the `agents.md` spec. If Claude / Cursor diverge sharply, we
  must add per-tool **extension blocks** rather than full overrides.
  This is anticipated and tracked in
  [`docs/decisions/open-questions.md`](../decisions/open-questions.md).
- Claude-specific features (skills, hooks, fast mode) cannot live in
  the canonical `AGENTS.md` without being meaningful to other tools.
  Mitigation: `templates/ai_tools/claude/extensions/` allows Claude-only
  appendices.

## Alternatives Considered

- **Per-tool hand authoring** — see Context strategy (B). Rejected: it
  is exactly the duplication aikata exists to remove.
- **Auto-symlink `CLAUDE.md → AGENTS.md`** — rejected. Symlinks are
  Windows-hostile and lose the ability to add Claude-only extensions.
- **Templating `CLAUDE.md` with `{{include AGENTS.md}}`** — equivalent
  to (A) but pushes templating into the file the user reads. Worse UX.
