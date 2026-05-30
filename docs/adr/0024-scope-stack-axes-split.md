---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-30
audience: [human, agent]
---

# ADR 0024 - Split the `--preset` enum into orthogonal `--scope` and `--stack` axes

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: aikata maintainers
- **Related**: ADR 0016 (aikata.yaml schema v2), ADR 0017 (post-init
  command taxonomy), ADR 0022 (v0.8.x security & governance hardening),
  supersedes the leading position of Q-DESIGN-04 in
  [`open-questions.md`](../decisions/open-questions.md)

## Context

`aikata init` exposes a single `--preset` flag whose value set mixes two
conceptually independent axes:

```
--preset = minimal | standard | flutter | typescript
```

- `minimal` / `standard` (and the reserved `extended`) describe a
  **scope** — how much documentation the scaffold emits.
- `flutter` / `typescript` describe a **stack** — the target technology,
  which in practice also pins the scope to a `standard`-equivalent body.

A third axis, document language, is **already** a separate flag
(`--lang en|ja`). So the surface is asymmetric: language is orthogonal,
but stack is fused into the scope enum.

The fusion collapses a `scope × stack` product into one dimension and
makes most combinations unreachable:

| Intent | Today |
| --- | --- |
| `minimal` Flutter project | unreachable (`flutter` implies standard) |
| `standard` TypeScript | only via `--preset typescript` (scope fixed) |
| Flutter **and** TypeScript (monorepo) | unreachable (presets are exclusive) |
| stack-agnostic `minimal` | `--preset minimal` |

The data model already disagrees with the CLI surface. `aikata.yaml`
stores `stacks` as a **list** (`[]string`, ADR 0016 schema v2), and
`aikata enable stack <name>` (ADR 0017) exists as a post-init escape
hatch precisely because `init` cannot treat stack as a first-class,
multi-valued axis. Internally, `init.go`'s `stacksForPreset(preset)`
re-derives a stack *from* a scope-shaped enum value — inverting the
intended dependency direction.

Q-DESIGN-04 left this open with a leading answer of "presets compose by
set-union of feature flags." That framing keeps `preset` as the
composition mechanism. This ADR takes the opposite, simpler route: stop
overloading `preset`, and make scope and stack independent inputs.

## Decision

Split the fused `--preset` enum into two orthogonal flags on
`aikata init`, alongside the existing `--lang`:

```
aikata init <name> \
  --scope  minimal | standard            # extended reserved (ADR 0017)
  --stack  flutter | typescript | ...     # repeatable; empty = stack-agnostic
  --lang   en | ja
```

Axis definitions:

- **`--scope`** — documentation breadth. Single-valued. Default
  `standard`. `extended` stays **reserved** (no behaviour change here;
  it remains a v1.0 item per ADR 0017).
- **`--stack`** — target technology. **Multi-valued in syntax**:
  repeatable (`--stack flutter --stack typescript`) and/or
  comma-separated (`--stack flutter,typescript`). Empty means
  stack-agnostic. Persisted to `aikata.yaml`'s existing `stacks` list —
  no schema bump. The multi-valued *shape* is locked now even though
  v0.8.2 accepts only the single-stack combinations that have a template
  tree (see Scope boundary), precisely so v1.0 never has to reshape the
  flag surface again.
- **`--lang`** — unchanged.

### Scope boundary — surface is orthogonal, deliverable combinations are bounded

This ADR was first written assuming the split also made the full
`scope × stack` product *buildable*. Inspection of the template tree
disproves that premise and the ADR is corrected here. The four preset
directories (`presets/{minimal,standard,flutter,typescript}/`) are
**independently authored full trees**, not a `standard` base plus a
stack overlay: `presets/standard/` never references `.Stacks`, while
`presets/flutter/` and `presets/typescript/` carry stack-specific prose
(e.g. "implement a new widget / screen" vs "module / endpoint") and a
divergent markdown-numbering convention inline in `AGENTS.md`,
`README.md`, `ARCHITECTURE.md`, etc. There is no template tree for
`minimal` + a stack, nor for multiple stacks.

v0.8.2 therefore delivers the orthogonal **flag surface** and the
**deprecation window** only. `(scope, stacks)` resolves to exactly the
four trees that exist today:

| Input | Template tree |
| --- | --- |
| `--scope minimal` (no stack) | `presets/minimal/` |
| `--scope standard` (no stack) | `presets/standard/` |
| `--scope standard --stack flutter` | `presets/flutter/` |
| `--scope standard --stack typescript` | `presets/typescript/` |

Every other combination — `minimal` with any stack, two or more stacks,
or `extended` — is an **explicit error** (e.g. *"scope `minimal` with a
stack is not yet supported; use `--scope standard`"*), never a silent
fallback to a minimal tree with an orphaned `docs/stacks/<x>.md` that no
`AGENTS.md` links. A half-wired combination is worse than a clear error.

Unlocking new combinations requires decomposing the stack-flavored
trees into a scope base plus stack partials and re-deriving the existing
flutter/typescript output without drift — a template-refactor **feature**
with golden-test rework, explicitly **deferred** as follow-up (see
Consequences). It is out of scope for a patch shipping alongside the
v0.8.3 data-loss fix. What this ADR still buys, in full: `stack` becomes
a first-class axis like `--lang` instead of a value fused into a scope
enum (the original motivation), and `--preset` enters its deprecation
window so v1.0 can drop it cleanly.

### `--preset` becomes a deprecated alias

`--preset` is **not removed** in v0.x. It maps deterministically:

| Deprecated input | Equivalent |
| --- | --- |
| `--preset minimal` | `--scope minimal` |
| `--preset standard` | `--scope standard` |
| `--preset flutter` | `--scope standard --stack flutter` |
| `--preset typescript` | `--scope standard --stack typescript` |

Rules:

- Using `--preset` prints a one-line deprecation notice to stderr
  pointing at `--scope` / `--stack`, then proceeds.
- `--preset` combined with `--scope` or `--stack` in the same
  invocation is an **error** (no silent precedence), so the alias and
  the new axes never have to be reconciled.
- The interactive prompt asks **scope**, then **stack** (multi-select,
  empty allowed), then **lang** — it never asks for "preset".
- Removal of the alias is deferred to **v1.0**, where breaking changes
  are permitted. This is the reason to ship the split *before* v1.0:
  introducing the new axes and the deprecation in the same release that
  removes the alias would leave no migration window.

### Placement: v0.8.2, a distinct "CLI surface" line

The work ships as **v0.8.2**. ADR 0022 scoped v0.8.x to *security &
governance hardening of aikata's own repository* with "no binary or
template change," so this user-facing CLI/template change does not
belong to that charter. It is interleaved into the 0.8.x **number
space** as its own line rather than renumbering downstream versions:
shifting the v0.9.x channel-publication line to v0.10.x would force
rewriting CHANGELOG's already-published forward references
("deferred to v0.9.x") — i.e. rewriting history — for no benefit. The
ROADMAP precedent is v0.7.4, a distinct "cleanup tail" living inside the
0.7.x space without sharing its core theme. v0.8.2 is labelled a
"pre-v1.0 stable-surface correction," not a security patch.

## Consequences

- `scope`, `stack`, and `lang` become three orthogonal, independently
  selectable **flags**. The deliverable combinations stay bounded by the
  template trees that exist today (see Scope boundary): `minimal`,
  `standard`, `standard + flutter`, `standard + typescript`. New
  combinations (`minimal` + a stack, multi-stack monorepos) are **not**
  unlocked by this release — they need a template refactor (scope base +
  stack partials, re-deriving flutter/typescript output without drift)
  that is **deferred** as a separate follow-up feature. v0.8.2 ships the
  final flag *shape* and the `--preset` deprecation window; it does not
  add buildable combinations.
- No `aikata.yaml` schema bump: `stacks` already exists as a list.
  `init` now writes it directly from `--stack` instead of via
  `stacksForPreset`.
- `stacksForPreset` is removed; the manifest seed stops re-deriving a
  stack from a scope-shaped value. The manifest's `Preset` field is
  reinterpreted as **scope** (default `standard`); whether to rename the
  field or keep it as a back-compat alias is an implementation detail
  for the v0.8.2 PR, recorded here as a tracked decision, not a schema
  migration for users.
- **GLOSSARY.md migration item** (do not edit in this planning pass — it
  describes shipped behaviour until the code lands): the `preset` entry
  states "stack-specific presets such as `flutter`," which becomes false
  once `flutter` is a stack. v0.8.2 must update the `preset` entry to
  describe it as a **deprecated alias for `--scope`**, add a `scope`
  entry, and align the `stack` / `stack-agnostic core` entries. The
  word "preset" survives only as a scope-synonym alias. `aikata doctor`
  runs a glossary-consistency check, so this is normative, not
  cosmetic.
- Other in-tree forward references to the `--preset minimal|standard|
  flutter|typescript` surface (README, SPEC §5.x, ARCHITECTURE config
  examples, `docs/` access layer) are updated in the v0.8.2 PR. Shipped
  ROADMAP/CHANGELOG history keeps its text; only live help and reference
  docs change.
- Q-DESIGN-04 is **superseded**: presets are no longer the composition
  mechanism, so "compose by set-union of flags" is moot. The entry is
  closed pointing at this ADR.

## Alternatives considered

- **Keep `preset`, make presets compose (Q-DESIGN-04 leading).**
  Rejected: composition rules (set-union + last-write-wins overrides)
  are more machinery than the problem needs, and still leave stack and
  scope entangled inside opaque bundle names. Two plain orthogonal flags
  express the same space with no composition algebra.
- **Remove `--preset` outright in v0.8.2.** Rejected: breaking change
  before v1.0 violates the ROADMAP's "never break Phase N users" rule.
  The alias costs little and is removed cleanly at v1.0.
- **Defer the whole split to v1.0.** Rejected: introducing the axes and
  removing the alias in one release leaves no deprecation window. The
  split must land first so v1.0 can drop the alias against an already-
  warned surface.
- **Renumber v0.9.x → v0.10.x and give the split a dedicated minor.**
  Rejected: forces rewriting published CHANGELOG forward references; the
  0.8.2-as-distinct-line option achieves the same thematic separation
  with no history rewrite (see Decision).
