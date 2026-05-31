---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0029 — Stack briefs carry a code-free canonical layout convention

- **Status**: Accepted
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0006 (locale and
  Japanese documentation policy), ADR 0017 (post-init command taxonomy),
  ADR 0024 (`scope` / `stack` axes), ADR 0028 (prioritize core-concept
  stabilization), [v0.9.0 design note](../decisions/v0.9-core-concept-stabilization.md);
  partially resolves Q-DESIGN-13 in
  [`open-questions.md`](../decisions/open-questions.md)

## Context

`aikata init --stack flutter|typescript` scaffolds a stack brief at
`docs/stacks/<stack>.md` — additive collaboration guidance layered on top
of `AGENTS.md`. Today the brief is documentation only: it states style,
testing, and architecture conventions and never emits source code.

Dogfooding raised a concrete pain. When a developer starts a stacked
project, the recurring friction is not "the brief is missing a rule" but
"I do not know where things go": where design tokens, the theme, and
non-UI constants should live. The blank-page hesitation is shared by the
human and the AI agent, and it produces a recognizable failure mode —
magic numbers and ad-hoc literals scatter across widgets / modules
because no canonical home was declared.

A tempting fix is to have `aikata init --stack <x>` **generate** the
skeleton: create `lib/theme/`, a `tokens` file, a `constants` directory,
and so on. This was considered and rejected. It crosses aikata's
identity line:

- aikata's unit of truth is **documents**, not code (SPEC §1.2 / §1.3).
  Generating source placeholders moves aikata into the code-scaffolder
  category (Mason, very_good_cli, create-next-app, Nx generators) where
  it has no differentiation.
- ADR 0028 D3 fixed, at the start of the v0.9.0 stabilization line, that
  stack support "does not become … a framework generator."
- Do-No-Harm (ADR 0003): a guessed `theme` / `tokens` / `constants` file
  that does not match the project's actual design-system and
  state-management choices **rots** — it imposes deletion / adaptation
  cost on a user who never asked for it. An empty document does not; a
  wrong code file does.

The remaining pain is therefore one of **placement knowledge**, not of
file existence. That is squarely a documentation problem, and the
v0.9.0 design note §8 requires the resolution to land in a focused ADR
before the template change.

## Decision

### D1 — Stack briefs include a "canonical layout convention" section

Each built-in stack brief gains a section that names where the stack's
foundational artifacts should live and the recurring AI-collaboration
failure mode that the convention prevents. For Flutter this covers the
theme, design tokens, and non-UI constants; for TypeScript the
constants, the design-token / theme location for frontend projects, and
module boundaries.

The section is written as **guidance, not a mandate**: a recommended
location, the failure mode it avoids, and a project-choice TODO for the
parts that legitimately vary (for example feature-first vs layer-first,
or frontend vs backend TypeScript). It must not harden into prescriptive
rules — that would reintroduce the best-practice bloat the stabilization
line is trying to reduce.

### D2 — aikata does not generate stack code; the agent scaffolds on demand

aikata emits no source files, directories, or placeholder code for a
stack. The canonical layout convention is shared **as a document** that
both the human and the AI agent read; the agent creates the actual
skeleton on demand, in the right place, following the convention. aikata
supplies the judgment (where, why, what fails), not the bytes. This
keeps aikata document-centered and Do-No-Harm-compliant.

### D3 — Briefs stay small; subtractive trimming is a separate, evidence-led step

The standing shape of a built-in brief is: high-value AI guardrails,
minimum verification commands, project-choice TODOs, and links to ADRs
for choices that vary by project. The canonical layout convention is an
**additive** application of that shape.

Reducing existing briefs by removing policy that is only valid for some
teams (for example committing every generated file, requiring a test for
every reusable widget, or pinning detailed compiler preferences) is the
**subtractive** half of Q-DESIGN-13. Which rules are collaboration-
critical guardrails versus team-specific opinions is a subjective call;
per the design note §2 ("if the answer is unclear, defer the feature
until dogfooding evidence sharpens it") it is **not** bundled into the
additive change, and not shipped inside a signed release on speculation.
It is deferred to a later v0.9.x increment with concrete evidence.

## Consequences

- The Flutter and TypeScript briefs (en + ja) gain a canonical layout
  convention section. Golden trees for the stacked presets are
  regenerated; `minimal` / `standard` outputs are unchanged (Do-No-Harm).
- Q-DESIGN-13 is **partially resolved**: its direction (code-free,
  layout-convention-bearing, small briefs) is settled here; its open
  part (the specific subtractive trims) stays tracked for a later
  increment.
- aikata's product boundary is reaffirmed in a citable place: stacks are
  documentation, the agent is the code author. Future "just generate the
  scaffold" requests have an ADR to point at.
- Bilingual parity (ADR 0006) is preserved: the ja briefs receive the
  same section structure, kept consistent for the `doctor`
  lang-consistency check.

## Alternatives Considered

- **Generate the stack skeleton at init time (the code-placeholder
  idea).** Rejected: crosses the document-centered identity (SPEC
  §1.2/§1.3), reverses ADR 0028 D3 the day it was accepted, and violates
  Do-No-Harm because a guessed code file rots when it does not match the
  project's real choices. The pain it targets is placement knowledge,
  which a document solves without generating code.
- **Only update the Q-DESIGN-13 leading position, no ADR.** Rejected:
  the design note §8 requires each stabilization answer to land in a
  focused ADR before the code change, precisely so the product-boundary
  reasoning is durable rather than buried in a template diff.
- **Trim the existing briefs in the same change.** Rejected for this
  increment: subjective subtraction does not belong in the same signed
  release as the additive convention; deferred to an evidence-led
  follow-up (D3).
