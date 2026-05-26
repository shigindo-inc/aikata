---
project: aikata
status: draft
version: 0.6.0
updated: 2026-05-26
audience: [human, agent]
---

# aikata — Claude Code Plugin

This directory holds the v0.6 Claude Code plugin scaffold. It bundles
the v0.3.1 skill (`skills/aikata.md`) with four slash commands so
Claude Code users can drive aikata without rote shell invocations.

## What ships in this plugin

| Component | Source | Purpose |
|---|---|---|
| `plugin.json` | `dist/claude-code/plugin/plugin.json` | Plugin manifest (name, version, requires) |
| `skills/aikata.md` | mirrors `dist/claude-code/skill/SKILL.md` | "When to call aikata" guidance |
| `commands/aikata-init.md` | new in v0.6 | `/aikata-init` slash command |
| `commands/aikata-generate.md` | new in v0.6 | `/aikata-generate` slash command |
| `commands/aikata-doctor.md` | new in v0.6 | `/aikata-doctor` slash command |
| `commands/aikata-sync.md` | new in v0.6 | `/aikata-sync` slash command |

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
install today (v0.6); marketplace listing lands in a v0.6.x or v1.0
release once the upstream flow is stable.

## Relationship to the v0.3.1 skill

`dist/claude-code/skill/SKILL.md` (the standalone, single-file skill
shipped since v0.3.1) continues to work as-is. The plugin **extends**
that skill with slash commands — installing the plugin is strictly
additive; uninstalling the plugin falls back cleanly to the standalone
skill.

The skill copy inside `plugin/skills/aikata.md` is byte-identical to
`dist/claude-code/skill/SKILL.md`. The release workflow asserts this
identity; do not edit the plugin copy directly.

## Versioning

`plugin.json`'s `version` tracks the aikata release that shipped the
plugin (so the v0.6.0 plugin matches the v0.6.0 CLI). `requires.minVersion`
enforces that the user's installed CLI is recent enough to honor every
slash command's documented invocation.
