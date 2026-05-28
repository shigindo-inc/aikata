---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0018 - Managed-block append for project-owned generic files

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy, §"Planned application:
  generic project-owned files"), ADR 0011 (sync design), ADR 0014
  (Manifest as living record), Q-INTEROP-04

## Context

Through v0.6.x aikata wrote `.gitignore` from scratch on every
init / generate. That worked for greenfield projects but lost user
content when running against an existing repository: any custom
ignore entries the user already had were silently overwritten by the
preset template. The Do-No-Harm Policy (ADR 0003) explicitly flags
this as a planned application — files that commonly pre-exist
("user-owned even when aikata needs to add a few required entries")
must be preserved.

Q-INTEROP-04 captured the leading direction: a shared managed-block
writer that appends or refreshes only the aikata-owned section while
the rest of the file remains untouched. v0.7.2 lands the writer and
its first integration; this ADR makes the contract explicit.

## Decision

aikata gains a `internal/managed/` package that exposes two
operations on byte slices:

```
ApplyBlock(existing, newBlock []byte) ([]byte, error)
HasBlock(existing []byte) bool
```

`ApplyBlock` returns the merged file contents:

- An empty / absent file produces just the framed block.
- A file with no existing markers grows by one framed block at EOF
  (separated by a single blank line from preceding content).
- A file with a `BlockStart ... BlockEnd` pair has only the body
  between the markers replaced; user content outside the markers
  is byte-preserved.
- Malformed input (duplicate `BlockStart`, `BlockStart` without
  `BlockEnd`, `BlockEnd` without `BlockStart`) is refused with an
  error rather than silently corrupting the file.

### Block markers

```
# >>> aikata managed >>>
... aikata-owned content ...
# <<< aikata managed <<<
```

The `>>> ... <<<` shape mirrors the conda-init / shell-init
convention so users with shell-init experience recognise the
visual signal. Once shipped these strings are effectively
permanent — changing them would force every existing aikata
project's `.gitignore` to migrate. A future ADR may revisit only
under unavoidable pressure.

### Match rules

A marker counts only when it is the entire trimmed contents of a
line. A stray substring inside a user comment ("see also # >>>
aikata managed >>> note") does **not** trigger detection. This
keeps the writer safe against accidental embedding.

### Target list (v0.7.2)

Just `.gitignore`. The writer is intentionally narrow; expanding
to UPPERCASE.md files (CONTRIBUTING.md, SECURITY.md, ...) is
out-of-scope for v0.7.2 because the merge rules for prose files
differ from rule files. Adding a new target requires:

- An update to this ADR's target list with the rationale.
- The corresponding template wrapped so the rendered content can
  flow through the managed writer.

### Suppression

`docs.generate_gitignore: false` in `.aikata/aikata.yaml`
suppresses the writer entirely for `.gitignore`, matching the
pre-v0.7.2 behaviour where the file was not emitted at all. Users
opting in to that (aikata's own repo does, per ADR 0003) keep
their `.gitignore` fully user-owned.

### Integration scope (v0.7.2)

The first integration uses the writer at scaffold-time in
`scaffold.writeAll`:

- Fresh write (target `.gitignore` does not exist): write the
  rendered template verbatim. Behaviour identical to v0.7.1.
- Target `.gitignore` exists: read the existing file, merge via
  `ApplyBlock`, write the merged content.

`aikata sync` is **not** routed through the managed writer in
v0.7.2. The 3-way merge against the manifest hash continues
unchanged; a user-edited `.gitignore` outside the managed block
falls under `StatusUserOnlyEdit` (template unchanged, user
content different from manifest) and is preserved. A later
release will route sync through `ApplyBlock` so user-only edits
outside the block can travel through a fresh upstream rendering
without conflict markers.

The `--force` opt-out continues to be the path for "init against
an existing directory". Without `--force`, `aikata init` against a
non-empty directory still falls back to `.aikata-proposed/`
(unchanged from v0.6.x). A future release may relax that for
managed-append paths specifically; v0.7.2 keeps the change
surface tight.

## Consequences

### Positive

- `aikata init --force` against a repository with a hand-written
  `.gitignore` no longer destroys the user's entries.
- The managed block is self-documenting: anyone reading the file
  sees the `aikata managed` markers and knows what aikata owns.
- The writer is centralised, so the second target (whenever it
  lands) plugs into the same primitives — no per-file ad-hoc
  parsing.

### Negative

- Two paths for `.gitignore` (fresh vs merge). The scaffold layer
  has to detect which path applies; the divergence is one helper
  function (`contentForWrite`) but it does add a fork.
- Users who manually copy aikata's `.gitignore.tmpl` content into
  their existing file (without markers) will end up with both
  copies after the next `aikata init --force`. This is a one-time
  cost; idempotent re-runs after the merge converge.
- The markers are permanent strings. Any future rename burns
  goodwill.

### Out of scope (deferred)

- **`aikata sync` integration**: the managed writer is not yet
  used during `aikata sync`. The 3-way merge already preserves
  user edits as `StatusUserOnlyEdit`, so the regression risk is
  limited; sync integration is a follow-up.
- **UPPERCASE.md targets** (CONTRIBUTING.md, SECURITY.md, ...):
  prose files have different merge semantics. Track future
  requests in open-questions.
- **Init against a non-empty directory without `--force`** that
  smart-merges `.gitignore` and falls back to
  `.aikata-proposed/` for everything else. The benefit is real
  but the change surface is broader than v0.7.2 wants. Tracked
  for a follow-up.

## Implementation map

- `internal/managed/managed.go` — `ApplyBlock`, `HasBlock`, and
  the marker primitives. `BlockStart` / `BlockEnd` exported so
  callers can reference the constants in error messages.
- `internal/managed/managed_test.go` — covers empty file, append
  to existing, replace block, idempotency, malformed-input
  refusal, marker-in-user-content non-matching.
- `internal/scaffold/scaffold.go` — `contentForWrite` +
  `isManagedAppendPath` route `.gitignore` through `ApplyBlock`
  when the target file already exists.
- `docs/adoption.md` — adoption guide cross-links the writer for
  the "I already have a .gitignore" scenario.
