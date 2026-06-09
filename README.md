---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-07
audience: [human, agent]
---

# aikata

> **aikata** (相方, *ai-kata*, /aɪˈkɑːtə/) — a lightweight CLI that
> scaffolds AI-readable markdown documents and per-AI-tool config files
> in a single command.

The name means "partner" in Japanese: a companion that helps humans and
LLMs collaborate as equals during development.

Japanese users can start from
[`docs/japanese-users.ja.md`](./docs/japanese-users.ja.md).

[![ci](https://github.com/shigindo-inc/aikata/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/shigindo-inc/aikata/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/shigindo-inc/aikata?display_name=tag&sort=semver)](https://github.com/shigindo-inc/aikata/releases)
[![license](https://img.shields.io/github/license/shigindo-inc/aikata)](./LICENSE)

> **Status — v0.11.1.**
> `aikata init` selects a documentation scope (`--scope minimal |
> standard`) and an optional target stack (`--stack flutter |
> typescript`) as orthogonal axes (ADR 0024; `--preset` is a deprecated
> alias), with bilingual templates (`--lang en|ja`), the long-term
> agent memory slot (`--with-memory`), and opt-in capabilities
> (`--with-prompts`, `--with-env`, …). `aikata fill` adopts an existing
> repo or tops up a managed one by writing only the **missing** canonical
> documents, never overwriting existing files (ADR 0042). `aikata
> generate` emits Claude Code (`CLAUDE.md`) and Cursor
> (`.cursor/rules/main.mdc`) artifacts; Codex passes through `AGENTS.md`
> directly. `aikata doctor` validates the aikata-managed document surface
> by default (`--all-markdown` audits every file). Adopting into a
> non-empty directory writes a reviewable `.aikata-proposed/` scaffold
> (ADR 0037), or use `aikata fill` for an in-place, non-destructive top-up.
> [ROADMAP.md](./ROADMAP.md) lists what comes next.

---

## Why aikata?

Modern projects must be readable by **both humans and multiple AI coding
agents** (Claude Code, Cursor, Codex, Gemini CLI, Copilot, Windsurf, …).
Today this means hand-maintaining several near-duplicate instruction
files. aikata fixes that by treating **markdown documents** as the
single source of truth and **generating** tool-specific files from them.

For the long-form rationale, read [SPEC.md](./SPEC.md) §1.

---

## Install

aikata is a single static binary. **Go is not required to use aikata** —
it is only required if you install from source.

### Recommended — pre-built binary (no Go toolchain)

Download the latest archive for your platform from the
[Releases page](https://github.com/shigindo-inc/aikata/releases/latest):

| OS | Architecture | Asset suffix |
|---|---|---|
| macOS | Apple Silicon | `_darwin_arm64.tar.gz` |
| macOS | Intel | `_darwin_amd64.tar.gz` |
| Linux | x86_64 | `_linux_amd64.tar.gz` |
| Linux | arm64 | `_linux_arm64.tar.gz` |
| Windows | x86_64 | `_windows_amd64.zip` |

Extract, verify the checksum, then move the binary onto your `PATH`:

```bash
tar -xzf aikata_*_<os>_<arch>.tar.gz
sha256sum -c checksums.txt 2>&1 | grep aikata    # optional but recommended
mv aikata "$HOME/.local/bin/"                     # or /usr/local/bin (sudo)
aikata --version
```

`checksums.txt` is published alongside the binaries in the same release.

### Verifying a release signature (v0.8.1+)

From v0.8.1, releases are signed with [cosign](https://docs.sigstore.dev/)
keyless signing. `checksums.txt` is signed via GitHub OIDC — no
long-lived key — and the release carries `checksums.txt.sigstore.json`,
a Sigstore bundle that combines the short-lived Fulcio certificate and
the signature (cosign v3 format). Each archive also ships a
syft-generated SBOM (`<archive>.sbom.json`).

To verify the checksum file's signature before trusting it (and, through
it, every archive it lists), download `checksums.txt` and
`checksums.txt.sigstore.json`, then run (cosign v2.4+/v3):

```bash
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/shigindo-inc/aikata/\.github/workflows/release\.yml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

A `Verified OK` confirms the checksum file was produced by aikata's
release workflow on a tag. After that, `sha256sum -c checksums.txt`
authenticates the archive itself. Verification is optional; the manual
`sha256sum -c` check above remains sufficient for integrity.

### Convenience — install script (Linux / macOS, v0.2.1+)

For a one-line install, the project ships a POSIX shell script that
detects your OS / architecture, downloads the matching archive, verifies
its SHA-256 against `checksums.txt`, and drops the binary into
`$HOME/.local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh
```

To pin a specific version (and avoid the unauthenticated GitHub API call
that resolves the latest tag):

```bash
curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh \
  | AIKATA_VERSION=v0.2.1 sh
```

Override the install location with `AIKATA_INSTALL_DIR`. The script
prints a warning if the install directory is not on your `PATH`. Windows
users: continue to use the manual download path above; an installer for
Windows is not planned for the v0.x series.

If you prefer to read the script before executing it (recommended for
any `curl | sh` pattern), download it first:

```bash
curl -fsSL -o install.sh https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh
less install.sh
sh install.sh
```

### From source — `go install` (requires Go 1.21+)

```bash
go install github.com/shigindo-inc/aikata/cmd/aikata@latest
aikata --version
```

`go install` writes the binary to `$GOBIN` when it is set, otherwise to
`$(go env GOPATH)/bin`. That directory must be in `PATH`; see
[`docs/troubleshooting.md`](./docs/troubleshooting.md) if `aikata` is
not found after install.

> A Homebrew tap (`shigindo-inc/tap/aikata`) and an `npx aikata` wrapper
> are deferred to v0.9.x — see [ROADMAP.md](./ROADMAP.md).

### Agent skills & plugins (optional, v0.3.1+)

aikata ships **two** first-party skills (ADR 0040): `aikata-cli` teaches
the agent when to call the CLI and how to parse `aikata doctor --json`,
and `aikata-context` teaches the daily in-repo context-maintenance loop
(which canonical docs to read, where new context belongs, what to check
before handoff). Both ship together as the single `aikata` plugin — there
is no separate install per skill. Pick the surface your agent uses:

```text
# Claude Code (self-hosted marketplace)
/plugin marketplace add shigindo-inc/aikata
/plugin install aikata@aikata
```

```bash
# Codex CLI 0.135.0+ (tracks the default branch — stays current)
codex plugin marketplace add shigindo-inc/aikata
codex plugin add aikata@aikata
```

```bash
# Universal (npx skills; any AGENTS.md-aware agent)
npx skills add https://github.com/shigindo-inc/aikata/tree/main/dist/universal-skill --agent universal
```

Each is a thin wrapper over the aikata CLI — no MCP server, sub-agent, or
app integration, and no slash commands. In Claude Code the two skills are
invoked (or appear in the `/` menu) as `/aikata:aikata-cli` and
`/aikata:aikata-context`; in Codex as `$aikata-cli` / `$aikata-context`.
To **update**: Claude Code `/plugin marketplace update aikata`; Codex
`codex plugin marketplace upgrade aikata && codex plugin add aikata@aikata`;
universal `npx skills update` (add `--global` for the `--agent universal`
install).

See [`dist/README.md`](./dist/README.md) for the full per-surface install,
update, reinstall/migration, standalone-skill, and tag-pinning steps.

### Shell completion (v0.3.1+)

`aikata completion <shell>` prints a completion script for `bash`, `zsh`,
`fish`, or `powershell`. Activate it once per shell:

```bash
# Bash (current shell):
source <(aikata completion bash)

# Zsh (system-wide, with `autoload -Uz compinit; compinit` already in .zshrc):
aikata completion zsh > "${fpath[1]}/_aikata"

# Fish (per-user, persistent):
aikata completion fish > ~/.config/fish/completions/aikata.fish

# PowerShell:
aikata completion powershell | Out-String | Invoke-Expression
```

Run `aikata completion --help` for additional install paths.

### Stay current (v0.4.2+)

```bash
aikata update --check
```

`aikata update --check` reads the GitHub Releases API and reports
whether a newer aikata is available; it never modifies the installed
binary. Add `--json` for the machine-readable envelope shared with
`aikata doctor --json`. Avoid running it in a CI loop — the
unauthenticated API rate limit (60 req/h per IP) is shaped for
human-invoked use. Native self-update is planned for v0.9.x.

### Keep in sync (v0.5+)

```bash
aikata sync
```

`aikata sync` performs a 3-way diff-merge between your project's
current state, the templates as they were when aikata first wrote
them (recorded in `.aikata/manifest.yaml`), and the freshly rendered
upstream templates. User edits are preserved; upstream-only changes
auto-apply; true conflicts are written back with git-merge-style
markers (`<<<<<<<`, `|||||||`, `=======`, `>>>>>>>`) for manual
resolution.

- `aikata sync --dry-run` previews the merge plan without writing.
- `aikata sync --rebaseline` seeds a manifest from the current upstream
  rendering for projects that pre-date v0.5 (one-shot). If a manifest
  already exists it is a no-op with a notice; use `--reseed` to
  deliberately re-anchor an existing manifest.
- `aikata sync --json` emits a machine-readable report with the same
  envelope shape as `doctor` / `list` / `update`.

**Keeping a file you rewrote (v0.8.3):** files you have intentionally
forked for your project are preserved across **every** sync — a
`user-only-edit` is never absorbed and overwritten on a later run
(ADR 0025). To stop aikata from even diff-comparing a file you have
fully taken over (e.g. `README.md`), list it under `sync.own` in
`.aikata/aikata.yaml`:

```yaml
sync:
  own:
    - README.md
    - .gitignore
```

Such paths report the `owned` status and are never compared,
conflict-markered, overwritten, or manifest-tracked.

See [ADR 0011](./docs/adr/0011-aikata-sync-design.md) and
[ADR 0025](./docs/adr/0025-sync-divergent-file-preservation.md) for the full
merge contract.

### Monorepo layout (v0.6+)

```bash
aikata init my-workspace --scope standard --monorepo --no-interactive
```

`--monorepo` adds workspace-style scaffolding on top of the chosen
preset: `apps/README.md` plus `apps/_example/AGENTS.md` (the per-app
rule template) and `docs/monorepo.md` (the convention explainer).
The root `AGENTS.md` still carries repository-wide invariants; per-app
`AGENTS.md` files layer on top inside each `apps/<name>/` tree.

Copy `apps/_example/` to `apps/<your-app>/` to bootstrap a new app:

```bash
cp -r apps/_example apps/web
$EDITOR apps/web/AGENTS.md   # fill in stack, test runner, deployment target
```

aikata does not regenerate per-app `AGENTS.md` files; they are
user-managed. `aikata generate` continues to write the root
`CLAUDE.md` / `.cursor/rules/main.mdc` only.

### Adopting aikata in an existing repository (v0.7.2+)

Already have `AGENTS.md`, a hand-written `CLAUDE.md`, or a
`.gitignore` you care about? See
[`docs/adoption.md`](./docs/adoption.md) for the recommended
migration paths. Highlights:

- `aikata init` without `--force` proposes its scaffold into
  `.aikata-proposed/` so nothing of yours is overwritten.
- `aikata init --force` against an existing `.gitignore` merges
  the aikata-owned block in place via the
  [ADR 0018](./docs/adr/0018-managed-append-for-project-owned-files.md)
  managed-block writer.
- `aikata sync` may add or refresh managed files but never
  silently deletes them when scope narrows
  ([ADR 0019](./docs/adr/0019-sync-missing-file-repair-semantics.md)).

## Quickstart

```bash
# Scaffold a new project (interactive when stdin is a TTY):
aikata init my-app

# Or non-interactively with explicit flags:
aikata init my-app --scope standard --no-interactive

# Add a target stack (v0.8.2 — orthogonal --scope / --stack axes,
# ADR 0024):
aikata init my-flutter-app --scope standard --stack flutter --no-interactive
aikata init my-ts-app --scope standard --stack typescript --no-interactive

# Japanese template set (v0.2):
aikata init my-app --scope standard --lang ja --no-interactive

# Opt in to the long-term agent memory slot (ADR 0004):
aikata init my-app --scope standard --with-memory --no-interactive

# --preset is a deprecated alias for --scope/--stack (removed in v1.0):
aikata init my-flutter-app --preset flutter --no-interactive

# Generate per-AI-tool files. Currently emits CLAUDE.md and
# .cursor/rules/main.mdc; Codex reads AGENTS.md directly:
aikata generate

# Check project consistency (v0.2):
aikata doctor
```

See [SPEC.md §4](./SPEC.md#4-functional-requirements-cli) for the full
command surface.

### Local development install

From a local checkout:

```bash
cd /path/to/aikata
go install ./cmd/aikata

# Or install into a directory that is often already in PATH:
GOBIN="$HOME/.local/bin" go install ./cmd/aikata
```

### Existing repository setup

From the repository root, `aikata init` refuses to write into a non-empty
directory unless `--force` is passed. Always inspect the dry run before
overwriting files:

```bash
aikata init my-project --scope standard --with-memory --no-interactive --dry-run --force

# If the proposed changes are acceptable:
aikata init my-project --scope standard --with-memory --no-interactive --force
```

For an existing repository that already has `AGENTS.md`, the safer minimal
path is to create `.aikata/aikata.yaml` manually:

```bash
mkdir -p .aikata
```

```yaml
version: 1
project:
    name: my-project
    lang: en
ai_tools:
    - claude
```

Then regenerate tool-specific artifacts:

```bash
aikata generate
```

`aikata generate` overwrites generated tool-specific artifacts such as
`CLAUDE.md`; keep canonical instructions in `AGENTS.md`.

---

## Project documents

| Read for… | Document |
|---|---|
| What & Why | [SPEC.md](./SPEC.md) |
| How (technical) | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| When (milestones) | [ROADMAP.md](./ROADMAP.md) |
| Terminology | [GLOSSARY.md](./GLOSSARY.md) |
| Agent / contributor rules | [AGENTS.md](./AGENTS.md) |
| Japanese users | [docs/japanese-users.ja.md](./docs/japanese-users.ja.md) |
| Release notes | [CHANGELOG.md](./CHANGELOG.md) |
| License | [LICENSE](./LICENSE) (MIT) |

### Decisions & design

- [`docs/adr/`](./docs/adr/) — Architecture Decision Records.
  - [0001 — Record Architecture Decisions](./docs/adr/0001-record-architecture-decisions.md)
  - [0002 — `AGENTS.md` is Canonical](./docs/adr/0002-agents-md-as-canonical.md)
  - [0003 — Do-No-Harm Policy](./docs/adr/0003-do-no-harm-policy.md)
  - [0004 — Long-Term Memory Slot](./docs/adr/0004-long-term-memory-slot.md)
  - [0005 — Cursor / Codex Pass-Through](./docs/adr/0005-cursor-codex-pass-through.md)
  - [0006 — Locale / Japanese Documentation Policy](./docs/adr/0006-locale-and-japanese-documentation-policy.md)
  - [0007 — Do Not Generate a Generic `DESIGN.md`](./docs/adr/0007-no-generic-design-md.md)
  - [0008 — aikata-Owned Config Directory](./docs/adr/0008-aikata-owned-config-directory.md)
  - [0009 — Reserve `aikata update` for CLI Version Updates](./docs/adr/0009-update-command-owns-cli-version-updates.md)
  - [0010 — Memory Projection Deferred to v0.6](./docs/adr/0010-memory-projection-deferred-to-v0-6.md)
  - [0011 — `aikata sync` Design](./docs/adr/0011-aikata-sync-design.md)
  - [0012 — Memory Projection Deferral Extended Past v0.6.0](./docs/adr/0012-memory-projection-deferral-extended.md)
  - [0013 — `aikata sync` Scope Derivation Hierarchy](./docs/adr/0013-sync-scope-derivation.md)
  - [0014 — `.aikata/manifest.yaml` is a Living Record](./docs/adr/0014-manifest-living-record.md)
  - [0015 — First-party Skill and Plugin Distribution](./docs/adr/0015-first-party-skill-plugin-distribution.md)
  - [0016 — `.aikata/aikata.yaml` schema v2](./docs/adr/0016-aikata-yaml-schema-v2.md)
  - [0017 — Post-init Command Taxonomy: enable / new (no add)](./docs/adr/0017-post-init-command-taxonomy.md)
  - [0018 — Managed-block Append for Project-owned Files](./docs/adr/0018-managed-append-for-project-owned-files.md)
  - [0019 — `aikata sync` Missing-file Repair Semantics](./docs/adr/0019-sync-missing-file-repair-semantics.md)
  - [0020 — Retire the legacy `.ai/aikata.yaml` config fallback](./docs/adr/0020-retire-ai-config-fallback.md)
  - [0021 — `doctor` Scope and Exclusion Semantics](./docs/adr/0021-doctor-scope-and-exclusion.md)
  - [0022 — v0.8.x Security & Governance Hardening](./docs/adr/0022-v0-8-security-governance-hardening.md)
  - [0023 — Release Signing and Supply-chain Hardening](./docs/adr/0023-release-signing-and-supply-chain.md)
  - [0024 — Split `--preset` into Orthogonal `--scope` and `--stack` Axes](./docs/adr/0024-scope-stack-axes-split.md)
  - [0025 — `aikata sync` Divergent-file Preservation](./docs/adr/0025-sync-divergent-file-preservation.md)
  - [0026 — Workflow Guides as Opt-in Collaboration Documents](./docs/adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)
  - [0027 — Verification Expectation in Generated Templates](./docs/adr/0027-verification-expectation-in-generated-templates.md)
  - [0028 — Prioritize Core-concept Stabilization](./docs/adr/0028-prioritize-core-concept-stabilization.md)
  - [0029 — Code-free Stack-brief Layout Convention](./docs/adr/0029-stack-brief-layout-convention.md)
  - [0030 — Trim Stack Briefs to Standard-aligned Guardrails](./docs/adr/0030-trim-stack-briefs-to-standard-guardrails.md)
  - [0031 — Brand Exploration Documents as One-off Artifacts](./docs/adr/0031-brand-exploration-documents-as-one-off-artifacts.md)
  - [0032 — Split the Channel-publication Line by Distribution Value](./docs/adr/0032-split-channel-publication-by-distribution-value.md)
  - [0033 — Direction for `doctor`'s Default Validation Scope](./docs/adr/0033-doctor-default-scope-direction.md)
  - [0034 — Move the Reusable-prompts Library to an Opt-in Capability](./docs/adr/0034-reusable-prompts-opt-in-capability.md)
  - [0035 — Native Self-update (`aikata update --apply`) Safety Model](./docs/adr/0035-native-self-update-safety.md)
  - [0036 — Ship Codex Native Distribution in v0.9.6](./docs/adr/0036-codex-native-distribution.md)
  - [0037 — Tighten Adoption Mutation Boundaries in v0.9.7](./docs/adr/0037-tighten-adoption-mutation-boundaries.md)
  - [0038 — Unify `.gitignore` on the Managed Block Across init and sync](./docs/adr/0038-unify-gitignore-managed-block-across-init-and-sync.md)
  - [0039 — Documentation Hygiene & Context Budget](./docs/adr/0039-documentation-hygiene-and-context-budget.md)
  - [0040 — Collaboration-operation Skill Split (aikata-cli + aikata-context)](./docs/adr/0040-collaboration-operation-skill-split.md)
  - [0041 — Skills-only Surface & Claude Code Plugin Skill Layout](./docs/adr/0041-skills-only-surface-and-plugin-skill-layout.md)
  - [0042 — `fill` Command for Canonical Document Completion](./docs/adr/0042-fill-command-for-canonical-document-completion.md)
- [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md) — what is **not** yet decided.
- [`docs/adoption.md`](./docs/adoption.md) — adopting aikata in an existing repository.

### AI-tool entry points

- [`CLAUDE.md`](./CLAUDE.md) — thin Phase 1 wrapper that points Claude Code at
  `AGENTS.md`. Will be regenerated by `aikata generate` once that command
  ships (see [ADR 0002](./docs/adr/0002-agents-md-as-canonical.md)).

---

## Differentiation in one sentence

> Existing AI-scaffolding tools center on **rules**; aikata centers on
> **documents that both humans and LLMs read** — opinionated where it
> reduces friction, silent where it would harm users who opt out.

For the comparison table, see [SPEC.md §1.3](./SPEC.md#13-differentiation).

---

## Contributing

External contributors: start from [`CONTRIBUTING.md`](./CONTRIBUTING.md).
It covers the quick-start build, where things live, the PR checklist,
and the ADR workflow.

Maintainers and AI agents: the canonical operational rules live in
[AGENTS.md](./AGENTS.md). CONTRIBUTING.md is the human-friendly summary;
AGENTS.md is what aikata itself dogfoods and what `aikata generate`
projects into per-tool config files.

---

## License

MIT — see [LICENSE](./LICENSE).
