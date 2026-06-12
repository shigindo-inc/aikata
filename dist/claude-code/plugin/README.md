---
project: aikata
status: draft
version: 0.12.0
updated: 2026-06-09
audience: [human, agent]
---

# aikata — Claude Code Plugin

This directory holds the Claude Code plugin. It bundles **four** skills
with capability names plus thin slash-command wrappers (ADR 0043,
ADR 0046): `manage-docs` (CLI invocation guidance), `track-context` (the
in-repo context-maintenance loop), `refresh-docs` (bring a repo up to the
latest aikata), and `migrate-structure` (relocate off-structure docs into
the recommended layout). The commands `/aikata:manage-docs`,
`/aikata:refresh-docs`, and `/aikata:migrate-structure` appear in the `/`
menu; the skills are `user-invocable: false` so they do not double-list.
`track-context` has no command and loads automatically when relevant.

## What ships in this plugin

| Component | Source | Purpose |
|---|---|---|
| `.claude-plugin/plugin.json` | — | Plugin manifest (metadata only; skills + commands auto-discover) |
| `commands/manage-docs.md` | — | Thin wrapper → invokes the `manage-docs` skill |
| `commands/refresh-docs.md` | — | Thin wrapper → invokes the `refresh-docs` skill |
| `commands/migrate-structure.md` | — | Thin wrapper → invokes the `migrate-structure` skill |
| `skills/manage-docs/SKILL.md` | mirrors `dist/universal-skill/manage-docs/SKILL.md` | "When and how to call the aikata CLI" |
| `skills/track-context/SKILL.md` | mirrors `dist/universal-skill/track-context/SKILL.md` | "Operating loop inside an aikata repo" |
| `skills/refresh-docs/SKILL.md` | mirrors `dist/universal-skill/refresh-docs/SKILL.md` | "Bring the repo's docs up to the latest aikata" |
| `skills/migrate-structure/SKILL.md` | mirrors `dist/universal-skill/migrate-structure/SKILL.md` | "Relocate off-structure docs into the recommended layout" |

Skills are auto-discovered from the `skills/<name>/SKILL.md` directories
and commands from `commands/*.md`; the manifest carries only metadata (no
`commands` key — `claude plugin validate` passes).

## Install

Install from the repo's self-hosted marketplace (the root
`.claude-plugin/marketplace.json` makes this repo a valid source):

```text
/plugin marketplace add shigindo-inc/aikata
/plugin install aikata@aikata
```

Or from a local checkout:

```bash
mkdir -p ~/.claude/plugins/aikata
cp -r dist/claude-code/plugin/* dist/claude-code/plugin/.claude-plugin ~/.claude/plugins/aikata/
```

Inspect the unpacked tree (no autoexec) before enabling. To update an
existing install: `/plugin marketplace update aikata` then reinstall.

## Marketplace listing (deferred)

Publishing to the upstream Claude Code plugin marketplace is gated on the
upstream review flow (ADR 0032 D1). The self-hosted marketplace install
above is the supported path today.

## Relationship to the standalone skills

The standalone skills under `dist/claude-code/skill/`
(`manage-docs/SKILL.md`, `track-context/SKILL.md`,
`refresh-docs/SKILL.md`, `migrate-structure/SKILL.md`) are the same four
skills without the plugin wrapper or commands. Both forms ship byte-identical content from the
canonical sources at `dist/universal-skill/<skill>/SKILL.md`
(`internal/repolint/distribution_test.go` asserts this); do not edit the
plugin copies directly. See ADR 0040, ADR 0041, and ADR 0043 for the copy
boundary, the directory-skill layout, and the command-wrapper surface.

## Versioning

`plugin.json`'s `version` tracks the aikata release that shipped the plugin
and stays in lockstep with the marketplace manifests.
