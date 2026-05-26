---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-26
audience: [human, agent]
---

# ADR 0012 - Memory generate-projection deferral extended past v0.6.0

- **Status**: Accepted
- **Date**: 2026-05-26
- **Deciders**: aikata maintainers
- **Related**: ADR 0004 (long-term memory slot, option δ), ADR 0010
  (memory projection deferred to v0.6), ADR 0003 (do-no-harm policy)
- **Supersedes**: the "ship in v0.6 if the layouts have stabilized"
  guidance from ADR 0010 and ROADMAP v0.6

## Context

ADR 0010 deferred memory generate-projection (ADR 0004 option δ) from
v0.4 to v0.6 specifically so the v0.6 Claude Code plugin work could
own the projection contract. The v0.6 release cycle is now in
progress, and the bundled scope decision has come due.

Two questions had to be answered:

1. Have the per-AI-tool memory channel layouts stabilized enough that
   aikata can write into them safely?
2. Is there documented dogfooding evidence that the lack of
   projection has caused drift between `docs/memory/` and the tool's
   native channel?

The answer to both is **no**:

- Claude Code's `~/.claude/` layout has continued to evolve through
  the v0.4 → v0.6 window (skills, plugins, memory slot; per-project
  vs per-user storage). Anchoring an `aikata generate` projection on
  any of those paths today is high-likelihood of becoming legacy
  before the v1.0 stable surface promise.
- Cursor's rule format is in the middle of a similar churn cycle
  (`alwaysApply` semantics, file location, MDC syntax).
- No user has reported drift between `docs/memory/` and their tool's
  memory channel during the v0.2 → v0.6 dogfooding window. The
  hypothesized pain point from ADR 0004 has not materialized.

## Decision

**Memory generate-projection does not ship in v0.6.0.**

The decision is a strict continuation of ADR 0010's deferral; aikata
keeps option δ open but does not commit to a v0.6.x or v0.7 timeline
either. The next concrete review point is:

- **Trigger**: a documented dogfooding case where a user reports that
  the per-tool channel for memory has diverged from `docs/memory/`
  and they want aikata to keep the two in sync.
- **Scope of that future ADR**: pick exactly one tool (Claude),
  define the projection contract on the then-stable layout, and ship
  it as a follow-up ADR that supersedes this one.

## Rationale

1. **Spec instability cost > projection benefit (re-confirmed).** The
   three reasons ADR 0010 deferred (per-tool churn, side-effect
   surface, YAGNI) are all stronger in v0.6 than they were in v0.4
   — the Claude plugin surface itself just landed, and Cursor's
   rule format absorbed another revision in the interim.

2. **The v0.6 Claude Code plugin already covers the discoverability
   gap.** The plugin's skill text explicitly tells Claude to read
   `docs/memory/*.md` when the conversation needs long-term
   context. The information is *projected at session time* by
   Claude itself, not at `aikata generate` time by aikata. This is
   strictly safer (no on-disk write to a tool-owned directory; no
   stale projection if upstream renames the path).

3. **ADR 0003 (Do-No-Harm).** A projection that writes into
   `~/.claude/...` or `.claude/...` competes with the user's own
   tool configuration. Skipping the projection keeps aikata's write
   set inside the canonical document tree.

## Consequences

- `aikata generate` continues to write **only** the per-tool wrapper
  set documented in ARCHITECTURE.md §3 (`CLAUDE.md`,
  `.cursor/rules/main.mdc`, etc.). It does not write to
  `.claude/memory/`, `.claude/skills/`, `.cursor/rules/long-term/`,
  or any other tool-owned subtree.
- `docs/memory/` remains the canonical long-term memory slot per
  ADR 0004. Discovery happens via the AGENTS.md read order plus
  the Claude Code plugin's skill text.
- ROADMAP v0.6 row updated to mark memory projection as **deferred
  again** (status: not shipped), with the trigger condition above
  recorded so a future ADR knows what evidence to look for.
- This ADR does not retract ADR 0010. ADR 0010's "if at all, ship
  in v0.6" guidance is superseded by this ADR's "defer indefinitely
  with a concrete revisit trigger".

## Alternatives considered

- **Ship Claude-only minimal projection in v0.6.0.** Rejected:
  upstream layout still moving (see Rationale §1); a v0.6.0
  projection would become deprecated within one or two Claude Code
  releases without an offsetting user-pain signal. Cost > benefit.
- **Drop option δ permanently.** Rejected for the same reason ADR
  0010 rejected it: the dogfooding case stays plausible. This ADR
  defers without abandoning.
- **Wait until v1.0 (a hard ban for v0.x).** Rejected because the
  cycle in which to ship projection is dictated by user evidence,
  not by version-number prestige; a v0.6.x release is fine if a
  dogfooder reports the gap.

## References

- ADR 0004 — Long-term memory slot, option δ enumeration
- ADR 0010 — Memory generate-projection deferred to v0.6 (this
  ADR's predecessor)
- ROADMAP.md §v0.6 — entry updated alongside this ADR to record the
  continued deferral
- v0.6 Claude Code plugin (`dist/claude-code/plugin/`) — the
  session-time alternative to projection
