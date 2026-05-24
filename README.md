---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
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

> **Status — v0.2.**
> `aikata init` ships four presets (`minimal` / `standard` / `flutter` /
> `typescript`), bilingual templates (`--lang en|ja`), and the long-term
> agent memory slot (`--with-memory`). `aikata generate` emits Claude
> Code (`CLAUDE.md`) and Cursor (`.cursor/rules/main.mdc`) artifacts;
> Codex passes through `AGENTS.md` directly. `aikata doctor` runs eight
> read-only consistency checks.
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
> land in v0.6 — see [ROADMAP.md](./ROADMAP.md).

### Claude Code skill (optional, v0.3.1+)

aikata ships a minimal Claude Code skill that teaches Claude when to
call the CLI and how to parse `aikata doctor --json`. It is a single
file — no slash commands, sub-agents, or hooks — so installing it is
one copy:

```bash
mkdir -p ~/.claude/skills
curl -fsSL -o ~/.claude/skills/aikata.md \
  https://github.com/shigindo-inc/aikata/releases/latest/download/aikata-skill.md
```

See [`dist/README.md`](./dist/README.md) for offline install from a
checkout and notes on the v0.6 plugin migration path.

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

## Quickstart

```bash
# Scaffold a new project (interactive when stdin is a TTY):
aikata init my-app

# Or non-interactively with explicit flags:
aikata init my-app --preset standard --no-interactive

# Stack-flavored presets (v0.2):
aikata init my-flutter-app --preset flutter --no-interactive
aikata init my-ts-app --preset typescript --no-interactive

# Japanese template set (v0.2):
aikata init my-app --preset standard --lang ja --no-interactive

# Opt in to the long-term agent memory slot (ADR 0004):
aikata init my-app --preset standard --with-memory --no-interactive

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
aikata init my-project --preset standard --with-memory --no-interactive --dry-run --force

# If the proposed changes are acceptable:
aikata init my-project --preset standard --with-memory --no-interactive --force
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
- [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md) — what is **not** yet decided.

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
