---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# Adopting aikata in an existing repository

aikata is happiest in a greenfield directory: `aikata init` writes a
coherent set of files and `.aikata/aikata.yaml` + `.aikata/manifest.yaml`
record what it did. Real projects, though, often already have an
`AGENTS.md`, a hand-written `CLAUDE.md`, a `.cursor/rules/` directory,
or a `.gitignore` that does important work for the team. This guide
covers the recommended migration paths so an existing repository can
adopt aikata without losing user content.

The contract aikata commits to during adoption is documented in
[ADR 0003](./adr/0003-do-no-harm-policy.md) (do-no-harm),
[ADR 0014](./adr/0014-manifest-living-record.md) (manifest contract),
[ADR 0018](./adr/0018-managed-append-for-project-owned-files.md)
(managed append for `.gitignore`), and
[ADR 0019](./adr/0019-sync-missing-file-repair-semantics.md)
(`aikata sync` never deletes silently).

## Scenario 1 — repo already has `AGENTS.md`

The canonical move is to keep your existing `AGENTS.md` and let
aikata wrap around it.

1. From the repo root, run `aikata init` **without** `--force`. A
   non-empty target directory triggers the proposal fallback: aikata
   writes its full scaffold into `.aikata-proposed/` instead of the
   current directory. Nothing of yours is touched.
2. Diff `.aikata-proposed/AGENTS.md` against your existing
   `AGENTS.md`. Pull the sections you want (operating rules, hard
   rules, navigation matrix) and merge into the canonical file by
   hand.
3. Move the rest of the proposed scaffold (`SPEC.md`,
   `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`, `.aikata/`,
   `docs/`) into the repo root.
4. `rm -rf .aikata-proposed/`.
5. Run `aikata sync --rebaseline`. This seeds
   `.aikata/manifest.yaml` from the now-current template rendering
   (not your on-disk bytes — see ADR 0011) so the next ordinary
   `aikata sync` sees your customisations as `user-only-edit` and
   preserves them.

A future `aikata adopt <file>` parser that hoists an existing
`AGENTS.md` into the canonical skeleton is tracked in
[Q-INTEROP-03](./decisions/open-questions.md#q-interop-03--adopting-a-users-existing-claudemd)
but intentionally not built yet — documentation-first per the
project's hypotheses.

## Scenario 2 — repo already has a hand-written `CLAUDE.md` or `.cursor/rules/`

Generated AI-tool artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`,
`GEMINI.md`, `.windsurfrules`) are by design **regenerable from
`AGENTS.md`**. The recommended adoption path is to migrate the
content of your hand-written files into `AGENTS.md` so the
generators have one source of truth.

1. Save copies of your hand-written `CLAUDE.md` /
   `.cursor/rules/main.mdc` outside the repo.
2. Migrate the operating rules into `AGENTS.md` §"Hard rules" and
   the tool-specific extensions (Claude Skills, Cursor globs) under
   `templates/ai_tools/<tool>/extensions/` in your project (planned
   v1.0; for v0.x simply keep tool-specific tips inside `AGENTS.md`
   sections that are clearly delimited).
3. Run `aikata generate`. The generated artifacts now reflect your
   content.
4. Decide whether to commit the generated artifacts or gitignore
   them. The target-project default is gitignored
   (`docs.generate_gitignore: true`); aikata's own repository
   commits them so contributors get a working Claude / Cursor
   experience without first installing aikata (ADR 0003 §6).

## Scenario 3 — repo already has a `.gitignore` you care about

aikata's `.gitignore.tmpl` adds a few aikata-owned entries
(`.aikata-proposed/` and the optional AI-tool artifact paths). As
of v0.7.2 the managed-append writer (ADR 0018) merges these into
an existing `.gitignore` instead of overwriting:

- Running `aikata init --force` against a directory that already
  has a `.gitignore` preserves every user-owned line and appends a
  clearly-marked aikata block:

  ```
  # your existing lines stay here

  # >>> aikata managed >>>
  # ...aikata-owned entries...
  # <<< aikata managed <<<
  ```

- Re-running is idempotent: the block is refreshed in place, not
  duplicated.

- Setting `docs.generate_gitignore: false` in
  `.aikata/aikata.yaml` suppresses the writer entirely. Use this
  when you want `.gitignore` to remain fully user-owned (aikata's
  own repository does, per ADR 0003 §6).

Without `--force`, `aikata init` against a non-empty directory
still falls back to `.aikata-proposed/`. The managed-append
integration with `--force` is the v0.7.2 surface; smarter behaviour
for the no-`--force` case is a follow-up tracked against the
ROADMAP.

## Scenario 4 — repo already has `docs/memory/`

`docs/memory/` is opt-in (ADR 0004). If your project already keeps
long-term agent memory under that path:

1. Run `aikata enable memory`. The renderer's `writeIfMissing`
   semantics preserve every existing file under `docs/memory/`; new
   files (any of `user.md`, `feedback.md`, `project.md`,
   `reference.md` that you do not yet have) are added.
2. `aikata enable memory` also flips `components.memory: true` in
   `.aikata/aikata.yaml` (ADR 0016 schema v2), so subsequent
   `aikata sync` runs recognise the slot is enabled.

If your existing memory files have different filenames or a
different layout (e.g. `memory/notes.md`), keep them as-is and
treat the aikata templates as a reference. The schema-v2
`components.memory` flag is the durable signal aikata cares about;
the file layout under `docs/memory/` is conventional.

## Scenario 5 — repo has an existing config under `.ai/`

aikata v0.3.2 moved its config from `.ai/aikata.yaml` to
`.aikata/aikata.yaml` (ADR 0008). The legacy path is still readable
through the v0.x line: `aikata generate`, `aikata sync`, and
`aikata doctor --fix` migrate the file automatically on the next
write. No manual action required unless you want to nudge the
migration earlier — `aikata doctor --fix` is the fast path.

## What aikata never does during adoption

- **Delete files** silently. `aikata sync` may add or refresh
  managed files; it never removes them when scope narrows (ADR
  0019). A future `aikata clean` would be opt-in and previewable.
- **Overwrite `AGENTS.md`** without a fallback. `aikata init`
  against a non-empty directory falls back to `.aikata-proposed/`
  unless `--force` is set.
- **Rewrite `.gitignore`** when entries it cares about are
  already there. The managed-append writer (ADR 0018) replaces
  only the aikata-owned block in `.gitignore`.

## Reporting an adoption rough edge

If you ran into a scenario this guide doesn't cover, please file
an issue on GitHub describing the repo state and the aikata
commands you ran. Adoption stories shape the v1.0 surface; see
[SPEC §7 Hypotheses to validate](../SPEC.md#7-hypotheses-to-validate).
