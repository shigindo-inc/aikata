---
description: Bring this repo into aikata's recommended layout — relocate off-structure docs into docs/layout.md homes, dry-run first, git mv only after you approve
---

The user invoked the slash command `/aikata:migrate-structure`.

Immediately invoke the `aikata:migrate-structure` skill with the Skill tool
and follow its instructions. No preamble or confirmation is needed to start —
the skill's guidance takes priority. The skill is observe → propose →
confirm: it shows a dry-run plan and moves files (with `git mv`) only after
you explicitly approve, and never rewrites document contents.

Arguments the user passed: $ARGUMENTS
