---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
audience: [human, agent]
---

# ADR 0036 - Ship Codex native distribution in v0.9.6

- **Status**: Accepted
- **Date**: 2026-06-01
- **Deciders**: aikata maintainers
- **Related**: ADR 0015 (first-party skill and plugin distribution),
  ADR 0032 (channel-publication split), Q-ECOSYSTEM-04

## Context

ADR 0015 deferred native wrappers beyond Claude Code to v1.0 so aikata
would not commit to unstable platform shapes. Codex has since stabilized
the parts aikata needs:

- Codex CLI `0.125.0` discovers a direct universal skill installed under
  `.agents/skills/`.
- Agent Skills accept optional `agents/openai.yaml` metadata for the
  Codex App skill card and implicit invocation policy.
- Codex CLI `0.135.0+` supports plugin manifests under `.codex-plugin/`
  and self-hosted marketplace catalogs under
  `.agents/plugins/marketplace.json`.

The Codex wrapper can remain thin: it teaches Codex when to use the
existing aikata CLI. It does not need an MCP server, app integration,
branding assets, or a Codex equivalent of Claude Code slash commands.

## Decision

### D1 - Advance the Codex skill-only native wrapper to v0.9.6

Ship `dist/codex/plugin/` as a first-party Codex plugin containing:

- `.codex-plugin/plugin.json`;
- `skills/aikata/SKILL.md`;
- `skills/aikata/agents/openai.yaml`.

The plugin-bundled `SKILL.md` and `agents/openai.yaml` stay
byte-identical to their canonical copies under `dist/universal-skill/`.
Repository tests enforce that invariant.

Cursor and Gemini CLI native wrappers remain deferred to v1.0. This ADR
refines ADR 0015 only for Codex, whose required distribution shape is now
stable enough and materially improves discoverability.

### D2 - Add minimal Codex App metadata

Add `dist/universal-skill/agents/openai.yaml` with:

- display name `aikata`;
- a concise UI description;
- a `$aikata` starter prompt;
- implicit invocation enabled.

Do not add speculative icons, colors, MCP dependencies, or branding
assets. The metadata is copied unchanged into the Codex plugin skill.

### D3 - Prefer the native plugin path on Codex CLI 0.135.0+

Expose the plugin from the repository root marketplace catalog at
`.agents/plugins/marketplace.json`. The preferred install flow is:

```bash
codex plugin marketplace add shigindo-inc/aikata --ref v0.9.6
codex plugin add aikata@aikata
```

Direct universal skill installation remains the fallback for older
Codex versions. Codex CLI `0.125.0` already discovers the resulting
`.agents/skills/aikata/SKILL.md`.

### D4 - Ship archive assets alongside the existing one-file skill

Keep the existing `aikata-universal-skill.md` release asset for backward
compatibility. Add generated archives:

- `aikata-universal-skill.tar.gz`;
- `aikata-codex-plugin.tar.gz`.

`scripts/package-distribution-assets.sh` creates both archives before
GoReleaser runs. CI checks that the archives contain the required files.

## Consequences

- Codex users get a native plugin install flow and Codex App metadata
  before v1.0 without changing the Go CLI.
- Existing direct skill installation remains valid for older Codex
  versions and other universal-skill consumers.
- The Codex plugin adds no executable code, remote dependency, MCP
  server, or project mutation. It satisfies ADR 0003's Do-No-Harm policy.
- Release preparation now keeps Claude and Codex plugin metadata in
  version lockstep.

## Alternatives Considered

- **Wait until v1.0.** Rejected: Codex's required skill-plugin surface is
  stable and the missing App metadata is a concrete discoverability gap.
- **Ship Codex-specific commands or an MCP server.** Rejected: the thin
  skill wrapper already covers the useful CLI invocation surface.
- **Add icons and colors now.** Rejected: aikata has no reviewed branding
  assets for the Codex card, so minimal metadata is the honest shape.
