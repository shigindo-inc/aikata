---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-27
audience: [human, agent]
---

# ADR 0014 - `.aikata/manifest.yaml` is a Living Record

- **Status**: Accepted
- **Date**: 2026-05-27
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0011 (`aikata sync`
  design), ADR 0008 (`.aikata/` config namespace)

## Context

ADR 0011 D4 introduced `.aikata/manifest.yaml` as the SHA-256 record of
every template-derived file at `aikata init` time, used as the common
ancestor in `aikata sync`'s 3-way merge. The wording framed the
manifest as an **init-time** snapshot.

Reality contradicted that framing within one release:

- `aikata add <component>` (memory / ui / api / tdd / changelog /
  stack) writes new template-derived files post-init. v0.6.0–v0.6.1
  did **not** update the manifest when these ran. The next `aikata
  sync` therefore saw the just-added files as `!hadAncestor &&
  hasCurrent && hasUpstream` → `StatusUpstreamAdded` (auto-applied) or
  worse `StatusConflict` if the user had already customised them.
- The same gap also explains why `inferFlags`
  (`internal/sync/plan.go`) could not pick up post-init opt-ins: there
  was nothing to infer from when the manifest never grew.

The framing of "init-time snapshot" needs to be replaced with
something that holds across all aikata commands that write
template-derived files.

## Decision

**`.aikata/manifest.yaml` is the living record of every
template-derived file aikata has authored in this project.** It is
updated by:

1. **`aikata init`** — writes the manifest as part of the initial
   scaffold (existing behaviour from ADR 0011).
2. **`aikata add <component>`** — appends the freshly-rendered
   entries via the new `components.RecordInManifest` helper. Hashes
   are the SHA-256 of the **template-rendered** content, not of the
   on-disk file. Files the user customised before running `add` keep
   their content; the manifest only records what aikata's template
   says they "should" be — which lets the next sync classify the
   diff as `user-only-edit`.
3. **`aikata sync`** — regenerates the manifest from the post-merge
   on-disk state when the merge completes without conflict
   (existing behaviour). Conflict-resolution runs leave the manifest
   stale until the user resolves and re-syncs.

The manifest is **never** edited by hand. `aikata generate`
artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`, …) are explicitly
out of scope (ADR 0011 D1) and do not appear in it.

### Invariants

- Every `Path` in `manifest.Files` was rendered by some `aikata` code
  path from an embedded `*.tmpl`. User-authored ADRs, `docs/plans/`,
  per-project READMEs, etc. are **not** in the manifest.
- `SHA256` is the digest of "what the template version of this file
  is at the time we wrote / appended this entry" — which is the
  ancestor `aikata sync` will compare against. It is **not** the
  digest of the file currently on disk (those can drift when the
  user customises).
- The manifest is `MarshalManifest`-deterministic: entries sorted by
  `Path` ascending, the schema's required fields populated. Repeated
  `add` invocations against the same component produce byte-identical
  manifest output (idempotency).

### Edge cases

- **`aikata add` against a directory with no `.aikata/aikata.yaml`** —
  the project was never init'd by aikata. `RecordInManifest` no-ops
  silently (minting a manifest into someone else's project would
  violate Do-No-Harm, ADR 0003). The file write still happens, so
  legacy behaviour is preserved.
- **`aikata add` against a v0.4.x project with `.aikata/aikata.yaml`
  but no `.aikata/manifest.yaml`** — seed a fresh manifest with
  `preset: standard`, `lang` from `aikata.yaml`. The same defaults
  the rebaseline path uses for legacy projects. Subsequent
  `aikata sync` proceeds normally; users who want the manifest's
  preset to differ from `standard` can run `aikata sync --rebaseline`
  beforehand.
- **`aikata add` overwriting a manifest entry** — if the user runs
  `add ui` twice (the second is a no-op for the file write because
  `writeIfMissing` skips), the manifest entry is rewritten with the
  current template's hash. This is correct: aikata's template may
  have evolved between v0.6.x releases, and the manifest should
  reflect the latest known ancestor.

## Consequences

### Positive

- The bug "`aikata add ui` followed by `aikata sync` overwrites the
  user's UI.md customisations" is structurally impossible after this
  change.
- `inferFlags` (manifest path → `--with-*`) can now reliably detect
  post-init opt-ins. v0.6.3 builds on this to extend scope
  derivation (ADR 0013).
- The manifest stays the **single source of truth** for the 3-way
  merge ancestor; no per-component side-channels needed.

### Negative

- One more thing to keep idempotent: every new component author must
  remember to call `RecordInManifest` after the file-write step. The
  existing six components flow through one of three call sites
  (`memory.go`, `singlefile.go`, `stack.go`), so the surface is
  small.
- Adding a component to a directory that lacks `.aikata/aikata.yaml`
  is now a soft failure: the file is written but the manifest is
  not, and a later `aikata sync` will treat the file as
  `upstream-added`. This is consistent with v0.6.x's overall stance
  ("sync only manages aikata-init'd projects") but worth calling out
  in component documentation.

### Out of scope (deferred)

- **Per-component opt-in flags in `aikata.yaml` schema** (e.g.
  `features.ui: true`). Tracked for v0.7+ as a schema v2 migration;
  until then, the manifest's path list is the only source for
  `inferFlags`.
- **Migrating a v0.4.x project to v0.6.x without running `aikata
  sync --rebaseline`** — if the user runs `aikata add ui` first, the
  fresh manifest will be seeded with `preset: standard`, which may
  not match the project's actual init-time preset. The user can fix
  this by editing the manifest's `preset:` field by hand or by
  running `aikata sync --rebaseline` afterwards.

## Implementation map

- `internal/components/component.go` — exports
  `RecordInManifest(targetDir, rendered)` and documents its
  contract.
- `internal/components/{memory,singlefile,stack}.go` — call
  `RecordInManifest` immediately after the file-write step.
- `internal/components/manifest_record_test.go` — asserts
  manifest growth, idempotency, preservation of prior entries,
  no-op behaviour without `aikata.yaml`, and stack guide
  registration.
- `internal/config/manifest.go` — unchanged; the existing
  `BuildManifest` / `SaveManifest` API is the natural fit.
