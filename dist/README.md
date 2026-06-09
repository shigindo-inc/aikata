# dist/

Shippable artifacts that ride alongside the aikata binary. Files here
are **not** compiled into the binary; the release workflow attaches
them as plain assets so users can copy them where they belong.

From v0.12.0 the first-party skill surface is **three** skills, named by
capability with thin Claude Code plugin command wrappers (ADR 0043):

- **`manage-docs`** — when and how to invoke the aikata CLI (`init`,
  `fill`, `generate`, `doctor`, `sync`, `enable`, `new`) and how to parse
  `aikata doctor --json`.
- **`track-context`** — the daily in-repo context-maintenance loop:
  which canonical documents to read before editing, where newly-learned
  context belongs, how to keep `docs/tasks/current.md` current, and what
  to check before handoff.
- **`refresh-docs`** — the downstream-maintenance loop that brings an
  aikata-managed repo up to the latest aikata: `aikata update` → `sync`
  → `fill` → `doctor` → retire deprecated docs → `generate`.

Every `SKILL.md` carries `user-invocable: false` in its frontmatter. In
the Claude Code **plugin** only, thin slash-command wrappers under
`commands/` (`commands/manage-docs.md`, `commands/refresh-docs.md`)
provide a clean user entry point — `/aikata:manage-docs` and
`/aikata:refresh-docs` appear in the `/` menu — while the skills
themselves stay out of the `/` menu (no double listing). `track-context`
has **no** command wrapper by design; it is meant to fire automatically
from its description. On Codex, Claude Code standalone, and the universal
install there are no commands, so `user-invocable: false` means a user
cannot launch a skill directly with `/` — those platforms rely on the
model auto-firing on the skill's description. Only the plugin gets a
first-class command entry point; ADR 0043 accepts this platform
asymmetry.

All three ship together from the single `aikata` marketplace entry and plugin.

## Layout

```
dist/
├── claude-code/
│   ├── skill/                              ← standalone skills (no slash commands)
│   │   ├── manage-docs/SKILL.md
│   │   ├── track-context/SKILL.md
│   │   └── refresh-docs/SKILL.md
│   └── plugin/
│       ├── .claude-plugin/plugin.json
│       ├── README.md
│       ├── commands/                       ← thin slash-command wrappers (plugin only)
│       │   ├── manage-docs.md
│       │   └── refresh-docs.md
│       └── skills/
│           ├── manage-docs/SKILL.md        ← byte-identical to the canonical SKILL.md
│           ├── track-context/SKILL.md      ← byte-identical
│           └── refresh-docs/SKILL.md       ← byte-identical
├── codex/
│   └── plugin/
│       ├── .codex-plugin/plugin.json
│       └── skills/
│           ├── manage-docs/
│           │   ├── SKILL.md                ← byte-identical to the canonical SKILL.md
│           │   └── agents/openai.yaml      ← byte-identical
│           ├── track-context/
│           │   ├── SKILL.md                ← byte-identical
│           │   └── agents/openai.yaml      ← byte-identical
│           └── refresh-docs/
│               ├── SKILL.md                ← byte-identical
│               └── agents/openai.yaml      ← byte-identical
└── universal-skill/                        ← CANONICAL source for all three skills
    ├── manage-docs/
    │   ├── SKILL.md
    │   └── agents/openai.yaml              ← Codex App UI metadata
    ├── track-context/
    │   ├── SKILL.md
    │   └── agents/openai.yaml
    └── refresh-docs/
        ├── SKILL.md
        └── agents/openai.yaml
```

`dist/universal-skill/<skill>/SKILL.md` is the **single canonical source**
of each skill's content; every per-platform file is byte-identical to it
(enforced by `internal/repolint/distribution_test.go`). All platforms use
the same `<base>/<skill>/SKILL.md` directory layout — Claude Code plugin
skills auto-discover from `skills/<name>/SKILL.md`, Codex reads
`skills/<name>/SKILL.md`, and the universal layout uses
`.agents/skills/<name>/SKILL.md`. Copies exist only for per-platform
discovery location, never for content (ADR 0040, ADR 0041, ADR 0043). The
Claude Code plugin additionally carries thin command wrappers under
`commands/` for `manage-docs` and `refresh-docs` (ADR 0043); the other
surfaces are skill-only.

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

If you installed via a plugin, update it so the new two-skill plugin is
picked up:

```text
# Claude Code
/plugin marketplace update aikata
/plugin install aikata@aikata
```

```bash
# Codex — if the marketplace tracks the default branch (recommended)
codex plugin marketplace upgrade aikata
codex plugin add aikata@aikata
```

If your Codex `aikata` marketplace was added with a pinned `--ref` (an
older tag), `marketplace upgrade` only re-pulls that same tag, so it will
not advance — you must remove and re-add it. The same is true if
`marketplace add` reports `marketplace 'aikata' is already added from a
different source`:

```bash
codex plugin remove aikata              # drop the stale installed plugin
codex plugin marketplace remove aikata  # drop the old marketplace entry
codex plugin marketplace add shigindo-inc/aikata   # re-add (default branch)
codex plugin add aikata@aikata
```

The full install instructions for each surface are below.

## Claude Code plugin (v0.6+)

The Claude Code plugin bundles the three skills plus thin command wrappers
(ADR 0043). `/aikata:manage-docs` and `/aikata:refresh-docs` appear in the
`/` menu; the skills themselves are `user-invocable: false` so they do not
double-list. `track-context` has no command and loads automatically when
relevant.

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

## Claude Code standalone skills (without the plugin)

For users who want the skills without installing the plugin, copy the
skill directories into `~/.claude/skills/` (personal skills use the same
`<name>/SKILL.md` layout):

```bash
mkdir -p ~/.claude/skills
cp -r dist/claude-code/skill/manage-docs    ~/.claude/skills/manage-docs
cp -r dist/claude-code/skill/track-context  ~/.claude/skills/track-context
cp -r dist/claude-code/skill/refresh-docs   ~/.claude/skills/refresh-docs
```

Restart Claude Code, then ask it about scaffolding/regenerating an aikata
project (`manage-docs`), start non-trivial work in an aikata repo
(`track-context`), or bring a repo up to the latest aikata
(`refresh-docs`) — it picks up the right skill automatically. The
standalone install has no command wrappers, and the skills are
`user-invocable: false`, so you cannot launch them with `/` directly; they
fire from the model on a matching request.

## Universal skill (v0.9.3+)

ADR 0015 schedules a first-party universal skill for installers such as
`npx skills`. The canonical tree lives at `dist/universal-skill/`, with
one directory per skill — the `.agents/skills/<name>/SKILL.md` layout that
Claude Code, Codex, Cursor, Gemini CLI, and other AGENTS.md-aware tools
read. It does not install third-party skills.

Install **all three** skills in one command by pointing the installer at
the `universal-skill` container directory — it walks the container one
level deep and discovers each `<skill>/SKILL.md`:

```bash
npx skills add https://github.com/shigindo-inc/aikata/tree/main/dist/universal-skill --agent universal
```

Point at the container, **not** an individual skill subdirectory: the
installer treats the given path as a container of skills, so a path like
`.../dist/universal-skill/track-context` finds nothing. Add `--skill
manage-docs` (or `track-context` / `refresh-docs`) to install just one.
`dist/universal-skill/` is the canonical source; no publication mirror
repository is required.

The `tree/main/...` URL tracks the default branch. To **update** an
installed skill to the latest, use the dedicated update command:

```bash
npx skills update                 # update all (interactive scope prompt)
npx skills update --global -y     # the --agent universal install is global
npx skills update manage-docs track-context refresh-docs   # or name them
```

For Codex CLI `0.125.0+`, direct installation into `.agents/skills/` is
also the fallback when native plugin commands are unavailable. Extract the
release tarball, or copy the directories from a checkout:

```bash
mkdir -p ~/.agents/skills
cp -r dist/universal-skill/manage-docs    ~/.agents/skills/manage-docs
cp -r dist/universal-skill/track-context  ~/.agents/skills/track-context
cp -r dist/universal-skill/refresh-docs   ~/.agents/skills/refresh-docs
```

Restart Codex after installation. The release ships
`aikata-universal-skill.tar.gz` for the complete three-skill directory.

## Codex plugin (v0.9.6+)

Codex CLI `0.135.0+` can install `dist/codex/plugin/` through the root
self-hosted marketplace. The recommended (update-friendly) form tracks the
default branch, so it always installs the latest skills:

```bash
codex plugin marketplace add shigindo-inc/aikata
codex plugin add aikata@aikata
```

Update later with:

```bash
codex plugin marketplace upgrade aikata
codex plugin add aikata@aikata
```

**Optional — pin a tag for reproducibility.** Add `--ref v0.10.2` to the
`marketplace add` to lock the plugin to a specific release. The trade-off
is updates are manual: `marketplace upgrade` re-pulls the *same* pinned
tag (it never advances), so moving to a newer release means
`codex plugin marketplace remove aikata` then re-adding with the new
`--ref` (see "Reinstalling" above).

The plugin is a thin wrapper over the CLI: all three skills, each with its
own `agents/openai.yaml` UI metadata, no MCP server, and no app
integration. Codex has no slash-command mechanism, so the command wrappers
are Claude Code-only; on Codex the skills are model-invoked.
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
