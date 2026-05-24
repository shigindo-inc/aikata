---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0005 — Cursor and Codex Pass-Through Strategy

- **Status**: Accepted
- **Date**: 2026-05-22
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0003 (Do-No-Harm)

## Context

ADR 0002 made `AGENTS.md` the canonical source for agent instructions
and put `aikata generate` in charge of emitting per-tool artifacts.
v0.1 shipped a single Provider (`claude` → `CLAUDE.md`). v0.2 needs to
add the next two most-requested targets — **Cursor** and **OpenAI
Codex** — and that means deciding *what each one actually needs*
rather than reflexively writing a translator per tool.

The 2026 state of those two tools is:

- **Codex (CLI and cloud)** reads `AGENTS.md` natively. The spec walks
  from the repo root down to the current directory, plus an optional
  global `~/.codex/AGENTS.md`, plus an optional `AGENTS.override.md`
  at each level. Every discovered file is injected as a user-role
  message before the user prompt. Source: `developers.openai.com/codex/guides/agents-md`.
- **Cursor** also reads `AGENTS.md` natively (root + subdirectories).
  In addition Cursor supports `.cursor/rules/*.mdc` — markdown with
  YAML frontmatter (`description`, `globs`, `alwaysApply`) — which
  enables per-glob scoping and file-pattern-specific rules that
  `AGENTS.md` cannot express. Source: `cursor.com/docs/rules`.

The naive translation would be to author parallel rule sets for each
tool. That is exactly the duplication aikata exists to eliminate
(SPEC §1.3, ADR 0002 Context strategy B).

## Decision

We adopt **pass-through with a discoverability wrapper for Cursor and
a no-op for Codex** as the v0.2 strategy.

Concretely:

1. **Codex is a no-op provider.** `CodexProvider.Files()` returns an
   empty map. `aikata generate` reports the no-op to stderr —
   `[codex] no files generated (reads AGENTS.md directly)` — so users
   who enable `codex` in `.aikata/aikata.yaml` are not left wondering
   whether the command ran. The canonical `AGENTS.md` already
   satisfies Codex's discovery chain.
2. **Cursor emits one file: `.cursor/rules/main.mdc`.** The body is a
   thin wrapper with `description: "Read AGENTS.md as the canonical
   source for project rules"` and `alwaysApply: true`. It includes
   an adaptive read-order section (the same `has()` existence checks
   used by `ClaudeProvider`) but states explicitly that `AGENTS.md`
   is canonical. This keeps the rule discoverable from the
   `.cursor/rules/` directory listing without duplicating content.
3. **No glob-scoped MDC split in v0.2.** A richer treatment — emitting
   one `.mdc` per rule category with appropriate `globs` — is
   deferred and tracked as Q-DESIGN-08 in
   [`docs/decisions/open-questions.md`](../decisions/open-questions.md).
   It will be revisited once we have field experience with the
   one-file wrapper.

### Concrete consequences in code

- `internal/generate/codex.go` registers `codex` so users can list it
  in `ai_tools` and `aikata list` enumerates it as a recognized tool.
- `internal/generate/cursor.go` mirrors `ClaudeProvider`'s shape
  (same `has()` helper, same template-rendering pattern).
- `generate.Run` now returns `(map[string]int, error)` so the cli
  layer can surface per-provider file counts and emit the no-op
  notice. Existing callers updated; the registry interface
  (`Provider.Files`) is unchanged.
- aikata's own `.aikata/aikata.yaml` opts into `cursor` so the repo
  commits `.cursor/rules/main.mdc` for contributors who open the
  project in Cursor (consistent with ADR 0002 §"Operational status"
  for self-generated artifacts).

## Consequences

**Positive**:

- One canonical source (`AGENTS.md`) continues to win every conflict.
- Cursor users get an in-IDE pointer with `alwaysApply: true` without
  paying the maintenance cost of a parallel rule set.
- Codex users get explicit acknowledgement that the pass-through is
  intentional, not a missing implementation.
- Provider interface validated by a second registered provider
  (`cursor`) before stack presets and i18n land — the abstraction
  carries its second weight before its tenth.

**Negative**:

- Cursor users who want fine-grained `globs` / file-pattern rules see
  only the wrapper in v0.2. Mitigation: Q-DESIGN-08 captures the
  follow-up; the wrapper does not block such files from being added
  manually (`.cursor/rules/<name>.mdc` is just a directory entry).
- Codex users may find the no-op surprising despite the stderr
  notice. Mitigation: `aikata list` (planned v0.2+) will mark
  `codex` as `pass-through (reads AGENTS.md directly)`.

## Alternatives Considered

- **Full per-tool rule generation** — write `.cursor/rules/<name>.mdc`
  for every rule category, write a Codex-specific `AGENTS.md` variant.
  Rejected as exactly the duplication aikata exists to remove. The
  cost would compound across stacks and languages.
- **Skip Cursor entirely (`AGENTS.md` is enough)** — rejected. While
  technically functional, this would hide aikata's awareness of
  Cursor from the `.cursor/rules/` listing and force Cursor users to
  consult `AGENTS.md` outside Cursor's normal rules surface.
- **Generate `AGENTS.override.md` for Codex with project-specific
  overrides** — rejected for v0.2. No concrete use case yet; the
  override is a tool we should pick up only when a documented need
  exists.
