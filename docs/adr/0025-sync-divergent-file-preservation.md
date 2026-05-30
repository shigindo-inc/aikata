---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-30
audience: [human, agent]
---

# ADR 0025 - `aikata sync` preserves intentionally-divergent files

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: aikata maintainers
- **Related**: ADR 0011 (`aikata sync` design — defines the 3-way merge
  and the "Rebaseline ancestor choice" principle this ADR extends to
  the post-clean-run regeneration path), ADR 0013 (scope derivation),
  ADR 0014 (manifest is a living record), ADR 0018 (managed-append for
  project-owned files such as `.gitignore`), ADR 0019 (sync
  missing-file repair semantics — no silent deletes), ADR 0021
  (`doctor.exclude` glob list — the config-shape precedent reused by
  `sync.own`)

## Context

A downstream dogfooder (`itteco`, a Flutter app on aikata `v0.8.1`)
reported that `aikata sync` cannot durably preserve a preset-managed
file that the user has intentionally rewritten for their project
(`README.md`, `.gitignore`, `docs/tasks/current.md`). The report is
reproducible against the shipped code; four distinct problems combine
into one bad experience.

### Problem 1 — re-baseline oscillation (data loss)

The 3-way merge itself is correct: a user-edited / upstream-unchanged
file classifies as `user-only-edit` and the user's bytes are preserved
on that run. The defect is in **what the manifest records afterwards**.

On a conflict-free run, `internal/sync/sync.go` rebuilds the manifest
from `postMergeSnapshot`, which re-reads the **on-disk bytes** of every
path. For a `user-only-edit` file the on-disk bytes are the *user's*
content, so the new ancestor hash becomes the user's content. The next
sync then sees `current == ancestor` and `upstream != ancestor`,
classifies the file as `upstream-applied`, and **silently overwrites
the user's content with the generic template**. Iterating sync thus
oscillates:

```
run N:   user-only-edit   (preserved)  → manifest ancestor := user bytes
run N+1: upstream-applied (overwritten) → manifest ancestor := upstream bytes
```

The decisive evidence in the report is the manifest's per-file
`sha256` flipping from the upstream hash to the user-content hash after
a "0 applied / exit 0" run, then the next run reporting `3 applied`
with no upstream change between the two runs.

The irony is that ADR 0011 already documents the correct rule under
*Rebaseline ancestor choice*: the manifest ancestor must be **the
upstream rendering, not the on-disk bytes**, precisely because
recording disk bytes "would cause the next sync to treat them as 'user
has no edits' … silently overwriting the user's customisations." The
`--rebaseline` path obeys that rule; the post-clean-run regeneration
path (`postMergeSnapshot`) contradicts it. The two manifest-write paths
disagree, and the disagreement is the bug.

### Problem 2 — `docs.generate_gitignore: false` is a dead flag

`config.Docs.GenerateGitignore` is defined and defaulted to `true` but
is **never read** anywhere outside tests (`rg GenerateGitignore`
confirms). Setting `generate_gitignore: false` therefore has no effect:
`.gitignore` is always rendered, recorded in the manifest, and treated
as a sync-managed file. The flag's stated intent (do not let aikata
manage `.gitignore`) and the actual behaviour diverge.

### Problem 3 — `--rebaseline` is a silent no-op when a manifest exists

The rebaseline path is gated on `!manifestPresent`. Passing
`--rebaseline` to a project that already has `.aikata/manifest.yaml`
falls straight through to the normal merge with no indication the flag
was ignored. There is also no supported way to re-seed an existing
manifest from the current upstream rendering.

### Problem 4 — manifest schema is `v1`, config schema is `v2`

`.aikata/aikata.yaml` migrates to `version: 2` (ADR 0016) but
`.aikata/manifest.yaml` stays at `version: 1`. The reporter flagged the
generation skew as possibly unintentional.

## Decisions

D1 is the data-loss root-cause fix and is independent of the rest: it
must ship even if D2 slips. D2 is a complementary UX feature whose
problem only exists **after** D1 removes the data loss. D3 and D4 are
small, self-contained corrections. The Problem 4 skew is a clarifying
note, not a decision (see Consequences).

### D1 — Post-clean-run manifest regeneration records the upstream rendering, not the on-disk snapshot

On a conflict-free run, regenerate the manifest by hashing the
**in-memory upstream rendering** for every path — the same content
`aikata init` and `aikata sync --rebaseline` already record — instead
of re-reading the post-merge on-disk bytes. Concretely: drop
`postMergeSnapshot` and build the new manifest directly from the
`upstream` map.

This unifies the two manifest-write paths under the single principle
ADR 0011 already states for rebaseline: **the ancestor is the upstream
rendering**. The result per status:

| Status | New ancestor under D1 | Effect |
| --- | --- | --- |
| `unchanged` | upstream rendering | identical to today |
| `upstream-applied` | upstream rendering | identical (disk == upstream) |
| `both-match` | upstream rendering | identical (disk == upstream) |
| `upstream-added` | upstream rendering | identical (disk == upstream) |
| `user-only-edit` | upstream rendering | **fixed** — ancestor stays at upstream, so the next sync re-classifies the file as `user-only-edit` and preserves it indefinitely |
| `user-deleted` | upstream rendering (entry kept) | **fixed** — see below |

For `user-only-edit` the new ancestor equals the *old* ancestor (a file
is classified `user-only-edit` exactly because `upstream == ancestor`),
so the user's divergence is preserved on every subsequent sync with no
oscillation and no silent overwrite. For every other written or
unchanged status, the on-disk bytes already equal the upstream
rendering after the run, so the recorded ancestor is unchanged from
today. **The only behavioural change versus the current code is the
`user-only-edit` path — which is exactly the bug.**

The reframing also closes a latent `user-deleted` defect for free:
`postMergeSnapshot` drops deleted paths from the manifest, so the next
sync sees the path as new-to-upstream and re-creates the file the user
deleted. Recording the upstream rendering keeps the manifest entry, so
the deletion stays respected across syncs (consistent with ADR 0019's
no-silent-resurrection stance). This is a noted side effect of the
unified principle, not a scope expansion.

### D2 — Per-file `owned` opt-out via `aikata.yaml` `sync.own:`

D1 stops data loss but leaves a residual: a file the user has *fully*
forked (e.g. `README.md`) still produces conflict markers whenever
**upstream** changes that same file, because all three sides diverge.
For files the user has deliberately taken ownership of, conflict
display is noise, not signal.

Add an optional top-level block to `.aikata/aikata.yaml`:

```yaml
sync:
  own:
    - README.md
    - .gitignore
    - docs/tasks/current.md
```

`sync.own` is a glob list (same matcher and additive semantics as
`doctor.exclude`, ADR 0021 — no schema bump; an absent block is the
empty list). A path matching `sync.own` is reported with a new
`owned` status and is **never** rendered-compared, conflict-markered,
or overwritten by `aikata sync`. Ownership is declared in the
user-facing config (`aikata.yaml`), not in the machine-owned manifest,
because ADR 0011 forbids hand-editing `.aikata/manifest.yaml`.

This is the durable, declarative replacement for the reporter's manual
`git restore <file>` two-pass workaround.

### D3 — Remove the dead `docs.generate_gitignore` flag

`config.Docs.GenerateGitignore` is removed rather than wired up. The
field never affected behaviour (it is defined, defaulted to `true`, and
never read), so removal is a behavioural **no-op** — `.gitignore` is
already, and continues to be, rendered and routed through the ADR 0018
managed-append writer, which owns only the aikata block and preserves
user-owned lines. aikata simply stops advertising a knob that did
nothing.

The original intent of the flag — "do not let aikata manage
`.gitignore` at all" — is now served by two existing, more general
mechanisms instead of a single-purpose field:

- ADR 0018 managed-append already makes the default non-destructive, so
  the common reason to disable generation (fear of clobbering hand-
  written entries) no longer applies.
- A user who wants `aikata sync` to never touch `.gitignore` lists it
  under `sync.own` (D2) — the same general opt-out used for any other
  file.

Old configs that still carry `generate_gitignore:` parse without error
(the unknown key is ignored); no schema bump or migration is required
because the field was inert.

### D4 — `--rebaseline` is explicit when a manifest is present; add `--reseed`

Passing `--rebaseline` to a project that already has a manifest emits a
notice instead of silently running a normal merge — for example:

```
rebaseline skipped: .aikata/manifest.yaml already present (use --reseed to re-seed from the current upstream rendering)
```

A new `--reseed` flag re-seeds an existing manifest from the current
upstream rendering (the same ancestor choice as the no-manifest
rebaseline path and as D1), for the rare case where a project's
baseline has drifted and the user wants to re-anchor it deliberately.
`--reseed` writes only `.aikata/manifest.yaml`; it never modifies
source files.

## Consequences

- The oscillation / silent-overwrite data-loss path is removed by a
  change to a single manifest-write site; no merge-classification logic
  changes. `user-only-edit` files are preserved across unlimited sync
  runs.
- aikata gains a first-class way to say "this file is mine now"
  (`sync.own`), aligned with the existing `doctor.exclude` precedent and
  requiring no schema bump.
- The inert `generate_gitignore` flag is gone; users who want aikata to
  leave `.gitignore` alone use `sync.own` (D2), and the ADR 0018
  managed-append default already preserves their hand-written lines.
  One opt-out mechanism, not two.
- **Manifest schema skew (Problem 4) is intentional and is recorded as a
  clarification, not a decision.** `ManifestVersion` and the
  `aikata.yaml` `version` field version **independently**: the manifest
  is an internal, regenerated, machine-owned record (ADR 0014) and the
  config is a user-owned, migrated document (ADR 0016). There is no
  required lockstep and no code change; the skew is documented so it
  stops reading as a bug. No `doctor` check is added.
- **Deferred to the implementation PR** (per the ADR 0024 precedent of
  not editing live reference docs in a planning pass): GLOSSARY gains an
  `owned` (sync status) term and a `sync.own` reference — `aikata doctor`
  runs a glossary-consistency check, so this is normative once the code
  lands; SPEC §4 (`aikata sync`) documents the `owned` status,
  `--reseed`, and the `generate_gitignore` behaviour; README / `docs/`
  help and examples are aligned. Shipped ROADMAP / CHANGELOG history is
  left intact.

## Alternatives considered

- **Keep `postMergeSnapshot`, special-case `user-only-edit` to retain
  its prior ancestor.** Functionally identical to D1 but keeps two
  manifest-write paths with divergent rules and more branching. D1's
  "ancestor = upstream rendering everywhere" is the same outcome with
  less code and one principle shared with `--rebaseline`. Rejected as
  strictly more complex.
- **`owned` marker without D1.** Insufficient: data loss persists for
  every edited-but-not-owned file, which is the common case (a user who
  tweaks two lines of `AGENTS.md` has not opted into `sync.own`). D1 is
  the necessary fix; D2 only removes residual conflict noise for fully
  forked files.
- **Per-file `owned: true` flag inside `manifest.yaml`.** Rejected:
  ownership is user intent and belongs in the user-editable
  `aikata.yaml`; the manifest is explicitly not hand-editable (ADR
  0011).
- **Honor `generate_gitignore: false` instead of removing it** (wire
  the flag up: exclude `.gitignore` from render and manifest, prune a
  stale entry). Rejected in favour of the simpler route: the flag was
  always dead, so wiring it up *adds* a code path and a second way to
  express "don't manage this file," which `sync.own` (D2) already
  covers generally. Removing the inert field is a behavioural no-op and
  leaves one opt-out mechanism instead of two. Fewer knobs, less to
  reconcile.
- **Bump the manifest schema to `v2` for parity with the config.**
  Rejected: a version bump with no field change is churn; the two
  schemas are independent by design (see Consequences).

## Implementation map

- `internal/sync/sync.go` — drop `postMergeSnapshot`; build the
  post-clean-run manifest from the `upstream` map (D1). Add `owned`
  classification gated on `sync.own` before the 3-way merge (D2). Add
  the `--rebaseline`-present notice and `--reseed` path (D4).
- `internal/config/aikata_yaml.go` — optional `sync.own []string`
  block (D2), mirroring the `doctor.exclude` shape.
- `internal/config/aikata_yaml.go` — remove the inert
  `Docs.GenerateGitignore` field (D3); confirm unknown-key tolerance so
  old configs carrying `generate_gitignore:` still parse.
- `internal/cli/sync.go` — `--reseed` flag; the manifest-present
  `--rebaseline` notice (D4).
- `internal/sync/sync_test.go` — regression tests: the two-run
  oscillation no longer overwrites (D1), `user-deleted` stays deleted
  across two syncs (D1 side effect), `owned` paths are skipped (D2),
  `generate_gitignore: false` removes tracking (D3), `--rebaseline`
  with a manifest emits the notice and `--reseed` re-anchors (D4).
- GLOSSARY / SPEC / README — deferred to the implementation PR (see
  Consequences).
