---
description: Pull newer aikata template content into the current project without losing user edits.
---

# /aikata-sync

Run `aikata sync` to perform a 3-way diff-merge between the current
project, the templates as they were when aikata first wrote them
(recorded in `.aikata/manifest.yaml`), and the freshly rendered
upstream templates. User edits are preserved; upstream-only changes
auto-apply; true conflicts get git-merge-style markers.

## When to run

- A new aikata release shipped and the user wants the project caught
  up on template improvements.
- `aikata doctor` warned about template drift.
- The user is preparing a release of their own and wants the latest
  upstream conventions integrated first.

## Step 1 — preview the plan

Always preview first so the user sees what is about to change:

```bash
aikata sync --dry-run
```

The output lists each file under one of these statuses:

- `unchanged` — no diff in any direction (skipped in summary).
- `upstream-applied` — user file matched the manifest; upstream
  evolution applied cleanly.
- `user-only-edit` — user edited; upstream unchanged; preserved.
- `both-match` — user and upstream coincidentally agree.
- `conflict` — user edited AND upstream changed AND the two diverge.
- `upstream-added` — new file from upstream.
- `upstream-removed` / `user-deleted` — informational.

## Step 2 — apply

If the dry-run looks safe:

```bash
aikata sync
```

Exit codes:

- `0` — clean (no conflicts) or dry-run.
- `2` — conflicts were written to disk; manual resolution required.

## Step 3 — resolve conflicts

Conflicts use git-merge markers (`<<<<<<<`, `|||||||`, `=======`,
`>>>>>>>`). Open each flagged file in the editor (or use
`git mergetool`), pick the desired hunks, then re-run `aikata sync`
to refresh the manifest.

## --rebaseline (one-shot)

For projects that pre-date v0.5 (no `.aikata/manifest.yaml`):

```bash
aikata sync --rebaseline
```

Seeds a manifest from the current on-disk state. The next normal
`aikata sync` then does a real 3-way merge.

## Notes

- `aikata sync` only touches files that came from preset templates.
  Generated artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`) stay
  the responsibility of `aikata generate` per ADR 0011.
- The manifest is regenerated only on conflict-free runs; resolve
  any conflicts before re-running.
- See `docs/adr/0011-aikata-sync-design.md` for the full merge
  contract.
