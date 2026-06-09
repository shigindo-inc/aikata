---
name: refresh-docs
user-invocable: false
description: Use when the user wants to bring an aikata-managed repository's documentation up to the latest aikata — check whether the installed aikata binary is current and update it, pull upstream template changes (`aikata sync`), add any missing canonical documents (`aikata fill`), reconcile consistency (`aikata doctor`), retire deprecated docs, and regenerate per-AI-tool configs. Triggers on "update this repo to the latest aikata", "refresh the docs", "is our aikata setup current?", "migrate our docs to the newest templates", or "clean up deprecated docs". This is a downstream-maintenance loop; for raw single-command invocation use `manage-docs`, and for the in-repo context-maintenance loop use `track-context`.
---

# refresh-docs

This skill brings a **downstream** aikata-managed repository's canonical
documentation up to date with the latest aikata. It does not enforce a
fixed pipeline — assess the repo's state and run only the steps that are
actually needed, confirming anything destructive with the user first.

You are in an aikata-managed repository when the root has an `AGENTS.md`
and/or a `.aikata/aikata.yaml`. If neither is present, this skill does not
apply (offer `aikata fill` via `manage-docs` to adopt the repo instead).

All command invocations below are the raw aikata CLI surface taught by the
**`manage-docs`** skill — hand off to it for flag detail. If `aikata` is
not on `PATH`, surface the install paths before doing anything else.

## The loop

Run these in order, skipping any step whose work is already done.

### 1. Is the aikata binary current?

```bash
aikata update --check        # reports installed vs. latest; exit 0 always
```

If an update is available, update the binary **after** telling the user
which version it moves to:

```bash
aikata update --apply        # prebuilt channels only (install-script, github-release)
```

For non-binary installs (`go install`, package managers) `--apply` prints
the channel-native upgrade command instead — relay it; do not force it.
Skip this whole step if the binary is already latest or the user only
wants doc changes, not a binary upgrade.

### 2. Pull upstream template changes

```bash
aikata sync                  # 3-way merge; preserves your edits and deletions
aikata sync --dry-run        # preview the plan first when the repo has many local edits
```

`sync` also forward-migrates `.aikata/aikata.yaml` (schema v1 → v2) in the
same run, so this is where doc-schema migration happens. If sync writes
conflict markers (`<<<<<<<` / `=======` / `>>>>>>>`), resolve them with
the user before continuing — do not auto-pick a side.

Note the boundary: `sync` respects deletions and will **not** restore a
doc the repo is missing — that is step 3's job. Do not reach for
`aikata sync --rebaseline` to add missing docs; it records absent docs as
deleted.

### 3. Add any missing canonical documents

```bash
aikata fill                  # writes only missing canonical docs; never overwrites
```

Use this when a newer aikata introduced a canonical document the repo
predates, or one was deleted and should come back. fill is idempotent and
records hand-written files so the next `sync` preserves them.

### 4. Reconcile consistency

```bash
aikata doctor --json         # machine-readable; parse summary.errors
aikata doctor --fix          # apply the trivially-fixable subset (stale dates, missing frontmatter)
aikata doctor --json         # re-run; report remaining issues as file:line
```

Report anything `doctor` cannot auto-fix with `file:line` references the
user can open.

### 5. Retire deprecated docs (judgement + confirmation)

`doctor` flags deprecated ADRs (`Status: deprecated`) and requires each to
reference its replacement ("Replaced by" / "Superseded by"). When a
deprecated doc is genuinely dead:

- Confirm with the user before deleting anything.
- Remove the deprecated file and update any cross-references that pointed
  at it (other docs, `docs/decisions/open-questions.md`, links in
  `AGENTS.md`).
- Do **not** delete a deprecated ADR that is still referenced as live
  rationale elsewhere; surface the conflict instead.

There is no `aikata` command for this cleanup by design — it needs
judgement. Keep it conservative.

### 6. Regenerate tool files

```bash
aikata generate              # rewrite CLAUDE.md / .cursor/rules/main.mdc from AGENTS.md
```

Run last, after any canonical doc changed in the steps above, so the
generated per-AI-tool artifacts match the refreshed canonical set.

## When you are done

- The binary is current (or the user declined the upgrade).
- `aikata doctor` reports no errors.
- Missing canonical docs are filled; conflicts from `sync` are resolved.
- Deprecated docs are retired or explicitly kept, with the user's
  confirmation.
- Generated artifacts are regenerated and consistent.

## Reference

- Raw CLI surface: `manage-docs` skill (flags, `doctor --json` schema).
- In-repo context loop: `track-context` skill.
- Repository: <https://github.com/shigindo-inc/aikata>
