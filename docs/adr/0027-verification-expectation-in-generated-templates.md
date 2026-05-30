---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0027 — Verification expectation in generated templates

- **Status**: **Accepted** — the maintainer approved the final rule
  wording (D1) and the opt-in recommendation prose (D3) when approving the
  v0.8.5 implementation plan on 2026-05-31. The human-in-the-loop gate is
  satisfied.
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0003 (Do-No-Harm
  Policy), ADR 0026 (workflow guides). This ADR resolved and replaced the
  former Q-DESIGN-11 in
  [`docs/decisions/open-questions.md`](../decisions/open-questions.md).

## Context

A dogfooding design discussion asked whether aikata's generated files
should take more of a TDD stance. The investigation found that the
question conflates three distinct things:

1. **Test existence** — "tests should exist for new code". This is
   **already a default**: the standard `AGENTS.md` template ships the
   hard rule *"Add tests for new code unless the change is
   documentation-only"*
   ([`testdata/golden/standard/AGENTS.md`](../../testdata/golden/standard/AGENTS.md)).
   Generated files are **not** silent on testing.
2. **Test ordering** — test-first / TDD red-green methodology. This is
   deliberately **opt-in** via `--with-tdd` / `enable tdd`, per
   [ADR 0003](./0003-do-no-harm-policy.md). Forcing test-first across all
   domains is genuinely wrong for some (exploratory, prototype, research
   code), so this stays opt-in.
3. **Verification** — "before claiming a change complete, run the
   project's checks and show the output". This is methodology-neutral and
   is the **actual gap**.

The concrete evidence for the gap is an asymmetry: aikata's **own**
`AGENTS.md` carries a verification gate (Hard Rule 7,
`make test && make lint`), but the **generated** standard template does
not project any equivalent. AI agents are prone to claiming "done"
without running anything; a verification gate is the cheapest objective
check, and its value rises specifically under AI collaboration — the
historical human cost of writing/running tests collapses for agents while
the benefit of a fast red-green feedback loop for autonomous iteration
rises.

## Decision

### D1 — Add a methodology-neutral verification rule to the standard template

The standard `AGENTS.md` template gains a conditional hard rule. Proposed
wording (final wording is the maintainer-approval gate):

> **Verify before claiming done.** Before claiming a change is complete,
> run the project's tests / build **if they exist** and show the output.

This serves AI collaboration without imposing any methodology choice.

### D2 — Test-first stays opt-in

Reaffirms [ADR 0003](./0003-do-no-harm-policy.md). No test-first /
red-green wording leaks into the default templates. The opt-in
`--with-tdd` / `enable tdd` surface is unchanged.

### D3 — TDD *recommendation* lives in opt-in docs, never in default `AGENTS.md`

Any recommendation to adopt test-first belongs in the opt-in
`docs/testing.md` template (currently a bare TODO skeleton in
`internal/templates/data/components/tdd/{en,ja}/tdd.md.tmpl`) — or, later,
a `docs/workflows/` guide ([ADR 0026](./0026-workflow-guides-as-opt-in-collaboration-docs.md)).
The recommendation may lean harder for AI collaboration than it would for
humans, for the cost/benefit reason in Context. It must never appear in
the canonical default `AGENTS.md`.

### D4 — Do-No-Harm: the rule must degrade gracefully

The "if they exist" condition in D1 is load-bearing. A project with zero
tests yet must not be handed an empty obligation, or this becomes a
Do-No-Harm violation in new clothes. The `minimal` scope must continue to
show **zero** optional-feature residue; the verification rule must read as
inert (not failing, not nagging) in a test-less project.

## Consequences

### Positive

- Closes the asymmetry: the verification discipline aikata practices on
  itself becomes available to every project it scaffolds.
- Captures the real win of AI collaboration (objective completion gate)
  without the harm of imposing test-first.
- Keeps the canonical `AGENTS.md` methodology-neutral and compact.

### Negative

- Changes the **default** generated output, so it ships to every project,
  not just opt-in users. This is why D1 is gated on maintainer approval.
- Adds one more hard rule to the standard template; must be worded tightly
  to avoid bloat.

## Alternatives Considered

- **Put test-first into the default `AGENTS.md`.** Rejected: violates
  ADR 0003; wrong for many domains.
- **Stay silent (status quo).** Rejected: the asymmetry leaves agents
  with no completion gate in generated projects, which is exactly where
  the "claims done without running anything" failure mode bites.
- **Only strengthen `docs/testing.md`.** Insufficient on its own: it
  reaches opt-in users only, while the verification gap exists for **all**
  generated projects. Adopted as D3 in addition to D1, not instead of it.

## Implementation plan (for out-of-session execution)

Two PRs, in this order:

- **PR-A — `docs/testing.md` strengthening (D3).** Opt-in only.
  Strengthen `internal/templates/data/components/tdd/{en,ja}/tdd.md.tmpl`
  from the bare TODO skeleton with a short rationale ("why testing matters
  for AI collaboration") and an optional, clearly-marked TDD
  recommendation. Regenerate affected golden fixtures. HITL: **light** —
  human reviews the shipped prose before merge. Not blocked on this ADR's
  acceptance.
- **PR-B — verification rule (D1).** Blocked on this ADR being
  `Accepted` and on the final rule wording being approved. HITL:
  **required**.
  - **Insertion point**: the `## 4. Hard rules` list of the standard
    preset template — `internal/templates/data/presets/standard/en/AGENTS.md.tmpl`
    (the existing `- **Add tests** for new code …` line, currently line
    54) **and** its `ja` sibling
    `internal/templates/data/presets/standard/ja/AGENTS.md.tmpl`. Both
    language variants must change together.
  - **Cross-preset scope (decide in PR-B)**: the `flutter` and
    `typescript` preset templates **already carry a stack-specific
    verification line** (`flutter/en/AGENTS.md.tmpl:56` "Run `flutter
    test` before declaring work complete"; `typescript/en/AGENTS.md.tmpl:55-56`
    "Run the test suite"). PR-B must decide whether the generic standard
    rule is *also* added there (risking duplication) or whether those
    presets are left as-is because they already satisfy the intent. Lean:
    leave the stack presets alone; they already have a stronger, more
    specific gate.
  - **`minimal` scope (decide in PR-B)**: `minimal/*/AGENTS.md.tmpl` has
    a deliberately lean hard-rules list. Decide whether the methodology-
    neutral rule belongs there too or whether minimal stays minimal. The
    D4 zero-residue assertion is about opt-in TDD identifiers, not about
    this default rule — keep the two concerns separate.
  - Then regenerate golden fixtures (`testdata/golden/standard/AGENTS.md`
    et al.) and run the standard gates.

Both PRs follow the standard gates: a `CHANGELOG.md` `[Unreleased]` entry,
`make test && make lint`, `aikata doctor` clean, and English commit / PR
text ([ADR 0006](./0006-locale-and-japanese-documentation-policy.md)).
Tracked in [ROADMAP.md](../../ROADMAP.md) under v0.8.5.

## Verification

- Golden tests for the standard / `--with-tdd` outputs reflect the new
  wording.
- A `minimal`-scope golden assertion shows zero verification-rule and zero
  TDD-recommendation residue.
- `aikata doctor` stays clean (frontmatter + links) on the strengthened
  `docs/testing.md`.
