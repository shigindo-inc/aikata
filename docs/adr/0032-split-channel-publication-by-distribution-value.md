---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0032 - Split the v0.9.x channel-publication line by distribution value

- **Status**: Accepted
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: ADR 0015 (first-party skill / plugin distribution),
  ADR 0022 (v0.8.x security & governance hardening — created the v0.9.x
  channel-publication line), ADR 0024 (scope / stack axes split — set the
  precedent for non-themed work and out-of-order numbering in a minor
  line), ADR 0028 (prioritize core-concept stabilization)

## Context

[ADR 0022](./0022-v0-8-security-governance-hardening.md) moved the
channel-publication work out of the v0.8.x number space and reserved a
single **v0.9.9** line for it. That line currently bundles five items of
**unequal value** and **differing external dependencies**:

1. Homebrew tap (`shigindo-inc/tap/aikata`).
2. npm wrapper (`npx aikata`).
3. Claude Code marketplace listing.
4. Universal `npx skills add` package (first-party usage guidance,
   ADR 0015 D3).
5. Native installer-managed `aikata update --apply`.

Treating these as one release conflates work that is high-value and
agent-doable today with work that is low-value-now and gated on
out-of-repo maintainer action:

- **`curl … | sh` already covers the "no Go required" install need**
  (shipped v0.2.1, with SHA-256 verification), and `go install` covers
  Go users. Homebrew and npm are therefore pure *convenience* channels —
  they fill no open gap, while adding standing cost: a new
  `shigindo-inc/homebrew-tap` repository plus `HOMEBREW_TAP_GITHUB_TOKEN`,
  npm org credentials (`NPM_TOKEN`), publish steps, and an ongoing
  breakage surface. This is consistent with how Claude Code itself steers
  users toward a native `curl` installer rather than a package manager.
- **Items 3 and 4 align with aikata's core identity** — agent-facing
  shared-context tooling discoverable where its users already work
  (ADR 0028). The Claude Code plugin scaffold already exists from v0.6.0,
  and the universal skill package's source artifacts are repo-local under
  `dist/` (ADR 0015 D2), so the buildable part is fully agent-doable.
- **Item 5 does not depend on items 1–2.** The ROADMAP's "ships once each
  channel above is real enough to test in CI" wording reads, in practice,
  as "fully deferred". It is not: the `install-script` / `go-install` /
  `github-release` self-update branches consume the
  `internal/install.Detect()` foundation shipped in v0.6.0 and the
  `aikata.install-source` marker `scripts/install.sh` already writes —
  they are testable today. Only the `homebrew` / `npm` branches genuinely
  wait on items 1–2 existing.

The v0.6.0 release set the honesty precedent for this kind of split: it
shipped the agent-doable subset of packaging and deferred the
maintainer-action channels rather than blocking on them.

## Decision

Refine the single v0.9.x channel-publication line defined in ADR 0022
into **three value-ordered lines**. This ADR is the canonical record of
the split; forward pointers (ROADMAP, the distribution-surface cadence
table, `open-questions.md`, README ADR index, CHANGELOG) are updated,
while prior ADR bodies are left intact as historical — notably ADR 0028's
"v0.9.9 channel-publication line" phrasing now denotes the line this ADR
splits, not a single release.

### D1 - v0.9.3: agent-ecosystem distribution (prioritized)

Ship the distribution surface that matches aikata's core identity:

- **Universal `npx skills add` package** (item 4) — first-party aikata
  usage guidance, source under `dist/universal-skill/` per ADR 0015. Fully
  agent-doable; this is the headline deliverable of v0.9.3.
- **Claude Code marketplace *readiness*** (item 3) — finalize the plugin
  manifest / listing metadata so aikata is submission-ready. The
  **submission act itself stays gated** on the upstream marketplace flow
  being available plus a maintainer submitting for review; per the v0.6.0
  precedent, that external step does not block the v0.9.3 release. The
  manual plugin-install path stays supported regardless.

### D2 - v0.9.4: native self-update for existing channels

Ship `aikata update --apply` covering only the channels that exist today:
`install-script`, `go-install`, and `github-release`. The
highest-value branch is **`install-script` self-update** — the `curl … |
sh` audience is aikata's main no-Go install path. The `homebrew` / `npm`
branches are stubbed with an actionable "use your package manager"
message until those channels are real (D3). Native self-update is a
convenience, not essential; v0.9.4 keeps it isolated from the v0.9.3
ecosystem work so neither blocks the other.

### D3 - v0.9.9: native package-manager channels (lowest priority)

Defer the convenience-only package-manager channels and their dependent
self-update branches to v0.9.9:

- Homebrew tap (item 1) and npm wrapper (item 2).
- The `homebrew` / `npm` branches of `aikata update --apply` (item 5),
  added once items 1–2 are real enough to test in CI.

These remain deferred until concrete demand (a user asking for
`brew install aikata` / `npx aikata`) justifies the standing maintenance
cost.

### D4 - Baseline and ordering invariants

- `go install`, GitHub Release, and `curl … | sh` remain the canonical
  install baseline; channels grow **monotonically** — adding a channel
  never breaks an earlier one.
- As with v0.8.3 shipping ahead of v0.8.2 (ADR 0024 precedent), **numeric
  order is direction, not ship order.** v0.9.3 is not blocked behind the
  still-unshipped v0.9.2 brand-exploration line; the two are independent.

## Consequences

- aikata becomes discoverable in the agent ecosystem (marketplace +
  universal skill) one or more minors sooner, instead of waiting behind
  convenience package-manager work it does not depend on.
- `curl … | sh` users gain `aikata update --apply` self-update in v0.9.4
  without waiting for Homebrew / npm to exist.
- The lowest-value, highest-maintenance channels (Homebrew, npm) are
  honestly parked at v0.9.9 with their cost stated, rather than padding an
  earlier release.
- The distribution-surface cadence table gains a **v0.9.3** row
  (marketplace + universal skill first appear); the v0.9.9 row now adds
  only the npm / Homebrew cells (marketplace / universal inherited,
  monotonic). No v0.9.4 row is added — the table tracks install-*surface*
  changes, and self-update is an update mechanism, not a new install
  source (the same reason v0.9.1 has no row).
- The v0.9.3 marketplace item is framed as readiness, so a slow or
  unavailable upstream submission flow cannot stall the release.

## Alternatives Considered

- **Keep all five items in a single v0.9.9.** Rejected: it couples
  high-value agent-ecosystem work to convenience channels with external
  credential dependencies, delaying the former for no technical reason.
- **Move native `update --apply` wholesale to v0.9.9.** Rejected: its
  `install-script` branch is high-value (the curl audience) and depends on
  nothing new, so deferring it whole would withhold a working feature.
  Only the brew / npm branches genuinely wait.
- **Ship Homebrew / npm earlier for completeness.** Rejected: they fill no
  open install gap (curl + go install already cover it) and add standing
  maintenance and breakage surface; demand-driven is the right trigger.
- **Renumber the whole v0.9.x line instead of splitting in place.**
  Rejected: ADR 0022 already established the v0.9.x channel line and
  updated forward pointers once; a second wholesale renumber churns
  references for no benefit. Splitting in place, recorded here, is
  cheaper and traceable.
