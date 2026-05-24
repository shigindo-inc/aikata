---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
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
│   └── tasks/
│       └── current.md           # Agent's working memory
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

---

## 3. Generated Project Structure

This is the layout aikata produces when a user runs `aikata init` on a
brand-new project (not aikata itself). The structure is the **product**;
the repository layout in §2 is the **producer**.

### 3.1 Default (`aikata init --preset standard`)

```
/<project>
├── README.md              # Thin nav for humans + LLMs
├── AGENTS.md              # Canonical agent instructions
├── SPEC.md                # What / Why
├── ARCHITECTURE.md        # How
├── GLOSSARY.md            # Terminology
├── .env.example
├── .gitignore
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/            # Populated per preset
│   ├── tasks/
│   │   └── current.md     # Agent's working memory
│   ├── troubleshooting.md
│   └── prompts.md
└── .aikata/
    └── aikata.yaml
```

### 3.2 Optional files

| File / directory | Triggered by | First shipped | Purpose |
|---|---|---|---|
| `UI.md` | `--with-ui` or UI-style preset | v0.4.1 | UI / UX / product-design guidelines |
| `API.md` | `--with-api` or API-style preset | v0.4.1 | API interface spec |
| `docs/testing.md` | `--with-tdd` | v0.4.1 | Test strategy |
| `CHANGELOG.md` | `--with-changelog` | v0.4.1 | Release notes |
| `docs/memory/` (5 files) | `--with-memory` | v0.2 | Long-term agent memory (`user`, `feedback`, `project`, `reference` + `README`). See [ADR 0004](./docs/adr/0004-long-term-memory-slot.md). |
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

There is intentionally no generic `DESIGN.md` in built-in presets. Product
requirements live in `SPEC.md`, technical design lives in
`ARCHITECTURE.md`, decision rationale lives in ADRs, and UI / UX guidance
belongs in optional `UI.md` when enabled. See
[ADR 0007](./docs/adr/0007-no-generic-design-md.md).

---

## 4. Configuration File: `.aikata/aikata.yaml`

> **Path note**: v0.3.2 onward writes `.aikata/aikata.yaml` per
> [ADR 0008](./docs/adr/0008-aikata-owned-config-directory.md). The
> legacy `.aikata/aikata.yaml` path remains read-only for projects from
> v0.2 / v0.3.0 / v0.3.1 and is migrated automatically by
> `aikata doctor --fix`; the fallback stays in place throughout the
> v0.x line.

### 4.1 Schema (v1)

```yaml
version: 1                       # Required. Migration anchor.
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

features:                        # All default false.
  tdd: false
  obsidian_hints: false
  monorepo: false

docs:
  generate_gitignore: true       # Add generated artifacts to .gitignore.
                                 # Default: true for `aikata init` output;
                                 # aikata's own repo sets this to false
                                 # (see Do-No-Harm §6).
  task_file_location: docs/tasks/current.md

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
- `aikata` must accept any future minor extension of v1 without crashing,
  unknown keys logged as warnings.
- `version: 2` triggers a migration path (no such migration in v0.x).

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
- `.gitignore` does **not** include `.aikata/` (or the legacy `.ai/`).
- Reason: a contributor cloning aikata must be able to open Claude Code /
  Cursor immediately. See
  [ADR 0003 — Do-No-Harm Policy](./docs/adr/0003-do-no-harm-policy.md).

### 6.2 Default for `aikata init` output

- Generated AI-tool artifacts and `.aikata/` **are** added to the target
  project's `.gitignore`.
- The flag `--no-gitignore-generated` opts out.

### 6.3 Binary distribution channels (planned)

- GitHub Releases (multi-arch binaries via GoReleaser).
- Homebrew tap: `brew install aikata-dev/tap/aikata`.
- `curl -sSL https://aikata.dev/install.sh | sh`.
- npm wrapper: `npx aikata` (post-v0.4).

### 6.4 CLI self-update model (planned)

- `aikata update` is reserved for updating the aikata CLI binary,
  matching the user-facing convention established by Claude Code's
  `claude update` (ADR 0009).
- Native / installer-managed installs may update themselves by fetching
  a signed or checksummed GitHub Release asset and replacing the aikata
  binary atomically.
- Package-manager installs remain package-manager-owned. `aikata update`
  may print or, behind explicit opt-in, run `brew upgrade`, npm install,
  or an equivalent manager command, but it must not overwrite a
  package-manager-managed binary directly.
- Installers should write small install-source metadata under the user's
  data directory so the CLI can distinguish native, Homebrew, npm, Go,
  and unknown installs before choosing an update path.

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
| `aikata init --preset standard` end-to-end | < 500 ms |
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

No business logic is in scope for Phase 2. The MVP (`aikata init --preset
minimal`) is **Phase 3** — see [ROADMAP.md](./ROADMAP.md).
