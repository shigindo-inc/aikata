---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# Changelog

All notable changes to **aikata** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/);
see [AGENTS.md](./AGENTS.md) for the project-specific rules.

## [Unreleased]

### Fixed

- **`aikata sync --rebaseline` now honours `aikata.yaml`'s `stacks`
  and `features.monorepo` / `features.tdd`** when seeding the
  manifest. Previously the rebaseline path hard-coded
  `preset="standard"` with all opt-ins false, silently dropping
  per-project stack and monorepo preferences from the manifest. The
  new scope-derivation hierarchy is documented in ADR 0013.

### Changed

- **`aikata sync` reads `aikata.yaml` even when the manifest is
  present** and OR-merges its stacks / monorepo / tdd preferences on
  top of `inferFlags`. This rescues post-init opt-ins the manifest
  may not have grown yet (especially for v0.4.x projects upgraded to
  v0.6.x). ADR 0013 spells out the precedence (CLI > manifest >
  aikata.yaml > defaults).
- **`aikata add stack <name>` now records the stack guide in the
  manifest even when `docs/stacks/<name>.md` already exists on
  disk** (e.g. a v0.4.x project with a hand-rolled stack guide). The
  manifest hash is the template's, not the on-disk content's — which
  makes any user customisation register as `user-only-edit` on the
  next sync rather than `upstream-added` / conflict. Restores
  consistency with the singleFile / memory components (ADR 0014).
- **`aikata sync` gained four one-off scope-override flags**:
  `--preset <name>` / `--lang <en|ja>` / `--stack <name>`
  (repeatable) / `--with-monorepo` (use `--with-monorepo=false` to
  force-disable). Overrides apply to a single invocation only and
  are never written back to the manifest or `aikata.yaml`; the
  prescribed way to make a change persistent stays
  `aikata add <component>` or hand-editing `aikata.yaml`. See
  ADR 0013.

## [0.6.2] - 2026-05-28

Two coordinated quality-of-life improvements for the v0.6 line: a
high-value `ROADMAP.md` preset template so scaffolded projects have a
first-class place to describe their direction, and a fix that makes
`aikata add <component>` write entries to `.aikata/manifest.yaml` so
subsequent `aikata sync` runs correctly preserve customisations
instead of treating add'd files as `upstream-added`. Scope-derivation
work (ADR 0013) and the related override flags shipped under
`[Unreleased]` and will tag as v0.6.3 next.

### Added

- **`ROADMAP.md` template** added to the `standard` / `flutter` /
  `typescript` presets (ja + en). Scaffolded projects now ship with a
  ROADMAP placeholder so agents have a first-class entry point for
  "where is this project going?" alongside SPEC / ARCHITECTURE /
  GLOSSARY. `minimal` preset is deliberately untouched. Existing
  projects pick up the template via `aikata sync` once the next normal
  3-way merge runs against the new manifest entry.

### Fixed

- **`aikata add <component>` now records the added files in
  `.aikata/manifest.yaml`.** Previously `add memory` / `add ui` / `add
  api` / `add tdd` / `add changelog` / `add stack <name>` wrote files
  to disk but left the manifest stale, causing the next `aikata sync`
  to treat those files as `upstream-added` and potentially overwrite
  any user customisation. With the manifest updated alongside the
  write, subsequent syncs correctly classify customisations as
  `user-only-edit` and preserve them. New ADR 0014 documents this as
  the "manifest is a living record" invariant.

## [0.6.1] - 2026-05-26

Unscheduled patch release. Headline fix is `aikata sync --rebaseline`:
v0.6.0 walked into a 2-way diff path with an empty ancestor table and
silently wrote `<<<<<<<` / `>>>>>>>` conflict markers into every
customised source file. v0.6.1 makes `--rebaseline` non-destructive
(seeds the manifest only; never touches source files) and fixes a
second `install.sh` footgun where an older `aikata` earlier on `$PATH`
would silently shadow the freshly installed binary.

The v0.6.1 ROADMAP slot was originally channel publication (Homebrew
tap, npm, marketplace, native `update --apply`); that work shifts to
v0.6.4 (v0.6.2 became "ROADMAP & manifest hygiene", v0.6.3 became
"sync scope derivation").

### Fixed

- **`aikata sync --rebaseline` no longer writes conflict markers into
  user-customised files.** In v0.6.0, running `--rebaseline` on a
  pre-v0.5 project performed a 2-way diff between current on-disk
  content and the upstream rendering, treating every customised file
  as a conflict and wrapping it in `<<<<<<<` / `>>>>>>>` markers.
  Rebaseline is now non-destructive: it writes the manifest from the
  current upstream rendering (matching what `aikata init` would write
  for a fresh project) and exits without touching any source file.
  The user's customisations then appear as `user-only-edit` on the
  next `aikata sync` and are preserved. See ADR 0011 (Rebaseline
  ancestor choice) for the design rationale. (#83)
- **`aikata sync` error message clarifies that `--rebaseline` is
  non-destructive**, removing the v0.6.0 footgun where users feared
  the recommended remediation would overwrite their files. (#83)
- **`scripts/install.sh` warns when another `aikata` shadows the
  freshly installed binary.** Previously a stale `~/go/bin/aikata`
  earlier on `$PATH` would silently win over the newly installed
  `~/.local/bin/aikata`, so `aikata --version` kept reporting the old
  version after a successful install. The script now compares
  `command -v aikata` against `${INSTALL_DIR}/${BINARY}` and prints a
  remediation hint when they differ. (#82)

### Changed

- `dist/README.md` clarifies the supported scope of the `dist/`
  directory for Codex skills and third-party skill marketplaces; no
  behaviour change. (#81)

## [0.6.0] - 2026-05-26

Packaging-and-distribution milestone (partial). Ships the
agent-doable subset of the v0.6 ROADMAP entry: monorepo
scaffolding, install-source metadata foundation, and a Claude Code
plugin scaffold. Homebrew tap, npm wrapper, and plugin-marketplace
listing are tracked under v0.6.4 because they require maintainer
actions outside this repo (originally scheduled for v0.6.1, then
v0.6.2; the v0.6.1 slot was reused for the `aikata sync --rebaseline`
regression fix, v0.6.2 for ROADMAP & manifest hygiene, and v0.6.3
for sync scope derivation — see the corresponding sections).

### Added

- `aikata init --monorepo` — scaffold a project as a monorepo with
  nested per-app `AGENTS.md` files (`apps/_example/AGENTS.md`),
  an `apps/README.md` quick-start, and `docs/monorepo.md` explainer.
  Flips `features.monorepo: true` in `.aikata/aikata.yaml`.
  Interactive prompt grew a matching question.
- `internal/install` — Source enum + `Detect(exePath)`
  reading build-time ldflag, sibling `aikata.install-source` file,
  `runtime/debug.ReadBuildInfo`, then falling back to `unknown`.
  aikata never writes the file; install channels own it.
- `scripts/install.sh` — writes
  `${INSTALL_DIR}/aikata.install-source` after placing the binary.
- `dist/claude-code/plugin/` — plugin scaffold bundling the
  v0.3.1 skill (byte-identical copy) with four slash commands:
  `/aikata-init`, `/aikata-generate`, `/aikata-doctor`,
  `/aikata-sync`. `plugin.json` enforces `requires.minVersion`
  v0.6.0.
- `docs/adr/0012-memory-projection-deferral-extended.md` — defers
  memory projection past v0.6 with a concrete revisit trigger.
  Supersedes ADR 0010's "ship in v0.6 if stable" guidance.

### Changed

- ROADMAP.md splits the v0.6 entry into v0.6.0 (this release) and
  v0.6.4 (user-blocker channels: Homebrew tap, npm, marketplace,
  `--apply` self-update). Distribution-surface table gains v0.6.0
  and v0.6.4 rows. (Originally v0.6.1; renumbered through v0.6.2
  and v0.6.3 as those slots were used for the rebaseline fix,
  ROADMAP & manifest hygiene, and scope derivation respectively.)
- README ADR index updated through ADR 0012.

### Notes

- `internal/install.Detect()` ships as the **recording** half only.
  Native `aikata update --apply` self-update lands in v0.6.4
  alongside Homebrew / npm channels so the upgrade flow can be
  tested against real installs.
- `--monorepo` is orthogonal to the v0.4 single-file components.
  A `--monorepo --with-memory` project gets both layers without
  special-case coupling.

## [0.5.0] - 2026-05-26

Headline feature: `aikata sync`. A long-lived aikata project can now
pull newer template content without losing user edits. The merge is
documented in ADR 0011 and lives in `internal/sync/`.

### Added

- `aikata sync` — 3-way diff-merge against `.aikata/manifest.yaml`.
  User edits are preserved; upstream-only changes auto-apply; true
  conflicts get git-merge-style file-level markers
  (`<<<<<<<`, `|||||||`, `=======`, `>>>>>>>`). Exit code 2 when
  conflicts were written so CI loops can distinguish "merge needed"
  from "merge failed".
- `aikata sync --rebaseline` — seeds a manifest from current on-disk
  state for projects that pre-date v0.5. One-shot operation; the
  next normal sync uses the just-seeded ancestor.
- `aikata sync --dry-run` — preview the merge plan without writing
  anything to disk. Exits 0 even when conflicts are reported so CI
  can sample drift safely.
- `aikata sync --json` — versioned envelope shared with `doctor` /
  `list` / `describe` / `update`:
  `{version: 1, kind: "sync", files: [...], summary: {...}}`.
- `aikata doctor --strict` — treats warning-level findings as
  exit-3 failures. Default behaviour (errors only) is unchanged; the
  flag is opt-in and is what the new dogfood CI gate uses.
- `internal/config/manifest.go` — `Manifest` / `ManifestFile` types,
  `BuildManifest` (deterministic, path-sorted), Marshal / Unmarshal,
  Save / Load (atomic write), `HashContent` (SHA-256, hex). Written
  by `aikata init` at scaffold time and consumed by `aikata sync`.
- `internal/config/schema_migrate.go` — `AikataYamlMigrator` registry
  and `MigrateAikataYaml` / `LoadMigrated` entry points. v0.5 ships
  with zero registered migrations (v1 is the only schema in the
  wild); v2+ lands as one map row plus the migrator function.
- `internal/sync/` — new package implementing the merge core
  (`Run`, `classifyAndMerge`, `mergeThreeWay`, `conflictMarkers`,
  `derivePlan` / `inferFlags`). The cobra layer is the only consumer.
- `internal/scaffold.Render(opts)` — render-only sibling of `Run`
  used by `aikata sync` to re-render upstream templates without
  touching the filesystem. `Run` continues to do the full
  scaffold-then-write flow via a shared `renderInto` core.
- `docs/adr/0011-aikata-sync-design.md` — accepted ADR locking in
  the four design decisions: templates-only scope (no generated
  artifacts), `.aikata.conflict`-style markers, schema migration
  built into `sync`, and `.aikata/manifest.yaml` as the 3-way
  ancestor.

### Changed

- `aikata init` now writes `.aikata/manifest.yaml` alongside
  `.aikata/aikata.yaml`. The manifest captures the SHA-256 of every
  template-derived file at render time so `sync` can do a real 3-way
  merge instead of a noisy 2-way diff.
- CI workflow runs `aikata doctor --strict` on every OS leg and now
  asserts `aikata generate` is byte-identical to the committed
  `CLAUDE.md` / `.cursor/rules/main.mdc` (Unix legs; Windows skipped
  for line-ending parity with the existing `go mod tidy` exception).
  This makes the v0.5 dogfooding milestone binding rather than
  advisory.

### Notes

- Files written by `aikata add <component>` post-init are not yet
  appended to the manifest. They are visible to `aikata sync` only
  through a 2-way diff against the latest template. Tracking issue
  / v0.5.x follow-up.
- Conflict markers are emitted at file granularity (the entire body
  is wrapped). Line-level diff3 is deferred to a v0.5.x follow-up
  pending real-world feedback.
- Dogfood gate currently runs after build + tests. If a contributor's
  edit forgets to rerun `aikata generate`, the byte-identical check
  blocks the merge with an actionable diff.

## [0.4.2] - 2026-05-25

Lightweight follow-up that lands the read-only half of the ADR 0009
`aikata update` surface. Self-update (binary overwrite) remains
scheduled for v0.6 alongside installer-source metadata.

### Added

- `aikata update --check` — opt-in release check against
  `api.github.com/repos/shigindo-inc/aikata/releases/latest`.
  Reports the running version, the latest published release, and a
  generic upgrade guidance block covering `go install`, the install
  script, and manual GitHub Release download. Exits 0 on every
  outcome (up-to-date / update-available / dev-build /
  ahead-of-latest); only network or parse failures exit 1.
- `aikata update --check --json` — versioned envelope matching the
  shape used by `doctor` / `list` / `describe`:
  `{version: 1, kind: "update-check", current, latest, status,
  release_url}`.
- `aikata update` (no flag) — prints a notice that self-update is
  planned for v0.6 and directs users at `--check`. Exits 0.
- New `internal/release` package — minimal in-house semver parser
  and GitHub Releases client with an injectable endpoint for
  tests. v0.6 self-update will reuse both.

### Changed

- ROADMAP: v0.4.1 follow-up list cleaned up. `--check` moved to the
  shipped v0.4.2 section; installer-source metadata and
  installer-managed self-update relocated to v0.6 packaging cycle so
  they ship alongside the Homebrew / npm channels they depend on.

### Notes

- Pre-release / pseudo-version inputs (e.g. local `go build` from a
  dirty checkout) are classified as dev-build for now. aikata does
  not ship semver pre-releases yet; the comparator learns the
  distinction when one does.
- The GitHub API is hit unauthenticated. The 60 req/h rate limit is
  ample for human-invoked use; avoid calling `aikata update --check`
  from CI loops.

## [0.4.1] - 2026-05-25

Second wave of v0.4 — closes the remaining five `aikata add`
components and brings `aikata init`'s optional-feature flag set to
flag / prompt parity. No breaking changes; defaults preserve the
v0.4.0 init tree byte-identically.

### Added

- `aikata add ai-tool <name>` — post-init counterpart of `aikata
  init --ai-tools`. Validates against `internal/generate.KnownTools()`,
  appends the tool to `cfg.AITools` (sorted), and persists via
  `config.Save`. Idempotent on re-run; the per-tool artifact is
  materialized by the next `aikata generate`.
- `aikata add ui` / `add api` / `add tdd` / `add changelog` — four
  single-file optional components emitting `UI.md` / `API.md` /
  `docs/testing.md` / `CHANGELOG.md` respectively. Each refuses to
  clobber an existing file; re-run is a no-op + notice on Stderr.
  Implemented through a shared `singleFile` helper so a future
  optional component is a one-line registry entry.
- `aikata init --with-ui` / `--with-api` / `--with-tdd` /
  `--with-changelog` flags. Each gates the corresponding component
  into the init-time tree. The matching interactive prompt asks one
  y/N question per flag (default N); explicit flags silently skip
  the question.

### Changed

- `scaffold.Run` dispatch for optional components is now a small
  table (`optionalSpecs`) covering memory plus the four new
  single-file components. Adding the next component is one row plus
  its existing `components.Render<Name>` shim — scaffold itself
  stays component-agnostic.
- `internal/cli/prompt.go` consolidates the y/N questions behind a
  shared `askYesNo` helper. The order matches `scaffold.Run`'s
  dispatch table.
- SPEC.md §4.1 and ARCHITECTURE.md §3.2 "First shipped: v0.4.1"
  markers replaced with the as-shipped wording.
- Repository dogfood: aikata's own configuration moved from
  `.ai/aikata.yaml` to `.aikata/aikata.yaml`. The legacy resolver
  fallback remains for downstream projects (retirement scheduled for
  v1.0 per ADR 0008).

### Verified

- Every v0.4.0 golden tree (`standard` / `minimal` /
  `flutter` / `typescript` / `*-with-memory*` / `*-ja*`) remains
  byte-identical. New `standard-with-extras/` fixture covers the
  four `--with-*` flags together.
- `aikata doctor` on the aikata repository reports zero errors and
  zero warnings post-merge (`config.legacy-path` warning falls
  silent).

## [0.4.0] - 2026-05-24

First wave of v0.4 — authoring ergonomics for existing aikata
projects. Users can now add ADRs, stack guides, and the long-term
memory slot post-init without re-running `aikata init`.

### Added

- `aikata add <component>` — new cobra parent that dispatches to a
  registry-driven set of components. The parent is open/closed: new
  components register themselves and the leaf subcommand appears
  automatically. Currently active components:
  - `aikata add adr "<title>"` — auto-numbered ADR skeleton under
    `docs/adr/NNNN-<slug>.md`. Title is slug-cased via the shared
    `templates.Kebab` helper; numbering uses the v0.3 `internal/adr`
    helper. Duplicate slugs (at any number) are refused.
  - `aikata add stack <name>` — adds `docs/stacks/<name>.md` and
    appends the stack to `.aikata/aikata.yaml` `stacks:`. Bundled
    stacks (`flutter`, `typescript`) reuse their preset template;
    re-running on an already-registered stack is a no-op + notice and
    user-edited files are never clobbered.
  - `aikata add memory` — opt-in equivalent of `aikata init
    --with-memory` for projects that did not enable it at init time.
    Idempotent: existing files are preserved.
- `aikata list components` — enumerates the components registry with
  the same versioned `--json` envelope as `list presets|stacks|ai-tools`.
  Reserved entries show a `(reserved)` suffix in text output.
- `aikata add --dry-run` — shared flag on every leaf subcommand;
  prints the plan to stdout without writing.
- `internal/components` — the SSOT for the components registry,
  consumed by `aikata add`, `aikata list components`, and the
  scaffold integration for `aikata init --with-memory`. The Memory
  renderer previously hard-coded in `scaffold.addMemoryArtifacts`
  now lives here so init-time and add-time emit byte-identical
  output.
- `internal/config.Save(root, cfg)` and `internal/config.Load(root)`
  — atomic write / load pair used by the new add subcommands when
  mutating `.aikata/aikata.yaml`.
- `internal/templates.LangDir(base, lang)` and `templates.Kebab(s)`
  — lifted out of scaffold so the components package can share one
  implementation.
- `docs/adr/0010-memory-projection-deferred-to-v0-6.md` — records
  the v0.4 investigation of ADR-0004 option δ (mirroring
  `docs/memory/` into Claude / Cursor memory channels). Decision:
  ship the authoring surface in v0.4 (this release), defer the
  projection itself to v0.6 where the per-tool plugin spec will
  own it.

### Changed

- `internal/scaffold.Stacks()` is removed. The bundled stack list is
  now sourced from `components.Stacks()` so listing and adding share
  one slice (SSOT).
- `scaffold.addMemoryArtifacts` is gone; `scaffold.Run` calls
  `components.RenderMemory` instead. Existing init golden trees
  remain byte-identical (verified by the unchanged scaffold tests).

## [0.3.2] - 2026-05-24

### Changed

- New projects scaffolded by `aikata init` now store their config at
  `.aikata/aikata.yaml` instead of `.ai/aikata.yaml`, completing the
  ADR 0008 migration scheduled in v0.3.0. Existing projects continue
  to load from `.ai/aikata.yaml`; readers prefer the new path when
  both are present.

### Added

- `internal/config` gains a path resolver
  (`Resolve(root) (path, isLegacy, error)`) plus atomic move helper
  `MoveLegacyToPrimary`, used by `aikata generate` and `aikata doctor`.
- `aikata generate` auto-migrates `.ai/aikata.yaml` to
  `.aikata/aikata.yaml` on first run when only the legacy path
  exists, printing a single `notice:` line to stderr. Failures
  degrade to a warning and the run continues against the legacy
  file so user work is never blocked.
- `aikata doctor` reports legacy-path projects as a warning with
  `Code: "config.legacy-path"`. `aikata doctor --fix` performs the
  same atomic move; the warning falls silent once the primary path
  exists.
- Template README / ARCHITECTURE / .gitignore strings for the
  standard, flutter, and typescript presets reflect the new path
  (English + Japanese template sets).

### Removed

- The hard-coded `.ai/aikata.yaml` string in `aikata generate`'s
  error path is replaced by `config.PrimaryPath(target)`; the legacy
  path keeps working through fallback only.

## [0.3.1] - 2026-05-24

### Added

- `aikata completion bash|zsh|fish|powershell` emits a cobra-generated
  completion script. The Long help embeds activation snippets for each
  shell so installation is one copy-paste.
- `aikata list presets | stacks | ai-tools` enumerates the identifiers
  init accepts; reserved `extended` is surfaced with a `(reserved)`
  suffix so its v1.0 status is discoverable today.
- `aikata describe preset <name>` returns the long-form description,
  status, and supported languages. Both `list` and `describe` accept
  `--json` with the same versioned envelope as
  [`aikata doctor --json`](./SPEC.md#43-aikata-doctor).
- `dist/claude-code/skill/SKILL.md` — a minimal Claude Code skill that
  teaches Claude when to call `aikata init / generate / doctor` and how
  to parse `aikata doctor --json`. Ships as the `aikata-skill.md`
  release asset for one-line install with `curl`.
- `CONTRIBUTING.md` adds a short human-friendly entry for external
  contributors with the quick-start build, PR checklist, and ADR
  workflow. `AGENTS.md` stays the canonical operational source.

### Changed

- Roadmap / spec command semantics updated: `aikata update` is reserved
  for updating the aikata CLI itself, while the future template
  diff-merge command is renamed to `aikata sync`
  ([ADR 0009](./docs/adr/0009-update-command-owns-cli-version-updates.md)).
- README "Install" gains Shell completion and Claude Code skill
  subsections; "Contributing" now points at the new CONTRIBUTING.md.

## [0.3.0] - 2026-05-24

### Added

- ADR 0007 records the document-taxonomy decision that built-in presets
  must not generate a generic `DESIGN.md`; product requirements stay in
  `SPEC.md`, technical design in `ARCHITECTURE.md`, decision rationale in
  ADRs, stack conventions in `docs/stacks/`, and UI / UX guidance in
  optional `UI.md`.
- ADR 0008 schedules a v0.3.x migration of aikata-owned configuration
  from `.ai/aikata.yaml` to `.aikata/aikata.yaml`. New `init` output
  will write to the new path; readers fall back to the legacy `.ai/`
  path with a deprecation warning through v0.x. `ARCHITECTURE.md` and
  `GLOSSARY.md` are updated to describe both paths and `ROADMAP.md`
  v0.3 lists the migration work item.
- New `internal/adr` package centralizes the `NNNN-slug.md` ADR
  filename convention. `Scan`, `Next`, and `Filename` will be reused
  by the upcoming `aikata add adr` command (v0.4) and `aikata doctor`.
- `aikata doctor` gains an `adr-numbering` check that reports
  duplicate ADR numbers and gaps in the `0001..max` range at
  `LevelInfo`. Findings stay advisory because the project may
  legitimately retire a number.
- `aikata init` exposes `--ai-tools` (comma-separated `claude | cursor
  | codex`, default `claude`). The value flows into the generated
  `.ai/aikata.yaml`. Unknown tools are rejected with a clear error.
- `aikata init` interactive prompt is brought to parity with the
  non-interactive flag surface for v0.3. It now asks about document
  language and AI tools in addition to project name, preset, and
  long-term memory. Questions whose flag was explicitly set on the
  command line are silently skipped. The optional-feature questions
  (UI / API / TDD / changelog) and OSS-intent question are recorded
  as Q-PROMPT-01 in `docs/decisions/open-questions.md` and land
  in v0.4 alongside their matching `--with-*` flags.
- `aikata doctor --json` emits a versioned machine-readable report on
  stdout. Schema version `1` with `issues[]` and `summary{errors,
  warnings, info}`; `line` and `code` are omitted when empty. The
  schema is documented in SPEC.md §4.3. Combine with `--fix` to get
  the post-fix report on a clean stream.
- `aikata doctor --fix` repairs the trivially-fixable subset of
  issues: missing-frontmatter blocks are scaffolded with placeholder
  values, missing required keys (`project`, `status`, `version`,
  `updated`, `audience`) are appended into the existing block, and
  stale `updated:` values are bumped to today. Combine with
  `--dry-run` to preview the count without writing files. Other
  findings (broken links, deprecated-ADR references, env-example
  drift) remain manual.

### Changed

- `ARCHITECTURE.md` and `ROADMAP.md` now describe `UI.md` as the future
  home for UI / UX / product-design guidance and explicitly steer future
  work away from a catch-all `DESIGN.md`.
- `aikata --version` output is now normalized to the v-prefixed tag
  form (`vX.Y.Z`) across install channels. GoReleaser binaries
  previously reported the bare semver string while `go install`
  reported the v-prefixed module version; both paths now agree.
  `cmd/aikata/main.go` guards the format defensively even when ldflags
  pass an unprefixed string.

## [0.2.1] - 2026-05-24

### Added

- v0.2.1 — onboarding patch: `scripts/install.sh` is a POSIX shell
  installer that detects the host OS / architecture, downloads the
  matching release archive, verifies its SHA-256 against
  `checksums.txt`, and drops the binary into `$HOME/.local/bin`
  (override with `AIKATA_INSTALL_DIR`). Pinning a tag via
  `AIKATA_VERSION=vX.Y.Z` skips the unauthenticated GitHub API call
  that resolves the latest release. Supported targets:
  linux/{amd64,arm64} and darwin/{amd64,arm64}; Windows users continue
  to use the manual download path.
- v0.2.1 — `README.md` Install section now documents the one-line
  convenience path
  (`curl -fsSL .../scripts/install.sh | sh`) alongside the manual
  download + checksum-verify path. `docs/japanese-users.ja.md` mirrors
  the addition.
- v0.2.1 — `.github/workflows/ci.yml` adds an `install-script` job
  that runs `scripts/install.sh` on `ubuntu-latest` and `macos-latest`
  and asserts `aikata --version` succeeds. The job pins
  `AIKATA_VERSION` to the previous release so PRs can verify the
  installer end-to-end without depending on an unreleased tag.
- ADR 0006 records the locale policy: aikata's own repository keeps
  English canonical documentation, `--lang ja` generated templates
  remain first-class, and Japanese repository documentation is a focused
  access layer rather than a full mirror.
- `docs/japanese-users.ja.md` provides a Japanese entry point for
  installation context, `--lang ja` usage, and support-language
  expectations.

### Changed

- `ROADMAP.md` rewritten to reflect shipped milestones (Phase 1, v0.1,
  v0.2 marked released with their actual deliverables) and resequenced
  to split the previously-overloaded v0.3 into focused phases:
  - **v0.2.1** — onboarding patch carrying only the `curl -fsSL ... | bash`
    install script + README "Install" rewrite. External-only, no CLI
    behaviour change, so it can ship ahead of the next feature cycle.
  - **v0.3** — lightweight quality-of-life follow-up: `doctor --fix` /
    `--json`, ADR auto-numbering helper, `aikata completion`,
    `aikata list` / `describe`, the minimal Claude Code `SKILL.md`,
    `CONTRIBUTING.md` for this repo.
  - **v0.4** — authoring ergonomics: `aikata add` split into a
    first-wave (`adr`, `stack`, `memory`) and second-wave (`ai-tool`,
    `ui`, `api`, `tdd`, `changelog`); the matching `--with-*` init
    flags; the memory generate-projection investigation.
  - **v0.5** — `aikata sync` interactive diff-merge gets its own
    release cycle (the single largest feature on the roadmap).
  - **v0.6** — packaging & distribution: `--monorepo`, npm wrapper,
    Homebrew tap, Claude Code plugin (slash commands), conditional
    memory-projection ship.
  - **v1.0** — `--oss` flag pack expanded to include
    `CODE_OF_CONDUCT.md` and GitHub issue / PR templates; plugin /
    skill distribution explicitly extended to Cursor / Gemini / VS
    Code.
- `ROADMAP.md` "Dogfooding milestone" gate moved from v0.3 to v0.5 and
  given three concrete pass criteria (doctor clean, `aikata init` build
  in CI, generated AI-tool artifacts byte-identical) so it stops being
  aspirational.
- `ROADMAP.md` "Out-of-scope" gains an explicit deflection for "aikata
  as a Claude Code *agent*" (distinct from skill or plugin) to prevent
  future scope creep into runtime-personality territory.
- `ROADMAP.md` "Distribution surface" matrix updated to track install
  channels across v0.1 → v1.0 including the new v0.2.1 / v0.5 / v0.6
  rows.
- `README.md` install section restructured: a new top-level "Install"
  block leads with the pre-built binary path (with per-platform asset
  table and explicit checksum step) and labels it "Recommended — no
  Go toolchain". `go install` is demoted to "From source" with the Go
  1.21+ requirement called out explicitly. This addresses the
  long-standing ambiguity that suggested Go was required to use aikata.
- `docs/japanese-users.ja.md` mirrors the new install guidance with a
  short Japanese summary pointing to the English table for per-asset
  detail.

## [0.2.0] - 2026-05-23

### Fixed

- `aikata --version` now falls back to Go build info when the release
  ldflags value is absent, so `go install github.com/shigindo-inc/aikata/cmd/aikata@latest`
  can report the resolved module version instead of `0.0.1-dev`.

### Added

- v0.2 Task 13 — `aikata doctor` initial release (read-only):
  - New `aikata doctor` subcommand runs eight consistency checks
    against the project at the current working directory: required
    frontmatter keys, AGENTS.md link existence, ADR Deprecated
    status (must reference `Replaced by` / `Superseded by`),
    memory `memory_type` ↔ filename match, frontmatter `updated:`
    staleness (> 365 days warning), `.env.example` variables
    cross-referenced from AGENTS.md / ARCHITECTURE.md, GLOSSARY
    terms referenced from at least one other markdown file. Exit
    code 0 (clean) / 3 (errors found); warnings and infos are
    advisory and never set a non-zero exit.
  - `internal/doctor/` exposes `Run(Options) ([]Issue, error)`,
    `Format`, and `HasErrors` so future commands (`add`, `update`)
    can reuse the checks programmatically.
  - The `.github/workflows/ci.yml` matrix runs `aikata doctor`
    against this repository on every PR — aikata is the first
    project to dogfood the check.
  - `--fix` and `--json` are deferred to v0.3 per the v0.2 Plan
    (option E5).
- v0.2 Task 12 — `--lang ja` parallel-directory templates:
  - All preset / memory / ai_tools templates moved from
    `<base>/<file>` to `<base>/en/<file>` and the matching Japanese
    set added under `<base>/ja/<file>`. The parallel directory
    structure was chosen over `{{if eq .Lang "ja"}}` branching so
    translations can evolve independently without bloating template
    files (ADR-equivalent decision documented as E3 in the Task 12
    Plan section).
  - `scaffold.resolveLangDir` selects `<base>/<lang>/`, falling back
    to `<base>/en/` with a one-line stdout notice when the requested
    language has no translations. The fallback is verified by
    `TestLangFallback_UnknownLangFallsBackToEn`.
  - `generate.resolveLangTemplate` does the same for AI-tool
    artifacts (`ai_tools/<tool>/<lang>/<file>`); the
    `generate.Context.Lang()` helper reads
    `cfg.Project.Lang` (default "en") so providers do not need to
    re-derive the language. Routing covered by
    `TestRun_LangRoutesToJaTemplate` and
    `TestRun_LangFallsBackToEnForUnknownLang`.
  - **47 new ja template files** (4 presets × ~12 files +
    5 memory + 2 ai_tools) authored as Japanese originals rather
    than mechanical translations. Frontmatter and template structure
    mirror the en counterparts exactly so existing tests stay
    invariant.
  - **5 new ja golden trees** (`minimal-ja/`, `standard-ja/`,
    `flutter-ja/`, `typescript-ja/`, `standard-ja-with-memory/`).
  - **`lang_consistency_test.go`** asserts en/ja file-name set
    equality for every preset, the memory directory, and each
    ai_tool — the cheap pre-doctor guard against translation drift.
  - Existing en golden trees are unchanged: the lang directory move
    only affects template paths, not output paths, so
    `TestGolden_Minimal` / `TestGolden_Standard` / `TestGolden_Flutter`
    / `TestGolden_Typescript` and their `*-with-memory` siblings
    remain byte-identical.
- v0.2 Task 11 — `--preset typescript`:
  - New `typescript` preset under
    `internal/templates/data/presets/typescript/` ships the same
    document set shape as flutter (README / AGENTS / SPEC / ARCHITECTURE
    / GLOSSARY / `.env.example` / `.gitignore` / ADR 0001 /
    docs/tasks/current / docs/troubleshooting / docs/prompts /
    `.ai/aikata.yaml`) plus a TypeScript-specific
    `docs/stacks/typescript.md` covering tsconfig strictness,
    module-format choice, package manager, lint discipline, type
    rules (no `any`, prefer `unknown` and narrow), test runner
    selection, error subclassing, and async hygiene.
  - `cli/init.go`'s `stacksForPreset` returns `["typescript"]` for the
    new preset; `scaffold.addPresetArtifacts` emits the same
    struct-driven `.ai/aikata.yaml` as standard / flutter with
    `stacks: [typescript]`.
  - `.gitignore` covers Node / TypeScript artifacts
    (`node_modules/`, `dist/`, `build/`, `*.tsbuildinfo`, `coverage/`,
    `.next/`, `.turbo/`, etc.).
  - Interactive prompt offers `typescript` as a fourth preset choice
    (`standard | minimal | flutter | typescript`).
  - 5 new unit tests (file presence, yaml stack content, AGENTS
    reference, gitignore Node coverage, OSS scrub) plus 2 golden trees
    (`testdata/golden/typescript/`,
    `testdata/golden/typescript-with-memory/`).
  - The Do-No-Harm regression test (formerly
    `TestRun_NonFlutter_NoFlutterFootprint`) is renamed
    `TestRun_NonStack_NoStackFootprint` and asserts that
    `minimal`/`standard` outputs contain neither `flutter` nor
    `typescript` nor `docs/stacks/`.
- v0.2 Task 10 — `--preset flutter`:
  - New `flutter` preset under
    `internal/templates/data/presets/flutter/` ships a full document
    set (README / AGENTS / SPEC / ARCHITECTURE / GLOSSARY / `.env.example`
    / `.gitignore` / ADR 0001 / docs/tasks/current / docs/troubleshooting
    / docs/prompts / `.ai/aikata.yaml` via `config.Default`) plus a
    Flutter-specific `docs/stacks/flutter.md` covering lints, null
    safety, state management, widget authoring rules, async / isolate
    budget, build_runner, testing, platform channels, and accessibility.
  - `scaffold.Options` gains a `Stacks []string` field; `templateData`
    exposes it as `{{.Stacks}}` so AGENTS.md templates can render
    `docs/stacks/<stack>.md` cross-references with
    `{{range .Stacks}}`. Stack-flavored presets (currently just
    flutter) seed this implicitly via cli/init's `stacksForPreset`
    helper. The minimal / standard presets stay stack-agnostic
    (Stacks=nil), preserving the Do-No-Harm guarantee.
  - `.gitignore` covers Flutter build outputs: `build/`, `.dart_tool/`,
    `ios/Pods/`, `pubspec.lock`, etc.
  - `AGENTS.md` Read-order list uses repeated `1.` so markdown
    renderers auto-number across optional sections (stacks, memory)
    without colliding explicit numbers.
  - Interactive prompt offers `flutter` as a third preset choice.
  - 6 new tests (file presence, yaml stack content, AGENTS reference,
    gitignore Flutter coverage, OSS scrub, Do-No-Harm regression on
    minimal/standard) plus 2 golden trees
    (`testdata/golden/flutter/`, `testdata/golden/flutter-with-memory/`).
- v0.2 Task 9 — Cursor pass-through + Codex no-op:
  - `CursorProvider` emits `.cursor/rules/main.mdc`, a thin wrapper
    with `alwaysApply: true` that defers to canonical `AGENTS.md` and
    mirrors the same adaptive Read-order rendering as `CLAUDE.md`.
  - `CodexProvider` is registered for symmetry but produces zero
    files; Codex reads `AGENTS.md` directly per the 2026 spec
    (`developers.openai.com/codex/guides/agents-md`). `aikata
    generate` now prints `[codex] no files generated (reads AGENTS.md
    directly)` to stderr so the no-op is visible.
  - `internal/generate/Run` now returns `(map[string]int, error)`
    reporting per-provider file counts so the cli layer can surface
    no-op providers. Existing callers updated; the `Provider`
    interface is unchanged.
  - `docs/adr/0005-cursor-codex-pass-through.md` formalizes the
    strategy; richer glob-scoped MDC generation is deferred and
    tracked in `docs/decisions/open-questions.md` as Q-DESIGN-08.
  - aikata's own `.ai/aikata.yaml` opts into `cursor` and `codex` so
    `.cursor/rules/main.mdc` is committed to the repository,
    consistent with ADR 0002 §"Operational status" (aikata's repo
    commits self-generated artifacts).
- v0.1 release plumbing (Task 8):
  - `docs/origin/` removed pre-public-flip. The folder held the
  pre-v0.1 planning notes (`initial-design.md`, `initial-setup.md`)
  whose substance was already split into SPEC / ARCHITECTURE /
  ROADMAP / GLOSSARY / ADR / open-questions. Keeping it around added
  duplicated noise and a redundant top of the navigation tree for new
  contributors. The notes themselves are preserved in git history
  (commit `ea48abf`). References from AGENTS / SPEC / ARCHITECTURE /
  ROADMAP / README / GLOSSARY / ADR 0002 / ADR 0003 / memory/project
  are either updated to point at the operational docs or removed.
- `.github/workflows/ci.yml` matrix expanded to macOS / Linux /
    Windows × Go 1.21 (lint and `go mod tidy` are Linux-only to keep
    the matrix focused).
  - `.github/workflows/release.yml` — tag-driven (`v*`) workflow that
    runs GoReleaser.
  - `.goreleaser.yml` — cross-platform builds for linux/darwin/windows
    on amd64/arm64 (no Windows ARM64), checksum file, GitHub-driven
    changelog filtering, `-trimpath` + `-X main.version`.
  - `.gitattributes` — forces LF line endings everywhere so the
    golden trees stay byte-identical across runners.
  - `CHANGELOG.md` — promoted to `[0.1.0]`, dated.
  - `README.md` — Quickstart "planned" notice removed; CI / Release
    badges added.

## [0.1.0] - 2026-05-22

This is the first tagged release of aikata. The `Added` section below
lists the cumulative scope of the v0.1 cycle (Tasks 0 → 8 of the
post-Phase-1 roadmap).

### Added

- Operational documents split from the pre-v0.1 planning notes
  (preserved in git history at commit `ea48abf`):
  - `README.md`, `AGENTS.md`, `SPEC.md`, `ARCHITECTURE.md`, `ROADMAP.md`,
    `GLOSSARY.md`, `LICENSE`, `.gitignore`
  - `docs/adr/0001-record-architecture-decisions.md`
  - `docs/adr/0002-agents-md-as-canonical.md`
  - `docs/adr/0003-do-no-harm-policy.md`
  - `docs/decisions/open-questions.md`
- `CLAUDE.md` — Phase 1 hand-written wrapper that defers to `AGENTS.md`.
  Will be regenerated by `aikata generate` in v0.1 (see ADR 0002).

### Changed

- `LICENSE` — copyright holder set to `Satoshi Minami` (was the
  placeholder `aikata contributors`).
- `docs/adr/0002-agents-md-as-canonical.md` — Phase 1 deviation section
  expanded to document the `CLAUDE.md` wrapper, its lifetime, and the
  explicit exception to the top-level-minimalism rule.
- `AGENTS.md` — references to `docs/tasks/current.md` rewritten as
  conditional ("once it exists; introduced in v0.1"). Until then, PR
  descriptions and commit messages are the working-memory of record.
- Long-term agent memory slot introduced under `docs/memory/`:
  - `docs/memory/{README,user,feedback,project,reference}.md` —
    aikata's own dogfooding seed; format spec lives in the README
    inside that directory.
  - `docs/adr/0004-long-term-memory-slot.md` — design rationale,
    type taxonomy, Do-No-Harm compliance, and the deferral of the
    generate-projection question (option δ) to v0.3+.
  - `SPEC.md` §3 — Design Principle #8 ("Rules ≠ memory ≠ working
    state") added; §4.1 — `--with-memory` flag listed (ships v0.2).
  - `ARCHITECTURE.md` §3.2 — `--with-memory` row added with v0.2 tag.
  - `ROADMAP.md` — v0.2 gains `--with-memory` implementation, v0.3
    gains the option-δ investigation, v0.4 gains the conditional ship.
  - `AGENTS.md` — Read order extended to include `docs/memory/`; new
    Navigation matrix row "Recall user/project context"; new §4a
    explaining the rules/memory/working-state distinction.
  - `GLOSSARY.md` — new entries `long-term memory`, `memory type`,
    `memory slot`.
  - `docs/decisions/open-questions.md` — Q-DESIGN-07 registered for
    option (δ) (memory generate-projection across AI-tool memory
    channels).
- Interactive `aikata init` mode (Task 6):
  - New file `internal/cli/prompt.go` — small bufio-based prompt that
    asks for project name (skipped if already supplied), preset
    (`standard`/`minimal`, `1`/`2` also accepted, blank keeps default),
    and `--with-memory` (`y`/`yes`/`n`/`no`, blank keeps default).
  - `internal/cli/init.go` runs the prompt when stdin is a TTY and
    `--no-interactive` is not passed. When stdin is not a terminal
    (CI, pipes, `go test`) the command auto-falls-back to non-
    interactive mode so it never hangs.
  - `isTTYFunc` is a package variable so tests can swap it out for a
    deterministic answer.
  - **No new dependencies.** The original plan called for
    `charmbracelet/huh` + `lipgloss`, but huh v1 requires Go 1.23+
    and its dependency tree spans bubbletea, bubbles, x/ansi, etc.
    Both would have violated aikata's "minimal external dependencies"
    rule (ARCHITECTURE.md §10) and bumped the Go floor. A bufio-based
    prompt fits aikata's "lightweight, opinionated but small" stance
    and keeps the go directive at 1.21.
  - Tests: 8 new unit tests in `prompt_test.go` (happy path, blank
    inputs keep defaults, name skipped when supplied, empty-name
    error, unknown preset error, numeric preset choices `1`/`2`, yes/
    no variants, unknown yes/no error) + 3 new `init_test.go` tests
    (TTY auto-fallback, interactive happy path, defaults accepted).
- `aikata generate` for Claude (Task 7 — ADR 0002 migration completed):
  - New package `internal/generate` with `Provider` interface,
    registry (`Get` / `KnownTools`), `Context` struct, all-or-nothing
    `Run`. The registry exposes `claude` only in v0.1; Cursor, Codex,
    Gemini, Copilot, Windsurf land in v0.2+.
  - `ClaudeProvider` reads canonical `AGENTS.md` plus existence
    checks on README/SPEC/ARCHITECTURE/GLOSSARY/docs/memory/
    open-questions, then emits a generated `CLAUDE.md` whose Read
    order adapts to whichever docs the project ships.
  - New CLI subcommand `aikata generate` reads `.ai/aikata.yaml`
    from cwd, iterates the enabled `ai_tools`, and writes each
    provider's artifacts. Missing config maps to exit 2; unknown
    tool maps to exit 2 with the list of known names.
  - New template `internal/templates/data/ai_tools/claude/CLAUDE.md.tmpl`
    backs the generator with the same dot-helper pattern used by
    `internal/templates`.
  - `cmd/aikata`'s root command now wires `generate` alongside `init`.
  - aikata's own repository ships `.ai/aikata.yaml` (overriding
    `docs.generate_gitignore` to `false` per ADR 0003) and migrates
    `CLAUDE.md` from the Task 1 hand-written wrapper to the
    generator output.
  - `docs/adr/0002-agents-md-as-canonical.md` updated to "Operational
    status (Task 7 shipped)", with the migration diff documented
    inline.
  - Tests: 8 unit tests in `internal/generate` (registry,
    unknown-tool error, empty-tools error, no-AGENTS-MD error,
    artifact production, adaptive read order, OSS scrub) + 4
    integration tests in `internal/cli` (`aikata generate` end-to-end
    after init, exit-code mapping, help mention).
- `aikata init --with-memory` opt-in (Task 5A — ADR 0004 (γ) ships):
  - 5 new memory templates under `internal/templates/data/memory/`
    (`README`, `user`, `feedback`, `project`, `reference` — each with
    matching `memory_type` frontmatter).
  - `scaffold.Options` gains `WithMemory bool`; the new
    `addMemoryArtifacts` walks `memory/` in the embedded FS and adds
    `docs/memory/*.md` to the rendered set.
  - `templateData` helper centralizes the template-fields contract.
  - `internal/cli/init.go` accepts `--with-memory` (boolean, default
    false). Composes with both `--preset minimal` and `--preset
    standard`.
  - Standard preset's `AGENTS.md.tmpl` gains `{{if .WithMemory}}`
    branches in Read order and Navigation matrix (`Recall user /
    project context` row). Minimal preset's `AGENTS.md.tmpl` gains a
    Read order branch for `docs/memory/`. Standard preset's
    `docs/tasks/current.md.tmpl` gains a conditional mention of
    long-term memory.
  - New tests: 4 scaffold tests (with-memory file set, memory_type
    matches filename, standard AGENTS references memory, Do-No-Harm
    regression for both presets), 2 init tests (with-memory generates
    dir, without-memory omits dir), 2 golden snapshots
    (`minimal-with-memory`, `standard-with-memory`).
  - `Makefile` gains `update-golden` target (`go test ... -update`).
  - `docs/decisions/open-questions.md` Q-DESIGN-07 updated: scope (γ)
    shipped; (δ) AI-tool memory-channel projection still deferred.
- `aikata init --preset standard` (Task 5):
  - New package `internal/config` with `AikataYaml` struct, `Default`,
    `Marshal`, `Unmarshal`, version constant; schema matches
    `ARCHITECTURE.md` §4.1.
  - 11 standard-preset templates under
    `internal/templates/data/presets/standard/` covering
    README/AGENTS/SPEC/ARCHITECTURE/GLOSSARY plus `.env.example`,
    `.gitignore`, `docs/adr/0001-...`, `docs/tasks/current.md`,
    `docs/troubleshooting.md`, `docs/prompts.md`.
  - `internal/scaffold/scaffold.go` gains `addPresetArtifacts` which
    injects `.ai/aikata.yaml` (struct-driven, not template-driven) for
    the standard preset.
  - `aikata init` default preset switched from `minimal` to `standard`
    (SPEC.md §4.1 default).
  - New dependency: `gopkg.in/yaml.v3 v3.0.1`.
  - Tests: 4 new standard-preset Run tests (file set, aikata.yaml
    content, gitignore content, OSS scrub) + 1 cli init test
    (`aikata.yaml` written under default preset) + 1 golden snapshot
    `testdata/golden/standard/` covering all 12 generated files.
- `aikata init --preset minimal` MVP (Task 4):
  - New package `internal/templates` — embed-backed template loader
    (`//go:embed all:data`), template helpers `now`, `joinSlash`,
    `kebab`, and `Render(path, data, clock)` with
    `missingkey=error` semantics.
  - New package `internal/scaffold` — `Options` + `Run` performing
    all-or-nothing template rendering before any filesystem write.
    `ErrTargetDirNotEmpty` sentinel for the no-`--force` path.
  - `internal/cli/init.go` — cobra `init` subcommand with flags
    `--preset`, `--name`, `--no-interactive`, `--force`, `--dry-run`,
    `--lang`. Positional name supported; `--name` wins when both.
  - `internal/cli/errors.go` — `ExitError{Code, Err}` for routing
    user-input failures (exit 2) through `cmd/aikata/main.go`.
  - `cmd/aikata/main.go` — exit-code translation via `errors.As`.
  - `internal/templates/data/presets/minimal/{README,AGENTS,SPEC}.md.tmpl`
    — 3 lightweight templates (~30 lines each) per D1.
  - `testdata/golden/minimal/` — golden snapshots for the minimal
    preset; `-update` flag on the package rewrites them from current
    output.
  - Tests: 10 unit tests in `scaffold`, 7 in `cli`, 6 in `templates`,
    plus a fresh golden-comparison test for the minimal preset.
- ARCHITECTURE.md §2 and §5.2 updated: template root is now
  `internal/templates/data/`, embedded with `//go:embed all:data`
  (D2 — keeps `..`-relative embed paths out of the build).
- OSS readiness scrub (Task 3A — pre-public-release hygiene):
  - `docs/memory/reference.md` — removed the maintainer's absolute
    local path (`/Users/...`) and replaced it with guidance to use
    `$REPO_ROOT` in shell snippets going forward. Old entry kept with
    a `(superseded ...)` marker per ADR 0004 conventions.
  - `docs/memory/project.md` — recorded the repository-visibility
    plan (private until v0.1, public at tag), the binding
    squash-merge policy, and the git-history scrub decision
    (HEAD-only, history left intact because the leaked string is a
    local path rather than a secret; revisit at Task 8).
  - GitHub repo settings — disabled merge-commit and rebase-merge,
    enabled `Automatically delete head branches`. Squash is now the
    only merge mode.
  - Cleaned up the merged `feat/phase-2-go-init` branch (local +
    remote).
- Go project skeleton (Phase 2):
  - `go.mod` (`module github.com/shigindo-inc/aikata`, `go 1.21`).
  - `cmd/aikata/main.go` — entry point that delegates to
    `internal/cli`.
  - `internal/cli/{root,doc,root_test}.go` — cobra root command with
    `--version` / `--help` and unit tests.
  - `internal/{scaffold,doctor,generate,config,presets,templates}/doc.go`
    placeholders documenting each package's responsibility.
  - `Makefile` (targets: `build`, `test`, `lint`, `install`, `run`,
    `clean`, `tidy`, `verify`).
  - `.golangci.yml` (golangci-lint v2 schema, default linters
    `errcheck` / `govet` / `ineffassign` / `revive` / `staticcheck` /
    `unused`; formatters `gofmt` / `goimports`).
  - `.github/workflows/ci.yml` — Linux + Go 1.21 CI running `go vet`,
    `go test -race`, `golangci-lint`, `go build`, and a `--version`
    smoke test.
  - Dependency added: `github.com/spf13/cobra` v1.10.2.

### Removed

- _(none)_

---

<!--
Release sections will follow this template:

## [0.1.0] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
