---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0019 - `aikata sync` missing-file repair semantics

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0011 (sync design), ADR 0014 (Manifest as living
  record)

## Context

`aikata sync`'s 3-way merge (ADR 0011) classifies every path in the
union of ancestor and upstream renderings into one of several
statuses (`StatusUpstreamApplied`, `StatusUserOnlyEdit`,
`StatusConflict`, `StatusUpstreamAdded`, …). The ROADMAP item
"missing-file repair semantics" surfaced two specific questions
once schema-v2 + enable made it easier to ask "what should this
project look like":

1. **What is sync allowed to RESTORE?** If a project's declared
   scope (preset + enabled capabilities) implies a file that is
   missing from disk, may sync re-render it?
2. **What must sync NEVER delete?** If a file is in the manifest
   (so aikata once authored it) but is no longer rendered by
   upstream — e.g. because the user narrowed scope — sync must not
   silently rm the on-disk copy.

The pre-v0.7.2 code already answers both questions correctly:
`StatusUpstreamAdded` re-renders missing files;
`StatusUpstreamRemoved` and `StatusUserDeleted` produce a notice
and emit no `merged[path]` entry, so no `os.WriteFile` runs and the
file is left alone. This ADR makes the behaviour an explicit
contract instead of an emergent property of the switch in
`classifyAndMerge`.

## Decision

`aikata sync` restores and refreshes managed files but **never
deletes**. Concretely:

1. **May add**: when ancestor lacks a path that upstream now
   renders (`!hadAncestor && hasCurrent==false && hasUpstream`),
   sync writes the file. Status: `StatusUpstreamAdded`. The user
   sees the new file in the next commit.
2. **May refresh**: when ancestor has a path, upstream has it with
   different content, and the user has not edited locally
   (`current == ancestor`), sync writes the upstream content.
   Status: `StatusUpstreamApplied`.
3. **Must preserve user edits**: when current and ancestor differ
   while upstream and ancestor agree, sync leaves the file
   untouched. Status: `StatusUserOnlyEdit`.
4. **Must surface conflicts**: when current, ancestor, and upstream
   all disagree, sync writes a file with git-merge-style conflict
   markers and increments `RunResult.Conflicts`. Status:
   `StatusConflict`. The user resolves manually.
5. **Must not delete on scope narrowing**: when ancestor has a
   path but upstream no longer renders it (scope shrank, preset
   changed, capability disabled), sync emits a
   `StatusUpstreamRemoved` notice and leaves the on-disk file
   alone. Same rule applies to `StatusUserDeleted` (ancestor has
   the path, current and upstream do not). No `os.WriteFile` and
   no `os.Remove` runs in either case.

### What "scope narrowing" means

A user can narrow scope by:

- Editing `.aikata/aikata.yaml` to flip a `components.*` flag
  back to `false`.
- Removing an entry from `stacks:` or `ai_tools:`.
- Running `aikata sync --preset minimal` (CLI override).

In all cases the manifest's record of "aikata rendered this file
once" is honoured. The file stays. Future `aikata clean` (or
similar explicit cleanup verb, if ever added) would have to be
an opt-in command, not a side-effect of sync.

### Out of scope

- **An explicit `aikata clean`** that removes manifest-tracked
  files no longer in scope. The verb is reserved by name; no
  implementation lands in v0.7.x. When and if it ships, it must
  be previewable (`--dry-run`), require an explicit confirmation
  step, and never run as part of `sync`.
- **Detection of files the manifest knows about but the user
  manually deleted**. Status `StatusUserDeleted` already covers
  this; the file is not re-created. A user who wants the file
  back can re-run the appropriate `aikata enable <capability>` /
  `aikata new <artifact>` / preset bump.

## Consequences

### Positive

- The "sync narrows scope and lost my UI.md" failure mode is
  structurally impossible. Sync is monotonic on the set of files
  the user has on disk: at worst it adds; at best it refreshes;
  it never subtracts.
- The contract is explicit in the ADR, so any future
  refactor of `classifyAndMerge` has a reference to check
  against. The new test
  `TestRun_UpstreamRemoved_DoesNotDelete` pins the invariant
  at the code level.

### Negative

- A user who narrowed scope but still wants the on-disk artifact
  gone has no built-in command for it. Manual `rm` is the answer
  until a future `aikata clean` lands.
- The behaviour was already correct pre-v0.7.2 but
  undocumented. Some users may not have realised sync was
  preserving on-disk files; the new ADR is also the place users
  will discover the behaviour, which is a documentation gap
  closure rather than a feature.

## Implementation map

- `internal/sync/sync.go` `classifyAndMerge` — the existing
  `hadAncestor && hasCurrent && !hasUpstream` /
  `hadAncestor && !hasCurrent && !hasUpstream` /
  `hadAncestor && !hasCurrent && hasUpstream` branches each
  produce `StatusUpstreamRemoved` / `StatusUserDeleted` with no
  `merged` entry. The write loop only touches paths present in
  `merged`, so no delete occurs.
- `internal/sync/sync_test.go` —
  `TestRun_UpstreamRemoved_DoesNotDelete` writes a manifest entry
  for a file no longer in scope, runs sync, and asserts the file
  still exists on disk with its original bytes.
