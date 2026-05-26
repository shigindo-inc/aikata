# dist/

Shippable artifacts that ride alongside the aikata binary. Files here
are **not** compiled into the binary; the release workflow attaches
them as plain assets so users can copy them where they belong.

## Layout

```
dist/
└── claude-code/
    ├── skill/
    │   └── SKILL.md   ← minimal Claude Code skill (v0.3.1+)
    └── plugin/
        ├── plugin.json
        ├── README.md
        ├── skills/aikata.md   ← byte-identical copy of skill/SKILL.md
        └── commands/
            ├── aikata-init.md
            ├── aikata-generate.md
            ├── aikata-doctor.md
            └── aikata-sync.md
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

## Claude Code plugin (v0.6+)

The v0.6 release adds `dist/claude-code/plugin/` — a Claude Code
plugin that bundles the same skill plus four slash commands:
`/aikata-init`, `/aikata-generate`, `/aikata-doctor`, `/aikata-sync`.

To install it from a local checkout:

```bash
mkdir -p ~/.claude/plugins/aikata
cp -r dist/claude-code/plugin/* ~/.claude/plugins/aikata/
```

Marketplace listing (one-click install) is deferred to a later v0.6.x
release once the upstream listing flow is stable; the manual install
above is the supported path today. The standalone skill at
`dist/claude-code/skill/SKILL.md` continues to work for users who do
not want the slash commands.

## Other tools

Cursor custom modes, Gemini CLI extensions, and a thin VS Code wrapper
are scoped to v1.0 (see [ROADMAP.md](../ROADMAP.md)). Until then, this
directory will grow only when a new tool's distribution shape is
stable enough to ship.
