# dist/

Shippable artifacts that ride alongside the aikata binary. Files here
are **not** compiled into the binary; the release workflow attaches
them as plain assets so users can copy them where they belong.

## Layout

```
dist/
├── claude-code/
│   ├── skill/
│   │   └── SKILL.md   ← minimal Claude Code skill (v0.3.1+)
│   └── plugin/
│       ├── plugin.json
│       ├── README.md
│       ├── skills/aikata.md   ← byte-identical copy of skill/SKILL.md
│       └── commands/
│           ├── aikata-init.md
│           ├── aikata-generate.md
│           ├── aikata-doctor.md
│           ├── aikata-sync.md
│           ├── aikata-new.md
│           └── aikata-enable.md
├── codex/
│   └── plugin/
│       ├── .codex-plugin/plugin.json
│       └── skills/aikata/
│           ├── SKILL.md   ← byte-identical copy of universal-skill/SKILL.md
│           └── agents/openai.yaml
└── universal-skill/
    ├── SKILL.md   ← tool-agnostic skill for `npx skills add ...` (v0.9.3+)
    └── agents/openai.yaml   ← Codex App UI metadata (v0.9.6+)
```

The repository root also carries `.claude-plugin/marketplace.json` and
`.agents/plugins/marketplace.json`. They list the Claude Code and Codex
plugins above so the repo is installable as a self-hosted marketplace.

Planned first-party wrapper directories (ADR 0015):

- `dist/cursor/`, `dist/gemini-cli/` — v1.0 native wrappers where the
  platform shape is stable enough.

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
plugin that bundles the same skill plus six slash commands:
`/aikata-init`, `/aikata-generate`, `/aikata-doctor`, `/aikata-sync`,
and (v0.9.5) `/aikata-new` and `/aikata-enable` for post-init authoring.

To install it as a self-hosted marketplace (v0.9.3+):

```text
/plugin marketplace add shigindo-inc/aikata
/plugin install aikata@aikata
```

The root `.claude-plugin/marketplace.json` makes the repo a valid
marketplace source, so the two commands above install the plugin without
a local checkout. Submission to Anthropic's canonical marketplace stays
gated on the upstream review flow (ADR 0032 D1); this self-hosted path is
the supported install today.

Or install from a local checkout:

```bash
mkdir -p ~/.claude/plugins/aikata
cp -r dist/claude-code/plugin/* ~/.claude/plugins/aikata/
```

The standalone skill at `dist/claude-code/skill/SKILL.md` continues to
work for users who do not want the slash commands.

## Universal skill (v0.9.3+)

ADR 0015 schedules a first-party universal skill for installers such as
`npx skills`. It lives at `dist/universal-skill/SKILL.md` — a
tool-agnostic skill that teaches any `AGENTS.md`-reading agent how to
invoke the aikata CLI. It does not install third-party skills.

Install it with the skill-specific tree path (the installer walks a
top-level `skills/` container, so point it at the directory directly):

```bash
npx skills add https://github.com/shigindo-inc/aikata/tree/main/dist/universal-skill --agent universal
```

This installs into the universal `.agents/skills/` layout, which Claude
Code, Codex, Cursor, Gemini CLI, and other AGENTS.md-aware tools read.
`dist/universal-skill/` is the canonical source; no publication mirror
repository is required.

For Codex CLI `0.125.0+`, direct installation into
`.agents/skills/aikata/` is also the fallback when native plugin commands
are unavailable:

```bash
mkdir -p ~/.agents/skills/aikata/agents
cp dist/universal-skill/SKILL.md ~/.agents/skills/aikata/SKILL.md
cp dist/universal-skill/agents/openai.yaml ~/.agents/skills/aikata/agents/openai.yaml
```

Restart Codex after installation. The release keeps the existing
`aikata-universal-skill.md` one-file asset and also ships
`aikata-universal-skill.tar.gz` for the complete directory.

## Codex plugin (v0.9.6+)

Codex CLI `0.135.0+` can install `dist/codex/plugin/` through the root
self-hosted marketplace:

```bash
codex plugin marketplace add shigindo-inc/aikata --ref v0.9.7
codex plugin add aikata@aikata
```

The plugin is a thin wrapper over the CLI: one skill, the same
`agents/openai.yaml` UI metadata, no MCP server, and no app integration.
The release also ships `aikata-codex-plugin.tar.gz` for offline use.

## Other tools

Cursor custom modes or rule packs, Gemini CLI extensions, and a thin VS
Code wrapper are scoped to v1.0 (see [ROADMAP.md](../ROADMAP.md)). Until
then, this directory will grow only when a new tool's distribution shape
is stable enough to ship.

This directory is for first-party aikata distribution artifacts. It does
not install or vendor third-party skill repositories such as those used
with `npx skills add ...`; that interop question is tracked in
[Q-ECOSYSTEM-04](../docs/decisions/open-questions.md#q-ecosystem-04--external-skill--plugin-marketplace-interop).
