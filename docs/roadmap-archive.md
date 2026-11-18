---
project: aikata
status: draft
version: 0.1.0
updated: 2026-11-18
audience: [human, agent]
---

# ROADMAP archive — released milestones (Phase 1 – v0.15.2)

> Historical record. These milestones are **all released**; their detailed
> per-version checklists were moved out of [ROADMAP.md](../ROADMAP.md) to keep
> the live roadmap focused on the current and upcoming surface
> ([ADR 0039](./adr/0039-documentation-hygiene-and-context-budget.md)). The
> binding decisions live in [`docs/adr/`](./adr/) and the shipped changes in
> [CHANGELOG.md](../CHANGELOG.md); this file preserves the planning detail
> verbatim. For current direction, see [ROADMAP.md](../ROADMAP.md).

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
  opt-in ([ADR 0004](./adr/0004-long-term-memory-slot.md)).
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
  ([ADR 0005](./adr/0005-cursor-codex-pass-through.md)); Codex
  reads `AGENTS.md` natively.
- `aikata doctor` initial release — eight read-only consistency checks
  (frontmatter / links / ADR / memory / updated / env / glossary /
  lang-consistency); exit code 0 or 3.
- Locale policy formalized in
  [ADR 0006](./adr/0006-locale-and-japanese-documentation-policy.md):
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

## v0.3 — Fast follow-up (lightweight) ✅ (released 2026-05-24)

Closed out the v0.2 surface with low-risk core-CLI quality-of-life
improvements. No template or preset content changed.

Shipped in v0.3.0:

- `aikata doctor --fix` repairs the trivially-fixable subset: missing
  frontmatter blocks are scaffolded, missing required keys are
  appended into the existing block, and stale `updated:` values are
  bumped to today. `--dry-run` previews the count without writing.
- `aikata doctor --json` emits a versioned (v1) machine-readable
  report on stdout; the schema is documented in
  [SPEC.md §4.3](../SPEC.md#43-aikata-doctor).
- `aikata doctor adr-numbering` reports duplicate / gap ADR numbers
  at `LevelInfo`. The numbering helper lives in `internal/adr/` and
  will be reused by `aikata add adr` in v0.4.
- `aikata --version` is normalized to `vX.Y.Z` across `go install`
  and GoReleaser binaries. `cmd/aikata/main.go` also normalizes
  defensively when ldflags pass a bare semver string.
- `aikata init` interactive prompt is at flag parity for v0.3
  (project name, preset, language, AI tools, long-term memory).
  Questions whose flag was explicitly set on the CLI are silently
  skipped. The `--ai-tools` flag itself is new (default `claude`).
- `ROADMAP.md` joined the `standard` preset surface as the durable
  when/sequence layer; `extended` reserved as the future heavier
  preset name.
- ADR 0007 (no generic `DESIGN.md`) and ADR 0008 (aikata-owned
  `.aikata/` config namespace; migration scheduled for v0.3.x).

---

## v0.3.1 — Discoverability & distribution surface ✅ (released 2026-05-24)

Lightweight follow-up that closes out the discoverability and
distribution items deferred from v0.3.0. No on-disk layout change.

Shipped:

- `aikata completion bash | zsh | fish | powershell` — cobra-generated
  shell completion. README "Install" documents the activation snippets.
- `aikata list presets | stacks | ai-tools` enumerates accepted
  identifiers; `aikata describe preset <name>` returns the long-form
  description. Both accept `--json` with the same versioned envelope
  as `aikata doctor --json`. Reserved `extended` is surfaced now so
  its v1.0 status is discoverable.
- **Minimal Claude Code skill** — single `SKILL.md` under
  `dist/claude-code/skill/`, shipped as the `aikata-skill.md` release
  asset. No slash commands, sub-agents, or hooks yet; the v0.6 plugin
  will extend rather than replace it.
- `CONTRIBUTING.md` adds a human-friendly entry for external
  contributors; `AGENTS.md` remains the canonical operational source.
- ADR 0009 records the `aikata update` (CLI version) vs
  `aikata sync` (project templates) split that v0.4 / v0.5 will land.

---

## v0.3.2 — Config namespace migration ✅ (released 2026-05-24)

Completes the ADR 0008 migration of aikata's own configuration file
from `.ai/aikata.yaml` to `.aikata/aikata.yaml`. New projects write
the new path immediately; existing projects continue to read from
the legacy path via fallback and are migrated automatically the
first time `aikata generate` or `aikata doctor --fix` runs.

Shipped:

- `aikata init` writes `.aikata/aikata.yaml`.
- Reader fallback (`internal/config.Resolve`) keeps v0.2 / v0.3.0 /
  v0.3.1 projects working unchanged; the primary path wins when both
  exist.
- `aikata generate` auto-migrates `.ai/` → `.aikata/` on first run
  (atomic move via temp + rename) and prints a single `notice:` to
  stderr. Migration failures degrade to a warning so a user's run is
  never blocked.
- `aikata doctor` reports legacy-path projects as a warning with
  `Code: "config.legacy-path"`. `aikata doctor --fix` performs the
  same atomic move. The warning falls silent once the primary path
  exists.
- Template README / ARCHITECTURE / .gitignore copy for the standard,
  flutter, and typescript presets (English and Japanese) was rewritten
  to mention the new path.

This fallback was later retired in the planned v0.7.4 cleanup
([ADR 0020](./adr/0020-retire-ai-config-fallback.md)).

---

## v0.4.0 — Authoring ergonomics, first wave ✅ (released 2026-05-24)

**Goal**: editing an aikata project is as ergonomic as creating one.

Shipped:

- `aikata add <component>` cobra parent + registry-driven dispatch
  (`internal/components`). The leaf subcommand for every new
  component appears automatically; no per-component switch in
  `add.go`.
- `aikata add adr "<title>"` — auto-numbered ADR skeleton under
  `docs/adr/NNNN-<slug>.md` using the v0.3 `internal/adr` helper.
  Refuses to clobber an existing ADR sharing the same slug.
- `aikata add stack <name>` — writes `docs/stacks/<name>.md` and
  appends the stack to `.aikata/aikata.yaml` `stacks:`. Bundled
  stacks (flutter, typescript) reuse their preset template.
  Idempotent on re-run; user-edited files survive.
- `aikata add memory` — opt-in equivalent of `aikata init
  --with-memory` for projects that did not enable the memory slot
  at init time.
- `aikata list components` — registry listing with the same
  versioned `--json` envelope as `list presets|stacks|ai-tools`.
- `internal/config.Save / Load` — atomic read/write pair behind the
  add commands' config mutations.
- [ADR 0010](./adr/0010-memory-projection-deferred-to-v0-6.md)
  — defers memory generate-projection (ADR-0004 option δ) to v0.6
  so the per-tool plugin work can own the spec.
- Command vocabulary cleanup carried over from v0.3.1 / ADR 0009:
  `aikata update` reserved for CLI self-update, future template
  diff-merge renamed to `aikata sync`. `aikata generate` continues
  to refuse to rewrite canonical project documents as a side
  effect.

## v0.4.1 — Authoring ergonomics, second wave ✅ (released 2026-05-25)

Closes the second wave of `aikata add` and brings `aikata init` to
flag / prompt parity.

Shipped:

- `aikata add <component>` — second wave: `ai-tool`, `ui`, `api`,
  `tdd`, `changelog`. The four single-file components share a
  `singleFile` helper so a future optional component is a one-line
  registry entry. `ui` emits `UI.md` for UI / UX / product-design
  guidance; the generic `DESIGN.md` is intentionally never
  introduced (ADR 0007).
- `aikata add ai-tool <name>` — post-init counterpart of
  `--ai-tools`; validates against `internal/generate.KnownTools()`,
  appends to `cfg.AITools` (sorted), and persists via `config.Save`.
- `aikata init --with-ui` / `--with-api` / `--with-tdd` /
  `--with-changelog` — flag parity with the second-wave components.
  Each new flag has a matching interactive prompt (default N) so
  flag / prompt parity stays in lockstep.
- `scaffold.Run` dispatch refactored from a single hand-written
  `WithMemory` if-block to a table covering memory + the four new
  components. Every v0.4.0 golden tree remains byte-identical.
- Repository dogfood: aikata's own configuration migrated from
  `.ai/aikata.yaml` to `.aikata/aikata.yaml`; `aikata doctor` now
  reports zero `config.legacy-path` warnings on the repo itself.

---

## v0.4.2 — Update check ✅ (released 2026-05-25)

Lightweight follow-up that lands the read-only half of the ADR 0009
`aikata update` surface so users can discover newer releases without
leaving the terminal. Binary self-update remains scheduled for v0.6
alongside the installer-source metadata layer.

Shipped:

- `aikata update --check` — opt-in release check against the GitHub
  Releases API. Reports current / latest version and prints generic
  upgrade guidance covering `go install`, the install script, and
  manual GitHub download. Does not modify the installed binary.
- `aikata update` (no flag) — prints a notice that self-update is
  planned for v0.6 and points users at `--check`; exits 0 so scripts
  don't surface false failures.
- `--json` envelope shared with `doctor` / `list` / `describe`:
  `{version: 1, kind: "update-check", current, latest, status,
  release_url}`.
- New `internal/release/` package owns the HTTP boundary plus the
  minimal semver comparison; v0.6 self-update will reuse both.

---

## v0.5 — `aikata sync` ✅ (released 2026-05-26)

**Goal**: keep an aikata project current with the canonical templates
without losing the user's edits.

- `aikata sync` 3-way diff-merge against `.aikata/manifest.yaml`. The
  init-time template hashes act as the common ancestor; user edits
  are preserved, upstream-only changes auto-apply, true conflicts get
  git-merge-style file-level markers. ADR 0011 documents the
  contract.
- `aikata sync --rebaseline` seeds a manifest from current on-disk
  state for projects that pre-date v0.5.
- `aikata sync --dry-run` previews the merge without writing.
- `--json` envelope joins the family used by `doctor` / `list` /
  `describe` / `update`:
  `{version: 1, kind: "sync", files: [...], summary: {...}}`.
- Schema migration framework (`internal/config.MigrateAikataYaml` /
  `LoadMigrated`) wired into `sync.Run` per ADR 0011 D3. v0.5 ships
  with the registry empty (only v1 exists); v2+ schemas land as a
  one-row addition.
- Dogfooding gate is now **binding**: `aikata doctor --strict` runs
  on every CI build and treats warnings as exit-3 failures, and a
  new `aikata generate is byte-identical` step `git diff --exit-code`
  s the committed `CLAUDE.md` / `.cursor/rules/main.mdc` against
  fresh `aikata generate` output.

Known limitations (v0.5.x follow-up candidates):

- `aikata add <component>` post-init does not yet append to the
  manifest, so files it adds are visible to sync only via 2-way
  diff. Init-time scaffolds cover the common drift case.
- Conflicts are written at file granularity (entire body wrapped in
  markers). Line-level diff3 is a follow-up if real-world feedback
  shows the file-level form is too coarse.
- `--rebaseline` assumes the `standard` preset when no manifest
  exists; explicit `--preset` override lands later if anyone reports
  a mismatch.

---

## v0.6.0 — Packaging & distribution (partial) ✅ (released 2026-05-26)

**Goal**: aikata is one-click installable in Claude Code,
one-shell-line elsewhere, and scales to a monorepo.

v0.6.0 ships the **agent-doable subset**. User-action channels
(Homebrew tap, npm wrapper, marketplace listing) are deferred to the
v0.9.9 channel-publication line so v0.6.x can close on repository-local
work. v0.6.1 / v0.6.2 / v0.6.3 are unscheduled patch releases
(rebaseline fix, ROADMAP & manifest hygiene, scope derivation) — see
their own sections.

Shipped:

- `aikata init --monorepo` — nested `apps/<name>/AGENTS.md` plus
  `docs/monorepo.md` explainer; `features.monorepo` flipped in
  `.aikata/aikata.yaml`. v0.4 single-file components stay
  orthogonal; users opt in independently. Per-app `AGENTS.md` files
  are user-managed (aikata does not regenerate them).
- `internal/install` — detection layer that records which channel
  placed the binary (`github-release`, `install-script`,
  `go-install`, `homebrew`, `npm`, `unknown`). Reads either a
  build-time ldflag or `<install-dir>/aikata.install-source` (written
  by `scripts/install.sh`). Foundation for a v0.9.9 native
  `aikata update --apply` self-update; the consuming side ships in
  the v0.9.9 channel-publication line once Homebrew / npm channels
  exist.
- `dist/claude-code/plugin/` — Claude Code plugin scaffold bundling
  the v0.3.1 skill with four slash commands (`/aikata-init`,
  `/aikata-generate`, `/aikata-doctor`, `/aikata-sync`). Installable
  manually today (`cp -r dist/claude-code/plugin/*
  ~/.claude/plugins/aikata/`); marketplace listing is deferred to
  v0.9.9.

Deferred again to v0.7+ (no projection in v0.6):

- ~~Memory generate-projection (ADR-0004 option δ).~~ **Deferred
  again** by [ADR 0012](./adr/0012-memory-projection-deferral-extended.md):
  the per-tool memory channel layouts have not stabilized between
  v0.4 and v0.6 and no dogfooding case has reported drift. The v0.6
  Claude Code plugin's skill text covers the discoverability gap
  without on-disk projection. Revisit when a documented dogfooder
  reports `docs/memory/` ↔ tool-channel drift; a follow-up ADR will
  pick the then-stable layout and supersede ADR 0012.

## v0.6.1 — `aikata sync --rebaseline` regression fix ✅ (released 2026-05-26)

Unscheduled patch release. v0.6.0's `--rebaseline` flag walked into
the 2-way diff branch with an empty ancestor table, classified every
customised file as a true conflict, and wrote `<<<<<<< / >>>>>>>`
markers into source files — the exact opposite of "non-destructive
seeding". v0.6.1 makes `--rebaseline` skip the merge entirely and
write only `.aikata/manifest.yaml`, with the manifest's ancestor
hashes seeded from the **upstream rendering** so the user's
customisations register as `user-only-edit` on the next sync.

Originally planned v0.6.1 work (channel publication) is deferred to
v0.9.9; nothing else lands in v0.6.1 so the fix can be tagged and
shipped without coupling.

- [x] `internal/sync/sync.go` — non-destructive rebaseline path.
- [x] `ErrNoManifest` message rewritten to spell out
      "non-destructive".
- [x] ADR 0011 gains a "Rebaseline ancestor choice" subsection
      explaining why ancestor = upstream rendering (not on-disk
      bytes).
- [x] Four new tests in `internal/sync/sync_test.go` covering
      byte-preservation, no-merge-ran invariant, dry-run, and the
      post-rebaseline integration scenario.
- [x] Hitchhiking on the same release: `scripts/install.sh` warns
      when another `aikata` earlier on `$PATH` shadows the freshly
      installed binary (#82); `dist/README.md` clarifies Codex /
      third-party skill scope (#81).

## v0.6.2 — ROADMAP template & manifest hygiene ✅ (released 2026-05-28)

Quality-of-life improvements built on the v0.6.1 base. Two
coordinated changes:

- A `ROADMAP.md` template is added to the `standard` / `flutter` /
  `typescript` presets (ja + en). Scaffolded projects now have a
  first-class place to describe their direction alongside SPEC /
  ARCHITECTURE / GLOSSARY. `minimal` stays minimal.
- `aikata add <component>` now writes entries to
  `.aikata/manifest.yaml`, so subsequent `aikata sync` runs treat
  add'd files as `user-only-edit` (preserving customisations)
  instead of `upstream-added` (potentially overwriting them).

ADR 0014 (`manifest is a living record`) ships alongside to make
the new invariant explicit.

- [x] `internal/templates/data/presets/{standard,flutter,typescript}/{en,ja}/ROADMAP.md.tmpl`
      — 6 new template files.
- [x] `internal/components/component.go` — new `RecordInManifest`
      helper called from `memory.go` / `singlefile.go` / `stack.go`.
- [x] `internal/components/manifest_record_test.go` — 6 new tests
      covering happy path, idempotency, preservation, no-op without
      `aikata.yaml`, and stack guide registration (both fresh-write
      and file-pre-exists branches).
- [x] `docs/adr/0014-manifest-living-record.md` — new ADR.

## v0.6.3 — Sync scope derivation & CLI overrides ✅ (released 2026-05-28)

Scope-derivation release. `aikata sync` now reads scope from a small
ordered hierarchy (defaults → `aikata.yaml` → manifest → CLI
overrides) instead of the hard-coded "preset=standard / no opt-ins"
rebaseline fallback. Four new one-off override flags (`--preset` /
`--lang` / `--stack` / `--with-monorepo`) round out the surface.

- [x] `internal/sync/plan.go` — `derivePlan` extended with
      `overrides` parameter and `aikata.yaml` reads; `inferFlags`
      gains a monorepo rule.
- [x] `internal/sync/sync.go` — `Options` gains four `Override*`
      pointer fields, threaded through to `scaffold.Options`.
- [x] `internal/cli/sync.go` — new `--preset` / `--lang` /
      `--stack` (repeatable) / `--with-monorepo` flags guarded by
      `cmd.Flags().Changed(...)`.
- [x] `internal/sync/derive_test.go` — 7 new test cases covering
      the hierarchy at the function level plus integration cases
      via `Run` (override changes scope; overrides do not mutate
      manifest).
- [x] `docs/adr/0013-sync-scope-derivation.md` — new ADR.
- [x] Stack consistency follow-up: `aikata add stack` records the
      manifest entry even when the on-disk stack guide already
      exists, matching the singleFile / memory pattern.

## v0.7.x — Schema & adoption hardening (in progress)

**Goal**: reduce inference before the v1.0 stable surface. v0.6.x made
`sync` safer by deriving scope from `aikata.yaml`, manifest paths, and
CLI overrides; v0.7.x should make durable intent explicit and simplify
the post-init command surface before the first stable CLI contract.

## v0.7.0 — Schema v2 ✅ (released 2026-05-29)

First sub-release of the v0.7.x line. Lands the schema bump so the
remaining v0.7.x items can build on a declarative source of truth for
optional template scope.

- [x] **`.aikata/aikata.yaml` schema v2** — a typed `components:`
      block records `memory`, `ui`, `api`, `tdd`, `changelog`, and
      `monorepo` as first-class fields. A v1 → v2 forward migrator
      lifts the legacy `features.tdd` / `features.monorepo` keys
      into the new block; migration is lazy (only writers persist),
      so v1 reads continue to work through the rest of the v0.x line.
- [x] `aikata sync` scope derivation OR-merges `cfg.Components.*`
      with manifest path inference and the legacy `features.*` keys
      so schema-v2 projects, in-flight v1 projects, and projects with
      stale manifests all converge to the same scope.
- [x] `internal/scaffold/scaffold.go` persists `Options.With*` into
      `Components` on init, removing the path-inference dependency
      for freshly-scaffolded projects.
- [x] `docs/adr/0016-aikata-yaml-schema-v2.md` — new ADR; Q-DESIGN-09
      resolved.
- [x] Every `testdata/golden/*/.aikata/aikata.yaml` fixture rewritten
      to the v2 shape; the schema bump is part of the golden tree.
- [x] aikata's own `.aikata/aikata.yaml` migrated to v2 (dogfood).

## v0.7.1 — Purpose-based CLI split ✅ (released 2026-05-29)

Second sub-release of v0.7.x. Lands the post-init command surface
that the schema-v2 `components:` block enables.

- [x] **`aikata enable <capability>`** — persists durable project
      capabilities (memory, ui, api, tdd, changelog, monorepo, stack
      `<name>`, ai-tool `<name>`). Each leaf renders the corresponding
      files, records them in the manifest (ADR 0014), and flips the
      matching schema-v2 `components.*` flag or appends to
      `ai_tools:` / `stacks:` in `.aikata/aikata.yaml`.
- [x] **`aikata new <artifact>`** — stamps one-off authoring scaffolds
      (`adr "<title>"`). No durable schema flip.
- [x] **`aikata add <component>` removed** without a compatibility
      alias. Per ADR 0017 the pre-v1.0 surface drops it cleanly.
- [x] `aikata list capabilities` / `aikata list artifacts` replace
      the pre-v0.7.1 `list components`, mirroring the split. Both
      keep the versioned `--json` envelope.
- [x] **`aikata enable monorepo`** — new `monorepoComponent`
      registered alongside the existing memory / ui / api / tdd /
      changelog / stack / ai-tool capabilities.
- [x] `enable`-tier flips the matching `components.*` flag via the
      new `EnableComponentInConfig` helper, so `aikata sync` no
      longer needs manifest path inference for components touched
      post-init.
- [x] `stack` / `ai-tool` switched from `config.Load` to
      `config.LoadMigrated` so the v1 → v2 schema migration
      (ADR 0016) runs as a side-effect of any enable-tier write.
- [x] `docs/adr/0017-post-init-command-taxonomy.md` — new ADR;
      Q-DESIGN-10 resolved.
- [x] **`aikata expand <tier>` deferred**: with only `standard` as a
      meaningful target and unclear semantics for minimal → standard
      growth, the verb is held until `extended` exists or a real
      project surfaces the need.

## v0.7.2 — Adoption, repair, managed append ✅ (released 2026-05-29)

Closes the v0.7.x line. Three loosely-coupled items shipped together:
the `aikata sync` no-silent-delete contract (ADR 0019), the existing-
repo adoption guide, and the managed-block append writer for
project-owned files like `.gitignore` (ADR 0018).

- [x] **Missing-file repair semantics** (ADR 0019) — `aikata sync`
      may add or refresh managed files via the existing
      `StatusUpstreamAdded` / `StatusUpstreamApplied` branches; it
      must never delete files merely because they fell out of scope.
      `StatusUpstreamRemoved` and `StatusUserDeleted` already emit no
      write — the ADR pins the contract and a new sync test
      (`TestRun_UpstreamRemoved_DoesNotDelete`) guards against
      regression.
- [x] **Existing-repo adoption guide** — `docs/adoption.md` covers
      the five concrete scenarios users actually hit
      (`AGENTS.md` already exists, hand-written `CLAUDE.md` or
      `.cursor/rules/`, hand-written `.gitignore`, pre-existing
      `docs/memory/`, historical `.ai/` config). Per the
      documentation-first stance no `aikata adopt` parser is built.
- [x] **Managed-block append for `.gitignore`** (ADR 0018) — new
      `internal/managed/` package writes the aikata-owned section
      between `# >>> aikata managed >>>` / `# <<< aikata managed <<<`
      markers; the scaffold layer routes `.gitignore` through it
      when the target file already exists. Idempotent re-runs
      converge.
- [x] `aikata init --force` against an existing `.gitignore` no
      longer destroys user-owned entries. The integration is
      intentionally narrow (init only); `aikata sync` continues to
      use the 3-way merge for `.gitignore`, with the user-only-edit
      preservation already working. Wider integration (sync + the
      no-`--force` path) tracked beyond v0.7.x.

## v0.7.3 — Doctor scope and exclusion ✅ (released 2026-05-29)

Patch release driven by a user report from `personal-skills`
(v0.6.1 baseline): `aikata doctor` flagged 62 spurious frontmatter
errors against a Claude Code plugin tree (`plugins/<name>/skills/<name>/SKILL.md`,
whose frontmatter is Anthropic's `name` + `description` only).
aikata itself sidesteps the same problem only because `dist/` is in
the hardcoded `skippedDirs` — a blind spot for users with the
mirror layout.

Shipped:

- **Configurable doctor exclusion** (ADR 0021): `.aikata/aikata.yaml`
  gains an optional top-level `doctor:` block with an `exclude:`
  glob list. Matching paths skip `checkFrontmatter` /
  `checkUpdated` / `checkGlossary` uniformly. User-supplied
  excludes are additive with the hardcoded
  `skippedDirs` / `skippedFiles` baselines. Zero default
  exclusions ship; ADR documents recommended snippets for Claude
  Code plugin layouts.
- `internal/doctor/glob.go` — small in-tree matcher (`*`, `**`,
  literals). No new external dependency. doublestar /
  `filepath.Match` recorded as Alternatives Considered.
- `internal/doctor.Options.Excludes []string`, threaded by
  `internal/cli/doctor.go` from `config.Load` (non-mutating —
  `aikata doctor` without `--fix` stays read-only).
- Q-DOCTOR-01 resolved by ADR 0021.

Out of v0.7.3 (deferred):

- `doctor.frontmatter_required_paths` (reverse-include
  specification). Held until a real user requests the symmetric
  "only check docs/**" knob.
- Auto-detection of known plugin layouts. Held until per-tool
  plugin specs stabilise (see ADR 0015).
- Severity downgrade for non-aikata subtrees. Interacts with
  `--strict` and deserves its own ADR if pursued.
- `aikata sync` exclusion. sync is manifest-driven; revisit if a
  user reports the analogous noise.

## v0.7.4 — Retire legacy `.ai/` config fallback ✅ (released 2026-05-29)

Pre-v1 cleanup release. v0.3.2 moved aikata-owned config to
`.aikata/aikata.yaml`; v0.7.4 removes the remaining path-level
compatibility branch so the stable surface has a single config
location.

- [x] **Remove `.ai/aikata.yaml` reads** — `internal/config.Resolve`
      accepts only `.aikata/aikata.yaml`; commands no longer fall
      back to `.ai/`.
- [x] **Delete automatic migration** — remove `MoveLegacyToPrimary`
      and the `aikata generate` / `aikata doctor --fix`
      auto-migration paths. Users with old projects manually move
      `.ai/aikata.yaml` to `.aikata/aikata.yaml`.
- [x] **Drop `config.legacy-path` doctor check** — doctor no longer
      warns about or fixes `.ai/`; `.ai/` is treated as user-owned or
      third-party state.
- [x] **Document the cleanup** — ADR 0020 records the decision, and
      the adoption guide keeps `.ai/` only as historical migration
      context.

v0.7.x is considered closed at v0.7.3 (with v0.7.4 as an optional
cleanup tail) unless a further critical patch is needed. v0.8.x
covers security & governance hardening of the aikata repository
itself; v0.9.0 covers core-concept stabilization; v0.9.9 covers channel
publication; v1.0 covers the stable surface (see below).

Out of v0.7.x intentionally:

- Memory projection (deferred again by ADR 0012).
- Third-party skill catalog management.
- New remote template fetching.
- An `aikata expand <tier>` verb — deferred per ADR 0017 until
  `extended` exists or a real project surfaces the need.

v0.9.0 is the core-concept stabilization line inserted after the v0.8.5
maintainer review (ADR 0028). v0.9.9 is the channel-publication line
(Homebrew / npm / marketplace / native self-update); that work was
previously numbered v0.8.x and moved back one minor when the security &
governance hardening line was inserted ahead of it (ADR 0022).

---

## v0.8.x — Security & governance hardening (pending)

**Goal**: bring the aikata *repository itself* up to the governance and
supply-chain bar expected of a publicly published OSS project before the
v0.9.9 channel-publication line widens its distribution footprint. Scope
is deliberately limited to aikata's own repo posture; the operational-
readiness *templates* that `aikata init --preset extended` scaffolds for
*user* projects (`SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`,
issue / PR templates) remain a v1.0 item and are intentionally not pulled
forward here. ADR 0022 records the re-sequencing decision and why this
line was inserted ahead of channel publication.

The security review that motivated this line found no exploitable
vulnerability in aikata's code, install script, or CI — the gaps are
governance and supply-chain hardening, mirroring guardrails already in
place in the sibling `personal-skills` repo.

### v0.8.0 — Governance & secret-scan ✅ (released 2026-05-29)

Low-risk repository guardrails. No change to the binary or templates.

- [x] **`SECURITY.md`** — private vulnerability disclosure via GitHub
      Security Advisories, security expectations (never commit `.env`,
      credentials, tokens, or private keys; use placeholders), and an
      **Agent Safety** section: AI agents must not push to protected
      branches, merge without human approval, weaken `CODEOWNERS` /
      validation, or add remote-code-execution behaviour without an
      ADR + review. Carries the standard aikata five-key frontmatter so
      `aikata doctor --strict` stays green without an exclusion.
- [x] **`.github/CODEOWNERS`** — require maintainer review on the
      security-sensitive surface: `/.github/`, `/AGENTS.md`,
      `/SECURITY.md`, `/.goreleaser.yml`, `/ROADMAP.md`, and
      `/docs/adr/`.
- [x] **Secret / privacy scan gate** — realised as a Go test in
      `internal/repolint` (no runtime code; not in the binary) that runs
      inside the existing `go test ./...` CI matrix on all three OSes,
      avoiding a separate workflow and shell-escaping fragility. It
      asserts `.env` / `.env.local` (and any `*.local*`) are not tracked
      and scans tracked files for key material (PEM `PRIVATE KEY`
      headers; `api_key` / `client_secret` / `refresh_token`
      assignments with a value), local user paths (`/Users/...`,
      `~/Workspace/...`), and private emails. Patterns are tightened to
      require the *shape of a real leak*, so the project's own prose
      documenting them does not self-trip; a table-driven self-test
      proves on every run that no pattern is silently dead. The
      `personal-skills` personal-profile denylist is **not** ported
      (aikata holds no personal data).
- [x] **`.github/dependabot.yml`** — weekly `github-actions` and
      `gomod` update checks.
- [x] **`.gitignore` hardening** — add `.env.local`, `*.local.yaml`,
      `*.local.yml` (the committed `.aikata/aikata.yaml` does not match
      these, so dogfooding is unaffected).
- [x] **`CONTRIBUTING.md`** — state "no direct pushes to `main`" as an
      explicit rule and add an Agent Contributions section that
      cross-references the SECURITY.md Agent Safety constraints.

### v0.8.1 — Supply-chain signing ✅ (released 2026-05-29)

Release-pipeline hardening. Split from v0.8.0 so a pipeline change
cannot destabilise the governance work, and tagged separately.

- [x] **Cosign keyless signing** of `checksums.txt` via GitHub OIDC
      (`id-token: write` added to `release.yml`; cosign installed in the
      job). Signing the checksum file transitively authenticates every
      archive. No long-lived key to manage.
- [x] **SBOM generation** (syft via GoReleaser `sboms:`), one SPDX SBOM
      per archive shipped as a release asset; syft installed in the job.
- [x] **SHA-pin GitHub Actions** — `actions/checkout`,
      `actions/setup-go`, `golangci/golangci-lint-action`,
      `goreleaser/goreleaser-action`, `sigstore/cosign-installer`, and
      `anchore/sbom-action` pinned to full commit SHAs (version in a
      trailing comment); Dependabot keeps them current. A new
      `goreleaser-check` CI job runs `goreleaser check` on every PR so a
      bad `.goreleaser.yml` fails in review, not at tag-push.
- [x] **Verification docs** — README "Verifying a release signature"
      section + a `scripts/install.sh` note (`cosign verify-blob`); the
      installer stays dependency-free.
- [x] **ADR 0023** — records the signing-mechanism decision (keyless
      cosign over GPG/long-lived key), SBOM, and SHA-pinning.

Out of v0.8.x intentionally:

- Code-level defense-in-depth (`filepath.IsLocal()`, `gosec`). The
  review found aikata's code already safe; revisit only if a concrete
  issue surfaces.
- `--preset extended` governance templates (stay at v1.0).

---

## v0.8.2 — CLI surface: scope × stack ✅ (released 2026-05-30)

**Goal**: a pre-v1.0 stable-surface correction, interleaved into the
0.8.x number space (not part of the security & governance theme of
v0.8.0 / v0.8.1; see [ADR 0024](./adr/0024-scope-stack-axes-split.md)
for why it is numbered here rather than renumbering the v0.9.9 line).

`aikata init`'s single `--preset minimal|standard|flutter|typescript`
flag fuses two orthogonal axes — documentation **scope** and target
**stack** — into one enum. Language is already a separate `--lang` axis;
stack should be too. v0.8.2 makes stack a first-class flag and opens the
`--preset` deprecation window. It delivers the orthogonal **flag
surface**, **not** new buildable combinations: the four preset template
trees are independently authored (ADR 0024 Scope boundary), so
`(scope, stacks)` resolves only to the four combinations that exist
today; everything else errors explicitly. Unlocking new combinations
(`minimal` + stack, multi-stack) needs a template refactor deferred as
follow-up.

- [x] **`--scope` flag** — `minimal | standard` (single-valued, default
      `standard`; `extended` stays reserved per ADR 0017).
- [x] **`--stack` flag** — multi-valued *in syntax* (repeatable and/or
      comma-separated); empty = stack-agnostic. Writes the existing
      `stacks` list in `aikata.yaml` directly (no schema bump). Removes
      `stacksForPreset`. v0.8.2 accepts only a single stack paired with
      `--scope standard`; other combinations error (see below).
- [x] **Bounded `(scope, stack)` resolution** — maps to the existing
      trees only: `minimal` / `standard` / `standard+flutter` /
      `standard+typescript`. `minimal`+stack, multi-stack, and
      `extended` return a clear "not yet supported" error rather than a
      silent half-wired fallback.
- [x] **`--preset` deprecated alias** — maps `minimal`/`standard` →
      `--scope`, `flutter`/`typescript` → `--scope standard --stack
      <name>`; prints a one-line deprecation notice; erroring if
      combined with `--scope`/`--stack`. Removed at v1.0.
- [x] **Interactive prompt** — ask scope, then stack (single, empty
      allowed), then lang; never prompt for "preset".
- [x] **Doc/glossary alignment** — update the GLOSSARY `preset` entry to
      "deprecated alias for `--scope`", add a `scope` entry, align
      `stack` entries (doctor runs a glossary-consistency check); update
      README / SPEC / ARCHITECTURE / `docs/` help and examples. Shipped
      ROADMAP/CHANGELOG history is left intact.
- [x] **Q-DESIGN-04 closed** — superseded by ADR 0024; presets are no
      longer a composition mechanism.

Out of v0.8.2 intentionally:

- **Template refactor for new `scope × stack` combinations** — scope
  base + stack partials, re-deriving flutter/typescript without drift.
  A separate follow-up feature; v0.8.2 ships only the flag shape and
  deprecation window (ADR 0024 Scope boundary).
- `extended` scope behaviour — stays reserved (v1.0 governance pack).
- Removing the `--preset` alias — deferred to v1.0 so the deprecation
  has a migration window.

---

## v0.8.3 — `aikata sync` divergent-file preservation ✅ (released 2026-05-30)

**Goal**: `aikata sync` durably preserves files a user has intentionally
rewritten, instead of oscillating between preserving and silently
overwriting them. Like v0.8.2, this is a pre-v1.0 stable-surface
correction interleaved into the 0.8.x number space — *not* part of the
v0.8.0 / v0.8.1 security & governance theme (ADR 0024 established the
precedent for non-security work living in this number space without a
renumber). Because the headline item is a **data-loss** fix, v0.8.3 may
land ahead of v0.8.2 despite the higher number; ROADMAP ordering is
direction, not sequence.

Motivated by a downstream dogfooding report (`itteco`, Flutter, aikata
`v0.8.1`): a preset-managed file the user rewrote for their project
(`README.md`, `.gitignore`, `docs/tasks/current.md`) could not be held
stable across repeated `aikata sync` runs. [ADR 0025](./adr/0025-sync-divergent-file-preservation.md)
records the decisions; the four reported problems map to the items
below.

- [x] **Re-baseline records the upstream rendering, not the on-disk
      snapshot** (ADR 0025 D1 — the data-loss root cause). On a
      conflict-free run, the manifest is regenerated from the in-memory
      upstream rendering rather than `postMergeSnapshot`'s on-disk
      re-read, unifying the post-clean-run path with the existing
      `--rebaseline` ancestor principle (ADR 0011). The only
      behavioural change is `user-only-edit`: its ancestor stays at the
      upstream rendering, so the file is preserved across unlimited
      syncs instead of being absorbed as the ancestor and overwritten
      next run. Also keeps `user-deleted` entries so a respected
      deletion is not silently re-created (ADR 0019). Independent of the
      items below — must ship even if the `owned` marker slips.
- [x] **Per-file `owned` opt-out** (ADR 0025 D2) — an optional
      `sync.own:` glob list in `.aikata/aikata.yaml` (same matcher and
      additive semantics as `doctor.exclude`, ADR 0021; no schema bump).
      Matching paths report an `owned` status and are never
      rendered-compared, conflict-markered, or overwritten. Removes the
      residual conflict noise D1 leaves on *fully* forked files and
      replaces the reporter's manual `git restore` workaround.
- [x] **Remove the dead `docs.generate_gitignore` flag** (ADR 0025 D3)
      — the field is defined but never read, so removal is a
      behavioural no-op. `.gitignore` stays managed by the ADR 0018
      managed-append writer (non-destructive by default); a user who
      wants sync to leave it alone uses `sync.own` (D2) rather than a
      single-purpose flag. Old configs carrying the key still parse.
- [x] **`--rebaseline` is explicit when a manifest exists; add
      `--reseed`** (ADR 0025 D4) — passing `--rebaseline` to a project
      that already has a manifest emits a notice instead of a silent
      no-op; `--reseed` re-anchors an existing manifest to the current
      upstream rendering (manifest-only write, no source files touched).
- [x] **GLOSSARY / SPEC / README alignment** — add the `owned` sync
      status and `sync.own` to GLOSSARY (doctor runs a glossary check),
      document `owned` / `--reseed` / the `generate_gitignore` behaviour
      in SPEC §4, and align README / `docs/` help. Deferred from this
      planning pass to the implementation PR per the ADR 0024 precedent.

Out of v0.8.3 intentionally:

- **Bumping `.aikata/manifest.yaml` to schema v2** for parity with the
  `aikata.yaml` v2 config. The two schemas version independently by
  design — the manifest is a machine-owned regenerated record (ADR
  0014), the config a user-owned migrated document (ADR 0016). No code
  change, documented as a clarification in ADR 0025 (Problem 4 / the
  reporter's Issue 4); no `doctor` check added.
- Line-level diff3 conflict markers (still file-granularity; a separate
  v0.x follow-up if real-world feedback shows file-level is too coarse).

---

## v0.8.4 — Workflow guide opt-in ✅ (released 2026-05-31)

**Goal**: give AI agents and humans a durable, project-local place to
read collaboration workflow policy without bloating `AGENTS.md` or
requiring the future `extended` governance pack. This is another
pre-v1.0 stable-surface correction in the v0.8.x number space: it
extends aikata's document-centered collaboration model, not the
security / governance hardening theme of v0.8.0 and v0.8.1.

Motivated by a small-team GitHub Flow policy used across personal and
team projects: short-lived branches, Conventional Commits, small PRs,
squash-only merges, SemVer release tags, trusted-committer review
rules, and CI gates. [ADR 0026](./adr/0026-workflow-guides-as-opt-in-collaboration-docs.md)
records the design: workflow guides are opt-in collaboration documents
under `docs/workflows/`, with Git as the first built-in domain.

- [x] **`aikata enable workflow git`** — new enable-tier command shape
      for workflow domains. The command intentionally uses the broader
      `workflow git` form rather than `git-workflow` so future release,
      deployment, incident, or review workflow guides can share the
      same category.
- [x] **`workflows:` config axis** — persist enabled workflow guides as
      a list in `.aikata/aikata.yaml`, for example `workflows: [git]`.
      This avoids adding one boolean per workflow under `components:`
      and mirrors the list-shaped `stacks:` / `ai_tools:` axes.
- [x] **`docs/workflows/git.md` template** — generate the first built-in
      workflow guide with portable Git policy: GitHub Flow, branch
      naming, Conventional Commits, PR size / squash merge rules,
      SemVer tags, hotfix / mobile release branch conventions where
      applicable, and CI gate expectations. The template must not
      hard-code personal account names, vault paths, private helper
      commands, or a specific paid GitHub plan assumption.
- [x] **Conditional `AGENTS.md` pointer** — when the Git workflow guide
      is enabled, add only a short reference from `AGENTS.md` to
      `docs/workflows/git.md`; do not inline the full policy into the
      canonical instruction file.
- [x] **Golden / config / doctor coverage** — assert that default and
      minimal projects have zero workflow residue; enabling the guide
      writes valid frontmatter, persists config, records the manifest,
      and keeps doctor checks green.
- [x] **Docs alignment** — update README / SPEC / ARCHITECTURE /
      GLOSSARY and generated template docs so users understand the new
      workflow-guide slot and its boundary with `AGENTS.md`,
      `CONTRIBUTING.md`, memory, and working state.

Out of v0.8.4 intentionally:

- **GitHub enforcement artifacts** — no `.github/CODEOWNERS`, PR
  template, Repository Rulesets JSON, or CI workflow generation in this
  release. Those files are useful but environment-specific and belong
  behind a later opt-in design such as `--with-github-files` or a
  separate GitHub operations capability.
- **Custom workflow import** — no `aikata import workflow --from <path>`
  yet. Vault-to-repo and team-template imports need a separate trust,
  template-variable, sync, and conflict model.
- **Contributor governance pack** — `CONTRIBUTING.md`, SECURITY /
  CODE_OF_CONDUCT, issue templates, and broader OSS readiness remain
  part of the v1.0 `extended` scope.

---

## v0.8.5 — Verification expectation in generated templates ✅ (released 2026-05-31)

**Goal**: project the verification discipline aikata practices on itself
(its own `AGENTS.md` Hard Rule 7, `make test && make lint`) into the
templates it generates, without imposing a test-first methodology.
[ADR 0027](./adr/0027-verification-expectation-in-generated-templates.md)
(Accepted 2026-05-31) records the design; it resolved the former
Q-DESIGN-11.

Framing: "test existence" is already a default rule, "test-first" stays
opt-in (ADR 0003), and the gap is methodology-neutral **verification** —
"run the checks and show the output before claiming done".

- [x] **PR-A — `docs/testing.md` strengthening** _(opt-in)_. Strengthened
      `internal/templates/data/components/tdd/{en,ja}/tdd.md.tmpl` from the
      bare TODO skeleton with a "Why this matters for AI collaboration"
      rationale and a clearly-marked opt-in test-first recommendation.
      Golden `standard-with-extras` regenerated.
- [x] **PR-B — verification rule in standard `AGENTS.md`** _(default
      output)_. Added the conditional, methodology-neutral hard rule
      ("**Verify before declaring done.** If the project has tests or a
      build, run them and show the output …") next to the existing
      `Add tests` rule in
      `internal/templates/data/presets/standard/{en,ja}/AGENTS.md.tmpl`.
      The `flutter` / `typescript` presets were left untouched (they
      already carry a stronger stack-specific verification line); `minimal`
      stays lean. Golden standard fixtures regenerated.
- [x] **Do-No-Harm coverage** — the existing byte-comparison golden tests
      for `minimal` / `minimal-ja` are unchanged after regeneration,
      proving zero verification-rule and zero TDD-recommendation residue
      in non-adopting scopes; the rule is conditional ("if they exist") so
      it reads as inert in a test-less project.
- [x] **Release-flow discoverability** — added a "Cut a release" row to
      the repo's own `AGENTS.md` navigation matrix pointing at
      `CONTRIBUTING.md` § Release flow and `ARCHITECTURE.md` §6.5, so an
      agent cutting a release finds the ritual from the first-read file.

All changes follow the standard gates: a `[Unreleased]` CHANGELOG entry,
`make test && make lint`, `aikata doctor` clean, the `aikata generate`
byte-identity check, and English commit / PR text.

Out of v0.8.5 intentionally:

- **Test-first as a default.** Not happening — ADR 0003 keeps test-first
  opt-in. v0.8.5 adds verification, not methodology.

---

## v0.9.2 — Scope discipline ✅ (released 2026-06-01)

Two scope-discipline changes that keep the default `standard` scaffold
lean: add opt-in brand-exploration artifacts, and move the empty
reusable-prompt library off the default surface. Both follow the v0.9.0
core-concept stabilization principle (ADR 0028) — the default scaffold
carries only documents with a distinct, non-latent role; convenience
documents are opt-in.

**Brand-exploration artifacts** ([ADR 0031](./adr/0031-brand-exploration-documents-as-one-off-artifacts.md)).
Mobile-app dogfooding showed that icon and mascot exploration documents
repeatedly save product-context reconstruction work, especially when
prompts must be passed to an external image-generation LLM that cannot
read the repository.

- [x] **`aikata new app-icon`** — stamp
      `docs/design/app-icon-concepts.md` with a concise bilingual starter
      structure: external-LLM product context, brand / technical
      constraints, concept comparison, image-generation prompts,
      negative prompts, and selection follow-up.
- [x] **`aikata new mascot`** — stamp
      `docs/design/mascot-character-ideas.md` with a concise bilingual
      starter structure: external-LLM product context, mascot role /
      tone, candidate comparison, image-generation prompts, intended
      product surfaces, and selection follow-up.
- [x] **One-off artifact semantics** — both registered under
      `aikata list artifacts`; no config flags, init prompts, preset
      defaults, or `.aikata/manifest.yaml` entries. After stamping, the
      project owns the files and `aikata sync` does not restore or merge
      them. Collision refuses rather than overwriting.

**Reusable-prompts opt-in** ([ADR 0034](./adr/0034-reusable-prompts-opt-in-capability.md),
Q-DESIGN-12). The empty `docs/prompts.md` skeleton is removed from the
default `standard` / `flutter` / `typescript` scopes and offered as an
opt-in capability.

- [x] **`aikata enable prompts` / `--with-prompts`** — single-file
      capability rendering `docs/prompts.md`; schema-v2
      `components.prompts`, manifest-recorded, `sync`-preserved. Removal
      from the default scaffold is sync-visible but non-destructive
      (ADR 0019).
- [x] **Verification** — component / CLI tests for en + ja rendering,
      collision refusal, dry-run output, artifact listing; golden trees
      confirm `docs/prompts.md` leaves the default scopes and appears only
      in `standard-with-extras`; `minimal` golden trees are unchanged.

Out of v0.9.2 intentionally:

- Default inclusion of brand artifacts in `standard`, `extended`,
  `--with-ui`, or stack selections.
- A branding hierarchy or speculative `new logo` / `new brand-guide`
  commands without repeated dogfooding evidence.

---

## v0.9.3 — Agent-ecosystem distribution ✅ (released 2026-05-31)

First of the three value-ordered channel-publication lines that
[ADR 0032](./adr/0032-split-channel-publication-by-distribution-value.md)
split out of the former single v0.9.9 line. v0.9.3 ships the distribution
surface that matches aikata's core identity — agent-facing shared-context
tooling, discoverable where its users already work (ADR 0028) — and is the
**prioritized** line. Numeric order is direction, not ship order: v0.9.3
is independent of the still-unshipped v0.9.2 brand-exploration line (the
v0.8.3-before-v0.8.2 precedent).

- [x] **Universal `npx skills add` package** — first-party aikata usage
      guidance at `dist/universal-skill/SKILL.md` per
      [ADR 0015](./adr/0015-first-party-skill-plugin-distribution.md).
      A tool-agnostic skill installed via
      `npx skills add …/tree/main/dist/universal-skill --agent universal`;
      `dist/universal-skill/` is canonical, so no publication mirror is
      required. Also shipped as the `aikata-universal-skill.md` release
      asset.
- [x] **Claude Code marketplace readiness** — a root
      `.claude-plugin/marketplace.json` lists the v0.6.0 plugin scaffold,
      making the repo installable as a self-hosted marketplace
      (`/plugin marketplace add shigindo-inc/aikata`); `plugin.json` is
      finalized for listing (version → 0.9.3, `category` / `keywords`).
      The **submission act stays gated** on the upstream marketplace flow
      plus a maintainer submitting for review; per the v0.6.0
      agent-doable-subset precedent that external step does not block the
      release. The manual plugin-install path stays supported regardless.

Out of v0.9.3 intentionally:

- Homebrew tap and npm wrapper (deferred to v0.9.9 — convenience-only).
- Native `aikata update --apply` (v0.9.4).

---

## v0.9.4 — Native self-update for existing channels ✅ (released 2026-06-01)

Second value-ordered line (ADR 0032 D2). Ships `aikata update --apply`
covering the channels that exist today. The foundation
(`internal/install.Detect()` and the `aikata.install-source` marker
written by `scripts/install.sh`) shipped in v0.6.0. The safety model is
recorded in [ADR 0035](./adr/0035-native-self-update-safety.md).

- [x] **`aikata update --apply`** ✅ — consumes `internal/install.Detect()`
      and picks the safe path per channel (ADR 0035 D1):
      `install-script` / `github-release` do an in-place **binary swap**
      (download → verify SHA-256 against `checksums.txt` → extract →
      atomic replace); `go-install` is shown the channel-native
      `go install …@latest`; `homebrew` / `npm` are stubbed with an
      actionable "use your package manager" message until those channels
      are real (v0.9.9); Windows and unknown installs get manual-download
      guidance. Verify-before-swap plus a tampered-archive regression
      test back the RCE-surface review (ADR 0035 D2 / `SECURITY.md`).

Native self-update is a convenience, not essential; v0.9.4 keeps it
isolated from the v0.9.3 ecosystem work so neither blocks the other.

---

## v0.9.5 — Plugin command surface + version lockstep ✅ (released 2026-06-01)

Rounds out the Claude Code plugin so it exposes aikata's authoring-growth
verbs, not just the init / sync / doctor / generate lifecycle loop, and
pins the distribution-metadata versioning policy. No CLI behaviour change;
the binary surface is unchanged.

- [x] **`/aikata-new` and `/aikata-enable` plugin slash commands** —
      thin `$ARGUMENTS` wrappers over `aikata new <artifact>` (adr /
      app-icon / mascot) and `aikata enable <capability>` (ui / api /
      tdd / changelog / prompts / memory / monorepo / stack / ai-tool /
      workflow). The plugin previously surfaced neither verb, so creating
      an ADR or enabling a capability had no plugin-level affordance.
- [x] **Both first-party skills document `enable` / `new`** —
      `dist/claude-code/skill/SKILL.md` and `dist/universal-skill/SKILL.md`
      gain a short post-init section (they previously omitted these
      commands entirely).
- [x] **`plugin.json` / `marketplace.json` version lockstep** — both bump
      with every release going forward (CONTRIBUTING § Release flow,
      ARCHITECTURE §6.5); set to `0.9.5`, resolving the v0.9.4 gap where
      the marketplace still showed `0.9.3`.

This is not a distribution *channel* change (no cadence-table row): the
plugin and `npx skills add` surfaces already exist from v0.6 / v0.9.3.

---

## v0.9.6 — Codex native distribution ✅ (released 2026-06-01)

Advances the Codex skill-only native wrapper from ADR 0015's v1.0
deferral now that the platform shape needed by aikata is stable. No Go
CLI behaviour changes: Codex gets a native install path and App metadata
around the same thin CLI wrapper already shipped as the universal skill.

- [x] **Codex App skill-card metadata** —
      `dist/universal-skill/agents/openai.yaml` adds the `aikata` display
      name, concise description, `$aikata` starter prompt, and implicit
      invocation policy without speculative icons, colors, dependencies,
      or branding assets.
- [x] **First-party Codex plugin** — `dist/codex/plugin/` contains
      `.codex-plugin/plugin.json` plus byte-identical copies of the
      universal skill and `agents/openai.yaml`. Root
      `.agents/plugins/marketplace.json` exposes the tracked plugin for
      the self-hosted Codex marketplace flow.
- [x] **Versioned native install guidance** — Codex CLI `0.135.0+`
      prefers `codex plugin marketplace add shigindo-inc/aikata --ref
      v0.9.6` followed by `codex plugin add aikata@aikata`. Direct
      universal-skill discovery remains the fallback for older Codex
      versions and works on CLI `0.125.0`.
- [x] **Release bundles + CI smoke coverage** — preserve
      `aikata-universal-skill.md`; add generated
      `aikata-universal-skill.tar.gz` and `aikata-codex-plugin.tar.gz`
      assets, with archive-entry smoke checks and repository lockstep
      tests.

Cursor and Gemini CLI native wrappers remain v1.0 work. The rationale
and Do-No-Harm analysis are recorded in
[ADR 0036](./adr/0036-codex-native-distribution.md).

---

## v0.9.7 — Adoption mutation boundaries ✅ (released 2026-06-02)

Dogfooding `aikata init` + `aikata doctor --fix` against the existing
`shigindo-flutter-skills` Agent Skills repository exposed an ownership
boundary failure: broad Markdown validation modified third-party
`skills/**` contracts and project-owned prose, while the scaffolded
`.gitignore` block carried stack and editor policies beyond aikata's
shared-context-document purpose. The public `.aikata-proposed/` adoption
fallback was also documented but unimplemented.

[ADR 0037](./adr/0037-tighten-adoption-mutation-boundaries.md) is
**Accepted**; all four maintainer-confirmed boundary fixes ship together.

- [x] **Managed-surface `doctor` default** — implement ADR 0033's
      deferred behavior flip: validate canonical names, known
      aikata-owned document directories (`docs/adr`, `docs/memory`,
      `docs/stacks`, `docs/tasks`, `docs/workflows`, `docs/design`, …),
      and manifest Markdown entries; leave arbitrary repository Markdown
      alone unless broad audit is explicitly selected. `doctor.exclude`
      stays additive (`internal/doctor/scope.go`,
      `Options.Includes`).
- [x] **Explicit broad audit mode** — `aikata doctor --all-markdown`
      restores the whole-repository walk. aikata's own CI gate uses
      `--all-markdown --strict` (the repo has no manifest and ships docs
      outside the managed surface).
- [x] **Minimal `.gitignore` managed block** — keep ADR 0018's
      non-destructive markers but remove future-tool, stack-build,
      editor / OS, coverage, and `*.local` rules. Emit only
      `/.aikata-proposed/`, the AI-tool artifacts, and a minimal
      **always-on** secret baseline (`.env`, `.env.local`). Stale
      `docs.generate_gitignore: false` template guidance removed.
- [x] **`.env.example` opt-in `env` capability** — `aikata enable env` /
      `aikata init --with-env`; schema-v2 `components.env`,
      manifest-tracked, `sync`-preserved. Only the example file is
      opt-in; the `.env` secret ignore stays unconditional. Canonical
      `AGENTS.md` / `README.md` mention `.env.example` as a pattern (no
      dangling hard link).
- [x] **`.aikata-proposed/` contract repair** — non-empty-directory
      `init` without `--force` renders the proposal under
      `.aikata-proposed/` and exits 0; a populated proposal tree is
      refused with `ErrProposalExists`.
- [x] **Regression and migration proof** — byte-identity tests for
      third-party `skills/**` + `CONTRIBUTING.md` under default scope,
      broad-audit coverage, proposal-tree creation/collision tests,
      `.gitignore`/`.env.example` golden updates, and a CHANGELOG
      migration note for projects carrying the previous broad block.

Out of v0.9.7 intentionally:

- Stack-native project scaffolding. Flutter / TypeScript build ignores
  belong to downstream project tooling unless a later opt-in capability
  demonstrates recurring value.
- Auto-detection of third-party Markdown schemas. The managed-surface
  default removes the need to chase external format registries.

---

## v0.9.8 — `.gitignore` managed-block unification + doc hygiene ✅ (released 2026-06-02)

Internal hardening, no new user-facing commands. Two follow-ups from the
v0.9.7 work, shipped together.

[ADR 0038](./adr/0038-unify-gitignore-managed-block-across-init-and-sync.md)
and [ADR 0039](./adr/0039-documentation-hygiene-and-context-budget.md)
are **Accepted**.

- [x] **`.gitignore` unified on the managed block** — `aikata sync` now
      refreshes only the `# >>> aikata managed >>>` block instead of the
      generic 3-way merge, so the file can never grow conflict markers.
      Fresh `aikata init` writes the framed block too (one
      representation), and a pre-0.9.8 markerless file is migrated in
      place on first sync without duplication. Shared
      `managed.IsAppendPath` keeps init and sync in lockstep and guards
      the (b) decision (never managed-append into prose). Closes
      Q-INTEROP-04 (a) and (b).
- [x] **Documentation hygiene rubric + pass** — ADR 0039 records a
      per-file-class policy (prune resolved first-read context each
      release; archive released ROADMAP/CHANGELOG detail by move-not-
      delete; never edit accepted ADR bodies; condense shipped design
      notes to pointers). Applied: eight resolved open-questions removed,
      Phase 1 – v0.8.5 ROADMAP detail archived to
      `docs/roadmap-archive.md`, the v0.9 stabilization note condensed.
      `CONTRIBUTING.md` links the rubric into the release ritual.

Out of v0.9.8 intentionally:

- Managed-append for prose (`CONTRIBUTING.md` / `SECURITY.md`). Rejected
  by ADR 0038; prose stays create-if-missing one-shot scaffolds.
- Archiving `CHANGELOG.md`. Not first-read context; Keep-a-Changelog
  favours completeness (ADR 0039).

---

## v0.10.0 — Collaboration-operation skills ✅ (released 2026-06-02)

**Goal**: turn aikata's document-centered collaboration model into an
agent-usable daily operating loop, not only a CLI bootstrap and maintenance
surface.

This milestone took priority over the convenience-only v0.9.9 package-manager
channels and shipped before them. The previous first-party skill taught agents
when and how to invoke the aikata CLI, but did not teach agents how to operate
an aikata-managed repository during ordinary development: which canonical
documents to read, where newly learned context belongs, how to maintain
working state, or what to check before handoff. That gap weakened the core
product claim that aikata reduces the human cost of maintaining shared context
for humans and AI coding agents. The decision is recorded in
[ADR 0040](./adr/0040-collaboration-operation-skill-split.md).

- [x] **Split the first-party skill surface by responsibility** — rename the
      existing CLI-wrapper responsibility to `aikata-cli`, covering `init`,
      `enable`, `new`, `generate`, `doctor`, `sync`, generated-artifact safety,
      and `doctor --json` parsing.
- [x] **Add the `aikata-context` MVP skill** — trigger for non-trivial work in
      an aikata-managed repository and teach the daily context-maintenance
      loop:
      - choose the relevant canonical documents to read before editing;
      - update `docs/tasks/current.md` at task start, during meaningful
        progress, and at completion when the slot exists;
      - classify durable information into the correct slot: invariant rules
        (`AGENTS.md`), requirements (`SPEC.md`), design decisions
        (`docs/adr/`), long-term facts and preferences (`docs/memory/`), or
        in-flight state (`docs/tasks/current.md`);
      - check documentation impact, verification results, unresolved
        questions, and handoff state before declaring work complete;
      - invoke `aikata-cli` when the operating loop needs `doctor`, `sync`,
        `generate`, or `new adr`.
- [x] **Keep one install surface** — ship `aikata-cli` and `aikata-context`
      together from the existing `aikata` marketplace entry and plugin. Do not
      create separate marketplace entries or plugins for the two skills.
      Preserve the `aikata-` prefix so globally installed skills remain
      attributable and avoid generic-name collisions.
- [x] **Apply the split consistently across first-party distributions** —
      update the canonical universal-skill tree, the Codex plugin copies, and
      the Claude Code standalone / plugin layouts; keep repository tests for
      canonical-copy boundaries and release archives.
- [x] **Record the design in a focused ADR before implementation** — refine
      ADR 0015's thin CLI-wrapper scope without introducing an "aikata agent",
      runtime personality, MCP server, or app integration. The selected model
      remains small composable skills used by the user's existing agent.
- [x] **No backward-compatibility layer for the old single-skill layout** —
      the only existing user is the maintainer, so v0.10.0 may replace the old
      `aikata` skill layout directly. Document the reinstall / migration step,
      but do not retain aliases, duplicate legacy skill files, or transitional
      packaging solely for compatibility.
- [x] **Dogfood the trigger boundary** — verify that ordinary non-trivial
      repository work activates `aikata-context`, pure CLI lifecycle requests
      activate `aikata-cli`, and the skills remain small enough that agents can
      select the correct responsibility without loading an all-purpose
      workflow manual.

Out of v0.10.0 intentionally:

- Splitting `aikata-context` into narrower `aikata-memory`, `aikata-adr`, or
  handoff skills before dogfooding demonstrates that the MVP is too broad.
- A separate aikata runtime agent, MCP server, or app integration.
- Third-party skill catalog management or native wrappers for additional AI
  tools.

---

## v0.13.0 — Doc Map (`docmap`) ✅ (released 2026-06-12)

**Goal**: give every aikata repository an always-current, machine-derived
map of its **document set** — inventory, cross-references, freshness, and a
managed/external split — so an agent or human can orient from one read
instead of several.

This extends the context-budget thesis of
[ADR 0039](./adr/0039-documentation-hygiene-and-context-budget.md). The
artifact is **doc-cartography**, a responsibility distinct from project
mission (`README.md` / `SPEC.md` / `AGENTS.md`) and from the hand-curated
Navigation Matrix (`AGENTS.md` §3, which keeps human judgement). It is built
from documents only — no source code is read — so the stack-agnostic core is
preserved. The decision is
[ADR 0044](./adr/0044-doc-map-derived-artifact.md); the design note is
[`docs/decisions/docmap-design.md`](./decisions/docmap-design.md).

Shipped in PR-sized phases (each ≤ ~400 lines, golden-tested):

- [x] **P0 — prescriptive structure index** ✅ — `docs/layout.md`, the
      single-page, cross-referenced index of the recommended layout (each
      path → role → governing capability/ADR → doctor-managed?). It is the
      **prescriptive** counterpart to the **descriptive** doc map and the
      reference the v0.14.0 migration assistant reconciles against; it does
      not duplicate `ARCHITECTURE.md` §3 (narrative) or `scope.go`
      (enforced set) but indexes them. Doc-only, no code.
- [x] **P1 — scan + data layer** ✅ — `internal/docmap` builds
      `.aikata/docmap.yaml` from frontmatter + link graph + filesystem tree;
      the shared `internal/docmeta` parser is extracted from `doctor` (and
      carries the shared scan-skip baseline) so link parsing cannot drift.
- [x] **P2 — readable view** ✅ — `.aikata/docmap.md`: directory tree + Mermaid
      `doc → doc` graph (adjacency-list degrade past a 40-node threshold) +
      `path → summary` index, with best-effort summary extraction.
- [x] **P3 — `aikata map` + isolated hooks** ✅ — the explicit verb plus a
      final, isolated rebuild step in `init` / `fill` / `enable` / `sync` /
      `generate`, decoupled from per-tool `generate` provider failures and
      suppressed on `--dry-run` / proposal mode.
- [x] **P4 — `doctor` freshness check** ✅ — in-memory rebuild compared by
      `HashContent` (date aligned so only structural drift counts); warning
      on drift, regenerated by `--fix`; `--json` code.
- [x] **P5 — config + optional renderers** ✅ — `docmap.{formats,targets,exclude}`
      and the optional `txt` / `json` / `mmd` outputs.

This is a core CLI behaviour change, not a distribution channel: it adds no
cadence-table row. Residual design questions are tracked under
[Q-DOCMAP](./decisions/open-questions.md#q-docmap).

Out of v0.13.0 intentionally:

- Reading source code / building code dependency graphs (a different tool
  category; breaks the stack-agnostic core).
- Generating an actual documentation website (the Markdown tree satisfies
  the sitemap-like intent).
- Runtime context selection or token counting (scaffold-time tool only).

---

## v0.14.0 — Structure-migration assistant (skill) ✅ (released 2026-06-12)

**Goal**: help a user bring an existing repository **into** the recommended
layout (`docs/layout.md`), with every file move gated on explicit user
approval.

This is the **reconcile** corner of the prescriptive / descriptive /
reconcile triad: `docs/layout.md` (v0.13.0 P0) defines the target, the doc
map (v0.13.0) reports which documents are `external` (off-structure), and
this skill proposes moving them into place. It depends on v0.13.0 shipping
first.

Boundary (to be fixed in its own ADR before implementation):

- Ships as a **skill**, not a CLI verb — keeps the CLI thin, matching the
  skill-guided deprecated-doc cleanup precedent
  ([ADR 0043](./adr/0043-command-wrapper-skill-surface-and-simple-skill-names.md) D5).
- The aikata CLI stays **observation-only**: the doc map *reports*
  off-structure documents; the CLI never moves user files. The agent
  performs moves under the skill, **dry-run by default**, and only after
  explicit per-move (or batched) user confirmation, preferring `git mv` to
  preserve history.
- No silent restructuring, no overwrite, no deletion — relocation
  suggestions only, applied on approval.

- [x] **ADR — migration mutation boundary** ✅ — [ADR 0046](./adr/0046-structure-migration-assistant-boundary.md)
      records the observe → propose → confirm-move contract, the
      observation-only CLI boundary, and why it is a skill, not a verb.
- [x] **`migrate-structure` skill** ✅ — reads the doc map's `external`
      (`managed: false`) set, maps each to its recommended destination per
      `docs/layout.md`, presents a dry-run plan, and applies approved moves
      (`git mv`), then rebuilds the doc map. Shipped as a first-party skill
      across the universal / Claude Code / Codex trees with a thin
      `/aikata:migrate-structure` plugin command (version lockstep bumped).
- [x] **Dogfood** ✅ — distribution canonical-copy and command-wrapper tests
      cover the new skill; the contract forbids any move without explicit
      approval and rebuilds the map after moves.

Out of v0.14.0 intentionally:

- Any automatic / unattended reorganization.
- Rewriting document *contents* (this is relocation only; content edits stay
  with the author / other skills).
- A CLI `aikata migrate` mutate verb (the skill + observation-only CLI is the
  chosen shape unless dogfooding shows it insufficient).

---

## v0.15.0 — Modeling capability + `model-feature` skill ✅ (released 2026-09-06)

**Goal**: close the gap between `SPEC.md` (what/why) and `ARCHITECTURE.md`
(how) where behaviour and structure actually get designed — *who does
what, triggered by what, reaching what outcome, failing how*, and the
entities/fields/invariants that outcome needs — so agents stop
fabricating fields and use cases drift less between design and code.

Dogfooding surfaced this as a concrete, not hypothetical, failure:
without a written record, data-model fields get invented speculatively
with nothing to say whether anything actually needs them. The decision
is recorded in
[ADR 0047](./adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md).

- [x] **`modeling` capability** ✅ — `aikata enable modeling` /
      `aikata init --with-modeling` scaffolds an opt-in document pair:
      `docs/usecases.md` (use-case ledger) and `docs/domain.md` (domain
      model, with a per-field `Related UC` column). The pair is enabled
      together, never separately, and joins the default `aikata doctor`
      managed surface with no new checks — existing frontmatter / link /
      unused-`GLOSSARY.md`-term validation applies for free.
- [x] **`model-feature` skill** ✅ — a new first-party per-feature design
      loop, distinct from `track-context` (different trigger condition):
      write a use case → propagate it into the domain model → fix new
      terms in `GLOSSARY.md` → confirm the bidirectional field/use-case
      check → hand off before implementation. Fires on exactly one
      criterion — externally observable behaviour changes — and is
      advisory, never gating. Shipped across the universal / Claude Code
      (plugin + standalone) / Codex distribution trees.
- [x] **`track-context` hand-off** ✅ — `track-context` now hands off to
      `model-feature` when the work about to start changes externally
      observable behaviour, so the two skills compose instead of
      overlapping.

Out of v0.15.0 intentionally:

- Any new CLI verb — `modeling` is scaffolding-only, reachable through
  the existing `enable` / `init` / `doctor` / `fill` / `map` surface.
- Extending traceability into code or tests — aikata does not read code,
  so such links would rot silently instead of being caught by `doctor`.
- Carrying the loop into implementation — `model-feature` ends at the
  hand-off; implementation stays with the project's normal development
  flow and the wider planning/implementation skill ecosystem.

This is not a distribution-channel change (no cadence-table row): it
ships across the same channels v0.10.0 established.

---

## v0.15.1 — Nested worktrees out of the doc map ✅ (released 2026-09-07)

**Goal**: stop `aikata map` (and the `generate` rebuild) from cataloguing
in-tree git worktrees as if they were the project's document set.

- [x] **Nested VCS skip** ✅ — `docmeta.SkipDir` prunes any directory
      other than the scan root that has its own `.git` (file or
      directory). `aikata map` and `aikata doctor` share the helper.
      Vendor-named trees such as `.claude` are not skipped as a whole;
      committed files under `.claude/skills/` stay on the map. Project-
      specific noise still uses `docmap.exclude` (ADR 0044 D7).

This is not a distribution-channel change (no cadence-table row).

---

## v0.15.2 — `CLAUDE.md` imports `AGENTS.md` ✅ (released 2026-10-05)

**Goal**: make the canonical `AGENTS.md` actually reach Claude Code
sessions, which skip `AGENTS.md` whenever a `CLAUDE.md` exists.

- [x] **`@AGENTS.md` import** ✅ — the generated `CLAUDE.md` (`en` / `ja`)
      imports `AGENTS.md` on its own line instead of only linking to it.
      Tests pin the import line for both languages.

This is not a distribution-channel change (no cadence-table row).

---
