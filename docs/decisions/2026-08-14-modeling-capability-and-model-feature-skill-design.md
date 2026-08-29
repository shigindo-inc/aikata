---
project: aikata
status: draft
version: 0.0.1
updated: 2026-08-14
audience: [human, agent]
---

# Design — `modeling` capability & the `model-feature` skill

> Design note (rationale source) for two things that ship together: an
> opt-in **`modeling` capability** that renders a document pair
> (`docs/usecases.md` + `docs/domain.md`), and a first-party
> **`model-feature` skill** that fills that pair one feature at a time.
> The durable record will be a boundary ADR under `docs/adr/` (in the
> spirit of [ADR 0046](../adr/0046-structure-migration-assistant-boundary.md));
> this note carries the reasoning it cites. Mirrors how
> [`2026-06-17-watched-external-doc-homes-design.md`](./2026-06-17-watched-external-doc-homes-design.md)
> backs its ADR pair.

---

## 1. Why this exists

Observed, not hypothesised. Dogfooding aikata on real app development
surfaced a concrete failure: **without a written record of who does what
and why, the data model and its parameters drift.** Fields get invented
speculatively, and no document says whether anything actually needs them.

The document set has a real hole at exactly that point. `SPEC.md` §2.3
names personas, §3 names functional requirements — but nothing owns the
band between them: *who, triggered by what, in what order, reaching what
outcome, failing how*. And one layer down, nothing owns the entities,
their fields, their invariants, and their state transitions. `GLOSSARY.md`
owns terms; `ARCHITECTURE.md` owns structure; neither owns the domain
model.

That second gap is the more expensive one for agents. Entity and field
truth is what an agent reaches for most often and fabricates most often.

The ordering between the two is not arbitrary: **behaviour first, data
second**. Modelling data before behaviour is what admits speculative
fields. But it is a **round trip, not a phase gate** — sketching the model
invalidates use cases, which sends work back upstream.

## 2. The judgement frames, applied

| Frame | Verdict |
|---|---|
| [ADR 0045](../adr/0045-documentation-value-model.md) — stock vs flow | ✅ Both documents are `canonical` knowledge stock. They change when the product changes, not daily. Unlike a backlog, they do not churn |
| Value surfaces (`generate`/`sync`/`doctor`/`docmap`) | ✅ "100% user content, therefore worthless to aikata" is the argument that excluded backlogs — but `SPEC.md` is 100% user content too. What aikata gives a *stock* document is a thinking scaffold, a canonical home, freshness/link checking, and a place in first-read context. That applies here |
| [ADR 0007](../adr/0007-no-generic-design-md.md) — prove no existing doc can own it | ✅ Discharged by **update-frequency asymmetry** (see D2) |
| [ADR 0028](../adr/0028-prioritize-core-concept-stabilization.md) — demand-driven | ✅ Observed pain, not a hypothetical |
| [Agile stance note](./2026-06-24-agile-lifecycle-scope-stance.md) §4 | ✅ Sequencing is process **policy**, not process **state**. Policy is the one designated opening |

## 3. Decisions

### D1 — One capability, one document pair

`aikata enable modeling` (and `init --with-modeling`) renders both
`docs/usecases.md` and `docs/domain.md`. They are never enabled
separately: they are read and edited as a pair, and half the pair cannot
discharge either completion condition in D4.

### D2 — Use cases live under `docs/`, not as a `SPEC.md` section

An earlier draft placed the use-case ledger in `SPEC.md`. Rejected once
the skill's granularity was fixed at one-feature-per-run:

1. **Update-frequency asymmetry.** `SPEC.md` states what and why and
   rarely changes; a use-case ledger changes on every feature. Co-locating
   them means SPEC's diff is permanently noisy with ledger appends, and
   real SPEC changes get lost in review.
2. **The pair round-trips.** Use cases and the domain model are edited
   back and forth (§1). Splitting them across `SPEC.md` and `docs/` means
   opening two distant files on every pass. Things that move together
   should sit together.

This asymmetry is also the ADR 0007 discharge: the content cannot live in
`SPEC.md` without degrading `SPEC.md`.

### D3 — A new first-party skill, `model-feature`

Not an extension of an existing skill. `track-context` fires at the start
of *every* session; this fires only when a feature is being designed.
Merging different trigger conditions into one skill degrades the firing
accuracy of both.

**Granularity: one feature per run.** Project kickoff is absorbed by
running the loop a few times; the reverse does not hold — a
kickoff-batch mode never learns incremental amendment, which is where
document rot starts.

**Terminal state: the hand-off just before implementation.** The skill
confirms the documents are complete and consistent, then hands off to the
project's normal development flow. Implementation is explicitly *not* its
responsibility — carrying it would collide with the general-purpose
planning/implementation skills in the wider ecosystem, and would bloat the
skill until its firing accuracy collapsed. Nothing about implementation
makes aikata more aikata.

### D4 — Four steps, with the bidirectional check at the centre

| # | Action | Destination | Completion condition |
|---|---|---|---|
| 1 | Add/update this feature's use case | `docs/usecases.md` | At least one UC states actor, trigger, and success state, **and names at least one exception path** |
| 2 | Propagate to the domain model; fix vocabulary | `docs/domain.md`, `GLOSSARY.md` | **Bidirectional** (below) |
| 3 | If a judgement call was made, record it | `docs/adr/` | Conditional — only where genuine alternatives existed. Skipped otherwise |
| 4 | Check and hand off | — | `aikata doctor` passes; no unused GLOSSARY terms |

Step 2's condition runs **both ways**:

- **Forward** — every piece of data the UC's success state requires exists
  in `docs/domain.md`.
- **Reverse** — no newly added field or state is unreachable from any use
  case.

The reverse direction is the point of the whole design. Forward-only
catches omissions but not drift; only the reverse direction catches the
speculative field that started §1.

### D5 — One firing criterion, advisory enforcement

Fires iff **externally observable behaviour changes**. Refactors, bug
fixes, copy tweaks, and UI position adjustments do not fire. A single
criterion beats a skip-condition table, and it is what keeps the loop from
degenerating into ceremony on trivial changes.

Unmet completion conditions are **reported, not gated** — consistent with
the Do-No-Harm posture and aikata's user-approval culture.

### D6 — Traceability closes inside `docs/`

Use cases carry IDs in their headings (`## UC-01 — …`). `docs/domain.md`
carries a **`related UC` column at field granularity**.

Field granularity, not entity granularity, is deliberate: entity-level
linkage is too coarse to detect an unreachable *field*, which is the exact
failure being targeted. The maintenance cost is bounded by writing only
*principal* fields — mechanical ones (`id`, `created_at`) are omitted.

Traceability is **not** extended into code or tests. aikata does not read
code ([SPEC §8](../../SPEC.md)), so it could not verify such links; they
would become hand-synchronised and rot. This matches the stance note §5.

### D7 — The name "DDD" is not used anywhere

Not cosmetic. LLMs carry a strong prior for **DDD = Domain-Driven
Design**. A workflow document or skill body that says "DDD" causes agents
reading it to import aggregates, repositories, and bounded contexts
unbidden — design that has nothing to do with the intent here.
`GLOSSARY.md` gains one entry pinning the relationship between
*document-centered* (structure, the existing canonical term) and the
process layer this skill adds.

The capability name `design` is likewise unusable: it collides with
`docs/design/`, already defined as the home for one-off brand-exploration
artifacts ([ADR 0031](../adr/0031-brand-exploration-documents-as-one-off-artifacts.md)).
Hence `modeling` — it covers both the model of behaviour (use cases) and
the model of structure (domain) without borrowing DDD vocabulary.

### D8 — Zero new CLI machinery

No `aikata modeling …` verb. `doctor` gains no new check. The skill calls
only the existing `fill` / `doctor` / `map`, reads the documents itself,
and reports. Content is written by the agent with user approval. This is
the ADR 0046 line held in a second place.

GLOSSARY consistency needs no implementation at all: putting entity names
in `GLOSSARY.md` makes the existing unused-term detection work as an
integrity check for free.

### D9 — `docs/workflows/design.md` is not created yet

An earlier draft paired the skill with an opt-in convention document.
Withdrawn. If the skill already carries the sequence and the completion
conditions, the convention file holds only project-specific deviations —
and whether any are needed is unknown (ADR 0028). A convention file left
at its defaults is exactly the empty stub that is worse than no file for
an agent-facing document set. The [ADR 0026](../adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)
slot stays open and unused until a real deviation appears.

## 4. Document shapes

`docs/usecases.md` — individual sections only; no summary table. A ledger
that keeps both a table and per-item sections always rots on one side.
Four required fields; preconditions are optional and added only when they
carry weight.

```markdown
## UC-01 — Cancel an order
- Actor: buyer
- Trigger: chooses "cancel" on the order detail screen
- Success state: order becomes `cancelled`; a refund is scheduled
- Main exception: already shipped ⇒ cancellation refused, routed to returns
```

`docs/domain.md` — per-entity, with the field table carrying the D6 link.

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

## 5. Wiring to the existing surface

- **`track-context`** gains one line: at the start of work, if externally
  observable behaviour will change, point at `model-feature`. No logic.
- **`model-feature`** ends by reporting readiness and handing off.
- **`doctor`**, **`manage-docs`**, **`refresh-docs`**, **`migrate-structure`**
  are untouched.

## 6. Out of scope

Deliberately excluded, with re-entry criteria. Recorded in
[`open-questions.md`](./open-questions.md), **not** in `ROADMAP.md` —
listing them as planned would be the speculative commitment ADR 0028
forbids.

| Item | Status | Re-entry criterion |
|---|---|---|
| Test strategy (extend the `tdd` component; `docs/workflows/testing.md`) | deferred | A concrete instance of agent/human disagreement about test layers or scope |
| Data handling & privacy (collected fields, permissions, store declarations) | deferred | Work actually blocked on a store declaration or permission design. Carries an extra unresolved question: generic capability vs. the `flutter` stack brief, since iOS and Android declaration formats differ |
| Screen/route inventory; external integration contracts | **permanently declined** | — Screens surface in UC triggers and success states; integration contracts belong to `ARCHITECTURE.md`. A separate document would duplicate and rot |

Both deferred items sit *on top of* this work rather than beside it, so
the ordering is correct:

- The mandatory **exception path** in each UC is directly the input to a
  test scenario. UC IDs (D6) are already the join key a future `tdd`
  extension would use.
- The **field-granular table** in `docs/domain.md` can take
  "personal data?" and "retention" columns later, deriving privacy
  declarations from the domain model. An entity-granular table would have
  closed that path.

## 7. Risks & mitigations

| Risk | Mitigation |
|---|---|
| **Granularity is wrong**, and it ships without local validation first (the recommended component-first / skill-later sequence was declined in favour of shipping both) | Initial scope frozen at four steps, the completion conditions, and the firing criterion — no options, no modes, no config keys. The boundary ADR states explicitly that granularity is pre-measurement and may be superseded. Validate on a real feature immediately after release |
| **Maintenance cost** of the field-granular `related UC` column | Principal fields only; mechanical fields omitted. This column is the feature's reason to exist (D6) |
| **Overlap** with general-purpose brainstorming/planning skills | This skill owns only the mapping from content to canonical destination. The thinking process is delegated, not reinvented |
| **Waterfall creep** | One firing criterion (D5), advisory not gating, one feature per run |
| **Progress state leaking into a new file** | "Which step are we on" is flow state. One line in `docs/tasks/current.md` is the ceiling; no dedicated progress file (agile stance note) |

## 8. Implementation map (to be confirmed during planning)

- `internal/templates/data/components/modeling/{ja,en}/` — `usecases.md.tmpl`,
  `domain.md.tmpl`. `internal/components/singlefile.go` is single-target by
  construction (`targetPath`, `tmplName`), and multi-file capabilities
  already use their own implementation (`memory.go`). So `modeling` needs
  either its own small renderer or a thin two-`singleFile` wrapper —
  decide during planning, but it is **not** a `singleFile` registration.
- `internal/components/registry.go` — enrol in `Capabilities`.
- Schema v2 `components.modeling` boolean (ADR 0016 OR-merge applies).
- New first-party skill `model-feature` under the plugin skill layout
  (ADR 0041 / ADR 0043 naming).
- One line added to `track-context`.
- Boundary ADR under `docs/adr/`.
- `GLOSSARY.md` — one entry (D7).
- `docs/layout.md` §2 — add the two paths.
- `open-questions.md` — two deferred entries (§6).
- `SPEC.md`, `ARCHITECTURE.md`, `CHANGELOG.md` — surface updates.

## 9. References

- [ADR 0007](../adr/0007-no-generic-design-md.md) — new documents must
  prove no existing document can own them.
- [ADR 0017](../adr/0017-post-init-command-taxonomy.md) — `enable` for
  capabilities; no compatibility aliases, so naming is decided up front.
- [ADR 0026](../adr/0026-workflow-guides-as-opt-in-collaboration-docs.md) —
  the workflow slot left open by D9.
- [ADR 0028](../adr/0028-prioritize-core-concept-stabilization.md) —
  demand-driven.
- [ADR 0031](../adr/0031-brand-exploration-documents-as-one-off-artifacts.md) —
  why `docs/design/` is taken.
- [ADR 0045](../adr/0045-documentation-value-model.md) — canonical /
  derived / log.
- [ADR 0046](../adr/0046-structure-migration-assistant-boundary.md) — the
  boundary-ADR pattern and the CLI/skill mutation line.
- [Agile lifecycle scope stance](./2026-06-24-agile-lifecycle-scope-stance.md) —
  process policy vs. process state.
- [`docs/layout.md`](../layout.md) — prescriptive document homes.
- [SPEC §8](../../SPEC.md) — aikata does not read code.
