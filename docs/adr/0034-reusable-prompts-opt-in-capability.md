---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
audience: [human, agent]
---

# ADR 0034 - Move the reusable-prompts library to an opt-in capability

- **Status**: Accepted
- **Date**: 2026-06-01
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0016 (aikata.yaml
  schema v2), ADR 0017 (post-init command taxonomy), ADR 0019 (`sync`
  missing-file repair semantics), ADR 0028 (prioritize core-concept
  stabilization)

## Context

`docs/prompts.md` shipped as a **default** file in the `standard`,
`flutter`, and `typescript` scopes (en + ja) through v0.9.1. The
generated document is an **empty skeleton**: a format explainer plus a
"_No prompts yet_" placeholder. It carries the standard five-key
frontmatter, so every scaffolded project pays its maintenance cost —
`doctor`'s `updated:` check, `sync` tracking, and a slot in the README /
ARCHITECTURE navigation — for a file with no content until the user fills
it in.

The v0.9.0 core-concept stabilization review
([ADR 0028](./0028-prioritize-core-concept-stabilization.md)) asked
whether each default-scaffolded file has a distinct role that repeatedly
saves context-reconstruction or maintenance work (Q-DESIGN-12). For an
empty reusable-prompt library the answer is no by default: a project that
has not yet validated any prompt gains nothing from the skeleton, while
every project pays the upkeep. A reusable-prompt library is genuinely
useful — but only once a team has prompts worth keeping, which is exactly
the signal an opt-in expresses.

"Generate by default; delete if unwanted" is not a valid opt-out under
aikata's maintenance model: a managed file the user deletes participates
in `sync` missing-file repair (ADR 0019), so deletion does not express a
durable opt-out.

## Decision

### D1 - `docs/prompts.md` becomes an opt-in single-file capability

Move the reusable-prompt library out of the default `standard` /
`flutter` / `typescript` scopes and behind the same opt-in surface as
`ui` / `api` / `tdd` / `changelog`:

- `aikata enable prompts` — post-init.
- `aikata init --with-prompts` (and the matching interactive prompt) —
  at scaffold time.

The capability template lives at `components/prompts/{en,ja}/` and renders
the same `docs/prompts.md` content as before. Enabling it records the
file in `.aikata/manifest.yaml` (ADR 0014) and flips the schema-v2
`components.prompts` flag, so `aikata sync` and post-init commands read
one declarative source.

### D2 - Schema: a new `components.prompts` boolean, no version bump

`components.prompts` is added as a new boolean field defaulting to
`false`. Adding a defaulted boolean is backward-compatible: pre-v0.9.2 v2
configs omit the key and read as `false`, and the v1→v2 migrator seeds
`prompts: false`. The schema version stays at 2 (no incompatible change).

### D3 - Existing projects keep their `docs/prompts.md`

This is sync-visible but **non-destructive** (ADR 0019): on an existing
project whose manifest still lists `docs/prompts.md`, the file is no
longer part of the default upstream rendering, so `aikata sync` treats it
as user-owned rather than deleting it. A user who wants the file tracked
as an aikata-managed capability again runs `aikata enable prompts`; a user
who no longer wants it deletes it and the deletion now sticks. The CHANGELOG
records the migration note.

## Consequences

- The default `standard` scope is leaner: it no longer ships an empty
  document whose value is purely latent.
- Projects that want a prompt library opt in with one command, and it is
  then a first-class managed capability (manifest-tracked, sync-preserved).
- The change is sync-visible to existing projects, but no file is deleted
  (ADR 0019). The only behavioral change for an existing project is that
  `docs/prompts.md` stops receiving upstream updates — acceptable, since
  the upstream content was an inert skeleton.
- One more capability appears under `aikata list capabilities` and one
  more `--with-*` flag on `aikata init`; the surface stays uniform with
  the existing single-file components.

## Alternatives Considered

- **Remove `docs/prompts.md` entirely** (no capability). Rejected: the
  reusable-prompt library is a legitimately useful document for teams that
  have validated prompts; removing it with no opt-in would strand them.
- **Keep it a default.** Rejected: an empty skeleton imposes maintenance
  cost on every project for latent value, which is the Q-DESIGN-12
  concern.
- **Gate it behind the future `extended` scope.** Rejected: `extended` is
  the operational-readiness governance pack, not a place for an
  authoring-convenience document; a per-capability opt-in is the right
  granularity and matches `ui` / `api` / `tdd` / `changelog`.
