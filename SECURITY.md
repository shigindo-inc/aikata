---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# Security Policy

## Supported Versions

aikata is pre-1.0. Security fixes target the latest `main` branch and the most
recent release tag. Older tags are not patched; upgrade to the latest release.

## Reporting a Vulnerability

Please do **not** open a public issue for suspected secret leakage, a
supply-chain concern (installer, release artifacts, CI), or any other security
problem.

Report privately through
[GitHub Security Advisories](https://github.com/shigindo-inc/aikata/security/advisories/new).
If advisories are unavailable, contact the maintainers through a private
channel. Include:

- the affected version or commit, and the affected file or command;
- reproduction steps;
- the expected impact;
- whether any secret, credential, or local path appears in public content.

We aim to acknowledge a report within a few days. Please allow a reasonable
period for a fix before any public disclosure.

## Security Expectations

- Never commit `.env`, `.env.local`, `*.local.yaml` / `*.local.yml`,
  credentials, tokens, private keys, or absolute local paths containing a
  user identifier. These are gitignored and a CI scan rejects them if they are
  force-added (see `internal/repolint`).
- The committed `.aikata/aikata.yaml` and the generated, dogfooded artifacts
  (`CLAUDE.md`, `.cursor/rules/`) are intentionally tracked; they contain no
  secrets.
- Changes to `/.github/`, `/AGENTS.md`, `/SECURITY.md`, `/.goreleaser.yml`,
  `/ROADMAP.md`, and `/docs/adr/` require maintainer review (`CODEOWNERS`).
- Release tags are immutable once published. Verify downloads against
  `checksums.txt`; `scripts/install.sh` does this automatically.

## Agent Safety

aikata is built with AI agents and expects AI-assisted contributions. To keep
that safe, an agent operating in this repository **must not**:

- push directly to `main` or any other protected branch;
- merge a pull request without explicit human approval;
- weaken `CODEOWNERS`, branch protection, the secret-scan gate, the
  `aikata doctor --strict` gate, or the `aikata generate` byte-identity check;
- introduce remote code execution, subprocess execution of untrusted input, or
  any new network call without a recorded ADR and human review;
- add AI attribution trailers to commits (see `AGENTS.md` rule 6).

These constraints are invariant; relaxing one requires a human-reviewed ADR.
