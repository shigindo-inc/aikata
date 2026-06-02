# dist/

Shippable artifacts that ride alongside the aikata binary. Files here
are **not** compiled into the binary; the release workflow attaches
them as plain assets so users can copy them where they belong.

From v0.10.0 the first-party skill surface is **two** skills (ADR 0040):

- **`aikata-cli`** — when and how to invoke the aikata CLI (`init`,
  `generate`, `doctor`, `sync`, `enable`, `new`) and how to parse
  `aikata doctor --json`.
- **`aikata-context`** — the daily in-repo context-maintenance loop:
  which canonical documents to read before editing, where newly-learned
  context belongs, how to keep `docs/tasks/current.md` current, and what
  to check before handoff.

Both ship together from the single `aikata` marketplace entry and plugin.

## Layout

```
dist/
├── claude-code/
│   ├── skill/                              ← standalone skills (no slash commands)
│   │   ├── aikata-cli.md
│   │   └── aikata-context.md
│   └── plugin/
│       ├── plugin.json
│       ├── README.md
│       ├── skills/
│       │   ├── aikata-cli.md               ← byte-identical to the canonical SKILL.md
│       │   └── aikata-context.md           ← byte-identical
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
│       └── skills/
│           ├── aikata-cli/
│           │   ├── SKILL.md                ← byte-identical to the canonical SKILL.md
│           │   └── agents/openai.yaml      ← byte-identical
│           └── aikata-context/
│               ├── SKILL.md                ← byte-identical
│               └── agents/openai.yaml      ← byte-identical
└── universal-skill/                        ← CANONICAL source for both skills
    ├── aikata-cli/
    │   ├── SKILL.md
    │   └── agents/openai.yaml              ← Codex App UI metadata
    └── aikata-context/
        ├── SKILL.md
        └── agents/openai.yaml
```

`dist/universal-skill/<skill>/SKILL.md` is the **single canonical source**
of each skill's content; every per-platform file is byte-identical to it
(enforced by `internal/repolint/distribution_test.go`). Copies exist only
for per-platform discovery location/format — Codex needs
`skills/<name>/SKILL.md`, Claude Code needs `skills/<name>.md` listed in
`plugin.json`, the universal layout needs `.agents/skills/<name>/SKILL.md`
— never for content (ADR 0040).

The repository root also carries `.claude-plugin/marketplace.json` and
`.agents/plugins/marketplace.json`. They list the Claude Code and Codex
plugins above so the repo is installable as a self-hosted marketplace.

Planned first-party wrapper directories (ADR 0015):

- `dist/cursor/`, `dist/gemini-cli/` — v1.0 native wrappers where the
  platform shape is stable enough.

## Reinstalling after the v0.10.0 split

If you previously installed the single `aikata` skill (v0.3.1–v0.9.x),
remove it before installing the two new skills — there are no
compatibility aliases (ADR 0040):

```bash
# Old standalone Claude Code skill
rm -f ~/.claude/skills/aikata.md
# Old universal / Codex skill directory
rm -rf ~/.agents/skills/aikata
```

If you installed via a plugin, refresh the marketplace, then reinstall so
the new two-skill plugin is picked up:

```text
# Claude Code
/plugin marketplace update aikata
/plugin install aikata@aikata
```

```bash
# Codex (re-point the marketplace at the new tag, then reinstall)
codex plugin marketplace add shigindo-inc/aikata --ref v0.10.1
codex plugin add aikata@aikata
```

The full install instructions for each surface are below.

## Claude Code plugin (v0.6+)

The Claude Code plugin bundles both skills plus six slash commands:
`/aikata-init`, `/aikata-generate`, `/aikata-doctor`, `/aikata-sync`,
`/aikata-new`, and `/aikata-enable`.

Install it as a self-hosted marketplace (v0.9.3+):

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

## Claude Code standalone skills (no slash commands)

For users who want the skills without the plugin's slash commands, copy
the two standalone files into `~/.claude/skills/`:

```bash
mkdir -p ~/.claude/skills
cp dist/claude-code/skill/aikata-cli.md     ~/.claude/skills/aikata-cli.md
cp dist/claude-code/skill/aikata-context.md ~/.claude/skills/aikata-context.md
```

Restart Claude Code, then ask it about scaffolding/regenerating an aikata
project (`aikata-cli`) or start non-trivial work in an aikata repo
(`aikata-context`) — it will pick up the right skill automatically.

## Universal skill (v0.9.3+)

ADR 0015 schedules a first-party universal skill for installers such as
`npx skills`. The canonical tree lives at `dist/universal-skill/`, with
one directory per skill — the `.agents/skills/<name>/SKILL.md` layout that
Claude Code, Codex, Cursor, Gemini CLI, and other AGENTS.md-aware tools
read. It does not install third-party skills.

Install **both** skills in one command by pointing the installer at the
`universal-skill` container directory — it walks the container one level
deep and discovers each `<skill>/SKILL.md`:

```bash
npx skills add https://github.com/shigindo-inc/aikata/tree/main/dist/universal-skill --agent universal
```

Point at the container, **not** an individual skill subdirectory: the
installer treats the given path as a container of skills, so a path like
`.../dist/universal-skill/aikata-context` finds nothing. Add `--skill
aikata-cli` (or `aikata-context`) to install just one. `dist/universal-skill/`
is the canonical source; no publication mirror repository is required.

For Codex CLI `0.125.0+`, direct installation into `.agents/skills/` is
also the fallback when native plugin commands are unavailable. Extract the
release tarball, or copy the directories from a checkout:

```bash
mkdir -p ~/.agents/skills
cp -r dist/universal-skill/aikata-cli     ~/.agents/skills/aikata-cli
cp -r dist/universal-skill/aikata-context ~/.agents/skills/aikata-context
```

Restart Codex after installation. The release ships
`aikata-universal-skill.tar.gz` for the complete two-skill directory.

## Codex plugin (v0.9.6+)

Codex CLI `0.135.0+` can install `dist/codex/plugin/` through the root
self-hosted marketplace:

```bash
codex plugin marketplace add shigindo-inc/aikata --ref v0.10.1
codex plugin add aikata@aikata
```

The plugin is a thin wrapper over the CLI: both skills, each with its own
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
