---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0016 - `.aikata/aikata.yaml` schema v2: explicit component flags

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0011 (`aikata sync`
  design, D3 migration framework), ADR 0013 (Sync scope derivation),
  ADR 0014 (Manifest as living record), Q-DESIGN-09

## Context

Through v0.6.x the optional-component scope of an aikata project lived
in three different places:

- `features.tdd` and `features.monorepo` in `.aikata/aikata.yaml`
  (legacy v1 schema).
- Path heuristics in `internal/sync/plan.go` `inferFlags` (the
  manifest told sync which components a project had, indirectly).
- Nowhere at all for `memory`, `ui`, `api`, `changelog` — only the
  manifest's file list signalled them.

ADR 0014 made the manifest reliable as the inference source, and
v0.6.3 / ADR 0013 made sync's scope-derivation hierarchy explicit
(defaults → aikata.yaml → manifest → CLI overrides). The remaining
gap is that the declarative source still leaks scope information
through path inference. Q-DESIGN-09 ("`.aikata/aikata.yaml` schema
v2") captured the resulting plan: lift the implicit signals into
first-class schema fields so post-init commands can read intent from
one declarative source.

## Decision

`.aikata/aikata.yaml` gains an explicit `components:` block at schema
**v2**:

```yaml
version: 2
project:
  name: ...
  lang: ...
ai_tools: [...]
stacks: [...]
components:
  memory:    false
  ui:        false
  api:       false
  tdd:       false
  changelog: false
  monorepo:  false
features:
  obsidian_hints: false
docs:
  generate_gitignore: true
  task_file_location: docs/tasks/current.md
```

`tdd` and `monorepo` move out of `features:` and into `components:`,
because they affect template scope (which files are rendered) rather
than project-style toggles. `features:` becomes the home for
non-scope-affecting flags only (`obsidian_hints` today; future
ergonomic toggles).

### Migration (v1 → v2)

A v1 → v2 migrator registers in
`internal/config/schema_migrate.go` per the framework introduced by
ADR 0011 D3. The migrator:

1. Reads `features.tdd` and `features.monorepo`; lifts them into
   `components.tdd` and `components.monorepo`.
2. Removes the two lifted keys from `features:`. Other keys
   (`obsidian_hints`, future flags) are left untouched.
3. Emits the full six-field `components:` block. Components that v1
   had no schema slot for (`memory`, `ui`, `api`, `changelog`) are
   seeded as `false`; `aikata sync`'s manifest path inference
   (`inferFlags`) continues to detect them at runtime so legacy
   projects do not lose their scope at the migration boundary.
4. Stamps `version: 2`.

The migrator runs lazily via `LoadMigrated`: a v1 file is only
rewritten when a writer (`aikata generate`, `aikata sync`,
`aikata doctor --fix`) saves it back. Read-only callers continue to
see v1 payloads in memory until then. The struct's zero-valued
`Components` field is a safe no-op under the OR-merge described
below.

### Sync scope derivation

`internal/sync/plan.go` `derivePlan` adds `cfg.Components.*` to its
OR-merge step, alongside the legacy `features.tdd` /
`features.monorepo` reads. Schema-v2 fields are the canonical signal;
the legacy `features.*` keys remain readable so projects whose config
has not yet been migrated by a writer continue to compute correct
scope. Manifest path inference (`inferFlags`) likewise stays as a
safety net.

The full priority order is unchanged from ADR 0013:

```
CLI overrides > manifest > aikata.yaml > defaults
```

`cfg.Components.*` and `cfg.Features.*` are both within the
"aikata.yaml" tier and OR-merge with `inferFlags(ancestor)` in the
manifest tier.

### Default policy

ADR 0003 (do-no-harm): defaults are all `false`. A `--preset minimal`
or default `--preset standard` user sees a six-field `components:`
block with every key explicitly `false`, mirroring the existing
`features:` block convention so the schema is discoverable from the
freshly-generated file.

## Consequences

### Positive

- Post-init commands (v0.7.1 `aikata enable`) have a single,
  declarative source of truth for "is this capability on for the
  project". No more reading manifests-as-a-database to answer the
  question.
- `aikata.yaml` becomes self-describing: a human reading the file can
  enumerate every capability the project has without cross-checking
  the manifest or running `aikata doctor`.
- The two-place lift removes the awkward "is `tdd` a feature or a
  component?" question — `features:` is now consistently "non-scope
  ergonomic toggles" and `components:` is "template scope".

### Negative

- Existing projects keep working but their on-disk config silently
  drifts to v2 on the next writer run. The `chore(release): prepare
  v0.7.0` PR body must call this out so users who diff their
  `aikata.yaml` aren't surprised.
- Two readers for "is tdd / monorepo on" (`Components.TDD` plus
  `Features["tdd"]`) for the duration of v0.x. The legacy reader is
  scheduled to drop at v1.0.

### Out of scope (deferred)

- **Validation** that `Components.<X>` matches the actual on-disk
  files (e.g. `Components.Memory == true` but `docs/memory/` absent
  → warning). This is `aikata doctor` territory; the schema bump
  itself stays mechanical.
- **`aikata doctor --fix` for schema migration** explicitly. The
  lazy LoadMigrated path covers the same ground; `--fix` already
  triggers a write via the existing rewrite pass.
- **Validation that legacy `features.{tdd,monorepo}` and v2
  `components.{tdd,monorepo}` agree** when both exist. The
  OR-merge means a mismatched pair never produces wrong scope; a
  cosmetic warning could land in v0.7.x or later if real projects
  hit the case.

## Implementation map

- `internal/config/aikata_yaml.go` — `Components` struct added,
  `AikataYaml.Components` field, `Version` bumped to `2`, `Default()`
  initialises the new block.
- `internal/config/schema_migrate.go` — `migrateV1ToV2` and helper
  primitives for raw YAML node mutation.
- `internal/sync/plan.go` — `derivePlan` OR-merges
  `cfg.Components.*` alongside the legacy `features.*` keys.
- `internal/scaffold/scaffold.go` — `addPresetArtifacts` persists
  the new components block from `Options.With*` flags so fresh
  scaffolds record their opt-ins.
- `testdata/golden/*/.aikata/aikata.yaml` — every fixture rewritten
  to the v2 shape; the schema bump is part of the golden tree.
- `internal/config/schema_migrate_test.go` — new tests cover the
  v1 → v2 lift (with and without `features:`), idempotency, and
  the persisted-file post-condition.
- `internal/sync/derive_test.go` — new tests cover the OR-merge
  semantics for schema-v2 components and the legacy `features.*`
  fallback.
