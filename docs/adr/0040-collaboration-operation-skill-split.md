---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-02
audience: [human, agent]
---

# ADR 0040 - Collaboration-operation skill split (aikata-cli + aikata-context)

- **Status**: Accepted
- **Date**: 2026-06-02
- **Deciders**: aikata maintainers
- **Related**: ADR 0015 (first-party skill & plugin distribution — this
  ADR refines its thin CLI-wrapper scope; it does **not** supersede it),
  ADR 0017 (post-init command taxonomy), ADR 0003 (Do-No-Harm Policy),
  ADR 0036 (Codex native distribution), ADR 0002 (`AGENTS.md` is
  canonical).

## Context

Through v0.9.x, aikata shipped a single first-party skill (`aikata`) in
four byte-tracked copies (universal, Codex plugin, Claude Code plugin,
Claude Code standalone). That skill teaches an agent *when and how to
invoke the aikata CLI* — `init`, `generate`, `doctor`, `sync`, `enable`,
`new` — and how to parse `doctor --json`.

The product thesis (ADR 0002, ADR 0028) is that a coherent set of
canonical Markdown lets humans and AI agents share project context
cheaply. But the skill surface only covered *bootstrapping and
maintaining* that surface through the CLI. It did not cover *operating*
an aikata-managed repository during ordinary development: which canonical
documents to read before editing, where newly-learned context belongs,
how to keep working state current, and what to verify before handoff.
That is the daily human + AI collaboration loop the product claims to
reduce the cost of — and it was undocumented as a skill.

ADR 0015 deliberately scoped first-party artifacts as thin CLI wrappers,
not an "aikata agent." The gap above is *not* a reason to add a runtime
agent; it is a reason to add a second small skill that teaches the
operating loop and delegates command execution to the CLI wrapper.

## Decision

### D1 — Split the skill surface by responsibility into two skills

Replace the single `aikata` skill with two:

- **`aikata-cli`** — the existing CLI-wrapper responsibility, renamed.
  Covers `init`, `enable`, `new`, `generate`, `doctor`, `sync`,
  generated-artifact safety, `doctor --json` (schema v1) parsing, and the
  PATH-missing fallback.
- **`aikata-context`** — new. Triggers when an agent begins non-trivial
  work in an aikata-managed repository (recognizable by `AGENTS.md` and/or
  `.aikata/aikata.yaml`) and teaches the context-maintenance loop: read
  the relevant canonical documents first; keep `docs/tasks/current.md`
  current at start / progress / completion when that slot exists; classify
  durable information into the correct slot (invariant rules → `AGENTS.md`,
  requirements → `SPEC.md`, design decisions → `docs/adr/`, durable
  facts/preferences → `docs/memory/`, in-flight state →
  `docs/tasks/current.md`); and check documentation impact, verification
  results, unresolved questions, and handoff state before declaring work
  complete. It invokes no commands itself — it hands off to `aikata-cli`
  for `doctor`, `sync`, `generate`, and `new adr`.

Both names carry the `aikata-` prefix so globally-installed skills stay
attributable and avoid generic-name collisions.

### D2 — One install surface; same marketplace entry and plugin

`aikata-cli` and `aikata-context` ship together from the existing single
`aikata` marketplace entry and plugin. No second marketplace entry, no
second plugin. The Claude Code plugin manifest lists both skill files; the
Codex plugin auto-discovers both from its `skills/` directory; the
universal tree carries both as sibling skill directories. The introduction
unit for a user remains one `aikata` plugin.

### D3 — `aikata-context` stays a single MVP skill (no further split)

`aikata-context` is intentionally *not* split into narrower
`aikata-memory`, `aikata-adr`, or handoff skills for the MVP. The loop is
short and an agent must be able to pick the right skill (cli vs. context)
without loading an all-purpose workflow manual. Splitting further is
deferred until dogfooding demonstrates the single skill is too broad —
not before.

### D4 — No runtime-agent boundary crossing

This refines ADR 0015's thin-wrapper scope; it does not breach it. No
"aikata agent," runtime personality, MCP server, sub-agent, or app
integration is introduced. The selected model remains small, composable
skills used by the user's existing agent. aikata stays a CLI plus
document conventions; the skills are guidance the user's chosen model
reads, not a competing assistant.

### D5 — No backward-compatibility layer for the old single-skill layout

The only existing user is the maintainer, so v0.10.0 replaces the old
`aikata` skill layout directly. No legacy aliases, no duplicate legacy
skill files, no transitional packaging are retained. The reinstall step
is documented in `dist/README.md` (remove the old
`~/.claude/skills/aikata.md` / `~/.agents/skills/aikata/` install, then
install the two new skills, or reinstall the plugin from the marketplace).

### D6 — Single canonical content per skill; copies for discovery only

`dist/universal-skill/<skill>/SKILL.md` is the **single canonical source**
of each skill's content (plus a Codex `agents/openai.yaml`). Every
per-platform file — the Codex plugin copy, the Claude Code plugin copy,
and the Claude Code standalone copy — is **byte-identical** to it. The
prior universal-vs-Claude-Code content divergence is retired: copies now
exist purely for per-platform discovery *location and format* (Codex needs
`skills/<name>/SKILL.md`; Claude Code needs `skills/<name>.md` listed in
`plugin.json`; the universal layout needs `.agents/skills/<name>/SKILL.md`),
never for content. A repository test enforces byte-identity of every copy
against the canonical source.

### D7 — Do-No-Harm compliance

Both skills are opt-in agent guidance installed by the user (ADR 0003). A
project that never installs them is unaffected; nothing in the aikata CLI,
scaffold output, or generated artifacts changes because of this ADR. The
skills only read canonical documents and recommend the existing CLI — they
mutate nothing on their own.

## Consequences

- The distribution layout becomes a per-skill directory tree:
  `dist/universal-skill/{aikata-cli,aikata-context}/SKILL.md` (each with
  `agents/openai.yaml`), mirrored byte-for-byte into the Codex and Claude
  Code copies. The byte-identity test (`TestSkillCopiesMatchCanonical`)
  and the CI archive assertions cover both skills.
- The single-file `.md` release assets (`aikata-skill.md`,
  `aikata-universal-skill.md`) are dropped — a single `.md` can no longer
  represent a two-skill surface. The universal and Codex tarballs plus the
  self-hosted marketplaces are the install path.
- Version lockstep continues across `.claude-plugin/marketplace.json`,
  `dist/claude-code/plugin/plugin.json`, and
  `dist/codex/plugin/.codex-plugin/plugin.json` (v0.10.0).
- The six Claude Code slash commands are unchanged; they back
  `aikata-cli`'s domain. `aikata-context` adds no commands (no CLI feature
  work is in scope).
- An agent now has two attributable responsibilities to select between.
  The trigger boundary is the success criterion and must be dogfooded.

### Trigger-boundary expectations (dogfood targets)

- `aikata-context` **fires**: "implement a payment flow in this aikata
  repo"; "refactor the doctor package" — non-trivial work in a repo with
  aikata markers.
- `aikata-context` **does not fire**: "what's 2+2"; "fix this typo in
  README" (trivial); a pure CLI lifecycle request (routes to `aikata-cli`);
  any repository without `AGENTS.md` / `.aikata/aikata.yaml`.
- `aikata-cli` **fires**: "regenerate CLAUDE.md"; "run aikata doctor
  --json"; "init a new aikata project"; "aikata sync".
- `aikata-cli` **does not fire**: "add a feature" (routes to
  `aikata-context`); source-only work in a non-aikata repository.

## Alternatives Considered

- **Keep one all-purpose skill.** Rejected: a single skill conflating CLI
  invocation with the operating loop is harder for an agent to select
  against and grows into a workflow manual, defeating the small-composable
  model (ADR 0015).
- **Split into many narrow skills now** (`aikata-memory`, `aikata-adr`,
  handoff). Rejected for the MVP (D3): premature before dogfooding shows
  the single context skill is too broad.
- **Keep the two per-platform content families** (universal vs. Claude
  Code). Rejected (D6): the divergence was only a framing paragraph, and
  collapsing to one canonical source makes the copy boundary total,
  trivially testable, and simpler to document.
- **A separate marketplace entry / plugin per skill.** Rejected (D2): it
  fractures the single introduction unit for no benefit to a single
  maintainer.
