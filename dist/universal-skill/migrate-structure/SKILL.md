---
name: migrate-structure
user-invocable: false
description: Use when the user wants to bring an existing repository into aikata's recommended document layout — relocating off-structure documents (an ADR sitting outside docs/adr/, memory notes outside docs/memory/, a stack brief outside docs/stacks/, etc.) into the homes defined by docs/layout.md, with every move shown as a dry-run plan first and applied only after explicit confirmation via git mv. Triggers on "tidy up our docs structure", "move these docs into the right aikata folders", "reconcile our repo with the recommended layout", "where should this document live?", or after adopting a repo with aikata fill leaves documents off-structure. Read-only proposal until you approve; never rewrites document contents. For adding missing canonical docs use manage-docs (aikata fill); for the daily context loop use track-context.
---

# migrate-structure

This skill helps bring an **existing** repository into aikata's recommended
document layout: it relocates *off-structure* documents into the homes that
[`docs/layout.md`](../../docs/layout.md) defines, one reviewed move at a
time. It is the **reconcile** corner of the prescriptive / descriptive /
reconcile triad (ADR 0046): `docs/layout.md` says where documents *should*
live, the doc map records where they *do* live, and this skill proposes
moving the difference into place.

You are in an aikata-managed repository when the root has an `AGENTS.md`
and/or a `.aikata/aikata.yaml`. If neither is present, this skill does not
apply — offer `aikata fill` via `manage-docs` to adopt the repo first.

## The contract (do not deviate)

Moving user files is a mutation with real blast radius. Follow
**observe → propose → confirm → move → rebuild** exactly:

1. **Observe** — read the doc map; never guess from a raw filesystem walk.
2. **Propose** — present every move as a dry-run plan and wait.
3. **Confirm** — apply only after explicit approval; default to doing nothing.
4. **Move** — `git mv` only; never overwrite a destination; never delete;
   never edit file contents.
5. **Rebuild** — `aikata map` so the map reflects reality.

The aikata CLI is **observation-only** here: there is no `aikata migrate`
verb. You perform the moves with `git mv`; aikata only *reports* (the doc
map) and *re-derives* (`aikata map`).

## The loop

### 1. Get a current doc map

```bash
aikata map                   # rebuild .aikata/docmap.{yaml,md}
```

Then read `.aikata/docmap.yaml`. Each entry has a `managed:` flag:

- `managed: true`  — already on the aikata-managed surface; leave it alone.
- `managed: false` — **external** (off-structure); a candidate to relocate.

The external set is your work list. If there are no `managed: false`
documents, the repo is already aligned — say so and stop.

### 2. Propose a destination per external document

For each `managed: false` document, decide its home from
[`docs/layout.md`](../../docs/layout.md). Common cases:

| The document is… | Recommended home |
|---|---|
| an architecture decision record (`# ADR NNNN …`, a `**Status**:` line) | `docs/adr/NNNN-*.md` |
| long-term memory (user / feedback / project / reference notes) | `docs/memory/<type>.md` |
| a stack / technology brief | `docs/stacks/<stack>.md` |
| a workflow / collaboration guide | `docs/workflows/<domain>.md` |
| in-flight working state | `docs/tasks/current.md` |
| a design / exploration note | `docs/design/` |
| a canonical top-level doc placed in a subfolder (e.g. `SPEC`, `GLOSSARY`) | repository root |

Apply judgement, not pattern-matching alone:

- If a document's correct home is **ambiguous**, leave it in place and ask —
  do not move on a guess (ADR 0046 D5).
- A canonical document that is merely *named* like an aikata doc but is
  actually project-owned prose is not automatically a move; confirm intent.
- Respect existing numbering: an ADR keeps its number; if the number
  collides with an existing `docs/adr/NNNN`, surface that rather than
  renumbering silently.

### 3. Show the dry-run plan and wait

Present the full plan and stop for approval — do not move anything yet:

```text
Proposed moves (dry run — nothing changed yet):
  docs-old/decisions/use-postgres.md   →  docs/adr/0007-use-postgres.md
  notes/team-prefs.md                  →  docs/memory/feedback.md
  FLUTTER.md                           →  docs/stacks/flutter.md

Leaving in place (ambiguous — tell me where these belong):
  misc/onboarding.md
```

Let the user approve the batch, approve individual moves, edit destinations,
or cancel. On silence or uncertainty, do nothing.

### 4. Apply approved moves with `git mv`

For each approved move, create the destination directory if needed and use
`git mv` to preserve history:

```bash
mkdir -p docs/adr
git mv docs-old/decisions/use-postgres.md docs/adr/0007-use-postgres.md
```

Hard rules:

- **Never overwrite.** If the destination already exists, refuse that move
  and surface the collision; do not clobber or merge.
- **Never delete.** This skill only relocates.
- **Never edit contents.** Do not rewrite the document, its frontmatter, or
  its links in this pass (ADR 0046 D4).
- Use `git mv` (not `mv`) so history follows the file and the change is a
  reviewable rename in the diff.

### 5. Rebuild the map and check consistency

```bash
aikata map                   # the moved docs should now be managed: true
aikata doctor                # surfaces links broken by the moves (file:line)
```

Moves commonly break relative links from *other* documents that pointed at
the old path. `aikata doctor` reports these as `broken link` findings.
Surface them with `file:line` for the user (or a follow-up skill) to fix —
do not silently rewrite links here.

## When you are done

- The intended `managed: false` documents are now under their `docs/layout.md`
  homes and read as `managed: true` in a fresh `aikata map`.
- Ambiguous documents were left in place and reported, not moved on a guess.
- No document was overwritten, deleted, or content-edited.
- `aikata doctor` link findings caused by the moves are surfaced for
  follow-up.

## Reference

- Target layout: [`docs/layout.md`](../../docs/layout.md).
- The mutation boundary: ADR 0046.
- Raw CLI surface (`map`, `doctor`): `manage-docs` skill.
- Repository: <https://github.com/shigindo-inc/aikata>
