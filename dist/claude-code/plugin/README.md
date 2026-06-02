---
project: aikata
status: draft
version: 0.10.3
updated: 2026-06-03
audience: [human, agent]
---

# aikata — Claude Code Plugin

This directory holds the Claude Code plugin. From v0.10.3 it bundles
**two** skills and nothing else (no slash commands — ADR 0041):
`aikata-cli` (CLI invocation guidance) and `aikata-context` (the in-repo
context-maintenance loop). They appear in the `/` menu and are invoked as
`/aikata:aikata-cli` and `/aikata:aikata-context`, or Claude loads them
automatically when relevant.

## What ships in this plugin

| Component | Source | Purpose |
|---|---|---|
| `.claude-plugin/plugin.json` | — | Plugin manifest (metadata only; skills auto-discover) |
| `skills/aikata-cli/SKILL.md` | mirrors `dist/universal-skill/aikata-cli/SKILL.md` | "When and how to call the aikata CLI" |
| `skills/aikata-context/SKILL.md` | mirrors `dist/universal-skill/aikata-context/SKILL.md` | "Operating loop inside an aikata repo" |

Skills are auto-discovered from the `skills/<name>/SKILL.md` directories;
the manifest carries only metadata (`claude plugin validate` passes).

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
(`aikata-cli/SKILL.md`, `aikata-context/SKILL.md`) are the same two skills
without the plugin wrapper. Both forms ship byte-identical content from the
canonical sources at `dist/universal-skill/<skill>/SKILL.md`
(`internal/repolint/distribution_test.go` asserts this); do not edit the
plugin copies directly. See ADR 0040 and ADR 0041 for the copy boundary
and the directory-skill layout.

## Versioning

`plugin.json`'s `version` tracks the aikata release that shipped the plugin
and stays in lockstep with the marketplace manifests.
