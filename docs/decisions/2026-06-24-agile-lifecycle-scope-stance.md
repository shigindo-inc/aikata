---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-24
audience: [human, agent]
---

# Stance — agile/scrum lifecycle coverage & where the boundary holds

> A **stance note**, not a decision: it records *why aikata does not absorb
> agile/scrum process-management*, so the proposal "let's add backlog /
> sprint / board management" does not get re-litigated from scratch each
> time it resurfaces. It promises nothing operationally. If real demand
> appears, the resolution path is a boundary ADR under `docs/adr/` (in the
> spirit of [ADR 0046](../adr/0046-structure-migration-assistant-boundary.md)),
> not this note. Sits underneath
> [SPEC §2.2](../../SPEC.md) (Non-Goal: no PM features) and
> [ADR 0045](../adr/0045-documentation-value-model.md) (the value model
> that already draws the line).

---

## 1. The question

Can aikata's managed documents carry an agile/scrum project across its
full lifecycle — upstream (vision, requirements, planning) through
downstream (implementation, test, release) — by document generation and
maintenance alone? And if there are gaps, should aikata grow to fill them?

## 2. What is already strong (the durable-knowledge spine)

The **knowledge-stock** axis is well covered and well-separated:

| Phase | Home | Note |
|---|---|---|
| Vision / why | `SPEC.md` §1, §7 hypotheses | strong |
| Ubiquitous language | `GLOSSARY.md` (doctor flags unused terms) | scrum/DDD-friendly |
| Requirements | `SPEC.md` §4 (goals / acceptance) | single canonical, no story granularity |
| Release direction | `ROADMAP.md` (version milestones) | coarse |
| Design / decisions | `ARCHITECTURE.md` + `docs/adr/` (immutable, supersede) | the standout |
| Implementation rules | `AGENTS.md` + `docs/stacks/` | strong |
| Working state | `docs/tasks/current.md` (backlog/archive forbidden) | one in-flight point only |
| Release / ops | `CHANGELOG.md`, `docs/troubleshooting.md` | adequate |
| Durable learning | `docs/memory/{feedback,project}` | adequate |

## 3. What is missing — and why that is correct, not an oversight

The gap is the **iteration / flow band**: product & sprint backlog,
sprint goal, DoR/DoD, review & retro records, requirement-level
traceability. There is no home for them between `ROADMAP.md` (coarse
direction) and `docs/tasks/current.md` (the single current point), which
explicitly forbids being a backlog.

This is by design. aikata manages **what a project knows** (a knowledge
stock), not **what iteration N is doing right now** (a flow). Three
reasons make *not widening* the active-correct call, not mere caution:

1. **aikata's machinery adds no value to flow artifacts.** All four value
   surfaces — `generate` (canonical → AI configs), `sync` (3-way merge of
   upstream templates), `doctor` (consistency), `docmap` (derived map) —
   assume "a canonical doc derives AI config and receives upstream
   updates." A backlog/board/burndown is 100% user-specific content with
   no upstream to merge, derives nothing, and churns daily — so `sync`
   no-ops, `generate` skips it, and the `init`-once model leaves a dead
   file none of the four surfaces ever curates again. Not "unsafe to add"
   — *no value to add*.
2. **[ADR 0045](../adr/0045-documentation-value-model.md) already draws
   the line.** Scrum flow artifacts are none of canonical / derived / log
   — they are high-churn operational state whose source of truth is
   naturally an external tracker (Jira / Linear / GitHub Projects /
   Obsidian). The existing value model says "keep them out" without
   needing a new judgment.
3. **It collides with the differentiation axis.** SPEC §1.3 positions
   aikata as *minimal core + presets*, explicitly **not** the heavy
   all-in-one config it competes against. Absorbing PM flow drifts into
   exactly that quadrant and breaks top-level minimalism, Do-No-Harm, and
   composable-not-monolithic at once. And by
   [ADR 0028](../adr/0028-prioritize-core-concept-stabilization.md)
   (demand-driven): aikata itself runs on ROADMAP + versions + ADRs with
   no sprints and **no observed residual pain** — its own evidence test
   says "do not build."

## 4. The boundary is not "agile yes/no" — split flow into three layers

| Layer | Examples | Verdict |
|---|---|---|
| Knowledge stock | requirements, design, glossary, decisions, rules | aikata's home turf — already covered |
| Process **policy** | DoD/DoR, ceremony cadence, "where the board lives" | the **only** opening — `enable workflow scrum` (ADR 0026) documents the *convention*, zero new machinery |
| Process **state** | live backlog, sprint board, burndown | **out** — delegate to an external tracker |

Draw the line at "does it absorb churning operational state?", not at
"does it mention agile?".

## 5. The one allowed path (conditional)

If real demand appears: a single opt-in `workflow scrum` guide under
`docs/workflows/` (the existing [ADR 0026](../adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)
pattern — the same way `AGENTS.md` documents git policy). It documents
*conventions* (DoD/DoR, cadence) and **delegates the board itself to an
external tool by reference**. It manages no board and adds no machinery.

Conditions, in order:

1. **No speculative build.** Write a boundary ADR first ("this is a
   convention document, not a PM feature") to fix the line, in the spirit
   of ADR 0046.
2. **Wait for ADR 0028 evidence** — dogfooding-observed pain, not a
   hypothetical.
3. **Traceability stays frontmatter-only** — hand-authored links, never a
   requirement-tracking engine (SPEC §8: aikata does not read code).

## 6. Recommendation to teams that want full-lifecycle coverage

Do not try to do it all in aikata. Use **aikata (knowledge canon + AI
config derivation) × an external tracker (flow state)**, and let a
`workflow scrum` guide hold only the *connection convention* between the
two. This keeps every aikata principle intact while still supporting real
scrum practice.

## 7. In-family edge worth noting (and why it is still not needed)

Retro/review records map cleanly onto ADR 0045's `log` bucket
(append-only, irreversible) — durable learning, not flow state — so a
dedicated log artifact would not violate the philosophy. But
`docs/memory/feedback.md` + `docs/adr/` already absorb that content, so a
new slot is "nice but unnecessary." Keep it as a deliberate non-addition.

## 8. References

- [SPEC.md §2.2](../../SPEC.md) — Non-Goal: no project-management features;
  §1.3 differentiation; §8 no code reading.
- [ADR 0045](../adr/0045-documentation-value-model.md) — canonical /
  derived / log value model (the line-drawing basis).
- [ADR 0028](../adr/0028-prioritize-core-concept-stabilization.md) —
  demand-driven, not architecture-driven.
- [ADR 0026](../adr/0026-workflow-guides-as-opt-in-collaboration-docs.md) —
  workflow guides as opt-in collaboration docs (the allowed path).
- [ADR 0046](../adr/0046-structure-migration-assistant-boundary.md) — the
  boundary-ADR pattern to follow if this is ever operationalized.
- [`docs/layout.md`](../layout.md) — the prescriptive document homes.
