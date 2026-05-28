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

Planned first-party wrapper directories (ADR 0015):

- `dist/universal-skill/` — v0.8.x package for `npx skills add ...`.
- `dist/codex/`, `dist/cursor/`, `dist/gemini-cli/` — v1.0 native
  wrappers where the platform shape is stable enough.

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

Marketplace listing (one-click install) is deferred to v0.8.x once the
upstream listing flow is stable; the manual install above is the
supported path today. The standalone skill at
`dist/claude-code/skill/SKILL.md` continues to work for users who do
not want the slash commands.

## Universal skill (planned v0.8.x)

ADR 0015 schedules a first-party universal skill package for installers
such as `npx skills add ... --agent universal`. It will live under
`dist/universal-skill/` unless the installer requires a publication
mirror repository. The package teaches agents how to invoke aikata; it
does not install third-party skills.

## Other tools

Codex skills / plugins, Cursor custom modes or rule packs, Gemini CLI
extensions, and a thin VS Code wrapper are scoped to v1.0 (see
[ROADMAP.md](../ROADMAP.md)). Until then, this directory will grow only
when a new tool's distribution shape is stable enough to ship.

This directory is for first-party aikata distribution artifacts. It does
not install or vendor third-party skill repositories such as those used
with `npx skills add ...`; that interop question is tracked in
[Q-ECOSYSTEM-04](../docs/decisions/open-questions.md#q-ecosystem-04--external-skill--plugin-marketplace-interop).
