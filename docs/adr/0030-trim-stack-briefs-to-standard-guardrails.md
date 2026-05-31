---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0030 — Trim stack briefs to standard-aligned guardrails

- **Status**: Accepted
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: **amends [ADR 0029](./0029-stack-brief-layout-convention.md)**
  (D3); ADR 0003 (Do-No-Harm), ADR 0028 (prioritize core-concept
  stabilization),
  [v0.9.0 design note](../decisions/v0.9-core-concept-stabilization.md) §4.4;
  fully resolves Q-DESIGN-13 in
  [`open-questions.md`](../decisions/open-questions.md)

## Context

ADR 0029 split Q-DESIGN-13 ("how small should built-in stack briefs
become?") into an additive half — a code-free canonical layout convention,
shipped in v0.9.0 — and a **subtractive** half: removing rules that read as
a general best-practice pack rather than collaboration-critical guardrails
(design note §4.4 names committing every generated Dart file, requiring a
test for every reusable widget, and pinning detailed TypeScript compiler
preferences).

ADR 0029 D3 deferred that subtractive half to "a later v0.9.x increment
**with concrete evidence**," and that "evidence-led" wording is now public
in three places (ADR 0029 D3, the open-questions entry, CHANGELOG [0.9.0]).
This ADR executes the trim in v0.9.1 and must reconcile that wording
honestly rather than contradict it.

## Decision

### D1 — Amend ADR 0029 D3: the basis is an editorial test, not usage data

The trim is decided by a **standard-vs-preference test**, applied by the
maintainer, **not** by new dogfooding usage data:

- **Keep** a rule when it is an official-standard / widely-recommended
  guardrail that prevents a common AI mistake (for example "use
  `dart format`", "`strict: true`", "no `any` without justification",
  accessibility basics).
- **Cut or soften** a rule when it is (a) an over-strict *mandate* beyond
  the standard, or (b) a one-team **preference** with no official backing.

This test is **decidable from the rule itself**, so it does not need the
usage evidence ADR 0029 gated on. ADR 0029 D3 is amended accordingly: the
evidence requirement applies only to *genuinely ambiguous* guardrails whose
status the rule text cannot settle — those stay until evidence arrives.
This is why the items in D3-KEEP below are retained rather than cut.

### D2 — Cuts (this increment)

Applied to the Flutter and TypeScript briefs, en + ja in parallel:

- **Flutter §6 (codegen)** — the "generated files are committed; CI fails
  if the tree is dirty" *mandate* becomes a project-choice TODO (commit vs
  gitignore generated code is a real per-project decision). The
  "generated files live next to source" naming convention stays.
- **Flutter §7 (testing)** — the blanket coverage mandates ("a widget test
  for *every* reusable widget", "golden tests for *any* UI", "integration
  tests for flows") are softened to "test what is likely to regress; run
  `flutter test` before declaring work done." The verification command (a
  guardrail) stays.
- **Flutter §1** — drop the 80-column pin and the "do not hand-tune
  whitespace" editorializing; keep "use `dart format`, CI enforces it."
- **Flutter §2** — drop "prefer `late` over `?`" (a preference); keep
  "avoid `!` without a justifying comment" (a guardrail).
- **TypeScript §1 (tsconfig)** — keep `strict: true` (+ "disabling a
  strict-family flag needs an ADR"); move the granular stricter-than-
  default pins (`noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`)
  to a "recommended; tune per project" note rather than a hard rule.
- **TypeScript §6** — drop "prefer `readonly` for inputs" (a preference);
  keep "no `any`", "avoid `as`", "avoid `!`", and `import type`.

### D3-KEEP — standards that fail the cut test (condense, do not delete)

- **Flutter §9 (accessibility)** — a11y is an official standard, not a
  preference; deleting it from a public tool is also poor guidance.
  Condensed to a short line, retained.
- **TypeScript §8 (errors)** — `Error` subclassing and `cause` chaining
  are standard modern JS/TS practice, not a strong opinion. Retained,
  lightly condensed.

## Consequences

- The four stack briefs shrink toward guardrails + verification commands +
  project-choice TODOs; the six stacked golden trees are regenerated.
  `minimal` / `standard` output is unchanged (Do-No-Harm).
- Q-DESIGN-13 is **fully resolved** (direction by ADR 0029, subtractive
  half by this ADR). The open-questions entry closes pointing at both.
- The public "evidence-led" wording is reconciled, not contradicted: the
  amendment states the honest basis and narrows the evidence requirement
  to ambiguous cases, which is itself why a11y and error handling stay.

## Alternatives Considered

- **Wait for dogfooding evidence before trimming (literal ADR 0029 D3).**
  Rejected for the standard-vs-preference subset: that classification is
  decidable from the rule text, so waiting adds no information. Evidence is
  still required for ambiguous guardrails (kept here).
- **Also cut accessibility / error handling to make briefs maximally
  small.** Rejected: both are standards, not preferences; cutting them
  fails the project's own removal test and worsens the guidance.
- **Fold the trim into ADR 0029 instead of a new ADR.** Rejected: ADR 0029
  is Accepted and shipped in a signed release; amending via a new,
  cross-referenced ADR keeps the record honest and auditable.
