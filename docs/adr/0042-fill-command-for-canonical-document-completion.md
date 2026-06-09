---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-05
audience: [human, agent]
---

# ADR 0042 - fill command for canonical document completion

- **Status**: Accepted
- **Date**: 2026-06-05
- **Deciders**: aikata maintainers
- **Related**: ADR 0017 (post-init command taxonomy — this ADR adds a peer
  verb), ADR 0011 (sync design — fill's "ancestor = upstream rendering"
  rule), ADR 0024 (scope/stack axes — `minimal` is config-lite), ADR 0003
  (do-no-harm). Resolves open question Q-INTEROP-03.

## Context

There was no low-friction way to bring an existing repository — one that
already has hand-written canonical documents (a bespoke `AGENTS.md`, say)
— under aikata while writing only the documents it is *missing* and never
touching what already exists. The three existing verbs each miss this:

- `init` scaffolds a **new** project. In a non-empty directory it diverts
  the whole tree to `.aikata-proposed/` for manual merge (SPEC §4.1,
  ADR 0037 D4) — safe, but high-effort.
- `sync` pulls upstream template **changes** and, by contract, **respects
  deletions**: a canonical doc the user removed is classified
  `user-deleted` and not restored (ADR 0019). It also hard-requires
  `.aikata/aikata.yaml` + a manifest, so it cannot bootstrap an unmanaged
  repo (the `--rebaseline` seeding run writes the manifest but no files,
  so absent docs read as `user-deleted` on the next sync).
- `enable <capability>` adds a **single** capability and likewise requires
  existing config.

This is exactly the gap recorded as Q-INTEROP-03 ("adopting a user's
existing repo"). The maintainer's constraint: solve it without growing the
flag surface — a clear, **option-free** verb is preferable to an
`init --fill` flag.

A related inconsistency surfaced while scoping this: `minimal`-scope
projects wrote `.aikata/manifest.yaml` but no `aikata.yaml`, so no command
could ever consume that manifest (`sync`/`enable`/`new` all fail at the
config load first). ADR 0011 ("init writes a manifest") and ADR 0024
("minimal is config-lite") disagreed for minimal.

## Decision

Add a new top-level verb **`aikata fill`** (no flags): render the
canonical document set the current repository's scope defines and write
only the files that are **missing**, never overwriting an existing file.
Idempotent.

Scope is inferred, not flagged:

- **Managed / partial** (`.aikata/manifest.yaml` and/or `aikata.yaml`
  present): preset/lang come from the manifest when present; the
  optional-component / stack / AI-tool set comes from `aikata.yaml`.
- **Unmanaged** (no `.aikata/`): default to scope `standard`, lang `en`,
  project name = the working-directory basename, and adopt the repo
  (write `aikata.yaml` + a manifest). `standard` is `init`'s own default,
  and because fill never overwrites, unwanted documents are simply pruned
  afterward.

The manifest is rebuilt from the **rendered (upstream) hashes** via the
existing `components.RecordInManifest` (merge-safe). A hand-edited file's
recorded ancestor is therefore the upstream rendering, so the next
`aikata sync` classifies the user's content as `user-only-edit` and
preserves it. Because fill writes every missing rendered file, nothing is
recorded-but-absent, so the bare-rebaseline "absent reads as user-deleted"
footgun cannot occur.

Companion fixes shipped with this ADR:

1. **Drop the manifest for `minimal`** — `presetHasStructuredConfig`
   gates `.aikata/aikata.yaml` and `.aikata/manifest.yaml` together, so
   minimal stays genuinely config-lite (ADR 0024 wins the conflict).
   `doctor` reads the manifest opportunistically and falls back to its
   static managed set, so this is behaviour-preserving there.
2. **`sync` error on missing config** — returns the actionable
   `ErrNotManaged` sentinel pointing at `aikata fill` / `--scope standard`
   instead of the raw "file does not exist".

### Why a peer verb, not a mode of init/enable/sync

`fill` operates on the **whole** canonical set and on a possibly-unmanaged
repo. It is not `init` (which is from-scratch and proposes, never merges
in place), not `enable` (single capability, requires config), and not
`sync` (changes, respects deletions). Folding it into any of them would
overload a contract users rely on. It is a distinct intent — "make this
repo a complete aikata project" — and earns its own verb.

### Why `fill`, not `adopt`

The `.aikata-proposed/` flow is already described in-code as "the adoption
fallback", and Q-INTEROP-03's leading term `adopt` was scoped more broadly
(parse a hand-written `CLAUDE.md` into the `AGENTS.md` skeleton). Naming
the new command `adopt` would (a) imply that parsing ambition, which this
ADR explicitly drops, and (b) conflate write-missing-in-place with
propose-everything. `fill` describes precisely what happens — write the
gaps — and is immediately legible (including to Japanese-speaking users).

## Rationale

The command reuses existing, tested machinery — `scaffold.Render`,
`components.WriteIfMissing`, `components.RecordInManifest` — so the new
surface is one cobra command and a thin core package. Option-free keeps
the cognitive load at zero, which was the entire motivation. The
manifest-from-upstream rule keeps it composable with `sync` rather than
fighting it.

## Consequences

Positive:

- Existing repos can be adopted, or partial projects topped up, with one
  zero-config command; nothing is ever overwritten.
- A `minimal` project that runs `fill` is additively upgraded to a managed
  `standard` project — an in-place, non-destructive realization of the
  minimal→standard upgrade that ADR 0011 had deferred.
- The minimal manifest/config inconsistency is resolved.

Negative / trade-offs:

- The command surface grows from 10 to 11 verbs. Justified by the gap and
  the maintainer's explicit preference for a clear verb over a flag.
- On an unmanaged repo fill assumes `standard`, which may write documents
  that do not fit a non-software repository (e.g. a portfolio). This is
  acceptable because fill is non-destructive (prune what doesn't fit), and
  such repos are better served by `init` or a hand-rolled minimal
  `.aikata/`. The default is not tuned to that edge case.
- A managed project with a deleted-on-purpose canonical doc will see fill
  restore it (fill does not consult deletion intent, unlike sync). This is
  the intended "complete the set" behaviour, not a regression.

## Alternatives considered

- **`init --fill` flag** — the user's first instinct. Cheapest faithful
  realization, reuses `WriteIfMissing`, no new subcommand. Rejected
  because the maintainer judged a clear, option-free verb more discoverable
  than overloading `init` with a behaviour-changing flag.
- **Managed-only `fill`** (require `.aikata/aikata.yaml`) — rejected: it
  abandons the originating unmanaged-repo use case and is nearly redundant
  with `sync` (which already auto-adds upstream-new docs; the only delta
  would be resurrecting deliberately-deleted files).
- **`aikata adopt` with CLAUDE.md parsing** (the old Q-INTEROP-03 lead) —
  deferred: the parsing ambition is unproven demand; write-missing covers
  the real pain today.
- **Do nothing, document `.aikata-proposed/` + `enable`** — rejected: the
  manual-merge path is the high-effort experience users were trying to
  avoid.

## References

- Q-INTEROP-03 in `docs/decisions/open-questions.md` (now resolved).
- ADR 0017 (taxonomy), ADR 0011 (sync / ancestor rule), ADR 0024
  (scope/stack), ADR 0019 (sync respects deletions), ADR 0003 (do-no-harm).
- Implementation: `internal/fill/`, `internal/cli/fill.go`,
  `internal/components` (`WriteIfMissing`, `RecordInManifest`),
  `internal/scaffold/scaffold.go` (`presetHasStructuredConfig`).
