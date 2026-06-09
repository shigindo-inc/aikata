---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-09
audience: [human, agent]
---

# ADR 0043 - Command-wrapper skill surface and simple skill names

- **Status**: Accepted
- **Date**: 2026-06-09
- **Deciders**: aikata maintainers
- **Related**: ADR 0040 (collaboration-operation skill split — this ADR
  supersedes its **D1** naming rule, the mandatory `aikata-` prefix), ADR
  0041 (skills-only surface — this ADR supersedes its **D1**, the removal
  of slash commands). Both ADR bodies stay immutable; this ADR supersedes
  exactly those two points, the same way ADR 0041 superseded a single
  point of ADR 0040. Also relates to ADR 0015 (first-party skill & plugin
  distribution), ADR 0036 (Codex native distribution), ADR 0003
  (Do-No-Harm).

## Context

Since v0.10.3 (ADR 0041) the first-party surface has been two skills,
`aikata-cli` and `aikata-context`, exposed identically across all four
distribution copies and invoked with no wrapping slash commands. ADR 0040
D1 additionally mandated the `aikata-` name prefix "so globally-installed
skills stay attributable and avoid generic-name collisions."

In daily use two ergonomic problems surfaced:

1. **The names are redundant and do not convey capability.** In the
   plugin the skills surface under the `aikata:` namespace, so the user
   sees `aikata:aikata-cli` / `aikata:aikata-context` — the tool name
   twice, and `cli` describes a *mechanism*, not what the skill does.
2. **No short, capability-named `/`-menu entry exists.** The maintainer
   wanted the pattern already used by a sibling personal plugin
   (`job-search`): a thin slash-command wrapper whose name states the
   capability, backed by a skill marked `user-invocable: false` so the two
   do not double-list in the `/` menu. Typing a short prefix then
   completes a clean command, and the skill carries the real instructions.

This is a deliberate reversal of two earlier points. ADR 0041 removed
commands partly because they duplicated the skill domain and Codex has no
slash-command mechanism (platform asymmetry). The maintainer has weighed
that asymmetry against the ergonomic gain and chosen the wrapper pattern,
accepting that only the Claude Code **plugin** gets clean commands. ADR
0040 D1's prefix rule is dropped because the command namespace
(`aikata:`) now carries attribution, and the sole user/maintainer makes
generic-name collision a theoretical, not practical, risk.

## Decision

### D1 — Skill names state the capability; no `aikata-` prefix

Rename the two skills to capability-stating names (supersedes ADR 0040
D1):

- `aikata-cli` → **`manage-docs`** — scaffold and maintain the
  AI-readable canonical documentation set via the aikata CLI.
- `aikata-context` → **`track-context`** — keep canonical docs and
  working state current while developing in an aikata repo.

The `description` text still names aikata freely (it explains the
capability); only the `name:` drops the prefix. The canonical
`dist/universal-skill/<name>/SKILL.md` remains the single source; all
copies stay byte-identical (ADR 0040 D6 unchanged).

### D2 — Thin Claude Code command wrappers (supersedes ADR 0041 D1)

Reintroduce slash commands in the Claude Code **plugin only**, as thin
wrappers that immediately invoke the backing skill and forward
`$ARGUMENTS`:

- `commands/manage-docs.md` → `/aikata:manage-docs`
- `commands/refresh-docs.md` → `/aikata:refresh-docs` (see D4)

Commands are auto-discovered from `commands/*.md`; the plugin manifest
gains no `commands` key (matches the sibling-plugin precedent and ADR
0041 D3's metadata-only manifest). The Codex plugin, the Claude Code
standalone skill, and the universal layout get **no** commands — they have
no command mechanism — so their surface stays skill-only.

### D3 — Skills are `user-invocable: false`

Every first-party `SKILL.md` adds `user-invocable: false`. In the plugin
this hides the skill from the `/` menu so the command is the single
user-facing entry (no double-listing). Model invocation is unaffected, so
a command's Skill-tool call still reaches the skill, and on Codex
(`allow_implicit_invocation: true` in `agents/openai.yaml`) the skills
remain model-invoked. Because byte-identity forces the flag into every
copy, the Claude Code **standalone** and universal installs lose direct
user `/`-invocation and rely on model auto-firing — the accepted cost of
one canonical skill file (see Consequences).

### D4 — `track-context` gets no command; add a `refresh-docs` skill+command

`track-context` fires by description-match when non-trivial work begins;
it is never something a user manually triggers, so a wrapper for it would
be the *only* user trigger for a skill nobody triggers — dead weight. It
ships skill-only.

A third skill, **`refresh-docs`**, is added (with a
`/aikata:refresh-docs` command). It teaches the agent the
bring-docs-up-to-date loop for a *downstream* aikata-managed repo: compare
installed vs. latest aikata (`aikata update --check`/`--apply`), pull
template changes (`aikata sync`, which also runs the schema migration of
ADR 0011 D3), fill missing canonical docs (`aikata fill`), reconcile
consistency (`aikata doctor --json`/`--fix`), retire deprecated docs by
judgement, and regenerate tool files (`aikata generate`). This is the
dogfooding signal ADR 0040 D3 deferred a third skill until; it is now
warranted.

### D5 — Deprecated-doc cleanup stays skill-guided, not a CLI verb

`refresh-docs` removes superseded/deprecated docs through the agent's
judgement with user confirmation (doctor surfaces them; the deprecated ADR
must already reference its replacement). No `aikata deprecate-cleanup` verb
is added — the CLI stays thin (ADR 0040 D4). If manual cleanup proves
insufficient under dogfooding, a dedicated verb is revisited in a later
ADR.

### D6 — Do-No-Harm

Distribution metadata, docs, and agent-guidance only. The aikata CLI,
scaffold output, and generated artifacts are unchanged; a project that
never installs the skills is unaffected (ADR 0003).

## Rationale

The command wrapper buys a short, capability-named, double-listing-free
`/`-menu entry where the maintainer actually works (the Claude Code
plugin), at the price of a per-platform asymmetry the maintainer
explicitly accepts. Keeping one canonical, byte-identical skill file is
worth more than per-platform user-invocability, because the model-invoked
path (the common one) is unaffected and the maintainer is the only user of
the standalone/Codex/universal copies. Naming by capability, now that the
`aikata:` command namespace carries attribution, removes the redundant
double-`aikata` the prefix rule had caused.

## Consequences

- The Claude Code plugin exposes `/aikata:manage-docs` and
  `/aikata:refresh-docs`; `track-context` remains model-fired. Skills no
  longer double-list thanks to `user-invocable: false`.
- **Platform asymmetry returns** (the thing ADR 0041 D1 removed): only the
  plugin has commands. Codex/standalone/universal are skill-only and, with
  `user-invocable: false`, depend on model auto-firing for invocation.
- Generic skill names (`manage-docs`, `track-context`, `refresh-docs`)
  carry a theoretical global-namespace collision risk; acceptable for a
  single maintainer, and skill selection is description-driven regardless.
- Tests change: `firstPartySkills` becomes the three new names;
  `TestClaudePluginHasNoCommands` inverts to assert the wrapper commands
  exist (the no-`components`-key check stays); the CI tarball-member
  assertions and `.goreleaser.yml` comment move to the new names.
- `refresh-docs` is the third first-party skill; ADR 0040 D3's deferral is
  now discharged.

## Alternatives considered

- **Keep skills-only, just rename.** Rejected: a rename alone still leaves
  `aikata:manage-docs` with no short command entry and does not deliver the
  wrapper ergonomics the maintainer asked for.
- **Per-verb commands** (`/aikata:doctor`, `/aikata:sync`, …). Rejected by
  the maintainer in planning: it re-fractures the surface ADR 0041 D1
  consolidated and multiplies wrappers over one umbrella skill.
- **Add a `deprecate-cleanup` CLI verb.** Deferred (D5): keeps the CLI thin;
  skill-guided manual cleanup is sufficient until dogfooding shows otherwise.
- **Keep the `aikata-` prefix and only add commands.** Rejected: the
  redundant double-`aikata` in `aikata:aikata-cli` was half the ergonomic
  complaint; the command namespace already attributes the skill.

## References

- ADR 0040 (skill split; D1 prefix rule superseded here, D3 deferral
  discharged, D4/D6 unchanged).
- ADR 0041 (skills-only; D1 command removal superseded here, D2/D3/D4
  unchanged).
- ADR 0011 (sync design; schema migration leaned on by `refresh-docs`).
- ADR 0042 (`fill` command; used by `refresh-docs`).
- Sibling precedent: the `job-search` personal plugin's thin
  command-wrapper + `user-invocable: false` skill pattern.
