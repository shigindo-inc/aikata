---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0023 - Release signing mechanism & supply-chain hardening (v0.8.1)

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0022 (v0.8.x security & governance hardening), which
  scheduled this work as the v0.8.1 sub-release

## Context

ADR 0022 split the v0.8.x line into v0.8.0 (governance, no binary or
pipeline change) and **v0.8.1 (supply-chain signing)**, and deferred the
signing-mechanism choice to "implementation time" — this ADR.

Before v0.8.1, a release was a GoReleaser build triggered by a `v*` tag,
publishing archives plus a SHA-256 `checksums.txt`. That gives integrity
(a user can detect a corrupted download) but **not provenance**: nothing
proves the artifacts were produced by aikata's own release pipeline
rather than substituted by an attacker who can write to the release.

Three gaps relative to the bar for a published OSS supply chain:

1. No cryptographic signature binding the artifacts to the release
   workflow's identity.
2. No Software Bill of Materials (SBOM) for downstream vulnerability
   scanning.
3. Third-party GitHub Actions referenced by mutable tags (`@v4`), so a
   compromised or force-moved tag could inject code into the release
   job.

## Decision

Harden the release pipeline along three axes.

### 1. Keyless cosign signing (not GPG)

Sign `checksums.txt` with [cosign](https://docs.sigstore.dev/) **keyless**
signing via GitHub OIDC. The release workflow gains `id-token: write`;
cosign mints a short-lived Fulcio certificate bound to the workflow
identity and records the signature in the Rekor transparency log. cosign
v3 (installed by `cosign-installer`) emits the new Sigstore **bundle**
format, so the release uploads a single `checksums.txt.sigstore.json`
that combines the certificate and signature (`--bundle`); the older
separate `--output-certificate` / `--output-signature` flags are
deprecated and ignored under v3.

Signing only `checksums.txt` is sufficient: it lists the SHA-256 of every
archive, so a verified checksum file transitively authenticates all
artifacts. This keeps the `signs:` config to a single entry.

**Keyless over GPG/long-lived cosign key** because:

- No private key to generate, store as a repository secret, rotate, or
  leak. The signing identity *is* the release workflow, verifiable by
  anyone without prior key exchange.
- Verification needs only the public Fulcio root and the expected
  identity (`.github/workflows/release.yml@refs/tags/v*` +
  `token.actions.githubusercontent.com`), both documented in the README.
- Matches the guardrails already used in the sibling `personal-skills`
  repository.

### 2. SBOM generation

Generate one SBOM per archive with syft via GoReleaser's `sboms:` block
(`<archive>.sbom.json`, SPDX). The release workflow installs syft before
invoking GoReleaser.

### 3. SHA-pinned Actions

Pin every third-party GitHub Action in `ci.yml` and `release.yml` to a
full commit SHA with a trailing version comment. Dependabot
(`github-actions`, added in v0.8.0) keeps the pins current via reviewed
PRs. A new `goreleaser-check` CI job runs `goreleaser check` on every PR
so a malformed `.goreleaser.yml` (e.g. a broken `signs:` / `sboms:`
block) fails in review rather than at tag-push time.

## Consequences

- Releases from v0.8.1 onward carry verifiable provenance and an SBOM;
  the README documents the `cosign verify-blob` command. Integrity for
  users who skip cosign is unchanged (`sha256sum -c checksums.txt` still
  works), so this is additive — no break for existing install paths, and
  `scripts/install.sh` stays dependency-free.
- The release pipeline gains an external dependency on the Sigstore
  public-good infrastructure (Fulcio, Rekor) at release time. Acceptable:
  it runs only on tag push, and a Sigstore outage delays a release rather
  than breaking installs.
- **v0.8.0 must be tagged before these changes land on `main`** (ADR
  0022's isolation rationale): the v0.8.0 governance release must ship on
  the *unchanged* pipeline. The first release exercising this pipeline is
  v0.8.1. Because keyless signing only runs under the Actions OIDC token,
  it cannot be validated locally; a `vX.Y.Z-rc.N` prerelease (the
  `.goreleaser.yml` `prerelease: auto` already supports this) is the
  intended way to exercise the full signed path before a final tag.

## Alternatives considered

- **Long-lived cosign key or GPG.** Rejected: introduces a secret to
  manage and rotate, and a key-distribution problem for verifiers, for no
  benefit over keyless in a CI-only signing flow.
- **Sign every archive individually.** Rejected as redundant: the signed
  checksum file already authenticates them transitively, at a fraction of
  the config and signing cost.
- **Keep mutable action tags.** Rejected: tag mutability is the concrete
  supply-chain hole this ADR closes; SHA pins + Dependabot give
  immutability without freezing versions forever.
