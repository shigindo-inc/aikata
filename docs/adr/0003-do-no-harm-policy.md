---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
audience: [human, agent]
---

# ADR 0003 — The Do-No-Harm Policy for Optional Features

- **Status**: Accepted
- **Date**: 2026-05-20
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (Canonical `AGENTS.md`);
  [`docs/origin/initial-design.md`](../origin/initial-design.md) §6

> **Note on numbering**: `docs/origin/initial-setup.md` §1.1 originally
> placed this ADR at `0001`. ADR 0001 was reserved for the meta-decision
> "use ADRs," and ADR 0002 was assigned to the canonical-`AGENTS.md`
> decision by the maintainer. This ADR was therefore renumbered to 0003.

## Context

aikata's positioning rests on being **opinionated but small**. Features
like Obsidian-compatible frontmatter, TDD scaffolding, Flutter presets,
and monorepo layout all sound nice in isolation, but each carries a tax:

- Obsidian's `[[wikilinks]]` are unreadable outside Obsidian.
- TDD rules in `AGENTS.md` push every project toward test-first even
  when that's wrong for the domain.
- A baked-in Flutter preset makes the core look stack-specific.
- A monorepo layout fragments `AGENTS.md` across nested directories.

Each tax silently penalizes users who do **not** opt in. The cumulative
effect of many "small" defaults is what made existing scaffolders feel
heavy.

## Decision

We adopt the **Do-No-Harm Policy** as a binding design constraint on
every feature that is not strictly required.

For each opt-in feature, the following must hold:

1. **Default-off.** The feature ships disabled. Enabling requires an
   explicit flag, preset selection, or `aikata add`.
2. **Zero residue when off.** With the feature disabled, no generated
   file mentions, references, or links to it. Grep should yield zero
   hits for its identifiers.
3. **Isolated when on.** When enabled, the feature's content lives in a
   dedicated location (e.g. `docs/stacks/<stack>.md`,
   `docs/testing.md`); top-level documents reference it conditionally,
   never unconditionally.
4. **Inert to outsiders.** Artifacts visible to non-adopting tools
   (Obsidian users opening a non-Obsidian project, Cursor users in a
   Claude-only repo) must remain readable and harmless.

### Concrete applications

| Feature | Rule |
|---|---|
| **Obsidian** | Standard markdown links only. No `[[wikilinks]]`. No `.obsidian/`. Dataview / Tasks queries, if ever included, live under `docs/obsidian-views/` and are gitignored by default. Frontmatter uses keys both Obsidian Properties and plain readers accept. |
| **TDD** | `docs/testing.md` only when `--with-tdd`. AGENTS.md references it inside a conditional block (`## Testing (if testing.md exists)`); no "tests first" rule leaks otherwise. |
| **Flutter / stack presets** | `--preset minimal` ships zero stack content. Stack rules live in `docs/stacks/<stack>.md`. Core code holds **no** Flutter knowledge. |
| **Monorepo** | Single-project layout by default. `--monorepo` introduces `apps/*/AGENTS.md`; the root structure does not change for non-monorepo users. |
| **AI tools** | `--ai-tools` defaults to `claude`. Cursor / Codex / etc. dirs are not created unless explicitly enabled. |
| **Language** | `--lang en` by default (maximum reach). `--lang ja` produces a parallel template set; the user is never forced into bilingual maintenance. |

### Consequences of the policy on aikata itself

- aikata's own repository **commits** generated artifacts (e.g.
  `CLAUDE.md` once `aikata generate` exists). This is the inverse of
  the user-project default and is intentional: a contributor cloning
  aikata gets a working Claude Code / Cursor experience without first
  installing aikata. See
  [ARCHITECTURE.md §6](../../ARCHITECTURE.md#6-distribution--generated-artifacts).

## Consequences

**Positive**:

- Reduces the long-term blast radius of every new feature.
- Lets aikata grow without becoming heavier from the user's perspective.
- Aligns with the seven design principles in
  [SPEC.md §3](../../SPEC.md#3-design-principles), especially Principle
  3 (Do No Harm) and Principle 6 (Stack-agnostic core).

**Negative**:

- Each PR that adds a feature must explicitly demonstrate compliance.
  We accept this overhead — it's the entire point.
- Some users may want stronger defaults (e.g. "ship TDD on by default
  for me"). The answer is a personal preset, not a change to the
  shipped defaults.

## Verification

- `aikata doctor` (v0.2+) checks that opt-out flags do not produce
  unreferenced files.
- Golden tests for `--preset minimal` must show **zero** mention of
  optional-feature identifiers.
