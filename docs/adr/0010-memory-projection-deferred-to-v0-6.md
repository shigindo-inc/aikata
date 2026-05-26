---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0010 - Memory generate-projection deferred to v0.6

- **Status**: Accepted
- **Date**: 2026-05-24
- **Deciders**: aikata maintainers
- **Related**: ADR 0004 (Long-term memory slot), option δ;
  ADR 0008 (`.aikata/` config namespace)

## Context

ADR 0004 introduced `docs/memory/` as the canonical long-term memory
slot for AI agents (option γ, shipped in v0.2). It explicitly left
**option δ** open: projecting that memory into each AI tool's
native channel so the tool sees the content without a custom
read order. The concrete projections under discussion are:

- Claude Code — copy / symlink relevant memory files to
  `.claude/memory/` (or whatever the tool's stable storage path is
  at the time projection ships).
- Cursor — emit a subset of memory entries under
  `.cursor/rules/long-term/` with `alwaysApply: true`.
- Codex, Gemini, Copilot, Windsurf — to be assessed individually.

The original ROADMAP entry for v0.4 read "Investigate the memory
projection cost; ship only if the cost is low." This ADR records
that investigation and the decision.

## Decision

**Do not ship memory generate-projection in v0.4.**
**Ship the projection (if at all) as part of the v0.6 packaging
& distribution cycle**, with the Claude Code plugin / multi-tool
plugin family as the natural delivery vehicle.

The v0.4 components surface (this release) covers `aikata add memory`
and `aikata init --with-memory`. Memory authoring is therefore
already first-class — agents can read `docs/memory/` via their
canonical AGENTS.md read order without per-tool synchronization.

## Rationale

Three observations pushed this out of v0.4 scope:

1. **Per-tool spec instability.** Claude Code's `.claude/` layout
   shifted twice in the last two minor releases (skills, plugins,
   memory slot). Cursor's rules format has comparable churn. A
   projection shipped in v0.4 risks emitting deprecated paths within
   one or two upstream releases, before the v0.6 plugin work that
   would naturally consume the same spec.

2. **Side-effect surface area.** Projection would mean `aikata
   generate` writes outside the canonical document set into
   tool-owned directories. That collides with two pre-existing
   contracts: ADR 0003 (Do-No-Harm — generated artifacts must not
   compete with the user's own tool configuration) and the v0.5
   `aikata sync` interactive diff-merge (the only path that
   touches tool-owned files without explicit prompts must be a
   user-initiated sync).

3. **YAGNI.** No user has reported drift between `docs/memory/` and
   their tool's memory channel in the v0.2 / v0.3 dogfooding window.
   Until that drift becomes a reported pain point, the marginal
   value of automatic projection does not justify the per-tool
   maintenance commitment.

## Consequences

- v0.4 ships `aikata add memory` (this release) and is **the
  feature-complete authoring surface** for memory until v0.6.
- The v0.6 Claude Code plugin work owns the first concrete
  projection. If it materializes, this ADR is superseded by a
  follow-up ADR that records the chosen projection contract.
- `aikata generate` continues to emit *only* the per-tool wrapper
  artifact set documented in ARCHITECTURE.md §3 (CLAUDE.md,
  `.cursor/rules/main.mdc`, etc.). It does not write to
  `.claude/memory/` or `.cursor/rules/long-term/` in v0.x.

## Alternatives considered

- **Ship Claude-only projection in v0.4** — rejected: the upstream
  `.claude/` layout is still moving and we would re-implement
  inside v0.5 / v0.6 anyway. Cost > benefit.
- **Ship all four target tools in v0.4** — rejected: doubles the
  v0.4 surface and pulls in plugin-shape decisions that belong in
  the v0.6 cycle.
- **Drop option δ permanently** — rejected: the dogfooding case is
  plausible enough to keep the option open. This ADR defers, not
  abandons.

## References

- ADR 0004 — `Long-term memory slot under docs/memory/`, option δ
  enumeration
- ROADMAP.md §v0.4 — original "Investigate" wording superseded by
  this ADR
- v0.6 plugin spec — TBD in the v0.6 cycle
