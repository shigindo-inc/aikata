---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-03
audience: [human, agent]
---

# ADR 0041 — Skills-only surface & Claude Code plugin skill layout

- **Status**: Accepted
- **Date**: 2026-06-03
- **Deciders**: aikata maintainers
- **Related**: ADR 0015 (first-party skill & plugin distribution), ADR 0036
  (Codex native distribution), ADR 0040 (collaboration-operation skill split —
  this ADR reverses its "commands unchanged" note). ADR 0040's body is
  immutable; this ADR supersedes that single point.

## Context

After the v0.10.0 skill split (ADR 0040), the Claude Code plugin shipped:

- two skills as **flat files** `skills/aikata-cli.md` and
  `skills/aikata-context.md`, listed under a non-standard `components` object
  in `plugin.json`; and
- six slash commands (`/aikata-init` … `/aikata-enable`) carried over unchanged
  from v0.6/v0.9.5.

Two problems surfaced in real use:

1. **The skills never loaded in Claude Code.** Per the Claude Code plugins
   reference, plugin skills must be `skills/<name>/SKILL.md` *directories* and
   are *auto-discovered* — "commands are simple markdown files." Claude Code
   ignores unknown manifest fields (so the `components` object did nothing) and
   auto-discovers `skills/<name>/SKILL.md` directories. Our flat
   `skills/<name>.md` files matched neither shape, so the two skills were
   silently dropped while the `commands/*.md` files (correct flat-file format)
   loaded. The user saw `/aikata-init` but never `/aikata:aikata-cli`. The
   manifest was also at the plugin root rather than the required
   `.claude-plugin/plugin.json`. The same flat-file shape broke the standalone
   `~/.claude/skills/` install (personal skills are also
   `~/.claude/skills/<name>/SKILL.md` directories).

2. **Split, asymmetric surface.** The six commands duplicated the `aikata-cli`
   skill's domain, and Codex has no slash-command mechanism, so the surfaces
   diverged per platform — contrary to the project's preference for a small,
   uniform command surface.

## Decision

### D1 — Skills-only surface; remove the slash commands

Remove all six first-party Claude Code slash commands. Both platforms (Claude
Code and Codex) — and the universal install — now expose exactly the two skills
`aikata-cli` and `aikata-context`. The skills already document every CLI verb
the commands wrapped; the commands added only redundant `$ARGUMENTS` scaffolding.

### D2 — Claude Code skills use the `<name>/SKILL.md` directory layout

Both Claude Code copies move from flat `.md` files to per-skill directories:

- plugin: `dist/claude-code/plugin/skills/<name>/SKILL.md`
- standalone: `dist/claude-code/skill/<name>/SKILL.md`

All four distribution locations (universal, Codex plugin, Claude Code plugin,
Claude Code standalone) now use the identical `<base>/<name>/SKILL.md` shape.
The content stays byte-identical to the canonical
`dist/universal-skill/<name>/SKILL.md` (repository test enforced).

### D3 — Standard, metadata-only Claude Code plugin manifest

`dist/claude-code/plugin/.claude-plugin/plugin.json` (moved into the required
`.claude-plugin/` directory) carries only metadata — `name`, `version`,
`description`, `author`, `homepage`, `repository`, `license`, `keywords` — plus
`$schema`. Skills are auto-discovered from `skills/<name>/SKILL.md`; the
non-standard `components` object and the unused `requires` object are dropped.
`claude plugin validate` passes.

### D4 — Invocation is plugin-namespaced

Claude Code surfaces plugin skills under the `plugin-name:skill-name` namespace,
so they are invoked as `/aikata:aikata-cli` and `/aikata:aikata-context` (and
appear in the `/` menu). This is inherent to plugin skills and is documented in
the install instructions. Nothing in the skill files disables invocation
(`disable-model-invocation` / `user-invocable` are left at their defaults).

### D5 — Do-No-Harm

Docs and distribution metadata only; the aikata CLI, scaffold output, and
generated artifacts are unchanged. A project that does not install the skills is
unaffected (ADR 0003).

## Consequences

- The Claude Code plugin and standalone skills load and are `/`-invocable for
  the first time; the surface matches Codex and universal.
- `plugin.json` moves to `.claude-plugin/plugin.json`; the version-lockstep
  test reads the new path. New tests guard the directory-skill layout
  (`TestClaudePluginSkillsAreAutoDiscoverable`) and the absence of commands
  (`TestClaudePluginHasNoCommands`).
- Users who installed the old plugin update via `/plugin marketplace update
  aikata` + reinstall; the slash commands disappear (intended).
- ADR 0040's statement that "the six Claude Code slash commands are unchanged"
  no longer holds.

## Alternatives Considered

- **Keep the commands, just fix the skill layout.** Rejected: it leaves the
  redundant, platform-asymmetric command surface the maintainer asked to remove.
- **Keep flat skill files and force them via a manifest key.** Rejected: the
  documented, supported shape is `skills/<name>/SKILL.md` auto-discovery;
  fighting it with non-standard manifest entries is what caused the bug.
