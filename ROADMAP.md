---
project: aikata
status: draft
version: 0.2.0
updated: 2026-05-24
audience: [human, agent]
---

# ROADMAP

> Direction, not promises. Versions are scoped by *what they unlock*, not by
> dates. The single ordering rule: **each milestone must leave aikata
> usable by its current user base** — never break Phase N users to ship
> Phase N+1.

For the why behind each item, see [SPEC.md](./SPEC.md). For the how, see
[ARCHITECTURE.md](./ARCHITECTURE.md). For open design questions affecting
sequencing, see [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md).

---

## Phase 1 — Documentation bootstrap ✅ (shipped)

aikata's repository is itself a coherent set of operational markdown
documents so that any AI agent and any new contributor can pick up the
project from `README.md` alone.

Shipped:

- Operational documents split out of the pre-v0.1 planning notes:
  README, AGENTS, SPEC, ARCHITECTURE, ROADMAP, GLOSSARY, CHANGELOG,
  LICENSE, ADRs, open-questions. Original notes preserved in git
  history at commit `ea48abf`.

---

## v0.1 — MVP ✅ (released 2026-05-22)

Validated the "scaffold from canonical templates" loop end-to-end with
two presets and Claude Code as the sole AI-tool target.

Shipped:

- Go project skeleton (`go.mod`, `cmd/aikata`, `internal/`); cobra wired.
- `embed.FS` template loader rooted at `internal/templates/data/`.
- `aikata init --preset minimal | standard` with the `--with-memory`
  opt-in ([ADR 0004](./docs/adr/0004-long-term-memory-slot.md)).
- Interactive `aikata init` via a bufio-based prompt — no huh/lipgloss
  so the Go 1.21 floor holds.
- Atomic generation; `.aikata-proposed/` fallback for non-empty dirs.
- `aikata generate` Claude target (`CLAUDE.md` from `AGENTS.md` + the
  canonical doc set).
- Golden tests for both presets.
- 3-OS CI matrix (macOS / Linux / Windows × Go 1.21) green.
- Tag `v0.1.0`, GoReleaser release, repository flipped to public.

---

## v0.2 — Stack & Language ✅ (released 2026-05-23)

aikata is useful to a Japanese-speaking Flutter or TypeScript developer
and exposes a `doctor` self-check.

Shipped:

- `--preset flutter` and `--preset typescript`, each with a stack
  brief at `docs/stacks/<stack>.md`.
- `--lang ja` via parallel-directory templates (`<base>/en/`,
  `<base>/ja/`) with en fallback and a one-line stdout notice.
- `aikata generate` Cursor target (`.cursor/rules/main.mdc`) and Codex
  pass-through no-op
  ([ADR 0005](./docs/adr/0005-cursor-codex-pass-through.md)); Codex
  reads `AGENTS.md` natively.
- `aikata doctor` initial release — eight read-only consistency checks
  (frontmatter / links / ADR / memory / updated / env / glossary /
  lang-consistency); exit code 0 or 3.
- Locale policy formalized in
  [ADR 0006](./docs/adr/0006-locale-and-japanese-documentation-policy.md):
  English canonical repo docs, Japanese access layer under `docs/`,
  `--lang ja` generated templates remain first-class.
- Tag `v0.2.0`, GoReleaser release.

---

## v0.2.1 — Onboarding patch ✅ (released 2026-05-24)

Removed the "Go required?" misperception by making a no-Go install one
shell line instead of "find the right asset on the Releases page".
External-only changes, no behavioural change to the CLI itself.

Shipped:

- `scripts/install.sh` — POSIX shell script that detects OS / arch,
  downloads the matching release archive, verifies its SHA-256 against
  `checksums.txt`, and drops the binary into `$HOME/.local/bin`
  (override with `AIKATA_INSTALL_DIR`). Pin a tag via
  `AIKATA_VERSION=vX.Y.Z` to skip the unauthenticated GitHub API call.
  Supported targets: linux/{amd64,arm64} and darwin/{amd64,arm64};
  Windows continues to use the manual download path. Served from
  `https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh`
  until an `aikata.dev` redirect exists.
- README "Install" section now documents both paths — manual download
  + checksum verify (secure) and `curl -fsSL ... | sh` (convenience).
  `docs/japanese-users.ja.md` mirrors the addition.
- CI gains an `install-script` job that runs the installer on
  `ubuntu-latest` and `macos-latest` and asserts `aikata --version`
  succeeds. The job pins `AIKATA_VERSION` to the previous release so
  PRs can verify the installer end-to-end without depending on an
  unreleased tag.

Patch release tag `v0.2.1`. No new core features, no schema bump.

---

## v0.3 — Fast follow-up (lightweight)

**Goal**: close out the v0.2 surface with low-risk quality-of-life
improvements, complete the scoped config-directory migration from ADR 0008,
and put a minimal Claude-Code distribution surface in place.

- [ ] `aikata doctor --fix` for the trivially-fixable subset (bump
      stale `updated:` to today, scaffold missing required frontmatter
      keys). Non-trivial fixes stay manual.
- [ ] `aikata doctor --json` machine-readable output.
- [ ] ADR auto-numbering (next number = max existing + 1) — wired into
      the `aikata add adr` command that lands in v0.4, but the numbering
      helper ships now so doctor / scaffold can share it.
- [ ] Normalize `aikata --version` output across install channels. `go
      install` currently reports Go module versions with a `v` prefix
      (`v0.2.1`), while GoReleaser binaries report `0.2.1`; prefer the
      tagged form (`vX.Y.Z`) everywhere, likely by passing GoReleaser's
      tag through ldflags or normalizing in `resolveVersion`.
- [ ] Bring interactive `aikata init` prompts back to parity with the
      supported non-interactive flags. At minimum, ask for document
      language (`en|ja`) and AI-tool targets instead of only project name,
      preset, and long-term memory. Audit the prompt against
      [SPEC.md §4.1](./SPEC.md#41-aikata-init-name) and explicitly mark
      any intentionally deferred questions.
- [ ] Migrate aikata-owned config from `.ai/aikata.yaml` to
      `.aikata/aikata.yaml` per
      [ADR 0008](./docs/adr/0008-aikata-owned-config-directory.md). New
      `init` output should use `.aikata/`; readers should fall back to the
      legacy `.ai/` path with a deprecation warning through v0.x.
- [ ] Promote `ROADMAP.md` into the `standard` preset. It represents the
      durable "when / sequence" layer alongside `SPEC.md` (what / why),
      `ARCHITECTURE.md` (how), and ADRs (decision rationale), so it should
      not be hidden behind a narrow `--with-roadmap` flag.
- [ ] Reserve `extended` as the future heavier preset name. For v0.3.x, it
      may be listed or accepted only as a future-use alias that prints a
      clear message and falls back to `standard`; do not add heavy generated
      content until its document set is justified.
- [ ] `aikata completion bash | zsh | fish | pwsh` — cobra-generated
      shell completion. Documented in README "Install".
- [ ] `aikata list presets | stacks | ai-tools` and
      `aikata describe preset <name>` for discoverability without
      touching templates.
- [ ] **Minimal Claude Code skill** — a single `SKILL.md` that teaches
      Claude when to call `aikata init / generate / doctor` and how to
      read their output. Lives under `dist/claude-code/skill/` in this
      repository and ships as a release asset. **No** slash commands,
      sub-agents, or hooks yet — distribution is "copy one file into
      `~/.claude/skills/`" so the surface stays tiny and reversible.
- [ ] `CONTRIBUTING.md` added to this repository directly (before the
      future `extended` preset grows an operational-readiness pack) so
      external contributors have a clear entry point now that aikata is
      public.

---

## v0.4 — Authoring ergonomics

**Goal**: editing an aikata project is as ergonomic as creating one.

- [ ] `aikata add <component>` — **first wave** (highest user value):
      - `adr` — auto-numbered ADR skeleton (uses the v0.3 numbering
        helper).
      - `stack` — adds `docs/stacks/<name>.md` and updates
        `.aikata/aikata.yaml` `stacks:`.
      - `memory` — opt-in equivalent of `--with-memory` for projects
        that did not enable it at init time.
- [ ] `aikata add <component>` — **second wave** (lower individual
      leverage, useful for completeness):
      - `ai-tool`, `ui`, `api`, `tdd`, `changelog`.
      - `ui` adds optional `UI.md` for UI / UX / product-design guidance;
        do not introduce a generic `DESIGN.md` (ADR 0007).
- [ ] `--with-ui`, `--with-api`, `--with-tdd`, `--with-changelog` flags
      on `aikata init`, matching the second-wave `aikata add` set.
      Skippable if `aikata add` is judged sufficient. When any new init
      flag lands, add the matching interactive prompt in the same change
      so flag / prompt parity does not drift again.
- [ ] Investigate memory generate-projection (ADR-0004 option δ): how
      to mirror `docs/memory/*` into tool-specific channels (Claude
      `.claude/memory/`, Cursor `.cursor/rules/long-term/`). Record
      findings in a new ADR; ship only if the cost is low.

---

## v0.5 — `aikata update`

**Goal**: keep an aikata project current with the canonical templates
without losing the user's edits.

- [ ] `aikata update` interactive diff-merge — the single largest
      feature on the roadmap. Carved out of v0.4 / v0.6 so it gets its
      own release cycle and doesn't drag packaging work with it.
- [ ] Migration framework for `.aikata/aikata.yaml` schema versions
      (needed so `update` can rewrite older configs forward-compatibly).
- [ ] Dogfooding gate becomes binding (see "Dogfooding milestone"
      below).

---

## v0.6 — Packaging & distribution

**Goal**: aikata is one-click installable in Claude Code, one-shell-line
elsewhere, and scales to a monorepo.

- [ ] `--monorepo` initialization with nested `AGENTS.md` per app.
- [ ] npm wrapper for `npx aikata` distribution.
- [ ] Homebrew tap (`shigindo-inc/tap/aikata`) published from the
      release workflow. The v0.2.1 `curl | bash` script remains the
      portable fallback.
- [ ] **Claude Code plugin** — bundle the v0.3 skill with
      `/aikata-init`, `/aikata-generate`, and `/aikata-doctor` slash
      commands under `dist/claude-code/plugin/`. Distributable through
      the public plugin marketplace once the upstream listing flow is
      stable; otherwise as a `git clone` + `.claude/plugins/` symlink.
- [ ] If v0.4 investigation justified it: ship
      `aikata generate --memory` for at least one AI-tool memory
      channel (ADR-0004 option δ).

---

## v1.0 — Stable surface

**Goal**: a surface that downstream tooling can depend on.

- [ ] Major AI tools all supported: Claude, Cursor, Codex, Gemini,
      Copilot, Windsurf.
- [ ] `--preset extended` adds the operational-readiness pack:
      - `CONTRIBUTING.md`, `SECURITY.md`
      - `CODE_OF_CONDUCT.md` (Contributor Covenant)
      - `.github/ISSUE_TEMPLATE/{bug_report,feature_request}.md`
      - `.github/PULL_REQUEST_TEMPLATE.md`
- [ ] Stable preset & template schema (semver guarantee).
- [ ] Official docs site (`aikata.dev`).
- [ ] External preset repositories (`aikata add stack github.com/foo/bar`).
- [ ] **Plugin / skill distribution beyond Claude** — Cursor custom modes
      or rule packs, Gemini CLI extensions, a VS Code extension that
      wraps the CLI. Per-tool scope is driven by H1 dogfooding evidence;
      the Claude plugin (v0.6) defines the surface shape and the others
      mirror it.

---

## v1.x — Beyond bootstrap

Speculative. Order and inclusion depend on validating
[hypotheses H1–H4](./SPEC.md#7-hypotheses-to-validate).

- LLM-API-assisted document drafting (`aikata draft <topic>`).
- VS Code / JetBrains extensions go beyond CLI wrapping (in-editor
  preview of generated docs, ADR scaffolder palette).
- Reverse-analysis of existing projects to suggest an aikata layout
  (agentsmesh-like).
- Bilingual document mode (Japanese for humans, English for LLMs in a
  single canonical file).
- Opt-in version-upgrade self-check (`aikata --version` queries the
  GitHub Releases API and prints "newer version available"). Opt-in
  because of the network call; would be the first time aikata reaches
  out to the network at all.

---

## Distribution surface — release-cadence summary

Cross-cutting view of where aikata can be installed from at each version.
Channels grow monotonically: adding a new channel must never break the
previous one (`go install` stays the canonical baseline).

| Version | go install | GitHub Release | curl \| bash | Claude skill | Claude plugin | npm | Homebrew | Other tools |
|---|---|---|---|---|---|---|---|---|
| v0.1 | ✅ | ✅ | — | — | — | — | — | — |
| v0.2 | ✅ | ✅ | — | — | — | — | — | — |
| v0.2.1 | ✅ | ✅ | ✅ | — | — | — | — | — |
| v0.3 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.4 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.5 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.6 | ✅ | ✅ | ✅ | ✅ | ✅ | `npx aikata` | tap | — |
| v1.0 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Cursor / Gemini / VS Code |

Plugin / skill scope grows monotonically too:

- **v0.3** — "Claude knows when to shell out to aikata." One SKILL.md,
  no commands, no agents.
- **v0.6** — adds `/aikata-init`, `/aikata-generate`, `/aikata-doctor`
  slash commands and is installable as a single plugin.
- **v1.0** — mirrors the v0.6 plugin shape into Cursor, Gemini CLI, and
  a thin VS Code wrapper. Per-tool feature parity is not promised; the
  promise is "you can discover and invoke aikata from your tool's native
  surface."

---

## Dogfooding milestone

A standing goal across phases. Becomes a binding **release gate from
v0.5 onward** (was v0.3 in the original draft; relaxed because v0.3 /
v0.4 still introduce new primitives that legitimately diverge from the
templates).

Pass criteria, all three must hold:

1. `aikata doctor` reports zero errors and zero warnings on the aikata
   repository at the release commit.
2. `aikata init --preset standard` in a clean directory produces a
   project that builds in CI on Linux without further edits.
3. The aikata repository's own `CLAUDE.md`, `.cursor/rules/main.mdc`,
   and any other generated AI-tool artifacts are byte-identical to what
   `aikata generate` produces at the release commit.

The aspirational long-form goal — "the aikata repository is fully
reproducible by `aikata init --preset extended --ai-tools claude,cursor`
plus a manual `git diff` review" — stays as the v1.0 target.

---

## Out-of-scope, indefinitely

These are documented here so future scope-creep proposals can be
deflected.

- IDE GUIs for editing aikata config.
- Real-time rule enforcement (linting source files for style violations).
- Task / issue tracker integration.
- Direct write access to remote git providers (no `aikata push`).
- **aikata as a Claude Code *agent*** (vs a skill or plugin). Skills and
  plugins are scoped distribution surfaces; an agent is a runtime
  personality that competes with the user's own choice of model and
  workflow. aikata is a CLI and ships shapes that wrap it.
