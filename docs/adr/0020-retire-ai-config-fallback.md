---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0020 - Retire `.ai/aikata.yaml` config fallback

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0008 (aikata-owned config directory), ADR 0017
  (post-init command taxonomy)

## Context

ADR 0008 moved aikata-owned config from the generic `.ai/` namespace
to `.aikata/aikata.yaml` in v0.3.2. To avoid breaking early v0.x
projects, it kept a read fallback for `.ai/aikata.yaml` and a
best-effort migration path in `aikata generate` / `aikata doctor
--fix`.

That compatibility layer now costs more than it returns:

- The project is still pre-v1, so the stable compatibility contract
  has not started.
- The old `.ai/` name is exactly the collision risk ADR 0008 tried to
  avoid.
- The extra resolver shape leaks into multiple commands and tests
  even though new projects have written `.aikata/aikata.yaml` since
  v0.3.2.
- User feedback for pre-v1 cleanup favours a smaller command surface
  over compatibility aliases and transitional paths.

## Decision

v0.7.4 removes `.ai/aikata.yaml` support entirely from the current
runtime:

- `internal/config.Resolve` accepts only `.aikata/aikata.yaml`.
- `internal/config.Load` returns only the parsed config and an error;
  it no longer reports whether a legacy path was selected.
- The `MoveLegacyToPrimary` helper and its tests are deleted.
- `aikata generate` no longer reads, warns about, or migrates
  `.ai/aikata.yaml`.
- `aikata doctor` no longer emits or fixes `config.legacy-path`.
- Documentation keeps `.ai/` only as historical context and tells
  users to move `.ai/aikata.yaml` manually if they still have one.

Schema migration support stays in place. This ADR only removes the
old filesystem location, not the v1 → v2 `aikata.yaml` migrator.

## Consequences

**Positive**:

- The config path is unambiguous before v1.0.
- Runtime code and tests no longer carry a path-level compatibility
  branch.
- `.ai/` is left fully available for user-owned or third-party state.

**Negative**:

- Projects that never ran the v0.3.2+ migration must manually move
  `.ai/aikata.yaml` to `.aikata/aikata.yaml` before using v0.7.4+.
- The v0.3.2 roadmap and changelog entries remain historically true
  for that release, but no longer describe current behaviour.

## Alternatives Considered

- **Keep fallback until v1.0**: rejected because aikata is still in
  the v0.x cleanup window and the compatibility branch works against
  the clear namespace goal.
- **Keep doctor-only migration**: rejected because it would still
  preserve special `.ai/` handling in the runtime. The manual `mv`
  path is simple and visible.
