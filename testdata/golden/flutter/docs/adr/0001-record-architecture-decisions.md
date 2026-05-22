---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# ADR 0001 — Record Architecture Decisions

- **Status**: Accepted
- **Date**: 2026-05-21
- **Deciders**: samplekata maintainers

## Context

We want a low-friction way to capture architecturally relevant decisions
so that:

- Future contributors (human or LLM) understand **why** a choice was
  made, not just **what** the current state is.
- Decisions can be superseded without losing their history.

A heavy decision-log process would discourage adoption. Free-form notes
under `docs/` would discourage discovery.

## Decision

Adopt **Architecture Decision Records (ADR)** in a simplified
Nygard / MADR-inspired format:

- One ADR per architecturally relevant decision.
- Stored at `docs/adr/NNNN-kebab-title.md` with monotonically
  increasing `NNNN`.
- Required frontmatter keys: `project`, `status`, `version`, `updated`,
  `audience`.
- Required body sections, in order: **Context**, **Decision**,
  **Consequences**. Optional: **Alternatives Considered**.
- Allowed `Status` values: `Proposed`, `Accepted`, `Deprecated`,
  `Superseded by ADR-NNNN`.
- Once an ADR is `Accepted`, **the body is not edited**. Changes are
  made by a new ADR that supersedes it.

## Consequences

**Positive**:

- New decisions cost about 10 minutes to record.
- Grep-friendly: `Status: Accepted` lists current decisions.
- Superseded decisions remain auditable.

**Negative**:

- Numbering creates merge conflicts on concurrent PRs. Mitigation: the
  PR author rebumps the number on rebase.
- The folder grows monotonically. Accepted trade-off for full history.

## Alternatives Considered

- **Single `decisions.md` log** — encourages monolithic growth, makes
  per-decision metadata hard.
- **Wiki / external system** — violates the "documents live with the
  code" principle.
