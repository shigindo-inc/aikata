---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# Contributing to aikata

Thanks for considering a contribution. aikata is a small Go CLI plus a
set of canonical markdown documents, so most contributions touch either
the `internal/` Go packages or the document set under repo root and
`docs/`.

> **Canonical operational rules live in [`AGENTS.md`](./AGENTS.md)** —
> read it first. CONTRIBUTING.md is the human-friendly entry point and
> intentionally shorter. If the two ever conflict, AGENTS.md wins
> ([ADR 0002](./docs/adr/0002-agents-md-as-canonical.md)).

---

## Quick start

```bash
git clone https://github.com/shigindo-inc/aikata.git
cd aikata

# Run the test suite (Go 1.21+ required):
go test ./...

# Build and self-check the repo:
go build -o /tmp/aikata ./cmd/aikata
/tmp/aikata doctor
```

`aikata doctor` should report zero error-level issues on `main`.
Info-level findings (orphan glossary entries, etc.) are acceptable —
the binding gate is errors and warnings.

Japanese-speaking contributors: see
[`docs/japanese-users.ja.md`](./docs/japanese-users.ja.md) for an
overview in 日本語. Public artifacts (issues, PRs, commits, code
comments) stay in English regardless of the contributor's language —
this is the canonical-locale policy in
[ADR 0006](./docs/adr/0006-locale-and-japanese-documentation-policy.md).

---

## What lives where

- `cmd/aikata/` — main entry point and `--version` handling.
- `internal/cli/` — cobra commands (one file per subcommand).
- `internal/scaffold/` — `aikata init` engine and preset metadata
  (`registry.go` is the single source for preset names / status).
- `internal/templates/data/` — embedded templates by preset / language.
- `internal/doctor/` — `aikata doctor` checks (`checks.go`),
  auto-fix engine (`fix.go`), JSON output (`json.go`).
- `internal/adr/` — shared ADR numbering helper.
- `internal/config/` — `.ai/aikata.yaml` schema and (from v0.3.2)
  `.aikata/aikata.yaml` migration helpers.
- `docs/adr/` — Architecture Decision Records, 4-digit zero-padded.
- `docs/decisions/open-questions.md` — undecided design items.
- `dist/` — shippable artifacts attached to releases (Claude Code
  skill in v0.3.1; plugin payload in v0.6).

---

## Pull-request checklist

1. **Branch off `main`.** PRs target `main` directly unless the
   maintainer points you at an integration branch (e.g. `feat/v0.3.x`).
2. **Commit and PR text are English.** This applies even if the
   in-issue conversation is in another language.
3. **Tests.** Add or update tests for code changes. New behaviour
   without a test will not merge unless explicitly waived.
4. **`aikata doctor` clean.** Run it on the repo before pushing.
5. **CHANGELOG.** Add an entry under `[Unreleased]` describing the
   user-visible change.
6. **CI green.** The 3-OS matrix (macOS / Linux / Windows × Go 1.21),
   `golangci-lint`, and the install-script smoke job must all pass.

For docs-only changes, the same English / CHANGELOG / CI rules apply,
but no tests are required.

---

## Architecture Decision Records

Significant design changes get a short ADR under `docs/adr/`.

- Filename: `NNNN-kebab-slug.md`, 4-digit zero-padded. The next
  available number lives in `internal/adr/numbering.go`; until
  `aikata add adr` ships in v0.4 you can run `ls docs/adr` to confirm.
- Frontmatter: the same 5-key block every aikata doc uses.
- Body sections: Context → Decision → Consequences → Alternatives.
- Status starts as `Proposed`; flip to `Accepted` (or `Rejected`,
  `Superseded by NNNN`) on merge.

Open design questions live in
[`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md);
promote one to a full ADR when consensus emerges.

---

## Release flow (for maintainers)

aikata uses GoReleaser triggered by `v*` tags on `main`. The release
flow is described in `ARCHITECTURE.md` §6 and is enforced by
`.github/workflows/`. Contributors should not push tags directly.

---

## Community

aikata is small enough that explicit community norms are kept short:
be kind, assume good faith, and prefer specific feedback over generic
criticism. A formal `CODE_OF_CONDUCT.md` (Contributor Covenant) ships
with the `extended` preset in v1.0 and will adopt it for the
repository at the same time.

If you have a question that does not fit an issue, open a
[GitHub Discussion](https://github.com/shigindo-inc/aikata/discussions)
instead.

---

## License

By contributing, you agree that your contribution is licensed under
the project's [MIT License](./LICENSE).
