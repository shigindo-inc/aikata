---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-28
audience: [human, agent]
---

# ADR 0015 - First-party Skill and Plugin Distribution

- **Status**: Accepted
- **Date**: 2026-05-28
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0003
  (Do-No-Harm Policy), ADR 0012 (memory projection deferred),
  Q-ECOSYSTEM-04

## Context

aikata is a CLI-first project that helps humans and LLM coding tools
share one documentation scaffold. The surrounding agent ecosystem now
has multiple native distribution shapes:

- Claude Code skills and plugins.
- Codex skills and plugins.
- Cursor rules / modes.
- Gemini CLI extensions.
- Universal `SKILL.md` distribution through installers such as
  `npx skills add ... --agent universal`.

The maintainer also has personal skill repositories outside this project
for private workflows. Those are useful experiments, but their scope is
different from aikata's public distribution surface.

The decision needed here is whether aikata should publish first-party
skills / plugins / extensions for LLM coding tools, where those files
should live, and whether aikata should be positioned as an "agent"
rather than a CLI.

## Decision

### 1. Ship first-party wrappers, not a new agent

aikata will publish first-party skills / plugins / extensions that teach
agent tools how to use the aikata CLI:

- when to run `aikata init`, `aikata add`, `aikata generate`,
  `aikata sync`, and `aikata doctor`;
- how to parse `aikata doctor --json`;
- why generated artifacts such as `CLAUDE.md` and
  `.cursor/rules/main.mdc` should not be hand-edited;
- how to respect `AGENTS.md` as the canonical source.

aikata will **not** ship "aikata agent" as a runtime personality. Agent
wrappers are distribution and invocation surfaces for the CLI, not a
separate autonomous assistant that competes with the user's chosen model
or workflow.

### 2. Put first-party distribution artifacts under `dist/`

The aikata repository's `dist/` directory is the source of truth for
first-party skill / plugin / extension artifacts:

```
dist/
├── claude-code/
│   ├── skill/
│   └── plugin/
├── universal-skill/        # planned v0.8.x
├── codex/                  # planned v1.0
├── cursor/                 # planned v1.0
└── gemini-cli/             # planned v1.0
```

Separate repositories may be created later only as **publication
mirrors** if a marketplace or installer requires a dedicated repository
layout. In that case, the generated / copied mirror must point back to
this repository as the canonical source. Personal skill repositories are
not part of aikata's release surface.

### 3. Add universal `npx skills add` distribution in v0.8.x

The v0.8.x channel-publication line will include a first-party universal
skill package installable via a command in the shape of:

```bash
npx skills add <aikata-skill-source> --agent universal
```

The exact source path and package name are decided when the installer
and marketplace requirements are known. The supported scope is still
first-party aikata usage guidance, not arbitrary third-party skill
installation.

### 4. Defer non-Claude native wrappers to v1.0

Claude Code already has a minimal skill and a plugin scaffold. Codex,
Cursor, Gemini CLI, and other native wrappers remain v1.0 work so each
can follow the platform's stable native shape instead of forcing a
lowest-common-denominator abstraction.

### 5. Do not manage third-party skill catalogs in v0.x

aikata will not install, update, remove, or curate arbitrary third-party
skills / plugins in v0.x. It may document recommended commands or
manifest locations. A future command such as `aikata add skill-source`
requires a separate ADR covering trust, pinning, update, removal, and
Do-No-Harm behaviour.

## Consequences

### Positive

- Users can discover aikata from inside Claude Code, Codex, Cursor, or
  other tools without aikata becoming a tool-specific product.
- `dist/` keeps release artifacts reviewable with the source code and
  avoids prematurely splitting a small project across repositories.
- Universal skill installation gives multi-agent users one low-friction
  path while native wrappers remain available where they add value.
- The boundary stays aligned with SPEC.md: aikata remains CLI-first and
  document-centered.

### Negative

- `dist/` will grow as more platforms are supported. Mitigation: keep
  one directory per platform and require every artifact to wrap the CLI
  rather than duplicate project rules.
- Some marketplaces may require repository shapes that differ from
  `dist/`. Mitigation: allow publication mirrors, but keep this repo
  canonical.
- Users may ask for "aikata agent" because platforms use that word
  loosely. Mitigation: documentation should say "skill", "plugin",
  "extension", or "wrapper" and reserve "agent" for the user's chosen
  runtime assistant.

## Alternatives Considered

- **Create a separate `aikata-skills` repository now** - rejected for
  v0.x. It adds release coordination before the artifact set is large
  enough to justify it.
- **Use the maintainer's personal skills repository** - rejected. That
  repository is for personal workflow experiments, not first-party
  aikata release artifacts.
- **Make `npx skills add` the only distribution path** - rejected.
  Universal installation is useful, but native marketplaces and
  platform-specific affordances are part of the value.
- **Ship an "aikata agent"** - rejected. It would blur the product
  boundary and compete with the user's model / tool choice.
