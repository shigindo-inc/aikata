---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-12
audience: [human, agent]
---

# ADR 0045 - Documentation value model: classify by source, not decay

- **Status**: Accepted
- **Date**: 2026-06-12
- **Deciders**: aikata maintainers
- **Related**: ADR 0039 (documentation hygiene & context budget — the
  *operational mechanism* this ADR sits underneath), ADR 0044 (doc map as
  a derived artifact), ADR 0037 (doctor managed-surface scope), ADR 0004
  (long-term memory slot — append-only / supersede-in-place), ADR 0028
  (prioritize core-concept stabilization — demand-driven, not
  architecture-driven), ADR 0002 (`AGENTS.md` is canonical), ADR 0001
  (ADR bodies are immutable).

## Context

LLMs compact their own context within a long session: stale or
superseded turns are summarized, dropped, or rewritten so the live window
stays small and high-signal. A natural question for a documentation tool:
should aikata do the *same* to the documents it manages, so that as a
repo's docs grow, an agent's per-session orientation cost (the
first-read context, ADR 0039) does not grow with them?

A brainstorm explored a concrete shape for this: a per-document
`retention: durable | perishable | snapshot` frontmatter hint, a `doctor`
extension that would stay silent on `durable`/`snapshot` and surface
stale `perishable` docs as "archive candidates," and possibly an
`aikata compact` command. The classification was first framed by
**question type** — WHY/HOW (design rationale, durable) versus
WHAT/WHEN/WHERE (current state, locations, versions; perishable).

Investigation showed the *need* is already met, and the *proposed
mechanism* is the wrong shape:

- **The pain is already addressed.** The measured pain — ~12k lines of
  docs entering first-read context every session — is what
  [ADR 0039](./0039-documentation-hygiene-and-context-budget.md) solved
  with a per-file-class hygiene rubric pruned every release, what
  [ADR 0044](./0044-doc-map-derived-artifact.md) cut further with an
  always-current doc map, and what [ADR 0037](./0037-tighten-adoption-mutation-boundaries.md)
  narrowed via doctor's managed surface. No residual pain is observed,
  so a new mechanism does not pass ADR 0028's demand-driven test.
- **What was missing is the *why*, not a feature.** ADR 0039 records
  *which method applies to which file class* but not the underlying
  model of *why* documents age differently. That conceptual model — and
  a record of why the per-document mechanism was rejected — is durable
  WHY-layer knowledge worth freezing so it is not re-litigated. This ADR
  fills that gap and binds nothing operationally beyond ADR 0039.

## Decision

### D1 — The axis is *source / regenerability*, a present fact

Classify documentation content by where its truth comes from — a
question answerable **now**, deterministically — not by predicted decay
speed (a guess about the future that goes stale the moment it is wrong).

| Bucket | Test (answerable today) | Correct treatment | Existing discipline |
|---|---|---|---|
| **canonical** | Is this the sole source of this fact? (lost ⇒ the fact is lost) | If stale, prompt an update (nag) | [GLOSSARY](../../GLOSSARY.md) *canonical source* |
| **derived** | Can it be regenerated from code, git, or another canonical doc? | If stale, regenerate; do not nag | ADR 0044 doc map; *lossy generation* |
| **log** | Is it an append-only record of an irreversible observation? | Freeze: do not edit, do not compress | ADR 0004 append-only / supersede-in-place |

The model is **descriptive of the design aikata already has**, not a new
surface. Each bucket maps onto an existing discipline; nothing new is
introduced to enforce it.

### D2 — Question type is a heuristic, not the axis

WHY/HOW/WHAT/WHEN/WHERE is kept only as an *explanatory heuristic* for
which bucket content tends to fall into (WHY tends canonical/log; WHAT
tends derived). It is **not** the classification axis, because type and
lifetime are correlated but not causal:

- **WHY also ages** — a superseded ADR is dead rationale; but it ages by
  *supersession* (a new ADR, ADR 0001), not by silent rot.
- **WHAT can be immutable** — a licence or module path is WHAT that never
  decays; a past event ("the repo was private until v0.1") is WHAT frozen
  the instant it was observed.
- **A single file mixes types** — `docs/memory/**` carries durable
  rationale, perishable state, and frozen observations as *bullets* in
  one file. A file-level label cannot express bullet-level mixing, so it
  would mislabel by construction.

### D3 — "Raw snapshot beats compression" is irreversibility, not slow decay

The intuition that a raw point-in-time record can be worth more than a
compressed summary is real, and its cause is **information
irreversibility**, not a slow decay rate. Summarizing lowers the
resolution of an observation that cannot be reconstructed; once lowered
it cannot be restored. Therefore append-only logs are *frozen*, never
mechanically compressed — already guaranteed by ADR 0004
(supersede-in-place, never delete), the no-silent-deletes discipline, and
ADR 0039's move-not-delete guardrail. Nothing new is needed to "keep raw
snapshots."

## Consequences

**Positive**:

- A durable rationale that forecloses re-litigating per-document
  retention metadata and runtime compaction; future explorers find the
  decision and the *why* instead of re-deriving them.
- ADR 0039's file classes gain their underlying model: classes follow
  source, and source is decided by file location (convention), which is
  why the rubric is keyed on path and not on per-file config.
- Zero code, zero new config key, zero derived-artifact residue — so it
  is trivially Do-No-Harm compliant (ADR 0003).

**Negative**:

- One more ADR adjacent to the first-read surface. Accepted: an ADR is
  immutable WHY (the cheapest durable slot), and this one *defers* scope
  rather than expanding it, so it shrinks future context churn rather
  than adding to it.

## Alternatives considered

- **Per-document `retention: durable|perishable|snapshot` frontmatter.**
  Rejected. It regresses convention→configuration (value is already
  decided by file class / location per ADR 0039); a file-level label
  cannot express the bullet-level WHY/WHAT mixing real in
  `docs/memory/**`; and it re-implements ADR 0039's rubric at a worse
  granularity.
- **Question type (WHY/HOW/WHAT/WHEN/WHERE) as the decay axis.**
  Rejected as a category error (D2): type and lifetime are not causally
  linked, and the counter-examples (superseded ADRs, immutable WHAT,
  mixed files) make it mislabel.
- **Runtime auto-compaction / `aikata compact`** (LLM summarization of
  the document set). Deferred. It collides with the SPEC non-goals —
  aikata calls no LLM API and performs no runtime context selection or
  token counting — and it fails ADR 0028's evidence test: the pain it
  would target is already handled by ADR 0039 + 0044 + 0037, with no
  observed residual.
- **Path-based `source:` inference in `doctor`** (no new frontmatter
  key; classify by `ManagedIncludeGlobs`-style globs). Deferred to a
  future phase, gated on observed evidence, tracked as
  [Q-CONTEXT-01](../decisions/open-questions.md#q-context). Preferred
  over a hand-written field if the model is ever operationalized, because
  derivable classifications should be derived, not hand-labelled (the
  ADR 0044 principle).

## References

- ADR 0039 (documentation hygiene & context budget — the operational
  mechanism).
- ADR 0044 (doc map — derived, not hand-curated; the `source: derived`
  precedent).
- ADR 0037 (doctor managed-surface scope).
- ADR 0004 (long-term memory — append-only / supersede-in-place; the
  `source: log` basis).
- ADR 0028 (demand-driven, not architecture-driven — the evidence test).
- [`GLOSSARY.md`](../../GLOSSARY.md) — *canonical source*, *lossy
  generation*, *doc map*.
- [`SPEC.md`](../../SPEC.md) §8 Out-of-Scope Examples — no LLM API, no
  runtime context selection.
- Open questions:
  [Q-CONTEXT-01](../decisions/open-questions.md#q-context).
