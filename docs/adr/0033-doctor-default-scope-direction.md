---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
audience: [human, agent]
---

# ADR 0033 - Direction for `doctor`'s default validation scope

- **Status**: Accepted
- **Date**: 2026-06-01
- **Deciders**: aikata maintainers
- **Related**: ADR 0021 (doctor scope & exclusion — added the
  `doctor.exclude` escape hatch), ADR 0028 (prioritize core-concept
  stabilization), ADR 0014 (manifest is a living record), ADR 0019
  (`sync` missing-file repair semantics)

## Context

[ADR 0021](./0021-doctor-scope-and-exclusion.md) made `aikata doctor`'s
walk *configurable* by adding an additive `doctor.exclude:` glob list:
matching paths skip `checkFrontmatter` / `checkUpdated` / `checkGlossary`.
That closed the immediate problem (a Claude Code plugin tree whose
`SKILL.md` carries Anthropic's `name` + `description` frontmatter tripped
62 spurious errors), but it left the **default** posture unchanged:
`doctor` still validates almost every Markdown file in the repository and
expects the *user* to subtract the files aikata does not own.

The v0.9.0 core-concept stabilization review
([ADR 0028](./0028-prioritize-core-concept-stabilization.md)) surfaced
this as a coherence question (tracked as Q-DOCTOR-02). aikata's value is
reducing the human cost of maintaining the **shared-context document
surface it manages** — `AGENTS.md`, the canonical doc set, ADRs, memory,
working-state. Validating arbitrary third-party Markdown by default
inverts that: it makes aikata responsible for documents it did not
generate and cannot reason about, and it pushes recurring exclusion churn
onto every adopter with a plugin tree, a vendored doc set, or a
third-party contract file.

The opposing risk is just as real. Flipping the default scope on a
pre-v1.0 tool is not free:

- **Adopted and pre-manifest projects have no manifest** to enumerate
  "aikata-managed" files (ADR 0014's manifest is seeded at init or via
  `--rebaseline`). A naive "validate only what the manifest lists"
  default would *silently stop checking* real project docs in exactly the
  projects that most need the safety net — a blind spot worse than noise.
- **`aikata doctor --strict` is a binding CI gate from v0.5 onward.**
  Narrowing what it inspects changes the meaning of a green run for every
  existing user without warning.

So the question is genuinely two-sided, and the safe answer separates
*direction* from *behavior change*.

## Decision

### D1 - Direction: managed-surface default, explicit broad audit mode

The intended end state is that `doctor`'s **default** walk validates
primarily the document surface aikata manages, with an **explicit broader
audit mode** (a flag and/or a `doctor.scope:` config value) for users who
want every Markdown file checked. `doctor.exclude` (ADR 0021) is retained
as an escape hatch where it remains useful even under the narrower
default. This ADR records that direction so v0.9.0's core-concept
stabilization line can close with the question resolved.

### D2 - Defer the behavior change to its own scoped step

This increment changes **no `doctor` behavior**. The default scope flip
ships later as an isolated change with **before/after coverage proof**
(golden / fixture evidence that adopted and pre-manifest projects still
receive useful validation, and that no currently-checked managed file
silently drops out). Bundling a strict-CI-affecting scope change into the
stabilization-tail planning pass would be exactly the kind of unproven,
hard-to-reverse move the v0.9.0 review warned against.

### D3 - Preconditions the future flip must satisfy

Before the behavior changes, the implementing change must establish:

- A definition of "aikata-managed" that **does not depend solely on the
  manifest**, so adopted and pre-manifest projects keep coherent
  validation (e.g. the canonical doc set by name plus `docs/adr/`,
  `docs/memory/`, and known managed paths, unioned with manifest entries
  when present).
- An **explicit, discoverable opt-in to the broad audit** so existing
  `--strict` CI users are never silently narrowed — the broad mode must
  be a documented one-liner.
- No regression to the ADR 0021 exclusion semantics or the glossary /
  frontmatter / updated checks themselves.

## Consequences

- v0.9.0's "doctor scope follow-up ADR" item is satisfied: the direction
  is decided and recorded, and Q-DOCTOR-02 moves to *Resolved (direction
  only; behavior deferred)*.
- No user-visible behavior changes now; existing `doctor` and
  `doctor --strict` runs are byte-for-byte unaffected.
- A later, narrowly-scoped change owns the actual flip and carries the
  burden of proving adopted / pre-manifest projects keep useful coverage.
- The exclusion knob from ADR 0021 stays valid under either posture, so
  users who configured it lose nothing.

## Alternatives Considered

- **Flip the default scope now.** Rejected: it would change the meaning of
  the binding `--strict` CI gate and risk silent blind spots in adopted /
  pre-manifest projects without coverage proof — the precise failure mode
  the v0.9.0 review flagged.
- **Leave Q-DOCTOR-02 fully open.** Rejected: the direction is clear
  enough to decide, and leaving it open blocks closing the v0.9.0
  stabilization line and invites repeated re-litigation.
- **Manifest-only default ("validate what `sync` manages").** Rejected as
  the *sole* definition: it strands adopted and pre-manifest projects with
  no managed set, which is why D3 requires a name-based union rather than
  a manifest-derived set alone.
- **Reverse-include knob only (`frontmatter_required_paths`).** This is
  the symmetric Q-DOCTOR-01 follow-up already deferred by ADR 0021 until a
  real user requests it; it solves a narrower problem than the default
  posture and is not a substitute for this direction.
