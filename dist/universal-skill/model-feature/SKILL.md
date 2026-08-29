---
name: model-feature
user-invocable: false
description: Use when designing one feature whose externally observable behaviour changes, in a repository that has aikata's modeling capability enabled (`docs/usecases.md` and `docs/domain.md` present). Walks a single feature from use case to domain model to a hand-off point just before implementation — writing the use case into docs/usecases.md, propagating required data into docs/domain.md, fixing new terms in GLOSSARY.md, and recording a decision in docs/adr/ when alternatives existed. Do not use for refactors, bug fixes, copy changes, or UI position adjustments (no observable behaviour change), in repositories aikata does not manage, or to carry out the implementation itself — this skill stops at the hand-off. For the daily context loop use `track-context`; for CLI invocation use `manage-docs`.
---

# model-feature

Design one feature at a time by writing it into the canonical documents,
in an order that keeps the data model honest: **behaviour first,
structure second**. Modelling data before behaviour is what admits
fields nothing actually needs.

This skill owns only the mapping from content to canonical destination.
It does not replace whatever thinking or planning process the project
already uses, and it stops before implementation.

## When this applies

Run this loop when **externally observable behaviour changes**. That is
the whole firing criterion — do not run it for refactors, bug fixes,
copy tweaks, or UI position adjustments.

It requires `docs/usecases.md` and `docs/domain.md`. If they are absent,
the repository has not enabled the capability; hand off to `manage-docs`
to run `aikata enable modeling` and confirm with the user before
proceeding.

Run it once per feature. A project kickoff is just this loop run several
times.

## The loop

### Step 1 — Write the use case

Add or update an entry in `docs/usecases.md`. Give a new use case the
next free `UC-NN` id; never reuse a retired id.

**Done when** the entry states an actor, a trigger, and a success state,
**and names at least one exception path**. The exception is required:
it is what distinguishes a use case from a feature list, and it is the
seed of a test scenario later.

### Step 2 — Propagate to the domain model

Update `docs/domain.md` with whatever entities, fields, invariants, and
state transitions this use case implies. Put every new term into
`GLOSSARY.md`.

**Done when both directions hold:**

- **Forward** — every piece of data the use case's success state needs
  exists in `docs/domain.md`.
- **Reverse** — no field or state you just added is unreachable from any
  use case.

The reverse direction is the point of this step. Forward-only catches
omissions but not drift; only the reverse direction catches a field that
nothing needs. When it fails you have two honest options — delete the
field, or write the use case that justifies it. Say which one you are
taking.

Expect to go **back to Step 1** here. Sketching the model routinely
invalidates a use case; that round trip is the loop working, not a
mistake.

### Step 3 — Record a decision, if there was one

If the shape you chose had real alternatives, write an ADR under
`docs/adr/` (hand off to `manage-docs` for `aikata new adr "<title>"`).

Skip this step when there was no genuine fork. Most features have none.

### Step 4 — Check and hand off

- Run `aikata doctor` (via `manage-docs`) and report the result.
- Check `GLOSSARY.md` has no unused terms.
- Note the in-flight feature in `docs/tasks/current.md` — **one line**.
  Step-by-step progress is not tracked anywhere; that is working state,
  not knowledge.

Then stop and hand off to the project's normal implementation flow.
Implementation is explicitly outside this skill.

## Reporting, not gating

Unmet conditions are **reported, not enforced**. State plainly what is
unsatisfied and let the human decide whether to proceed. Never block,
and never silently fill a gap with a guess.

## Document shapes

`docs/usecases.md` — one section per use case, four required fields:

```markdown
## UC-01 — Cancel an order
- Actor: buyer
- Trigger: chooses "cancel" on the order detail screen
- Success state: order becomes `cancelled`; a refund is scheduled
- Main exception: already shipped ⇒ cancellation refused, routed to returns
```

`docs/domain.md` — per entity, with `Related UC` at **field**
granularity. That granularity is deliberate: entity-level linkage is too
coarse to reveal an unreachable field, which is the failure this exists
to catch.

```markdown
### Order
- Description: one confirmed purchase by a buyer
- Invariant: an order that reached `cancelled` cannot leave it

| Field | Type | Meaning | Related UC |
|---|---|---|---|
| status | enum | progress state of the order | UC-01, UC-03 |
| cancel_reason | text | why it was cancelled | UC-01 |

- Transitions: placed → shipped (UC-02) / placed → cancelled (UC-01)
```

List only fields that carry meaning. Mechanical ones (`id`,
`created_at`) are omitted.

Use the field labels and column heading already present in the file —
Japanese projects render these in Japanese; do not translate them or add
a parallel column.

## Boundaries

- **Traceability stays inside `docs/`.** Do not push use-case ids into
  test names or source comments. aikata does not read code, so such
  links cannot be verified and will rot.
- **No new files.** Everything lands in `docs/usecases.md`,
  `docs/domain.md`, `GLOSSARY.md`, `docs/adr/`, or the one line in
  `docs/tasks/current.md`.
- **No progress file.** "Which step are we on" is working state.
- Sibling skills: `track-context` (daily context loop), `manage-docs`
  (raw CLI surface).
