---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-18
audience: [human, agent]
---

# ADR 0048 — Accept sync conflicts per file

- **Status**: Proposed
- **Date**: 2026-11-18

## Context

Retrying sync wraps existing conflict markers in another conflict block.
Editing the marked file to retain both local and upstream changes also
causes another conflict because the manifest still records the old hash.

## Decision

Keep unresolved files byte-identical on retry. After editing a conflict
and removing aikata's markers, the user can run
`aikata sync --accept-conflict <path>` with the same render overrides.
The command validates that the selected path still diverges from both
baseline and current upstream. It records the current upstream hash
for that path, preserving the edited file and unrelated manifest entries.
A batch with any invalid path leaves the manifest unchanged.

## Consequences

Ordinary sync subsequently preserves accepted edits as user edits until
their upstream template changes again. Choosing upstream verbatim works
with ordinary sync. Adopting a baseline for a single resolution no
longer requires reseeding the entire manifest.

## Alternatives Considered

- Reseed the whole manifest. This also accepts unrelated upstream changes.
- Infer acceptance from missing markers. This cannot distinguish a new
  local edit from an intentional conflict resolution.
