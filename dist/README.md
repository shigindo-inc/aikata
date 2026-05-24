# dist/

Shippable artifacts that ride alongside the aikata binary. Files here
are **not** compiled into the binary; the release workflow attaches
them as plain assets so users can copy them where they belong.

## Layout

```
dist/
└── claude-code/
    └── skill/
        └── SKILL.md   ← minimal Claude Code skill (v0.3.1+)
```

## Claude Code skill

The single `SKILL.md` teaches Claude when to call `aikata init`,
`aikata generate`, and `aikata doctor`, and how to parse
`aikata doctor --json`. It is intentionally tiny — no slash commands,
sub-agents, or hooks — so it stays a one-file copy.

To install it:

```bash
mkdir -p ~/.claude/skills
cp dist/claude-code/skill/SKILL.md ~/.claude/skills/aikata.md
```

Or, from the GitHub release asset (no local checkout required):

```bash
curl -fsSL -o ~/.claude/skills/aikata.md \
  https://github.com/shigindo-inc/aikata/releases/latest/download/aikata-skill.md
```

Restart Claude Code, then ask it about scaffolding or regenerating an
aikata project — it will pick up the skill automatically.

A fuller distribution (slash commands `/aikata-init`, `/aikata-generate`,
`/aikata-doctor`; sub-agents; hooks) lands in v0.6 as a Claude Code
*plugin* under `dist/claude-code/plugin/`. The skill above is forward
compatible — the plugin extends it rather than replacing it.

## Other tools

Cursor custom modes, Gemini CLI extensions, and a thin VS Code wrapper
are scoped to v1.0 (see [ROADMAP.md](../ROADMAP.md)). Until then, this
directory will grow only when a new tool's distribution shape is
stable enough to ship.
