---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-02
audience: [human, agent]
---

# ADR 0038 - Unify `.gitignore` on the managed block across init and sync

- **Status**: Accepted
- **Date**: 2026-06-02
- **Deciders**: aikata maintainers
- **Related**: ADR 0018 (managed-block append for project-owned files),
  ADR 0011 (`aikata sync` design), ADR 0019 (sync missing-file repair
  semantics), ADR 0025 (sync divergent-file preservation / `sync.own`),
  ADR 0037 (tighten adoption mutation boundaries). Closes the
  Q-INTEROP-04 (a) leading position.

## Context

`.gitignore` is the one project-owned generic file aikata only
contributes a small section to. ADR 0018 introduced the managed-block
writer (`# >>> aikata managed >>>` … `# <<< aikata managed <<<`) so
`aikata init` could splice that section in without owning the whole
file. But the two commands that touch `.gitignore` used **two different
mechanisms**:

- `aikata init` wrote it through the managed-block writer
  (`internal/managed`), refreshing only the marked block.
- `aikata sync` classified it with the **generic 3-way merge** (write
  discipline #7, `internal/sync`) — the same path used for prose like
  `SPEC.md`. That merge compares the whole file against the manifest
  ancestor and, on divergence, can emit git-style `<<<<<<<` conflict
  markers.

Two consequences followed. First, a future change to aikata's
`.gitignore` block could land **conflict markers inside a user's
`.gitignore`** on `aikata sync` — exactly the ownership overreach
ADR 0037 pushed back against, just on a different code path. Second,
there was a representation split: a **fresh** `aikata init` wrote the
block **without** the markers (`scaffold.contentForWrite` returned the
raw rendered template; only `init --force` over an existing file added
markers, because that branch ran `ApplyBlock`). A markerless on-disk
file is the trap — if `sync` were naively switched to call `ApplyBlock`
on it, the writer would find no markers and **append a second, framed
copy**, duplicating every rule.

Q-INTEROP-04 recorded the leading position to unify on managed-append
(a) and to **never** managed-append into prose (b). This ADR implements
(a). (b) needs no new code beyond a guard test; see Consequences.

## Decision

Route `.gitignore` through the managed block in **both** `init` and
`sync`, on a single shared representation.

**D1 — Always frame the block (one representation).** A fresh
`aikata init` now writes the **framed** form (markers + body), identical
to what `init --force` and `sync` produce. `internal/managed` exposes
`Frame(body []byte) []byte` (the standalone framed block, equivalent to
`ApplyBlock` against empty content), and `scaffold.contentForWrite`
calls it for managed-append paths on fresh writes. The on-disk file
therefore *always* carries the markers — self-documenting (ADR 0018)
and the precondition for a clean in-place refresh on sync.

**D2 — Single shared path list.** The set of managed-append paths lives
in `internal/managed` as `IsAppendPath(rel string) bool` (today exactly
`{".gitignore"}`). Both `internal/scaffold` and `internal/sync` call it,
so the two stay in lockstep and there is one tested place to reason
about membership.

**D3 — sync refreshes the block, never conflicts.** In
`sync.classifyAndMerge`, a managed-append path is special-cased
**before** the generic hash 3-way:

- file present on disk → `merged := managed.ApplyBlock(current,
  upstreamBody)`; if `merged == current` → `unchanged` (no write), else
  `upstream-applied` + write `merged`. Only the aikata block is
  replaced; user lines outside the markers are byte-preserved.
- file absent + previously in the manifest → `user-deleted` (respect the
  deletion, consistent with ADR 0019).
- file absent + not in the manifest → `upstream-added` + write the
  framed standalone block.

A managed-append path **never** produces conflict markers, by
construction.

**D4 — The manifest hash is not consulted in steady state, with one
narrow migration exception.** The manifest still records `.gitignore`,
but in steady state the sync branch bypasses the hash comparison. This
is deliberate and **not a bug to fix**: the on-disk file carries the
framed block (markers + body) while the manifest records the **raw**
upstream body, so a hash compare would always mismatch. `ApplyBlock`
idempotency (`merged == current`) is the real "did anything change"
signal, so `saveManifestFromUpstream` / `BuildManifest` need no change.

The exception is the **legacy migration** (D5). The ancestor hash is the
*only* signal that tells "pristine pre-0.9.8 aikata output, safe to
replace" apart from "user content," so the migration branch — and only
it — compares `currentHash == ancestor`.

**D5 — Migrate pre-0.9.8 markerless files in place, never duplicate.**
A project scaffolded by v0.9.7 or earlier has a **markerless**
`.gitignore` (fresh init returned the raw template before this ADR).
Running `ApplyBlock` on a markerless file finds no markers and appends a
**second** framed copy — silently doubling every rule on the most common
upgrade path, even when the template is byte-identical. So before
falling through to `ApplyBlock`, the sync branch special-cases
`!managed.HasBlock(current)`:

- **pristine legacy** (`hadAncestor && currentHash == ancestor`):
  replace the file wholesale with `managed.Frame(upstreamBody)` — a
  clean in-place migration to the framed form, no duplication, status
  `upstream-applied` (or `unchanged` if already equal).
- **diverged / user-edited legacy** (`currentHash != ancestor`): fall
  through to `ApplyBlock`. This appends our framed block and leaves the
  user's lines — including any old aikata rules they have since edited
  around — untouched. Appending is the data-preserving choice when we
  cannot prove the content is still ours; a wholesale replace there
  would risk discarding user edits.

`sync.own` (ADR 0025 D2) continues to short-circuit *before*
classification, so an owner who lists `.gitignore` under `sync.own`
keeps full control — managed-append is subordinate to `owned`.

## Consequences

- `aikata sync` can refresh aikata's `.gitignore` block on every future
  release without any risk of conflict markers in a user's file. This is
  the property ADR 0037 wanted for the adoption story, now true on the
  sync path too.
- Every `testdata/golden/**/.gitignore` gains the marker lines
  (regenerated via `make update-golden`); this is a cosmetic,
  self-documenting change to fresh output.
- Write discipline #3 in ARCHITECTURE §3.4 now governs **both** init and
  sync (previously init-only); the table and the manifest-tracking
  modifier note are updated accordingly.
- **Q-INTEROP-04 (b) is enforced by a guard test**: `IsAppendPath` is
  asserted to contain only `.gitignore` and reject prose
  (`CONTRIBUTING.md`, `SECURITY.md`, …). Splicing a managed block into
  prose was rejected — prose merge is context-sensitive and reintroduces
  the ownership drift ADR 0037 fought. Prose stays create-if-missing
  one-shot scaffolds (write discipline #4). Adding a prose path to the
  list now fails the test and forces a decision review.
- The manifest-hash-irrelevant note (D4) is recorded here and in
  ARCHITECTURE §3.4 so a future reader does not "fix" the harmless
  mismatch and accidentally reroute `.gitignore` back through the 3-way.

## Alternatives considered

- **Keep fresh init raw; only `ApplyBlock` when `HasBlock(current)`.**
  Avoids the golden churn but leaves the representation split: the first
  divergent sync would introduce markers as a one-time surprise, and the
  sync logic carries a `HasBlock` branch forever. Rejected — messier and
  surprising; the always-framed form is simpler and self-documenting.
- **Add `.gitignore` to the default `sync.own` set.** Would stop sync
  from ever touching it, but also stop aikata from delivering useful
  updates to its own block — the opposite of the goal. `sync.own`
  remains the per-project opt-out, not the default.
