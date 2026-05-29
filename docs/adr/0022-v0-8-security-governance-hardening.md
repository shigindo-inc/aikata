---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0022 - v0.8.x reassigned to security & governance hardening

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0006 (locale / English-canonical docs), ADR 0015
  (first-party skill / plugin distribution), ADR 0019 (sync
  no-silent-delete), ADR 0021 (doctor scope and exclusion)

## Context

A security review of the public aikata repository (prompted by the
question "is aikata adequately secured now that it is published as
OSS?") found **no exploitable vulnerability** in the binary, the
`scripts/install.sh` installer, or the CI / release workflows:

- File writes are constrained to embedded-template or hardcoded paths;
  no user-controlled string from `.aikata/aikata.yaml` reaches
  `filepath.Join` / `os.WriteFile`.
- YAML parsing uses `gopkg.in/yaml.v3` with no custom `UnmarshalYAML`.
- Generation uses `text/template` for string interpolation only; no
  subprocess execution (`os/exec` is absent from the tree).
- The only network call is the opt-in `aikata update --check` GitHub
  Releases request (10s timeout, 1 MiB cap, no credentials).
- `install.sh` verifies SHA-256 against `checksums.txt` before
  install; CI workflows use least-privilege `permissions:` and contain
  no `pull_request_target` script-injection surface.

What the review **did** find is a gap in *governance* and
*supply-chain* posture relative to the bar expected of a published OSS
project — and relative to the guardrails already in place in the
sibling `personal-skills` repository, which the maintainer flagged as a
reference. aikata is currently missing: a `SECURITY.md` disclosure
policy, a `CODEOWNERS` file, a secret / privacy scanning CI gate, a
`dependabot.yml`, release-artifact signing, and SBOM generation.

The ROADMAP previously reserved the v0.8.x line for **channel
publication** (Homebrew tap, npm wrapper, Claude Code marketplace,
native `aikata update --apply`) and stated that no v0.9.x line was
reserved — the next milestone after v0.8.x was v1.0.

A separate, pre-existing item must not be confused with this work: the
v1.0 `--preset extended` pack scaffolds governance *templates*
(`SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, issue / PR
templates) into **user** projects. That is a generated-output feature
and is distinct from hardening aikata's **own** repository, which is
what this ADR concerns.

## Decision

Reassign the v0.8.x line to **security & governance hardening of the
aikata repository itself**, and move the previously-planned channel
publication work to a new **v0.9.x** line. The resulting sequence is:

> v0.7.3 (shipped) → **v0.8.x security & governance** → **v0.9.x
> channel publication** → v1.0 stable surface.

Scope of v0.8.x is limited to aikata's own repository posture. The
`--preset extended` governance templates stay at v1.0 and are
explicitly **not** pulled forward.

The line splits into two sub-releases so a release-pipeline change
cannot destabilise the low-risk governance work:

**v0.8.0 — Governance & secret-scan** (no binary or template change):

- `SECURITY.md` — private disclosure via GitHub Security Advisories,
  no-secrets-in-repo expectations, and an **Agent Safety** section
  (agents must not push to protected branches, merge without human
  approval, weaken `CODEOWNERS` / validation, or add remote-code-
  execution behaviour without an ADR + review). It carries the
  standard aikata five-key frontmatter so `aikata doctor --strict`
  stays green without an exclusion.
- `.github/CODEOWNERS` — maintainer review required on `/.github/`,
  `/AGENTS.md`, `/SECURITY.md`, `/.goreleaser.yml`, `/ROADMAP.md`, and
  `/docs/adr/`.
- A secret / privacy scan CI gate asserting `.env` / `.env.local` are
  absent and grepping tracked files for key material (`BEGIN
  (RSA|OPENSSH|PRIVATE)`, `api_key=`, `client_secret=`,
  `refresh_token=`), local user paths (`/Users/...`, `~/Workspace`),
  and private emails. Tailored to aikata; the `personal-skills`
  personal-profile denylist is **not** ported because aikata holds no
  personal data. Pattern definitions are placed so the scanner does
  not flag its own source.
- `.github/dependabot.yml` — weekly `github-actions` and `gomod`
  checks.
- `.gitignore` hardening — `.env.local`, `*.local.yaml`, `*.local.yml`
  (the committed `.aikata/aikata.yaml` does not match these).
- `CONTRIBUTING.md` — explicit "no direct pushes to `main`" rule plus
  an Agent Contributions section cross-referencing SECURITY.md.

**v0.8.1 — Supply-chain signing** (release-pipeline change):

- Cosign keyless signing of release artifacts + `checksums.txt` via
  GitHub OIDC (`id-token: write` in `release.yml`).
- SBOM generation (syft via GoReleaser `sboms:`).
- SHA-pin the GitHub Actions used by CI / release to full commit SHAs,
  kept current by Dependabot.
- README + `install.sh` signature-verification notes.
- A follow-up ADR (0023) will record the signing-mechanism choice
  (keyless cosign vs GPG) at implementation time.

## Consequences

- The published repository gains a documented disclosure path, review
  gates on its sensitive surface, an automated secret-leak guard, and
  signed / SBOM-bearing releases before its distribution footprint
  widens in v0.9.x — the correct order, since broader distribution
  raises the cost of a governance gap.
- Channel publication slips one minor version. This is acceptable: it
  needs out-of-repo maintainer action (tap repo, npm / Homebrew
  credentials, marketplace submission) that is not yet staged, so it
  was not on a near-term critical path.
- The ROADMAP, the distribution-surface cadence table, and forward
  references in ADR 0015, `open-questions.md`, `README.md`, and
  `docs/japanese-users.ja.md` are updated from "v0.8.x" to "v0.9.x"
  for the channel-publication items. This ADR is the canonical record
  of the renumber; the historical shipped-release sections in ROADMAP
  keep their text except for these forward pointers.
- Code-level defense-in-depth (`filepath.IsLocal()`, the `gosec`
  linter) is intentionally **out of scope**: the review found the code
  already safe, so these are deferred unless a concrete issue surfaces.

## Alternatives considered

- **Keep channel publication as v0.8.x and bolt security onto v1.0.**
  Rejected: v1.0 is the stable-surface milestone and is already large;
  shipping wider distribution channels before basic disclosure /
  review / signing guardrails inverts the risk order.
- **Fold everything into a single v0.8.0.** Rejected: the release-
  pipeline signing change carries real release-breakage risk and
  should be isolated from the otherwise risk-free governance files.
- **Pull the `--preset extended` governance templates forward.**
  Rejected: that is a user-facing generated-output feature on a
  different theme; mixing it in would blur the line's scope. It stays
  at v1.0.
