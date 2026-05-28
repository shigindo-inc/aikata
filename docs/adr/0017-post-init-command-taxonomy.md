---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0017 - Post-init command taxonomy: enable / new (no add)

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0014 (Manifest as living record), ADR 0016
  (Schema v2), Q-DESIGN-10

## Context

`aikata add <component>` accumulated three different intents during
v0.4 – v0.6:

- **Capability**: `add memory`, `add ui`, `add api`, `add tdd`,
  `add changelog`, `add stack <name>`, `add ai-tool <name>` — these
  persist a project feature for the lifetime of the project.
- **Authoring artifact**: `add adr "<title>"` — a one-off scaffold
  for a single document.
- **Tier / scope expansion**: there was talk of `add monorepo` or
  `add extended`, both of which would expand the document tier (or
  layout) of the project. Neither shipped under the `add` umbrella.

Mixing the three reads as "add stuff" but they have wildly different
durability / scope semantics:

- Enable-tier writes to `.aikata/aikata.yaml` and renders files. The
  schema-v2 `components.*` block (ADR 0016) is its declarative record.
- Authoring-tier writes one new file and does not touch any project
  flag. Re-running with a different title creates another file.

The pre-release surface conflates these and leaks implementation
detail ("which component is it again?") into the user-facing prompt.

## Decision

The post-init surface splits by intent. v0.7.1 lands two of the
three intent-scoped commands; the third is deferred to a later
release whose semantics are tractable.

| Purpose | Command | Status |
|---|---|---|
| Maintain and restore the already-declared aikata surface (no scope change) | `aikata sync` | shipped v0.5 |
| Persist a durable project capability (memory, ui, api, tdd, changelog, monorepo, stack `<name>`, ai-tool `<name>`) | `aikata enable <capability>` | **v0.7.1** |
| Create a one-off authoring artifact (adr `"<title>"`) | `aikata new <artifact>` | **v0.7.1** |
| Persist a broader document tier (`standard`, `extended`) | `aikata expand <tier>` | deferred |

### Why `enable` for capabilities

Every enable-target either (a) flips a schema-v2
`components.<x>` boolean (memory, ui, api, tdd, changelog,
monorepo), or (b) appends to a list field (stacks, ai_tools). In all
cases the durable record lives in `.aikata/aikata.yaml`. "Enable"
matches the user intent: a one-time switch that flips state for the
project. The corresponding `aikata sync` re-render reads the same
record (ADR 0016 OR-merge), so enabling a capability and then
syncing does not require a second declarative step.

### Why `new` for artifacts

`aikata new adr "<title>"` reads as one-shot authoring. Re-running
with a different title produces another file; re-running with the
same title is refused by the existing slug-collision guard. There is
no durable flag to flip — re-running `new` ten times leaves the
project with ten ADRs, not "ADRs enabled".

### Why no compatibility alias for `add`

Pre-v1.0 surfaces are explicitly mutable. Adding a translation layer
("`add memory` is now `enable memory`") would:

- Double the help-text surface area.
- Force users to learn two spellings.
- Make the v1.0 stable surface harder to define.

The aikata user base in the v0.x window is small enough that a clean
rename produces a net-positive Day-1 experience for v1.0 users at
the cost of a one-line CHANGELOG note for early adopters.

### Why `expand` is deferred

`expand standard` is the only semantically clear use case in v0.7.x
(the other proposed tier, `extended`, lands in v1.0). With only one
target the verb is mostly plumbing. Its semantics are also
under-specified for the case "what does `expand standard` do when
the project was init'd with `aikata init --preset minimal` (no
.aikata/aikata.yaml at all)?" The cleanest path is to defer until
either:

- `extended` exists and the verb has two real targets.
- A real project surfaces the need to grow a `minimal` skeleton into
  `standard` without re-running `aikata init`.

Until then, users who want to expand can re-run `aikata init` with a
higher preset and accept the proposal-fallback flow.

## Consequences

### Positive

- Help text is shorter and scannable: `aikata enable --help` lists
  only durable-state subcommands; `aikata new --help` lists only
  authoring scaffolds.
- The schema-v2 OR-merge (ADR 0016) and the enable flow have a 1:1
  field-to-command mapping (`enable memory` ↔ `components.memory`).
  Future schema additions extend this matrix uniformly.
- The `add` command is gone, so no future contributor adds a
  half-third intent to the same parent and inadvertently grows the
  conflation back.

### Negative

- Anyone scripting against `aikata add <component>` breaks. The
  v0.7.0 → v0.7.1 transition is explicit in the CHANGELOG; the user
  base is currently small enough to accept the rename without an
  alias.
- Two parent commands (`enable`, `new`) plus `sync` plus the deferred
  `expand` means the post-init surface is broader than the single
  `add`. The shortness of each leaf list is the compensating signal.
- `aikata expand` exists as a name but not a command until later.
  Users who try it get cobra's standard "unknown command" message.
  That is acceptable for now; the ROADMAP lists `expand` as deferred
  with the open semantic questions.

### Out of scope (deferred)

- **`aikata expand <tier>`** — see §"Why `expand` is deferred"
  above. v1.0 adds it together with `extended`.
- **A `kata` or `aikata` shell-completion entry that explicitly
  marks `add` as removed** — cobra's "unknown command" message is
  the v0.7.1 stopgap. If users misfire it often enough, a custom
  ValidArgsFunction can land as a follow-up.
- **Backward-compatible aliases for v0.7.1's leaf names** (e.g.,
  `enable claude` instead of `enable ai-tool claude`). Out of scope;
  the current shape matches the schema and the manifest.

## Implementation map

- `internal/cli/enable.go` — new parent with one leaf per
  `components.Capabilities()`. Mirrors the loop-driven dispatch
  pattern from the deleted `internal/cli/add.go`.
- `internal/cli/new.go` — new parent with one leaf per
  `components.Artifacts()`.
- `internal/cli/root.go` — removes `newAddCmd`, adds `newEnableCmd`
  and `newNewCmd`.
- `internal/components/registry.go` — `Capabilities` and
  `Artifacts` slices replace the single `registry`. `GetCapability`
  / `GetArtifact` / `ActiveCapabilityNames` / `ActiveArtifactNames`
  replace `Get` / `ActiveNames`. `All()` retained for read-only
  internals.
- `internal/components/monorepo.go` — new `monorepoComponent`
  wrapper around the existing `RenderMonorepo` renderer, enrolled in
  the capabilities registry.
- `internal/components/component.go` — new
  `EnableComponentInConfig(targetDir, field)` helper flips a
  schema-v2 `components.*` boolean and persists. Called by
  `memory.Add`, `singleFile.Add`, `monorepo.Add` immediately after
  the file-write / manifest-record step.
- `internal/components/memory.go`,
  `internal/components/singlefile.go`,
  `internal/components/monorepo.go` — call
  `EnableComponentInConfig` from `Add`.
- `internal/components/stack.go`,
  `internal/components/ai_tool.go` — switch from `config.Load` to
  `config.LoadMigrated` so the v1 → v2 lazy migration (ADR 0016)
  runs as a side-effect of any enable-tier write.
- `internal/cli/list.go` — replaces `list components` with
  `list capabilities` and `list artifacts`, each with the existing
  versioned `--json` envelope.
- `internal/cli/list_components_test.go`,
  `internal/cli/enable_test.go`,
  `internal/cli/new_test.go` — coverage for the new surface plus the
  v1 → v2 migration via `enable`.
