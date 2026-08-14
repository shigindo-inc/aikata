---
project: aikata
status: draft
version: 0.0.1
updated: 2026-08-14
audience: [human, agent]
---

# ARCHITECTURE — How

> This document explains **how** aikata is built and how it produces its
> output. For **what / why**, read [SPEC.md](./SPEC.md). For the **when**,
> read [ROADMAP.md](./ROADMAP.md). Individual decisions live under
> [`docs/adr/`](./docs/adr/).

The goal of this document is that a new contributor (human or LLM) can
start a Go project initialization (`go mod init`, directory scaffolding,
`cmd/aikata/main.go`) **without re-reading the origin documents**. All
prerequisite decisions are captured here.

---

## 1. Implementation Language

**Decision**: Go 1.21+.

**Rationale**:

- Single-binary distribution (`curl | sh`, Homebrew tap, GitHub Releases).
- Cross-compilation to macOS / Linux / Windows is first-class.
- Mature CLI ecosystem (`cobra`, `charmbracelet/huh`, `lipgloss`).
- aikata's lightweight philosophy maps better to a self-contained Go binary
  than to a Node-dependent CLI.
- An npm wrapper can be published later (`npx aikata`) by shelling out to
  the Go binary — Go-first does not foreclose npm distribution.

**Alternatives considered**: Rust (steeper learning curve, no clear win),
TypeScript (Node dependency hurts the "lightweight" story), shell
(insufficient for templated generation).

**Minimum version**: Go 1.21+ (stable, has `embed.FS` mature, `slog` available).

---

## 2. Repository Layout

```
aikata/
├── cmd/
│   └── aikata/
│       └── main.go              # Entry point only — no business logic
├── internal/                    # Not importable from outside the module
│   ├── cli/                     # cobra command definitions
│   ├── scaffold/                # File generation (init / add)
│   ├── doctor/                  # Consistency checks
│   ├── generate/                # AI-tool-facing artifact generation
│   ├── config/                  # .aikata/aikata.yaml read/write + path resolver
│   ├── presets/                 # Preset registry
│   └── templates/               # Template loader (wraps embed.FS)
│       └── data/                # Embedded markdown templates (//go:embed all:data)
│           ├── base/            # Shared partials
│           ├── presets/
│           │   ├── minimal/
│           │   ├── standard/
│           │   └── flutter/     # (planned v0.2)
│           └── ai_tools/
│               ├── claude/
│               ├── cursor/
│               └── codex/
├── docs/
│   ├── adr/                     # Architecture Decision Records
│   ├── decisions/               # Open questions, design notes
│   ├── origin/                  # Historical planning docs (do not edit)
│   ├── stacks/                  # (planned) per-stack guides
│   ├── workflows/               # (optional) collaboration workflow guides
│   └── tasks/
│       └── current.md           # Agent's short-term working state
├── examples/                    # Real-world `aikata init` outputs
├── testdata/
│   └── golden/                  # Golden-test expected outputs
├── .github/
│   └── workflows/
├── SPEC.md
├── ARCHITECTURE.md
├── ROADMAP.md
├── GLOSSARY.md
├── AGENTS.md
├── README.md
├── CHANGELOG.md
├── LICENSE
├── go.mod                       # (created during Phase 2 — not yet present)
├── go.sum                       # (created during Phase 2)
├── Makefile                     # (created during Phase 2)
├── .golangci.yml                # (created during Phase 2)
└── .gitignore
```

> **Phase 1 vs Phase 2**: As of Phase 1 (this commit), only the markdown
> documents, `LICENSE`, and `.gitignore` are present. `go.mod`, `cmd/`,
> `internal/`, `templates/`, `testdata/`, `.github/`, `Makefile`,
> `.golangci.yml` are introduced in Phase 2 (Go project initialization,
> tracked separately).

### 2.1 Package responsibilities

| Package | Responsibility | Boundaries |
|---|---|---|
| `cmd/aikata` | `main()`, flag parsing entrypoint | Calls into `internal/cli`. Must contain **no business logic** (testable target). |
| `internal/cli` | cobra commands, flag definitions, user-facing messages | Translates flags into typed requests for lower layers. No file I/O. |
| `internal/scaffold` | Generate files for `init` and `add` | Reads templates, writes atomically. No knowledge of specific AI tools. |
| `internal/doctor` | Consistency checks | Read-only by default; `--fix` writes. |
| `internal/generate` | Produce AI-tool-facing artifacts | Per-tool plug-in interface. |
| `internal/config` | Parse / write `.aikata/aikata.yaml` | YAML schema and migration logic. |
| `internal/presets` | Preset registry; merges presets with flags | Pure logic, no I/O. |
| `internal/templates` | Wraps `embed.FS`, performs Go-template rendering | Knows about template syntax; not about presets. |
| `internal/docmeta` | Shared document-metadata parsing (frontmatter + Markdown link extraction) | Pure logic, no I/O; reused by `doctor` and `docmap` so link parsing cannot drift. |
| `internal/docmap` | Build the doc map (`.aikata/docmap.yaml` / `.aikata/docmap.md`) from the document surface | Reads documents only — no source code. See §13. |

---

## 3. Generated Project Structure

This is the layout aikata produces when a user runs `aikata init` on a
brand-new project (not aikata itself). The structure is the **product**;
the repository layout in §2 is the **producer**.

> For a single-page, cross-referenced **index** of every recommended path
> (role · governing capability/ADR · whether `doctor` validates it by
> default), see [`docs/layout.md`](./docs/layout.md). This section is the
> narrative; `layout.md` is the tabular governance index.

### 3.1 Default (`aikata init --scope standard`)

```
/<project>
├── README.md              # Thin nav for humans + LLMs
├── AGENTS.md              # Canonical agent instructions
├── SPEC.md                # What / Why
├── ARCHITECTURE.md        # How
├── GLOSSARY.md            # Terminology
├── .gitignore
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/            # Populated per preset
│   ├── tasks/
│   │   └── current.md     # Agent's short-term working state
│   └── troubleshooting.md
└── .aikata/
    ├── aikata.yaml        # Project config (schema v2)
    ├── manifest.yaml      # Rendered-file hashes for `sync` (ADR 0014)
    ├── docmap.yaml        # Doc map — structured data (derived; ADR 0044)
    └── docmap.md          # Doc map — tree + Mermaid link-graph (derived)
```

> The `.aikata/` machine zone is written by `init` itself: `aikata.yaml` +
> `manifest.yaml` during the scaffold, and the two `docmap.*` renderings as
> the final, isolated step (also re-run by `fill` / `enable` / `sync` /
> `generate`; see §13). `docmap.*` are aikata-owned derived state — not
> manifest-tracked, not subject to `sync` (§3.4).

> `.env.example` (`--with-env` / `enable env`, ADR 0037) and
> `docs/prompts.md` (`--with-prompts` / `enable prompts`, ADR 0034) are
> opt-in capabilities and are **not** part of the default `standard`
> scaffold; see §3.2.

### 3.2 Optional files

| File / directory | Triggered by | First shipped | Purpose |
|---|---|---|---|
| `UI.md` | `--with-ui` or UI-style preset | v0.4.1 | UI / UX / product-design guidelines |
| `API.md` | `--with-api` or API-style preset | v0.4.1 | API interface spec |
| `docs/testing.md` | `--with-tdd` | v0.4.1 | Test strategy |
| `CHANGELOG.md` | `--with-changelog` | v0.4.1 | Release notes |
| `docs/prompts.md` | `--with-prompts` / `enable prompts` | v0.9.2 | Reusable-prompt library (opt-in; was a default through v0.9.1). See [ADR 0034](./docs/adr/0034-reusable-prompts-opt-in-capability.md). |
| `.env.example` | `--with-env` / `enable env` | v0.9.7 | Environment-variable template (opt-in; was a default through v0.9.6). The `.env` secret ignore is unconditional and independent of this capability. See [ADR 0037](./docs/adr/0037-tighten-adoption-mutation-boundaries.md). |
| `docs/memory/` (5 files) | `--with-memory` | v0.2 | Long-term agent memory (`user`, `feedback`, `project`, `reference` + `README`). See [ADR 0004](./docs/adr/0004-long-term-memory-slot.md). |
| `docs/usecases.md` + `docs/domain.md` | `--with-modeling` / `enable modeling` | Unreleased | Opt-in use-case ledger (behaviour) + domain model (structure), rendered together as a pair. Filled one feature at a time by the `model-feature` skill. See [ADR 0047](./docs/adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md). |
| `docs/workflows/<domain>.md` | `enable workflow <domain>` | v0.8.4 | Opt-in collaboration policy (git first). `AGENTS.md` gains a short pointer only, never the policy body. See [ADR 0026](./docs/adr/0026-workflow-guides-as-opt-in-collaboration-docs.md). |
| `docs/monorepo.md` + `apps/**/AGENTS.md` | `--monorepo` / `enable monorepo` | v0.7.1 | Nested per-app instructions (see §3.5). |
| `docs/design/*.md` | `new app-icon` / `new mascot` | v0.9.2 | One-off brand-exploration artifacts (not a capability; not manifest-tracked, `sync` does not restore them). See [ADR 0031](./docs/adr/0031-brand-exploration-documents-as-one-off-artifacts.md). |
| `CONTRIBUTING.md` | `--oss` | v1.0 | Contributor guide |
| `SECURITY.md` | `--oss` | v1.0 | Security policy |
| `ROADMAP.md` | `--oss` | v1.0 | Roadmap |

### 3.3 File-level responsibilities

(Summarized; each file's contract is enforced by `aikata doctor` once it ships.)

- **`README.md`** — human-primary; ≤ 100 lines; navigation + quickstart.
- **`AGENTS.md`** — agent-primary; ≤ 200 lines; canonical instructions.
- **`SPEC.md`** — What / Why; no implementation detail.
- **`ARCHITECTURE.md`** — How; no individual decision logs (those are ADRs).
- **`GLOSSARY.md`** — terminology pin; ja/en bilingual when applicable.
- **`docs/adr/`** — one ADR per decision; never edited after `Accepted`.
- **`docs/tasks/current.md`** — frequently rewritten by the agent; isolated.
- **`docs/troubleshooting.md`** — common failures & fixes; the first stop
  when stuck (per `AGENTS.md`).
- **`docs/stacks/<name>.md`** — per-stack brief; code-free, included into
  `AGENTS.md` only when the stack is enabled. See
  [ADR 0029](./docs/adr/0029-stack-brief-layout-convention.md) /
  [ADR 0030](./docs/adr/0030-trim-stack-briefs-to-standard-guardrails.md).
- **`docs/memory/`** — long-term agent memory, opt-in (`--with-memory`).
  Four content files (`user`, `feedback`, `project`, `reference`) + a
  `README`. Mutable but **append-only / supersede-in-place**: facts are
  added or superseded, never silently deleted — the `log`-class lifetime of
  [ADR 0045](./docs/adr/0045-documentation-value-model.md), defined in
  [ADR 0004](./docs/adr/0004-long-term-memory-slot.md). Distinct from the
  short-term, freely-rewritten `docs/tasks/current.md`.
- **`docs/workflows/<domain>.md`** — opt-in collaboration policy; the full
  policy lives here while `AGENTS.md` carries only a pointer
  ([ADR 0026](./docs/adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)).
- **`docs/design/*.md`** — one-off brand-exploration scaffolds from
  `new app-icon` / `new mascot`; authored once, not regenerated or synced
  ([ADR 0031](./docs/adr/0031-brand-exploration-documents-as-one-off-artifacts.md)).

There is intentionally no generic `DESIGN.md` in built-in presets. Product
requirements live in `SPEC.md`, technical design lives in
`ARCHITECTURE.md`, decision rationale lives in ADRs, and UI / UX guidance
belongs in optional `UI.md` when enabled. See
[ADR 0007](./docs/adr/0007-no-generic-design-md.md).

### 3.4 File write disciplines

Every command decides per file **how** it touches an existing path. There
are exactly seven disciplines; this table is the single reference (the
behaviour is otherwise spread across ADRs 0002 / 0011 / 0018 / 0019 /
0025 / 0031 / 0037 / 0038 and the `internal/scaffold`,
`internal/components`, `internal/sync`, and `internal/generate`
packages). When adding a code path that writes a project file, pick one
of these — do not invent an eighth.

| # | Discipline | On absent | On existing | Used by | Code | ADR |
|---|---|---|---|---|---|---|
| 1 | **Overwrite (disposable)** | write | **overwrite** unconditionally | `aikata generate` artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`) | `internal/generate` `writeAll` | 0002 |
| 2 | **Atomic full-tree write** | write all-or-nothing | overwrite (only reached with `--force`; else → #6) | `aikata init` greenfield scaffold | `scaffold.writeAll` | — |
| 3 | **Managed-append (block)** | write the framed block (markers + body) | merge: replace only the `# >>> aikata managed >>>` block, byte-preserve user lines | `.gitignore` at init **and** `aikata sync` time | `scaffold.contentForWrite` + `sync.classifyAndMerge` + `internal/managed` (`ApplyBlock` / `Frame` / `IsAppendPath`) | 0018 / 0038 |
| 4 | **Create-or-skip (`writeIfMissing`)** | write | **skip** + notice (never overwrite) | single-file capabilities (`enable ui/api/tdd/changelog/prompts/env`, `memory`) | `singleFile.Add` → `writeIfMissing` | 0004 / 0034 / 0037 |
| 5 | **Refuse-on-collision** | write | **error** (refuse; leave untouched) | one-off artifacts (`new adr/app-icon/mascot`) | `oneOffArtifact.Add` | 0031 |
| 6 | **Proposal fallback** | n/a | render the whole scaffold under `.aikata-proposed/`, exit 0; refuse if that tree is non-empty (`ErrProposalExists`) | `aikata init` in a non-empty dir without `--force` | `scaffold.Run` | 0037 |
| 7 | **3-way merge** | re-create unless `user-deleted` (0019) | merge vs manifest ancestor: `user-only-edit` preserved, conflicts get git-style markers | `aikata sync` | `internal/sync` | 0011 / 0025 |

Two cross-cutting modifiers sit on top of the table:

- **`owned` / skip** — any path matching `sync.own` is exempted from #7
  entirely (never compared, merged, conflict-markered, or
  manifest-tracked). ADR 0025 D2.
- **Manifest tracking** — disciplines #2, #3, #4 record the rendered
  path in `.aikata/manifest.yaml` so #7 has an ancestor; #5 deliberately
  records nothing (the artifact becomes project-owned immediately); #1
  artifacts are never manifest-tracked (disposable). ADR 0014.
  Managed-append paths (#3) are an exception on `sync`: they re-run the
  block merge directly rather than the #7 hash 3-way, so in steady state
  the manifest hash is recorded but not consulted for them — the on-disk
  file carries the framed block while the manifest holds the raw body, so
  a hash compare would always mismatch. The one place the hash *is* read
  is the one-time migration of a pre-0.9.8 markerless file. Both are
  intentional; see ADR 0038.

Config files (`.aikata/aikata.yaml`, `.aikata/manifest.yaml`) are always
written atomically (temp + rename, `internal/config`), with lazy
schema-migration rewrites; they are aikata-owned state, not subject to
the disciplines above. The doc-map outputs (`.aikata/docmap.yaml`,
`.aikata/docmap.md`, ADR 0044) are the same kind of aikata-owned state:
written atomically, regenerated rather than merged, **not** manifest-tracked
and **not** subject to discipline #7 — their freshness is guaranteed by the
`aikata doctor` check (§13), not by the manifest ancestor.

### 3.5 Monorepo layout (`--monorepo` / `enable monorepo`)

When the monorepo capability is enabled, the structure gains a nested
per-app tier on top of §3.1: the root `AGENTS.md` holds shared,
cross-app rules and each app carries its own `AGENTS.md` for local
instructions.

```
/<project>
├── AGENTS.md              # Root: shared, cross-app rules
├── docs/
│   └── monorepo.md        # How the nested layout works
└── apps/
    ├── README.md
    └── <app>/
        └── AGENTS.md       # Per-app instructions (extends the root)
```

Default scope still caps the project root per top-level minimalism; the
nested `apps/**/AGENTS.md` files are not root-level and do not count
against it.

---

## 4. Configuration File: `.aikata/aikata.yaml`

> **Path note**: v0.3.2 onward writes `.aikata/aikata.yaml` per
> [ADR 0008](./docs/adr/0008-aikata-owned-config-directory.md).
> v0.7.4 removes the pre-v0.3.2 `.ai/aikata.yaml` fallback per
> [ADR 0020](./docs/adr/0020-retire-ai-config-fallback.md); the only
> supported config path is `.aikata/aikata.yaml`.

### 4.1 Schema (v2)

```yaml
version: 2                       # Required. Migration anchor.
project:
  name: my-app                   # Required. From `aikata init [name]`.
  lang: en                       # en | ja. Default: en.
  description: "Short summary"   # Optional.

ai_tools:                        # Empty list = no `aikata generate` output.
  - claude
  - cursor
  - codex

stacks:                          # Empty list = stack-agnostic.
  - flutter

workflows:                       # Empty/absent = no workflow guides.
  - git                          # Opt-in collaboration guides under
                                 # docs/workflows/<domain>.md (ADR 0026).
                                 # List axis like stacks / ai_tools, not a
                                 # components: boolean. Enable via
                                 # `aikata enable workflow <domain>`.

components:                      # Schema-v2 explicit template-scope flags.
  memory: false                  # See ADR 0016. All default false.
  ui: false
  api: false
  tdd: false
  changelog: false
  monorepo: false
  prompts: false                 # Opt-in reusable-prompt library (ADR 0034).

features:                        # Non-scope ergonomic toggles.
  obsidian_hints: false

docs:
  task_file_location: docs/tasks/current.md

docmap:                          # Optional. Doc map surface & formats (ADR 0044).
  formats: [yaml, md]            # Renderings to emit. txt/json/mmd addable.
  targets: ["**/*.md"]           # Documents to catalog (default: all Markdown).
  exclude: [".aikata/**"]        # Additional skips; generated AI-tool paths
                                 # and docmap's own outputs are always excluded.
                                 # Same matcher as doctor.exclude / sync.own.

sync:                            # Optional. `aikata sync` preferences (ADR 0025).
  own:                           # Globs the user has taken ownership of;
    - README.md                  # reported `owned`, never compared, conflict-
    - .gitignore                 # markered, overwritten, or manifest-tracked.
                                 # Same matcher as doctor.exclude (ADR 0021).

overrides:                       # Per-tool fine-tuning.
  claude:
    output: CLAUDE.md
    include: [AGENTS.md, SPEC.md, ARCHITECTURE.md, GLOSSARY.md]
  cursor:
    output_dir: .cursor/rules/
    split_by: domain
```

### 4.2 Compatibility rules

- `version` is **required**. Missing → error.
- `aikata` must accept any future minor extension of the current version
  without crashing; unknown keys are logged as warnings.
- `version: 1` triggers the v1 → v2 forward migration in
  `internal/config/schema_migrate.go` (ADR 0016). The migrator lifts
  `features.tdd` / `features.monorepo` into `components:` and is applied
  lazily — read-only callers see the in-memory v2 shape, and the
  on-disk file is rewritten only when a writer (`aikata generate`,
  `aikata sync`, `aikata doctor --fix`) persists it.
- A future `version: 3` will land the same way: a forward migrator at
  the corresponding registry slot and lazy rewrite.

---

## 5. Templates

### 5.1 Engine

- Go's standard `text/template` (NOT `html/template`; the output is
  markdown).
- Delimiters: default (`{{` … `}}`). No escaping plugins.
- Functions exposed: `lower`, `upper`, `title`, `joinSlash`, `now` (frozen
  during `--dry-run` for golden-test stability).

### 5.2 Embedding

- All templates live under `internal/templates/data/` and are embedded
  by the `internal/templates` package via `//go:embed all:data`. Keeping
  the embed root inside the package avoids `..`-relative embed paths,
  which the Go `embed` directive rejects.
- Loaded once at startup; no runtime filesystem lookup.
- This implies: rebuilding the binary is required to update templates
  (acceptable for v0.x; `--templates-dir` override is a v1 candidate).

### 5.3 Frontmatter convention

Every generated markdown file carries:

```yaml
---
project: <name>
status: draft|active|archived
version: <semver>
updated: <YYYY-MM-DD>      # Computed at generation time
audience: [human, agent]   # `agent` only for AGENTS.md
---
```

`aikata doctor` enforces the presence of these keys.

---

## 6. Distribution & Generated Artifacts

### 6.1 aikata's own repository

- `CLAUDE.md`, `.cursor/rules/`, etc. are **committed** when they exist.
- `.gitignore` does **not** include `.aikata/`.
- Reason: a contributor cloning aikata must be able to open Claude Code /
  Cursor immediately. See
  [ADR 0003 — Do-No-Harm Policy](./docs/adr/0003-do-no-harm-policy.md).

### 6.2 Default for `aikata init` output

- The scaffolded `.gitignore` managed block (ADR 0018 markers) ignores
  only the aikata-owned residue: `/.aikata-proposed/`, the generated
  AI-tool artifact paths, and a minimal **always-on** secret baseline
  (`.env`, `.env.local`). The secret baseline is independent of the `env`
  capability (ADR 0037 D2): a `.env` ignore for an absent file is
  harmless, and aikata already preaches "never commit secrets."
- `.aikata/` is **not** ignored: its config and manifest are project
  state the user should commit so `aikata sync` has a stable baseline
  (ADR 0037 D2).
- Stack build outputs, editor / OS files, and coverage directories are
  **not** managed by aikata — those policies belong to the downstream
  project (ADR 0037 D2).
- To stop aikata from managing `.gitignore` at all, list it under
  `sync.own` ([ADR 0025](./docs/adr/0025-sync-divergent-file-preservation.md));
  the inert `docs.generate_gitignore` flag was removed in v0.8.3.

### 6.3 Binary distribution channels (planned)

- GitHub Releases (multi-arch binaries via GoReleaser).
- Homebrew tap: `brew install aikata-dev/tap/aikata`.
- `curl -sSL https://aikata.dev/install.sh | sh`.
- npm wrapper: `npx aikata` (post-v0.4).

### 6.3.1 Agent skill and plugin distribution

- The first-party surface is **three capability-named skills with thin
  Claude Code command wrappers** (ADR 0040, ADR 0041, ADR 0043):
  `manage-docs` (CLI invocation guidance), `track-context` (the in-repo
  context-maintenance loop), and `refresh-docs` (bring a downstream repo
  up to the latest aikata). All ship from the single `aikata` marketplace
  entry and plugin. The plugin adds `commands/manage-docs.md` and
  `commands/refresh-docs.md`; the skills are `user-invocable: false` to
  avoid double-listing. Codex / standalone / universal are skill-only and
  model-invoked (the command surface is Claude Code plugin-only — an
  accepted platform asymmetry, ADR 0043).
- `dist/universal-skill/<skill>/SKILL.md` is the **single canonical
  source** of each skill's content. Each carries an `agents/openai.yaml`
  with minimal Codex App UI metadata while keeping the skill tool-agnostic.
- **All platforms use the `<base>/<skill>/SKILL.md` directory layout**
  (ADR 0041): the Claude Code plugin auto-discovers `skills/<name>/SKILL.md`
  (manifest at `.claude-plugin/plugin.json`, metadata only) and surfaces the
  skills namespaced as `/aikata:<skill>`; Codex reads `skills/<name>/SKILL.md`;
  the universal layout uses `.agents/skills/<name>/SKILL.md`.
- `dist/codex/plugin/` (Codex CLI `0.135.0+`) and
  `dist/claude-code/{skill,plugin}/` are byte-identical copies of the
  canonical sources; copies exist only for per-platform discovery
  location, never for content. Repository tests
  (`TestSkillCopiesMatchCanonical`, `TestClaudePluginSkillsAreAutoDiscoverable`,
  `TestClaudePluginHasCommands`) enforce the copy boundary, layout, and
  command-wrapper surface.
- `.agents/plugins/marketplace.json` exposes the tracked Codex plugin as a
  self-hosted marketplace plugin. Older Codex versions keep using direct
  `.agents/skills/` discovery, which works on CLI `0.125.0`.
- `scripts/package-distribution-assets.sh` creates ignored release
  archives for the complete two-skill universal and Codex plugin trees
  before GoReleaser uploads them. See
  [ADR 0036](./docs/adr/0036-codex-native-distribution.md) and
  [ADR 0040](./docs/adr/0040-collaboration-operation-skill-split.md).

### 6.4 CLI self-update model

- `aikata update` is reserved for updating the aikata CLI binary,
  matching the user-facing convention established by Claude Code's
  `claude update` (ADR 0009). `--check` reports availability; `--apply`
  (shipped v0.9.4, [ADR 0035](./docs/adr/0035-native-self-update-safety.md))
  performs the update, routed by `internal/install.Detect()`.
- **Verify before swap** (`install-script` / `github-release`): download
  the host release archive and `checksums.txt`, verify the archive's
  SHA-256 **before extracting**, extract the `aikata` entry, then replace
  the running binary atomically (temp file in the same directory →
  `os.Rename`; POSIX keeps the live process's old inode). SHA-256 over
  HTTPS-to-github is the transport trust anchor; cosign verification is a
  documented non-goal (ADR 0035 D2). The download/verify boundary lives in
  `internal/release`; extraction and the atomic replace live in
  `internal/selfupdate`, which always takes an injected `exePath`.
- **`go install`** installs are shown the channel-native
  `go install …@latest` rather than swapping a GoReleaser binary over a
  toolchain-built one.
- **Package-manager installs** remain package-manager-owned: `aikata
  update --apply` prints an actionable `brew upgrade` / npm command
  (Homebrew / npm channels themselves are deferred to v0.9.9) and never
  overwrites a package-manager-managed binary directly.
- **Windows** cannot overwrite a running `.exe`, so `--apply` returns
  manual-download / `go install` guidance before any download. Unknown
  installs and permission-denied paths likewise get actionable messages.
- `scripts/install.sh` writes install-source metadata
  (`aikata.install-source`) next to the binary so the CLI can distinguish
  native, Homebrew, npm, Go, and unknown installs before choosing a path.

### 6.5 Versioning & the release ritual

**There is no version constant in source.** The binary version is the
git tag, injected at build time:

- `Makefile` sets `VERSION ?= $(shell git describe --tags --dirty || echo
  "0.0.1-dev")` and passes it via `LDFLAGS := -X main.version=$(VERSION)`.
- `cmd/aikata/main.go` declares `var version = devVersion` where
  `devVersion = "0.0.1-dev"`; GoReleaser / `make` overwrite it through the
  ldflags above. GoReleaser builds (triggered by a `v*` tag) inject the
  bare semver string.

The practical consequence: **cutting a release never edits a version field
in Go source.** The tag is the source of truth. What a release *does*
touch is documentation, and it must be kept in sync at tag time:

| At `git tag vX.Y.Z` time, update | How |
|---|---|
| `CHANGELOG.md` | Promote the `## [Unreleased]` entries into a new `## [X.Y.Z] - YYYY-MM-DD` section with a one-paragraph summary. Leave a fresh empty `[Unreleased]`. |
| `ROADMAP.md` | Flip the milestone heading from `(pending)` / `(planned)` to `✅ (released YYYY-MM-DD)`. |
| `plugin.json` + `marketplace.json` | Bump `version` in `dist/claude-code/plugin/.claude-plugin/plugin.json` (1 place), `.claude-plugin/marketplace.json` (2 places), and `dist/codex/plugin/.codex-plugin/plugin.json` (1 place) to the release semver — **lockstep** with every release so marketplace listings match. |
| Agent distribution archives | Run `scripts/package-distribution-assets.sh` before GoReleaser. It creates ignored `dist/aikata-universal-skill.tar.gz` and `dist/aikata-codex-plugin.tar.gz` release assets (each carrying both skills). From v0.10.0 the single-file `.md` skill assets are dropped — a single file can no longer represent the two-skill surface (ADR 0040). |
| Binary version | **Nothing** — `git describe` picks up the new tag automatically. |

This ritual is performed in a `chore(release): prepare vX.Y.Z` PR that
merges to `main` immediately before the tag is pushed (see the
maintainer release flow in
[CONTRIBUTING.md](./CONTRIBUTING.md#release-flow-for-maintainers)).
Contributors never push tags directly.

---

## 7. Error Handling & Logging

### 7.1 Errors

- **No `panic`** in production code paths. Lib usage from other Go
  programs is allowed.
- Wrap errors with context: `fmt.Errorf("scaffold %s: %w", name, err)`.
- User-facing messages are **actionable**: name the file / flag /
  config-key that needs attention.

### 7.2 Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Generic runtime error |
| 2 | User input error (bad flag, malformed YAML, missing file) |
| 3 | `aikata doctor` found errors |

### 7.3 Logging

- stdout: human-readable user output (suppressible with `--quiet`).
- stderr: warnings, errors, and `--verbose` traces.
- `--json`: machine-readable output for CI integration (planned v0.3+).
- Use `log/slog` for structured logs (Go 1.21+).

---

## 8. Testing Strategy

### 8.1 Layers

| Layer | Scope | Tool |
|---|---|---|
| Unit | Pure logic in `internal/...` | `testing` + `testify/assert` |
| Golden | Whole-output of `aikata init`/`generate` per preset | `testing` + `testdata/golden/<preset>/` |
| Integration | Run compiled binary, assert filesystem state | `testing` + `os/exec` |
| Lint | Style, common bugs | `golangci-lint` |

### 8.2 Coverage target

- Core packages (`scaffold`, `doctor`, `generate`, `templates`,
  `presets`): **≥ 80 %**.
- `cmd/`, `internal/cli`: smoke-tested through integration tests.

### 8.3 CI

- GitHub Actions matrix: macOS, Linux, Windows × Go 1.21 (Phase 2 task).

---

## 9. Security Considerations

- **No network I/O** in core flows (`init`, `doctor`, `generate`).
  `aikata sync` may fetch templates over HTTPS only. `aikata update`
  may fetch release metadata and release assets only when the user
  explicitly asks to update or check for updates.
- Pin all `go.sum` entries; reject `replace` directives in releases.
- File generation refuses to write outside the working directory.
- Templates execute with `text/template` (markdown output), not
  `html/template` — no implicit script execution.
- Secrets: aikata never reads `.env`. `.env.example` is a placeholder only.

---

## 10. Dependencies

Minimal, listed with rationale. Additions require a CHANGELOG entry.

| Module | Purpose | Notes |
|---|---|---|
| `github.com/spf13/cobra` | CLI framework | De-facto Go CLI standard. |
| `github.com/charmbracelet/huh` | Interactive prompts | Used only when stdin is a TTY and `--no-interactive` is not set. |
| `gopkg.in/yaml.v3` | YAML parser | For `.aikata/aikata.yaml`. |
| `github.com/charmbracelet/lipgloss` | Terminal styling | Behind `--no-color` opt-out (`NO_COLOR` env supported). |
| `github.com/stretchr/testify` | Test assertions | Test-only (`require` / `assert`). |

**Forbidden categories**: heavy ORMs, GUI frameworks, generic LLM API
clients, dynamic template engines that allow script execution.

---

## 11. Performance Budget

| Operation | Budget |
|---|---|
| `aikata init --scope standard` end-to-end | < 500 ms |
| `aikata doctor` on a project with 50 markdown files | < 200 ms |
| Binary size (stripped) | < 20 MB |
| Memory peak during init | < 64 MB |

These are budgets, not guarantees. CI tracks regressions with simple
wall-clock measurements.

---

## 12. Phase 2 Bootstrap Checklist (for the next task)

When the Go-project-init task begins, the following artifacts must be
produced. This list is the executable form of the Phase 2 checklist.

- [ ] `go mod init github.com/<owner>/aikata` (replace `<owner>`).
- [ ] `cmd/aikata/main.go` with `--version` and `--help` only.
- [ ] cobra dependency added.
- [ ] `internal/` skeleton directories with `doc.go` placeholders.
- [ ] `Makefile` targets: `build`, `test`, `lint`, `install`.
- [ ] `.golangci.yml` with the standard linter set.
- [ ] `.github/workflows/ci.yml` running `make test` and `make lint`.
- [ ] `make build` produces a working binary.

No business logic is in scope for Phase 2. The MVP (`aikata init --scope
minimal`) is **Phase 3** — see [ROADMAP.md](./ROADMAP.md).

---

## 13. Doc Map (`docmap`)

The **doc map** is a derived artifact describing the *document set itself*
— inventory, cross-references, freshness, and a managed/external split. It
is **doc-cartography**, a responsibility distinct from project mission
(`README.md` / `SPEC.md` / `AGENTS.md`) and from the hand-curated
Navigation Matrix (`AGENTS.md` §3). The decision and its rationale are
[ADR 0044](./docs/adr/0044-doc-map-derived-artifact.md); the full design is
[`docs/decisions/docmap-design.md`](./docs/decisions/docmap-design.md).

### 13.1 Outputs

One scan of the document surface produces two mandatory renderings under
the aikata-owned machine zone:

- `.aikata/docmap.yaml` — the structured data layer (single source of the
  map's truth; `docs` sorted by `path` for diff-stable output).
- `.aikata/docmap.md` — the readable view: a directory tree, a Mermaid
  `doc → doc` link-graph (degrading to an adjacency list past a node
  threshold), and a `path → summary` index.

`txt` / `json` / `mmd` are optional, config-gated additional renderers
(`docmap.formats`). The map excludes its own outputs and the generated
AI-tool artifacts from the tracked set.

### 13.2 Data sources (documents only)

The map reads **no source code**. It is built from: frontmatter metadata
(`internal/docmeta`), the Markdown link graph (`internal/docmeta`, shared
with `doctor`), the filesystem tree, and the managed-surface set
(`internal/doctor/scope.go` `ManagedIncludeGlobs`, supplying the `managed`
flag). Per-document summaries degrade gracefully (`summary:` frontmatter →
leading `>` blockquote → first body paragraph → H1 → filename), so an
adopted repository needs no document refactor.

### 13.3 Triggers and isolation

- `aikata map` regenerates on demand.
- `init` / `fill` / `enable` / `sync` / `generate` each run the same
  rebuild as a **final, isolated step**. It is decoupled from per-tool
  `generate` provider failures: a failing provider must never leave the
  map stale, and the map step reports its own status (it does not abort the
  surrounding command on a non-fatal map issue).
- `aikata doctor` carries a **freshness check**: it rebuilds the map in
  memory and compares the `HashContent` of the result to the on-disk
  `docmap.yaml`. A mismatch is a `warning` (`--json` issue code under the
  versioned schema); `aikata doctor --fix` regenerates the map.

### 13.4 State class

`docmap.*` are aikata-owned derived state. They are **not** manifest-tracked
and **not** subject to `aikata sync`'s 3-way merge (§3.4); they are
regenerated, never merged. They are written atomically via
`internal/config`'s `writeAtomic`. Configuration lives under the `docmap`
key of `.aikata/aikata.yaml` (§4.1).
