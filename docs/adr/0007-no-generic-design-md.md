---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0007 - Do Not Generate a Generic `DESIGN.md`

- **Status**: Accepted
- **Date**: 2026-05-24
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy)

## Context

aikata's generated document set already separates product intent,
technical structure, decision history, and optional domain-specific
guidance:

- `SPEC.md` records what the project does and why.
- `ARCHITECTURE.md` records how the project is built.
- `docs/adr/*.md` records durable decisions and their rationale.
- `docs/stacks/<stack>.md` records stack-specific conventions.
- Optional files such as `UI.md` and `API.md` are addable when the
  project explicitly needs that surface.

The name `DESIGN.md` is attractive but ambiguous. It can mean product
design, UI design, technical design, system design, or decision rationale.
Adding it by default would overlap with the existing document taxonomy and
increase top-level surface area.

## Decision

aikata will not generate a generic top-level `DESIGN.md` in built-in
presets.

Design-related content belongs in the more specific document that owns the
subject:

- Product behavior, user outcomes, and requirements go in `SPEC.md`.
- Technical structure and implementation approach go in `ARCHITECTURE.md`.
- Durable trade-offs and accepted decisions go in `docs/adr/*.md`.
- Stack-specific conventions go in `docs/stacks/<stack>.md`.
- UI, UX, interaction, and product-design guidelines go in optional
  `UI.md` when the project opts into that component.
- API interface guidelines go in optional `API.md` when the project opts
  into that component.

If a future preset appears to need a `DESIGN.md`, it must first prove why
the content cannot be owned by one of the existing documents above and must
record that exception in a new ADR.

## Consequences

**Positive**:

- Keeps the generated root small and aligned with top-level minimalism.
- Avoids a vague catch-all document that would drift into duplicated
  `SPEC.md`, `ARCHITECTURE.md`, or ADR content.
- Gives future `aikata add ui` work a clear home for UI / UX guidance.

**Negative**:

- Users familiar with `DESIGN.md` need to learn aikata's split between
  product requirements, architecture, and decision records.
- Some projects may still want a generic design document. They can add one
  manually, but built-in presets should not normalize it.

