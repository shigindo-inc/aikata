---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0028 — Prioritize core-concept stabilization before ecosystem expansion

- **Status**: Accepted
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0015 (first-party skill
  and plugin distribution), ADR 0021 (`doctor` scope and exclusions),
  ADR 0024 (`scope` / `stack` axes), ADR 0026 (workflow guide opt-in),
  [`v0.9 stabilization design note`](../decisions/v0.9-core-concept-stabilization.md)

## Context

aikata has grown from a document scaffold into a small maintenance loop:
canonical project documents, per-tool artifact generation, consistency
checks, non-destructive sync, and opt-in context slots.

The maintainer raised a product-boundary concern after reviewing future
stack-axis extensions: it is valid to describe an extensible long-term
shape, but adding axes, stack combinations, registries, wrappers, and
optional categories too quickly risks producing a tool whose value for
AI collaboration is hard to explain. A documentation-maintenance tool
must not become another configuration system that humans have to
maintain.

Dogfooding also surfaced live documentation drift in aikata itself:
released behavior is not consistently reflected across README, SPEC,
ROADMAP, adoption docs, and dogfood config. This is direct evidence that
the next improvement should be convergence, not additional breadth.

## Decision

v0.9.0 prioritizes **core-concept stabilization** before ecosystem
expansion. The existing v0.9.9 channel-publication line remains
separate: it improves installation and discoverability for the current
product without widening its conceptual surface.

The core concept is:

> aikata reduces the human cost of maintaining project context while
> giving humans and AI coding agents a shared, coherent document
> structure and a common source of truth.

The essential product loop remains:

```text
init -> canonical documents -> generate -> doctor -> sync
```

### D1 — Stabilize the existing surface first

Before adding new capability categories, built-in stacks, composition
rules, or plugin ecosystems, v0.9.0 audits and simplifies the shipped
surface:

- converge stale live documentation;
- audit default standard-scope files for distinct value;
- investigate narrowing `doctor` to the surface aikata manages;
- shorten stack guidance toward collaboration-critical guardrails;
- reassess v1.0 requirements against the core concept.

Detailed implementation choices land in focused follow-up ADRs. The
design note linked above records the work packages and current leading
positions.

### D2 — Keep justified internal complexity behind a small user model

`doctor`, `sync`, and manifests stay part of the product. Their internal
complexity is justified where it detects context drift or prevents loss
of user edits.

Normal users should not need to understand manifest internals, merge
classification, schema migration, or exclusion mechanics to complete
the default flow. New controls are added only after evidence shows that
a simpler default cannot solve the problem.

### D3 — Stack guidance stays intentionally small

Built-in stack support provides a short additive brief:

- common AI failure modes;
- minimum validation commands;
- project-choice TODOs;
- links to ADRs for decisions that vary by project.

It does not become a comprehensive best-practice pack or a framework
generator. A scope-base / stack-partial refactor is permitted when it
reduces duplication and drift, but not as a pretext for stack
proliferation.

### D4 — Ecosystem expansion is demand-driven

The following are removed from the critical path unless concrete
dogfooding or user evidence justifies them:

- external preset or stack repositories;
- third-party skill / plugin catalog management;
- new workflow-guide domains;
- broad native wrapper proliferation;
- multi-stack composition.

Distribution work that makes the existing CLI easier to install and
discover remains valid v0.9.9 work.

## Consequences

### Positive

- The product story becomes easier to explain and evaluate.
- Dogfooding pressure is applied to the exact value aikata claims to
  provide: coherent, maintainable shared context.
- New features face an explicit evidence test instead of accumulating
  because the architecture can support them.
- Existing `doctor` and `sync` investment continues to protect users
  while their complexity is kept out of the default mental model.

### Negative

- Some planned ecosystem work moves later even if its long-term shape is
  already plausible.
- Simplifying generated templates can create sync-visible upstream
  changes for existing projects and requires careful golden tests.
- Narrowing `doctor` needs a follow-up ADR because adoption and
  pre-manifest projects still need a coherent validation story.

## Alternatives Considered

- **Continue the existing feature roadmap unchanged.** Rejected:
  architectural extensibility is not sufficient evidence of user value,
  and dogfooding drift shows that convergence work is already due.
- **Remove `doctor` or `sync` to make the CLI smaller.** Rejected: both
  directly reduce manual maintenance and protect the canonical document
  model. Their user-facing defaults should become simpler, not their
  value discarded.
- **Freeze all additions until v1.0.** Rejected: installation and
  discoverability improvements can make the existing product easier to
  adopt without widening its conceptual surface.
