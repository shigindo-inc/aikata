---
name: track-context
user-invocable: false
description: Use when beginning non-trivial work (a feature, refactor, investigation, or multi-step change) in a repository that aikata manages — recognizable by an `AGENTS.md` at the root and/or a `.aikata/aikata.yaml`. Teaches the daily context-maintenance loop — which canonical documents to read before editing, where newly-learned information belongs (AGENTS.md, SPEC.md, docs/adr/, docs/memory/, docs/tasks/current.md), how to keep working state current, and what to check before handoff. Do not use for trivial one-line edits, for questions unrelated to the project, in repositories without aikata markers, or for raw CLI invocation (use `manage-docs` for init, generate, doctor, sync, or new).
---

# track-context

aikata-managed repositories keep their shared human + AI context in a
small, fixed set of canonical markdown documents. This skill is the
operating loop for working **inside** such a repository: read the right
documents before editing, record what you learn in the right slot, keep
working state current, and verify the context is consistent before you
hand off.

You are recognizably in an aikata repository when the root has an
`AGENTS.md` (canonical operating rules) and usually a `.aikata/aikata.yaml`
manifest. If neither is present, this skill does not apply.

This skill does not invoke the CLI directly — when the loop needs
`doctor`, `sync`, `generate`, or `new adr`, hand off to the
**`manage-docs`** skill.

## 1. Read before you edit

Before changing code or docs for a non-trivial task, read the canonical
documents that bound the work — do not guess project rules:

1. `AGENTS.md` — invariant operating rules (read first; it links the rest).
2. `SPEC.md` — what the project is and why (requirements, scope).
3. `ARCHITECTURE.md` — how it is built (layout, patterns, constraints).
4. `GLOSSARY.md` — project-specific terminology.
5. The **relevant** `docs/adr/NNNN-*.md` and `docs/memory/` entries for
   the area you are touching — not all of them; the ones the task names
   or that govern the files you will change.

Stop and surface a conflict if the task contradicts a Hard Rule, an
Accepted ADR, or a recorded preference, rather than silently choosing.

## 2. Keep working state current

If `docs/tasks/current.md` exists, treat it as the shared working-memory
slot and update it:

- **At task start** — record what you are about to do and why.
- **At meaningful progress** — decisions made, blockers hit, what is
  done vs. pending.
- **At completion** — final state and any handoff notes.

`docs/tasks/current.md` is ephemeral and rewriteable; it is not a
substitute for the durable slots below. If the file does not exist, do
not create it unless the user asks (it is an opt-in standard-scope file).

## 3. Put new information in the right slot

When you learn something durable, classify it before writing it down —
each slot has a distinct lifetime and audience:

| What you learned | Slot | Notes |
| --- | --- | --- |
| An invariant rule everyone must follow | `AGENTS.md` | Edit canonical source; never edit generated `CLAUDE.md`. |
| A requirement (what/why the product does X) | `SPEC.md` | Scope and intent, not implementation. |
| A design decision with trade-offs | `docs/adr/` | Stamp via `aikata new adr`; accepted ADR bodies are immutable. |
| A durable fact or preference | `docs/memory/` | Dated, append/supersede; never delete. |
| In-flight state for this task | `docs/tasks/current.md` | Short-lived working state; not a backlog or archive. |

Rules > memory > working state when they conflict. A one-off
implementation detail that the code already expresses belongs in neither
— do not record what the repository already records.

These slots already sort content by the *source* of its truth — canonical
(the sole source), derived (regenerable), and append-only log (memory's
dated entries) — which is why durable rationale is superseded in place,
never summarized away (see `docs/adr/0039-*` and `docs/adr/0045-*`).

## 4. Check before you declare done

Before claiming the work is complete, confirm:

- **Documentation impact** — did an invariant, requirement, or design
  decision change? If so, is it recorded in the correct slot above, and
  were generated artifacts regenerated (hand off to `manage-docs` for
  `aikata generate`)?
- **Verification results** — tests/lint/build run and reported honestly
  (failures stated, skips named). Run `aikata doctor` via `manage-docs`
  to confirm documentation consistency.
- **Unresolved questions** — anything still open captured in
  `docs/decisions/open-questions.md` (or surfaced to the user), not left
  implicit.
- **Handoff state** — `docs/tasks/current.md` reflects the final state so
  the next human or agent can resume without re-deriving context.

## When to hand off to `manage-docs`

Use the `manage-docs` skill for the actual command invocations this loop
calls for:

- `aikata doctor` / `aikata doctor --json` — documentation self-check.
- `aikata sync` — pull newer upstream template content safely.
- `aikata fill` — write any **missing** canonical document into the repo
  without overwriting (adopt an existing repo, or restore a deleted doc).
- `aikata generate` — regenerate `CLAUDE.md` / `.cursor/rules/main.mdc`
  after editing `AGENTS.md` or another canonical doc.
- `aikata new adr "<title>"` — stamp an auto-numbered ADR for a design
  decision.

## When to hand off to `model-feature`

If the work about to start changes **externally observable behaviour**
(a new or altered feature, not a refactor or bug fix), use the
`model-feature` skill first: it writes the use case into
`docs/usecases.md`, propagates the required data into `docs/domain.md`,
and hands back just before implementation. Skip it for refactors, bug
fixes, copy changes, and UI position adjustments.

## Reference

- Canonical docs: `AGENTS.md`, `SPEC.md`, `ARCHITECTURE.md`, `GLOSSARY.md`.
- Decisions: `docs/adr/NNNN-*.md`; open items in
  `docs/decisions/open-questions.md`.
- Working state: `docs/tasks/current.md` (when present).
- Sibling skill: `manage-docs` (the raw CLI surface).
- Sibling skill: `model-feature` (per-feature design loop into `docs/usecases.md` + `docs/domain.md`).
