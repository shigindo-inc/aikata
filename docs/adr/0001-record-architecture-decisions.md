---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR 0001 — Record Architecture Decisions

- **Status**: Accepted
- **Date**: 2026-05-20
- **Deciders**: aikata maintainers

## Context

aikata is a tool **about** documentation; it must hold itself to its own
standards. We need a low-friction way to capture architecturally relevant
decisions so that:

- Future contributors (human or LLM) understand **why** a choice was made,
  not just **what** the current state is.
- `aikata doctor` can mechanically check decision metadata (status, date).
- Decisions can be superseded without losing their history.

A heavy decision-log process would discourage adoption. Free-form notes
under `docs/` would discourage discovery.

## Decision

We adopt **Architecture Decision Records (ADR)** in a simplified
[Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
/ [MADR](https://adr.github.io/madr/)-inspired format:

- One ADR per architecturally relevant decision.
- Stored at `docs/adr/NNNN-kebab-title.md` with monotonically increasing
  `NNNN`.
- Required frontmatter keys: `project`, `status`, `version`, `updated`,
  `audience` — same as every other markdown file in the repo.
- Required body sections, in order: **Context**, **Decision**,
  **Consequences**. Optional: **Alternatives Considered**.
- Allowed `Status` values: `Proposed`, `Accepted`, `Deprecated`,
  `Superseded by ADR-NNNN`.
- Once an ADR is `Accepted`, **the body is not edited**. Changes are made
  by a new ADR that supersedes it.

## Consequences

**Positive**:

- New decisions cost ~10 minutes to record.
- `aikata doctor` can flag stale or orphaned ADRs.
- Grep-friendly: searching `Status: Accepted` lists current decisions.

**Negative**:

- Numbering creates merge conflicts on concurrent PRs. Mitigation: the
  PR author rebumps the number on rebase; `aikata add adr <title>` will
  auto-pick the next free number in v0.3.
- Superseded ADRs remain on disk forever, increasing surface area. We
  accept this in exchange for full history.

## Alternatives Considered

- **Single `decisions.md` log**: rejected — encourages monolithic
  growth, makes per-decision metadata hard.
- **Wiki / external system**: rejected — violates the "documents live
  with the code" principle.
- **Full MADR template (with Pros/Cons lists per option)**: rejected as
  too heavy for v0.x. May be reconsidered for v1.0+.
