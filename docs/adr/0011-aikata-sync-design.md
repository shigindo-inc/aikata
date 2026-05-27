---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-26
audience: [human, agent]
---

# ADR 0011 - `aikata sync` Design

- **Status**: Accepted
- **Date**: 2026-05-26
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0008
  (`.aikata/` config namespace), ADR 0009 (`aikata update` reserved for
  CLI version updates), ADR 0013 (scope derivation hierarchy — defines
  how preset / lang / opt-in flags / stacks are picked across manifest,
  aikata.yaml, and CLI overrides), ADR 0014 (manifest is a living
  record — refines D4's "init-time ancestor" framing)

## Context

ADR 0009 reserved `aikata update` for CLI binary updates and routed the
project template diff-merge command to `aikata sync`. The feature itself
remained unscoped. v0.5 is the milestone that builds `aikata sync`, so
the API and merge contract need to be fixed before any code lands.

A long-lived aikata project drifts from the bundled templates over time:

1. The user edits scaffolded files (`AGENTS.md`, `SPEC.md`,
   `ARCHITECTURE.md`, etc.) as the project evolves.
2. Upstream aikata releases ship newer template content (typo fixes,
   new sections, additional ADR scaffolds).

Without a merge step, users either lose their edits when re-running
`aikata init --force` or never receive upstream improvements. `aikata
sync` closes that gap.

## Decisions

### D1 — Scope: templates only, not generated artifacts

`aikata sync` operates on the **preset template surface**: the files that
`aikata init` originally wrote into the project. It does not touch
generated artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`); those are
the responsibility of `aikata generate`. The two commands share no
overlapping write set.

**Why**: Separation of concerns. `generate` is byte-deterministic from
`AGENTS.md`; a 3-way merge there would be a regression in clarity.
`sync` owns "did upstream templates evolve and how should we adopt
that?"; `generate` owns "project this canonical doc into per-tool
files." Their boundaries stay sharp.

**Surface today**: every file under
`internal/templates/data/presets/<preset>/<lang>/` plus any optional
component the project originally opted into via `--with-*`. Generated
artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`) are explicitly out.

### D2 — Conflict UX: `.aikata.conflict` markers (git-merge style)

When the same region was edited by both the user and upstream, `aikata
sync` writes git-merge-style markers (`<<<<<<<`, `=======`, `>>>>>>>`)
back to the working tree and reports the affected paths. The user
resolves conflicts in their editor, exactly like a git merge conflict.

**Why**: Works in TTY and non-TTY (CI, agent) environments uniformly;
no separate `--no-interactive` flag needed. Tests can run against
fixture trees without mocking stdin. Users already know how to resolve
this shape from `git`. Avoids the "interactive prompt deadlocks an
agent" failure mode.

**Considered and rejected**:

- Interactive prompts per hunk — breaks under CI/agent execution,
  requires a `--no-interactive` mode and an additional bail strategy.
- Bail + report — safest but unhelpful: the user has to hand-merge from
  the report, which is exactly what conflict markers already encode.

### D3 — Schema migration is built into `sync`

`.aikata/aikata.yaml` carries a `version:` field
(`internal/config.Version`). When that schema version is bumped, the
forward-migration runs **inside** `aikata sync`. There is no separate
`aikata migrate` subcommand.

**Why**: `sync` already promises "bring this project up to the current
aikata version." Schema migration is the structured-data half of that
promise; templates are the unstructured half. A separate `migrate`
command would either be forgotten (users run `sync`, miss `migrate`,
hit confusing errors) or always-paired (in which case it is one
command, not two — YAGNI).

**Auto-run on every command** was considered and rejected: silent
rewrites of `aikata.yaml` outside `sync` would violate the Do-No-Harm
Policy (ADR 0003). `doctor` and `generate` instead fail with an
actionable error pointing at `aikata sync` when they encounter a
config schema older than `config.Version`.

### D4 — `.aikata/manifest.yaml` records init-time template hashes (3-way merge)

`aikata init` (and `aikata add`) writes `.aikata/manifest.yaml` recording
the SHA-256 of every template file as rendered at scaffold time, keyed
by the target-relative path. `aikata sync` reads this manifest, the
current on-disk content, and the freshly rendered upstream template;
the merge is a true 3-way merge with the manifest hash as the common
ancestor.

**Why**: Without an ancestor, `sync` can only do 2-way diff, which
would mark every divergence as a conflict — including upstream-only
changes that should auto-apply (the common case after upstream typo
fixes). 3-way merge is the only path that produces "auto-apply
upstream-only changes, mark only true conflicts" UX, matching git's
behavior.

**Manifest shape** (illustrative, schema lives in
`internal/config/manifest.go`):

```yaml
version: 1
preset: standard
lang: en
files:
  AGENTS.md:
    sha256: "abcd…"
    source: "presets/standard/en/AGENTS.md.tmpl"
  SPEC.md:
    sha256: "ef12…"
    source: "presets/standard/en/SPEC.md.tmpl"
```

The manifest is **regenerated** (not appended) on each `aikata sync`
run that completes without conflict; conflicted runs leave it stale
until the user resolves and re-runs.

**Privacy / portability**: the manifest contains only filenames and
content hashes; no user data leaks. It is committed to the project's
git repo so future `sync` runs (potentially by other contributors) can
reproduce the same 3-way merge.

## Consequences

### Positive

- `sync` ships with a merge story (3-way, like git) instead of a fragile
  whole-file overwrite or noisy 2-way diff.
- `generate` keeps its single-purpose role (canonical → per-tool); the
  failure modes of the two commands stay independent.
- Schema migration has a single, predictable home (`aikata sync`); no
  surprise rewrites outside it.
- Conflict markers compose with existing editor tooling (`git
  mergetool`, VS Code, Neovim) for free.

### Negative

- Existing v0.4.x projects do not have `.aikata/manifest.yaml`. `aikata
  sync` errors out for projects without a manifest and points the user
  at `aikata sync --rebaseline`, which seeds the manifest from the
  current **upstream rendering** (not from on-disk bytes — see
  *Rebaseline ancestor choice* below). `--rebaseline` is intentionally
  non-destructive: it writes only `.aikata/manifest.yaml` and never
  modifies source files. Documented as a known limitation in the v0.5.0
  release notes and revised in v0.6.1 after the v0.6.0 behaviour wrote
  conflict markers into customised files.

#### Rebaseline ancestor choice

`--rebaseline` records the ancestor as **the upstream rendering at
that moment** (the same content `aikata init` would have written for a
fresh project at this aikata version), *not* the current on-disk bytes.

Recording on-disk bytes as the ancestor would cause the next sync to
treat them as "user has no edits" (because `current == ancestor`), and
any upstream-only change would auto-apply — silently overwriting the
user's customisations. Recording the upstream rendering as the ancestor
makes those customisations register as `user-only-edit` on the next
sync, which preserves them. This refines the parenthetical wording in
earlier drafts ("trust current on-disk as the new ancestor"): the
*files* on disk are trusted as-is (never touched by rebaseline), but
the *manifest's ancestor hashes* must reflect upstream, not disk.
- The manifest is one more file to commit; the user must not edit it
  by hand. This is documented in `.aikata/manifest.yaml`'s header
  comment.
- `sync` requires a successful in-process render of the target preset.
  Embedded templates always render in current aikata, so this is not a
  practical risk, but is recorded as a precondition.

### Out of scope (deferred)

- Three-way merge across **preset switches** (e.g., user ran `init
  --preset minimal`, later wants to upgrade to `standard`). v0.5 only
  syncs within the same preset; cross-preset upgrade is a `v0.5.x`
  follow-up if demand emerges.
- Cherry-picking individual hunks (`sync --interactive`). Deferred
  until the merge core has shipped and we have evidence interactive
  selection is needed.
- Touching generated artifacts. Explicitly D1's exclusion; revisit only
  if `generate` evolves to a non-byte-deterministic shape.

## Implementation map

- `internal/sync/` — new package; owns the 3-way merge and the
  `.aikata.conflict` writer.
- `internal/config/manifest.go` — manifest read/write/regen.
- `internal/scaffold/scaffold.go` — write manifest after a successful
  init/add.
- `internal/cli/sync.go` — cobra subcommand wiring; flags are kept
  minimal (`--dry-run`, `--rebaseline`).
- `internal/doctor/` — gain a check that fails when
  `.aikata/aikata.yaml`'s `version` is older than `config.Version` and
  points at `aikata sync`.
