---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
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
/tmp/aikata doctor --all-markdown
```

`aikata doctor` validates the aikata-managed document surface by
default; this repository ships many docs outside that surface
(`CONTRIBUTING.md`, `SECURITY.md`, `docs/decisions/`, `dist/`), so its
own gate uses `--all-markdown` to audit every Markdown file (ADR 0037).
`aikata doctor --all-markdown` should report zero error-level issues on `main`.
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
- `internal/config/` — `.aikata/aikata.yaml` schema and legacy
  `.ai/aikata.yaml` migration helpers.
- `docs/adr/` — Architecture Decision Records, 4-digit zero-padded.
- `docs/decisions/open-questions.md` — undecided design items.
- `dist/` — shippable artifacts attached to releases (the `aikata-cli`
  and `aikata-context` skills + Claude Code / Codex plugins). The
  canonical skill source is `dist/universal-skill/<skill>/SKILL.md`; all
  per-platform copies are byte-identical to it and must not be edited
  directly (ADR 0040).

---

## Pull-request checklist

1. **Branch off `main`; never push to it directly.** Every public
   change — maintainers included — lands through a reviewed PR. PRs
   target `main` unless the maintainer points you at an integration
   branch (e.g. `feat/v0.3.x`). `CODEOWNERS` gates the sensitive
   surface (`/.github/`, `/AGENTS.md`, `/SECURITY.md`,
   `/.goreleaser.yml`, `/ROADMAP.md`, `/docs/adr/`).
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
  available number lives in `internal/adr/numbering.go`; or run
  `aikata new adr "<title>"` to scaffold an auto-numbered ADR.
- Frontmatter: the same 5-key block every aikata doc uses.
- Body sections: Context → Decision → Consequences → Alternatives.
- Status starts as `Proposed`; flip to `Accepted` (or `Rejected`,
  `Superseded by NNNN`) on merge.

Open design questions live in
[`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md);
promote one to a full ADR when consensus emerges. When a question is
resolved, **remove** its entry (the ADR is the durable record) rather
than leaving a pointer — see the documentation-hygiene rubric in
[ADR 0039](./docs/adr/0039-documentation-hygiene-and-context-budget.md),
which release authors apply each release (prune resolved first-read
context; archive released ROADMAP/CHANGELOG detail to `docs/*-archive.md`
by move, not delete; never edit an accepted ADR body).

---

## Release flow (for maintainers)

aikata uses GoReleaser triggered by `v*` tags on `main`, enforced by
`.github/workflows/`. The versioning mechanism is documented in
[`ARCHITECTURE.md` §6.5](./ARCHITECTURE.md#65-versioning--the-release-ritual).
Contributors should not push tags directly.

**The binary version comes from the git tag** (`git describe --tags` →
ldflags), so there is **no version constant to bump in Go source**. A
release is a `chore(release): prepare vX.Y.Z` PR that updates docs and
distribution metadata, merged to `main` immediately before the tag is
pushed. At tag time:

1. **`CHANGELOG.md`** — promote the `## [Unreleased]` entries into a new
   `## [X.Y.Z] - YYYY-MM-DD` section with a short summary paragraph; leave
   a fresh empty `[Unreleased]`.
2. **`ROADMAP.md`** — flip the milestone heading from `(pending)` /
   `(planned)` to `✅ (released YYYY-MM-DD)`.
3. **Plugin distribution metadata** — bump the `version` field in
   `dist/claude-code/plugin/.claude-plugin/plugin.json` (1 place),
   `.claude-plugin/marketplace.json` (2 places: root + the `plugins[0]`
   entry), and `dist/codex/plugin/.codex-plugin/plugin.json` (1 place) to
   the release semver. These stay in **lockstep** with every release so
   marketplace listings reflect the current version.
4. **Agent distribution archives** — run
   `scripts/package-distribution-assets.sh`. The generated `.tar.gz`
   files are ignored locally and attached by GoReleaser.
5. **Binary version** — nothing to edit; `git describe` picks up the new
   tag automatically once it is pushed.
6. **Tag & push** — `git tag vX.Y.Z && git push --tags`; GoReleaser does
   the rest (signed checksums, SBOMs, multi-arch archives).

---

## Never commit secrets

Do not add `.env`, `.env.local`, `*.local.yaml` / `*.local.yml`,
credentials, tokens, private keys, or absolute local paths containing a
user identifier. `.gitignore` keeps the common cases out, and a CI
secret-scan (`internal/repolint`, run under `go test ./...`) fails the
build if any are force-added. Use `.env.example`-style placeholders for
illustrative configuration. See [SECURITY.md](./SECURITY.md) for the
full policy and private disclosure path.

---

## Agent Contributions

aikata is built with AI agents and welcomes AI-assisted changes **when a
human maintainer reviews the final diff before merge.** Agents must
follow `AGENTS.md` and the Agent Safety section of
[SECURITY.md](./SECURITY.md): no direct pushes to protected branches, no
self-merge without human approval, no weakening of the review or
validation gates (`CODEOWNERS`, the secret-scan, `aikata doctor
--strict`, the `aikata generate` byte-identity check), and no AI
attribution trailers in commit messages (`AGENTS.md` rule 6).

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
