---
project: aikata
status: draft
version: 0.10.0
updated: 2026-06-02
audience: [human, agent]
---

# aikata — Claude Code Plugin

This directory holds the Claude Code plugin. From v0.10.0 it bundles
**two** skills — `aikata-cli` (CLI invocation guidance) and
`aikata-context` (the in-repo context-maintenance loop, ADR 0040) — plus
six slash commands so Claude Code users can drive aikata without rote
shell invocations.

## What ships in this plugin

| Component | Source | Purpose |
|---|---|---|
| `plugin.json` | `dist/claude-code/plugin/plugin.json` | Plugin manifest (name, version, requires) |
| `skills/aikata-cli.md` | mirrors `dist/universal-skill/aikata-cli/SKILL.md` | "When and how to call the aikata CLI" |
| `skills/aikata-context.md` | mirrors `dist/universal-skill/aikata-context/SKILL.md` | "Operating loop inside an aikata repo" |
| `commands/aikata-init.md` | `/aikata-init` slash command |
| `commands/aikata-generate.md` | `/aikata-generate` slash command |
| `commands/aikata-doctor.md` | `/aikata-doctor` slash command |
| `commands/aikata-sync.md` | `/aikata-sync` slash command |
| `commands/aikata-new.md` | `/aikata-new` slash command (v0.9.5) |
| `commands/aikata-enable.md` | `/aikata-enable` slash command (v0.9.5) |

## Install

```bash
# Replace with the final published path once the v0.6 release ships
# the plugin tarball.
mkdir -p ~/.claude/plugins/aikata
cd ~/.claude/plugins/aikata
curl -fsSL https://github.com/shigindo-inc/aikata/releases/latest/download/aikata-plugin.tar.gz | tar -xz
```

Then enable it in Claude Code via the plugin manager (UI or
configuration). Inspect the unpacked tree (no autoexec) before
running anything.

## Marketplace listing (deferred)

Publishing this plugin to the upstream Claude Code plugin
marketplace requires:

1. The marketplace listing flow to be available (Anthropic-side).
2. A maintainer to submit aikata for review.

The plugin is functional as a manual `~/.claude/plugins/aikata`
install today (v0.6); marketplace listing is deferred to v0.8.x or
later once the upstream flow is stable.

## Relationship to the standalone skills

The standalone skills under `dist/claude-code/skill/` (`aikata-cli.md`,
`aikata-context.md`) continue to work as-is. The plugin **extends** them
with slash commands — installing the plugin is strictly additive;
uninstalling the plugin falls back cleanly to the standalone skills.

The skill copies inside `plugin/skills/` are byte-identical to the
canonical sources at `dist/universal-skill/<skill>/SKILL.md`
(`internal/repolint/distribution_test.go` asserts this); do not edit the
plugin copies directly. See ADR 0040 for the copy boundary.

## Versioning

`plugin.json`'s `version` tracks the aikata release that shipped the
plugin (so the v0.6.0 plugin matches the v0.6.0 CLI). `requires.minVersion`
enforces that the user's installed CLI is recent enough to honor every
slash command's documented invocation.
