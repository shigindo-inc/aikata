---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-12
audience: [human, agent]
---

# ADR 0046 - Structure-migration assistant: observe → propose → confirm-move boundary

- **Status**: Accepted
- **Date**: 2026-06-12
- **Deciders**: aikata maintainers
- **Related**: ADR 0044 (doc map — the descriptive inventory this assistant
  reconciles), ADR 0045 (documentation value model), ADR 0043 (command-wrapper
  skill surface & simple skill names — the skill-not-verb precedent, D5
  deprecated-doc cleanup), ADR 0037 (tighten adoption mutation boundaries),
  ADR 0019 (sync missing-file repair semantics — non-destructive moves),
  ADR 0002 (`AGENTS.md` is canonical), `docs/layout.md` (the prescriptive
  target layout).

## Context

`docs/layout.md` is the **prescriptive** index of where each kind of document
belongs in an aikata project. The doc map (ADR 0044) is the **descriptive**
record of what actually exists, flagging off-structure documents as
`managed: false` (external). The third corner — **reconcile**: bring an
existing repository *into* the recommended layout — has been unowned.

A repository adopted with `aikata fill` (ADR 0042) keeps every pre-existing
document exactly where it was; `fill` only adds missing canonical documents
and never moves anything. So an adopted repo commonly carries documents that
*should* live under `docs/adr/`, `docs/memory/`, `docs/stacks/`, etc. but sit
elsewhere. Relocating them is mechanical to *propose* (the doc map already
computes the external set; `docs/layout.md` already defines destinations) but
**moving user files is a mutation with real blast radius** — it rewrites
history pointers, breaks relative links, and can clobber a file at the
destination.

The question this ADR settles: **what is allowed to move user files, and
under what contract?**

## Decision

### D1 — It ships as a skill, not a CLI verb

The reconcile capability ships as the first-party **`migrate-structure`**
skill, not as an `aikata migrate` mutate verb. This follows the ADR 0043 D5
precedent (skill-guided deprecated-doc cleanup is a skill because it needs
judgement, not a verb). Mapping an arbitrary off-structure document to its
"right" home, and deciding whether a move is safe, needs judgement the thin
stack-agnostic CLI deliberately does not encode.

### D2 — The aikata CLI stays observation-only

The CLI **never moves user files**. Its role in this workflow is limited to
*reporting*: `aikata map` produces the `managed: false` set the skill reads,
and `aikata map` rebuilds the map after moves. There is no `aikata migrate`,
no `--move`, no `aikata mv`. This preserves the "no `aikata push`, no
silent restructuring" boundary (ROADMAP "Out-of-scope, indefinitely") and
keeps the core stack-agnostic and side-effect-light.

### D3 — observe → propose → confirm-move → rebuild

The skill follows a fixed contract:

1. **Observe** — read the doc map's external (`managed: false`) set; never
   guess from a fresh filesystem walk.
2. **Propose** — map each external document to its recommended destination
   per `docs/layout.md`; present a **dry-run plan** (every `from → to`)
   first.
3. **Confirm** — apply moves only after explicit user confirmation
   (per-move or an explicit batch approval). Default to dry-run; do nothing
   on silence.
4. **Move** — relocate with `git mv` to preserve history; never overwrite an
   existing destination (refuse and surface the collision); never delete.
5. **Rebuild** — run `aikata map` so the doc map reflects the new reality.

### D4 — Relocation only; never content rewrite

The assistant **moves** documents; it does not edit their contents. Fixing
relative links that a move breaks, or rewriting prose, is out of scope here
(it belongs to the author or other skills). The skill surfaces newly-broken
links (via `aikata doctor`) for follow-up rather than silently rewriting
them.

### D5 — Conservative by default

No automatic or unattended reorganization. A document whose correct
destination is ambiguous is **left in place and reported**, not moved on a
guess. The skill prefers under-moving (surfacing a question) to over-moving.

## Consequences

- The reconcile corner of the prescriptive / descriptive / reconcile triad is
  now owned, without adding a mutate verb or weakening the observation-only
  CLI boundary.
- Distribution mirrors the existing first-party skills (canonical
  `dist/universal-skill/migrate-structure`, byte-identical copies under the
  Claude Code and Codex trees, a thin `/aikata:migrate-structure` plugin
  command). Version lockstep applies.
- The assistant depends on the doc map (v0.13.0) shipping first; it consumes
  the `managed` flag and the `external` set.
- Because moves are `git mv` and gated on confirmation, the blast radius is
  reviewable in a normal diff and reversible before commit.

## Alternatives considered

- **`aikata migrate` mutate verb.** Rejected (D1/D2): puts file-moving
  judgement and write access into the thin CLI, contradicting the
  observation-only boundary and the skill-not-verb precedent (ADR 0043 D5).
- **Auto-relocate during `fill` / `sync`.** Rejected: `fill` and `sync` are
  non-destructive by contract (ADR 0019, ADR 0037); silently moving user
  files inside them would violate the adoption mutation boundary.
- **Content-aware link rewriting in the same pass.** Deferred (D4): mixing
  relocation with content edits enlarges the blast radius; doctor already
  surfaces broken links for a separate, reviewable follow-up.
