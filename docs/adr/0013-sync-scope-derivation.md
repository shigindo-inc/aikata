---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-27
audience: [human, agent]
---

# ADR 0013 - `aikata sync` Scope Derivation Hierarchy

- **Status**: Accepted
- **Date**: 2026-05-27
- **Deciders**: aikata maintainers
- **Related**: ADR 0011 (`aikata sync` design), ADR 0014 (manifest is
  a living record), ADR 0008 (`.aikata/` config namespace), ADR 0003
  (Do-No-Harm Policy)

## Context

ADR 0011 D4 defined how `aikata sync` performs its 3-way merge but
left "which files / which preset / which optional components are in
scope?" implicit. In v0.5–v0.6.1 the answer was effectively:

- **In manifest-present mode**, the manifest's `preset` + `lang` and
  the manifest's path list (via `inferFlags`) were the sole inputs.
- **In rebaseline mode (manifest absent)**, the scope was
  hard-coded: `preset="standard"`, `lang` from `aikata.yaml`, every
  `--with-*` flag false, `stacks` and `features` ignored.

This caused real footguns:

1. A project that ran `aikata init --monorepo` had `features.monorepo:
   true` saved in `aikata.yaml`, but `aikata sync --rebaseline` after
   a manifest loss wrote a manifest that omitted the `apps/` /
   `docs/monorepo.md` surface.
2. A project that ran `aikata init --preset flutter` (so `aikata.yaml`
   has `stacks: [flutter]`) had `docs/stacks/flutter.md` in its
   project — but `aikata sync --rebaseline` always seeded with
   `preset="standard"` and dropped that file from the manifest.
3. Users had no one-off escape hatch ("just for this run, behave as
   if I'd passed `--with-monorepo`") without permanently editing
   `aikata.yaml` first.

The fix needs a small, ordered hierarchy of scope inputs.

## Decision

`aikata sync` derives the rendering scope (preset, lang, opt-in
component flags, stack list) from the following sources, in
**ascending precedence**:

1. **Defaults**: `preset="standard"`, `lang="en"`, no opt-in flags.
2. **`.aikata/aikata.yaml`** (the project's saved preferences):
   - `project.lang` → `lang`
   - `stacks: […]` → `stacks`
   - `features.monorepo: true` → `WithMonorepo`
   - `features.tdd: true` → `WithTDD`
   - Other keys (`features.obsidian_hints`, `docs.*`) are ignored at
     this layer because they have no scaffold-side effect today.
3. **`.aikata/manifest.yaml`** (when present): `preset`, `lang`, and
   `inferFlags(manifest.Files)` (path-presence → `--with-*`). For
   ambiguous cases the manifest's signal **wins** for `preset` /
   `lang`, and **OR-merges** with aikata.yaml for opt-in flags and
   stacks (manifest = "files that exist"; aikata.yaml = "what the
   user intends to enable"; either signal turns the flag on).
4. **CLI override flags** on `aikata sync`:
   - `--preset <name>` — narrow / widen the preset
   - `--lang <en|ja>` — render in a different language
   - `--stack <name>` (repeatable) — override the stack list
   - `--with-monorepo` (use `--with-monorepo=false` to force-disable)

   These are the **highest priority** and are intentionally
   **transient**: they apply to one invocation only and are never
   written back to the manifest or `aikata.yaml`. Users who want a
   change to be persistent should use `aikata add <component>` or
   hand-edit `aikata.yaml`.

This precedence is encoded in `internal/sync/plan.go:derivePlan`, and
the CLI wiring lives in `internal/cli/sync.go`.

### Why aikata.yaml participates even when manifest is present

ADR 0014 introduced the "manifest is a living record" framing —
`aikata add` updates the manifest alongside the file write. But
`aikata.yaml` is *also* a record of user intent (set by `init`,
`add stack`, `add ai-tool`, manual edit), and the two can drift
when:

- A user hand-edits `aikata.yaml` (e.g. flips `features.monorepo:
  true`) without running `aikata add`. Without OR-merging, the next
  sync would silently ignore that intent.
- A v0.4.x project upgraded to v0.6.x with a fresh `aikata sync
  --rebaseline` has a manifest that lacks any `--with-*` history.
  Reading `aikata.yaml` here is the only path to recovering the
  user's pre-v0.5 opt-ins (until the v0.7 schema bump adds explicit
  per-component flags).

OR-merging is conservative — it can only *expand* scope relative to
the manifest, never narrow it. To narrow scope for one run, use the
CLI override flags.

### Why CLI overrides are transient by default

The alternative — "make `--preset minimal` rewrite the manifest" —
breaks the model in two ways:

- The next sync would be **silently** different from the one before,
  because the manifest now lies about the project's history. Reviewers
  reading `aikata sync` output couldn't tell that the scope had
  changed across runs.
- A one-off override (e.g. `--preset minimal --dry-run` to peek at a
  narrower scope) would surprise-mutate state. That contradicts the
  Do-No-Harm Policy (ADR 0003).

If a user genuinely wants to change the preset persistently, the
v0.7 path is `aikata add preset <name>` (a future command) or
hand-editing `.aikata/manifest.yaml` (documented as supported when
explicitly intended).

## Consequences

### Positive

- `aikata sync --rebaseline` now honours `aikata.yaml`'s stacks and
  monorepo / tdd preferences. The "manifest loss leads to a smaller
  scope than the user wanted" footgun is closed.
- Users can run `aikata sync --preset minimal --dry-run` to preview
  a narrower scope without modifying any persisted state.
- `aikata sync --with-monorepo` becomes the prescribed one-off way
  to onboard the monorepo layout for a project that init'd without
  it (paired with a follow-up `aikata add monorepo` once that
  command lands).

### Negative

- One more source of truth for reviewers to consider when reading a
  sync output. Mitigated by listing the resolved scope in
  `result.Notes` (TODO follow-up if it shows up as confusing in
  practice).
- The CLI flag surface grew. The new flags are opt-in and the
  default behaviour is unchanged for projects that don't pass them.

### Out of scope (deferred)

- Per-component override flags on sync (`--with-ui`, `--with-api`,
  `--with-memory`, `--with-changelog`). `aikata.yaml` doesn't have
  per-component features today; until the v0.7 schema bump
  introduces them, the recommended path is `aikata add <component>`
  (which updates both the file set and the manifest, per ADR 0014).
- Validating that the override preset / stack name exists in the
  embedded template tree before rendering. Currently
  `scaffold.Render` returns a clear error from `templates.FS` if the
  preset is unknown; making the CLI parse-time check more friendly
  is a small follow-up.
- Recording the resolved scope in the JSON envelope output
  (`--json`). Useful for CI debuggers but not blocking.

## Implementation map

- `internal/sync/plan.go` — `derivePlan` extended to accept an
  `overrides` struct and to OR-merge `aikata.yaml` signals.
  `inferFlags` gained a monorepo detection rule.
- `internal/sync/sync.go` — `Options` gained four `Override*`
  pointer fields; `Run` threads them through `derivePlan` and onto
  `scaffold.Options`.
- `internal/cli/sync.go` — new `--preset` / `--lang` / `--stack`
  (repeatable) / `--with-monorepo` flags. `cmd.Flags().Changed(...)`
  guards each pointer so absent flags don't accidentally clear the
  inferred value.
- `internal/sync/derive_test.go` — exercises the hierarchy at the
  function level, plus integration cases via `Run` (override preset
  narrows scope; overrides do not mutate manifest).
