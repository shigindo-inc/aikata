---
project: aikata
status: draft
version: 0.0.1
updated: 2026-08-14
audience: [human, agent]
---

# ADR 0047 - Modeling capability and per-feature design loop: boundary

- **Status**: Accepted
- **Date**: 2026-08-14
- **Deciders**: aikata maintainers
- **Related**: ADR 0007 (no generic `DESIGN.md` — the "prove no existing
  document can own it" test this decision discharges), ADR 0017
  (post-init command taxonomy — `enable` for capabilities), ADR 0026
  (workflow guides as opt-in collaboration docs — the slot this decision
  deliberately leaves unused), ADR 0028 (prioritize core-concept
  stabilization — demand-driven, not speculative), ADR 0031 (brand
  exploration documents as one-off artifacts — why `docs/design/` is
  already taken), ADR 0043 (command-wrapper skill surface & simple skill
  names — the skill-not-verb precedent), ADR 0045 (documentation value
  model — both new documents are canonical stock), ADR 0046 (structure-
  migration assistant boundary — the sibling boundary-ADR pattern this
  one follows).

## Context

Dogfooding aikata on real feature work surfaced a concrete failure, not a
hypothetical one: **without a written record of who does what and why,
the data model and its parameters drift.** Fields get invented
speculatively, and no document says whether anything actually needs
them.

The generated document set has a real hole at exactly that point. The
standard `SPEC.md` template names users (§2.3) and functional
requirements (§3) — but nothing owns the band between them: *who,
triggered by what, in what order, reaching what outcome, failing how*.
One layer down, nothing owns the entities, their fields, their
invariants, and their state transitions. `GLOSSARY.md` owns terms;
`ARCHITECTURE.md` owns structure; neither owns this.

That second gap is the more expensive one for agents. Entity and field
truth is what an agent reaches for most often, and fabricates most
often, when no document owns it.

The ordering between the two matters: **behaviour first, data second**.
Modelling data before behaviour is what admits speculative fields. But
it is a round trip, not a phase gate — sketching the model routinely
invalidates a use case and sends work back upstream.

The full rationale — including the judgement-frame table (ADR 0045
stock/flow, the value-surfaces test, the ADR 0007 discharge, ADR 0028
demand-driven test, and the agile-lifecycle stance) — is recorded in the
design note that backs this ADR:
[`docs/decisions/2026-08-14-modeling-capability-and-model-feature-skill-design.md`](../decisions/2026-08-14-modeling-capability-and-model-feature-skill-design.md).
This ADR is the durable record; the note carries the working-out.

## Decision

### D1 — One opt-in capability, one document pair, one deliberately unused name

`aikata enable modeling` (and `aikata init --with-modeling`) renders both
`docs/usecases.md` (behaviour) and `docs/domain.md` (structure) as a
pair. They are never enabled separately — they are read and edited back
and forth, and half the pair cannot discharge either direction of the
traceability check in D4.

The capability is named `modeling`, not the more obvious label for
"data model plus its rules." That established label is avoided
everywhere in this feature's documents and skill body, deliberately: an
LLM reading it carries a strong prior for a specific, heavyweight
design school (aggregates, repositories, bounded contexts), and would
start importing that design unbidden into a project that asked for
none of it. `modeling` was chosen instead because it covers both the
model of behaviour (use cases) and the model of structure (domain)
without borrowing that vocabulary. The obvious alternative name,
`design`, is also unusable — it collides with `docs/design/`, already
defined as the home for one-off brand-exploration artifacts (ADR 0031).
`GLOSSARY.md` gains one entry (**per-feature design loop**) that pins
this process layer apart from the existing structural term
**document-centered**, so the relationship is written down once instead
of re-explained in every skill body.

### D2 — The use-case ledger lives under `docs/`, not as a `SPEC.md` section

An earlier draft placed the use-case ledger inside `SPEC.md`. Rejected
once the skill's granularity was fixed at one-feature-per-run, on
**update-frequency asymmetry**: `SPEC.md` states what and why and rarely
changes, while a use-case ledger changes on every feature. Co-locating
them would make `SPEC.md`'s diff permanently noisy with ledger appends,
burying the SPEC changes that actually matter in review. It would also
split the pair that is edited back and forth (D1) across two distant
files.

This asymmetry is also how this decision discharges ADR 0007's test:
"prove no existing document can own it before adding a new one." The
content cannot live in `SPEC.md` without degrading `SPEC.md`, so a new
canonical home is justified.

### D3 — A new first-party skill, `model-feature`, that ends before implementation

Not an extension of the existing `track-context` skill. `track-context`
fires at the start of *every* session; `model-feature` fires only when a
feature with observable behaviour change is being designed. The two
have different trigger conditions, and merging them into one skill body
would degrade the firing accuracy of both — this is the same reasoning
ADR 0043 already applies to keep individual skills narrow.

`model-feature` is a **per-feature design loop**: one feature per run,
walking use case → domain model → decision → hand-off, with the
bidirectional check at its centre (every field the use case's success
state needs must exist in `docs/domain.md`, and no field just added may
be unreachable from any use case — the reverse direction is what catches
the speculative field from the Context section above).

It fires on exactly **one criterion**: externally observable behaviour
changes. Refactors, bug fixes, copy tweaks, and UI position adjustments
do not fire it. A single criterion beats a skip-condition table, and it
is what keeps the loop from degenerating into ceremony on trivial
changes.

Its **terminal state is the hand-off just before implementation**. The
skill confirms the documents are complete and consistent, then hands
off to the project's normal development flow — it never carries out the
implementation itself. Carrying implementation would collide with the
general-purpose planning/implementation skills already in the wider
ecosystem and would bloat the skill until its firing accuracy
collapsed.

Enforcement is **advisory, not gating**: unmet completion conditions are
reported, not blocked, consistent with the Do-No-Harm posture and
aikata's user-approval culture.

### D4 — Traceability closes inside `docs/`, at field granularity, and goes no further

Use cases carry IDs in their headings (`## UC-01 — …`). `docs/domain.md`
carries a `Related UC` column at **field** granularity, not entity
granularity — entity-level linkage is too coarse to reveal an
unreachable *field*, which is the exact failure D3's bidirectional check
targets. The maintenance cost is bounded by writing only *principal*
fields; mechanical ones (`id`, `created_at`) are omitted.

Traceability is **never extended into code or tests**. aikata does not
read code ([SPEC §8](../../SPEC.md#8-out-of-scope-examples-for-clarity)),
so it could not verify such links; they would become hand-synchronised
and rot silently instead of being caught by `doctor`.

### D5 — Zero new CLI machinery

No `aikata modeling …` verb. `aikata doctor` gains no new check — the
two documents simply join the existing managed surface
(`internal/doctor/scope.go`), so the existing frontmatter / link /
unused-GLOSSARY-term checks apply to them for free. The skill calls only
the existing `fill` / `doctor` / `map` surface, reads the documents
itself, and reports; content is written by the agent with user
approval. This is the ADR 0046 CLI/skill mutation line held in a second
place: the CLI stays observation-and-scaffolding-only, judgement stays
in the skill.

## Consequences

**Positive**:

- The band between users and functional requirements, and the domain
  model one layer below it, now have a canonical, doctor-validated home
  — closing the gap in the Context section without adding a document
  that duplicates `SPEC.md`, `ARCHITECTURE.md`, or an ADR.
- The bidirectional check (D4) gives aikata a mechanical way to catch a
  speculative field — the actual failure mode observed in dogfooding —
  rather than relying on review discipline alone.
- Zero new CLI surface, zero new `doctor` check implementation, zero new
  config machinery beyond one schema-v2 boolean: the feature rides
  entirely on disciplines aikata already has (create-or-skip rendering,
  managed-surface globs, manifest tracking).
- `GLOSSARY.md`'s existing unused-term detection becomes a free integrity
  check on `docs/domain.md` vocabulary, once entity names are entered
  there — no new implementation needed for it.

**Negative**:

- **The per-feature granularity is a pre-measurement design decision and
  may be superseded.** This shipped without first running the loop on a
  real feature to validate that one-feature-per-run is the right unit —
  the recommended sequence of landing the `modeling` component alone,
  observing it, and adding the skill later was consciously declined in
  favour of shipping both together. If dogfooding shows the granularity
  is wrong (too fine, too coarse, or wrong at project kickoff), this ADR
  will need a successor rather than a quiet edit, since ADR bodies are
  immutable once Accepted (ADR 0001).
- The field-granular `Related UC` column (D4) is an ongoing authoring
  cost on every domain-model edit, not a one-time cost. Its justification
  is narrow — it is the column's entire reason to exist — so if the
  bidirectional check turns out not to catch real drift in practice, the
  cost has no fallback justification.
- `docs/usecases.md` and `docs/domain.md` are two more documents in the
  first-read context budget for a project that enables `modeling`.
  Accepted because both are canonical **stock** (ADR 0045) — they change
  when the product's behaviour changes, not on every session — so they
  do not carry the churn cost a flow document would.
- Choosing `modeling` and `per-feature design loop` over the
  industry-standard label means a reader arriving with that label in
  mind will not find it by search. Accepted: the cost of a reader
  missing a search hit is smaller than the cost of every agent silently
  importing an unrequested design school into every project that enables
  this capability.

## Alternatives considered

- **Use cases as a `SPEC.md` section.** Rejected (D2) on update-frequency
  asymmetry: a ledger that changes every feature cannot share a file with
  a document that changes rarely without burying real `SPEC.md` changes
  in ledger noise. This is also how ADR 0007's "prove no existing
  document can own it" test is discharged for this decision.
- **A paired `docs/workflows/design.md` convention document.** Deferred,
  not rejected outright. If the skill already carries the sequence and
  the completion conditions, a convention file would hold only
  project-specific deviations from them — and whether any project
  actually needs one is unknown (ADR 0028's demand-driven test). A
  convention file left at its defaults is an empty stub, and an empty
  stub is worse than an absent file in an agent-facing document set: an
  agent that finds it has to read it to learn there is nothing there.
  The ADR 0026 workflow-guide slot stays open and unused until a real
  deviation appears.
- **Extending `track-context` instead of adding a new skill.**
  Rejected (D3). The two skills have different trigger conditions —
  every session start versus one feature with observable behaviour
  change — and merging them would degrade the firing accuracy of both,
  the same reasoning that already keeps aikata's other skills narrow
  (ADR 0043).
- **Entity-granular rather than field-granular traceability.**
  Rejected (D4). Entity granularity cannot reveal an unreachable
  *field* — a use case can legitimately touch an `Order` while never
  needing `Order.cancel_reason` — and an unreachable field is precisely
  the failure this feature exists to catch. Field granularity was kept
  affordable by listing only principal fields.

## References

- [Design note](../decisions/2026-08-14-modeling-capability-and-model-feature-skill-design.md)
  — the rationale source for this ADR; carries the judgement-frame table,
  the document-shape examples, and the full risk/mitigation table.
- [`GLOSSARY.md`](../../GLOSSARY.md) — *document-centered*, *per-feature
  design loop*.
- [`docs/layout.md`](../layout.md) §2 — the two new document homes.
- [`SPEC.md`](../../SPEC.md) §4.2, §8 — the `enable modeling` capability
  target; aikata does not read code.
- [`ARCHITECTURE.md`](../../ARCHITECTURE.md) §3 — generated-structure
  coverage for the opt-in pair.
- Open questions:
  [Test strategy](../decisions/open-questions.md#q-modeling-01--test-strategy-for-the-per-feature-design-loop)
  and
  [Data handling & privacy](../decisions/open-questions.md#q-modeling-02--data-handling--privacy-declarations)
  — the two items deliberately deferred out of this decision's scope.
