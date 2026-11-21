---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-21
audience: [human, agent]
---

# ADR 0050 — Doc map off by default

- **Status**: Accepted
- **Date**: 2026-11-21
- **Related**: ADR 0044 (doc map as a derived artifact; its D5 "rebuild on
  every lifecycle command" now applies only when enabled)

## Context

ADR 0044 made every lifecycle command (`init`, `fill`, `enable`, `sync`,
`generate`) rebuild `.aikata/docmap.{yaml,md}`, and `doctor` warn when the
map drifts. In a downstream project with 518 documents the map reached
about 880 KB. No rule, hook, Makefile target, or skill in that project read
it; 49 commits touched it, and `doctor` kept warning that it was out of
date. Two smaller projects showed the same pattern: worktree-local
rebuilds produced map-only diffs that triggered automated reviews.

The only reader in aikata itself is the `migrate-structure` skill, which
runs `aikata map` and reads the result in the same session.

## Decision

1. A new `docmap.enabled` key in `.aikata/aikata.yaml` turns the map on.
   It defaults to off.
2. When off, lifecycle commands do not rebuild the map and `doctor` does
   not check its freshness.
3. `aikata map` builds the map on demand whether or not it is enabled, so
   `migrate-structure` keeps working.
4. When off but map files exist, `doctor` warns (`docmap.unmaintained`)
   that nothing keeps them current, and names both ways out: enable the
   key, or delete the files. `--fix` does neither; the choice is the
   owner's.

There is no compatibility branch: an absent key means off, including for
projects that already have a map.

## Consequences

- Breaking for projects that rely on the automatic rebuild: they add
  `docmap: {enabled: true}` once. The `docmap.unmaintained` warning tells
  them on the first `doctor` run after upgrading.
- Projects that never read the map stop paying for it (repository size,
  diff noise, a standing doctor warning).
- aikata's own repository enables it, because `docs/layout.md` and the
  `migrate-structure` skill describe it against this repository.

## Alternatives Considered

- Keep it on and make it smaller (drop the link graph). The cost falls,
  but there is still no reader.
- Keep it on with an opt-out key. Projects that do not know the map exists
  keep paying, which is the situation that prompted this ADR.
- Infer "enabled" from a non-empty `docmap:` block. An implicit rule is
  harder to explain than one key.
