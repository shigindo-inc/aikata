---
project: aikata
status: draft
version: 0.4.0
updated: 2026-11-21
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

## Released milestones — summary

The detailed per-version checklists through v0.15.2
moved to [`docs/roadmap-archive.md`](./docs/roadmap-archive.md) to keep this
roadmap focused on current and upcoming work
([ADR 0039](./docs/adr/0039-documentation-hygiene-and-context-budget.md)). The
binding decisions are in [`docs/adr/`](./docs/adr/); shipped changes in
[CHANGELOG.md](./CHANGELOG.md).

| Milestone | Released | Unlocked |
|---|---|---|
| Phase 1 | — | Repo-as-operational-docs bootstrap |
| v0.1 | 2026-05-22 | MVP: `init` / `generate`, Claude wrapper, memory slot |
| v0.2 | 2026-05-23 | Stack & language axes; `ja` templates |
| v0.2.1 | 2026-05-24 | Onboarding patch |
| v0.3 | 2026-05-24 | Lightweight fast-follow; interactive prompt |
| v0.3.1 | 2026-05-24 | Discoverability & distribution surface |
| v0.3.2 | 2026-05-24 | `.aikata/` config namespace migration |
| v0.4.0 | 2026-05-24 | Authoring ergonomics, first wave |
| v0.4.1 | 2026-05-25 | Authoring ergonomics, second wave |
| v0.4.2 | 2026-05-25 | `aikata update` version check |
| v0.5 | 2026-05-26 | `aikata sync` (manifest 3-way merge) |
| v0.6.0 | 2026-05-26 | Packaging & distribution (partial) |
| v0.6.1 | 2026-05-26 | `sync --rebaseline` regression fix |
| v0.6.2 | 2026-05-28 | ROADMAP template & manifest hygiene |
| v0.6.3 | 2026-05-28 | Sync scope derivation & CLI overrides |
| v0.7.0 | 2026-05-29 | `.aikata/aikata.yaml` schema v2 (ADR 0016) |
| v0.7.1 | 2026-05-29 | Purpose-based CLI split (`enable` / `new`) |
| v0.7.2 | 2026-05-29 | Adoption, repair, managed append (ADR 0018) |
| v0.7.3 | 2026-05-29 | `doctor` scope & exclusion (ADR 0021) |
| v0.7.4 | 2026-05-29 | Retire legacy `.ai/` config fallback (ADR 0020) |
| v0.8.0 | 2026-05-29 | Governance & secret-scan |
| v0.8.1 | 2026-05-29 | Supply-chain signing (cosign, SBOM) |
| v0.8.2 | 2026-05-30 | CLI surface: `--scope` × `--stack` (ADR 0024) |
| v0.8.3 | 2026-05-30 | `sync` divergent-file preservation (ADR 0025) |
| v0.8.4 | 2026-05-31 | Workflow guide opt-in (ADR 0026) |
| v0.8.5 | 2026-05-31 | Verification expectation in templates (ADR 0027) |
| v0.9.0 | 2026-05-31 | Core-concept stabilization (ADR 0028) |
| v0.9.2 | 2026-06-01 | Scope discipline |
| v0.9.3 | 2026-05-31 | Agent-ecosystem distribution |
| v0.9.4 | 2026-06-01 | Native self-update for existing channels |
| v0.9.5 | 2026-06-01 | Plugin command surface + version lockstep |
| v0.9.6 | 2026-06-01 | Codex native distribution |
| v0.9.7 | 2026-06-02 | Adoption mutation boundaries |
| v0.9.8 | 2026-06-02 | `.gitignore` managed-block unification + doc hygiene |
| v0.10.0 | 2026-06-02 | Collaboration-operation skills |
| v0.13.0 | 2026-06-12 | Doc Map (`docmap`) |
| v0.14.0 | 2026-06-12 | Structure-migration assistant (skill) |
| v0.15.0 | 2026-09-06 | Modeling capability + `model-feature` skill |
| v0.15.1 | 2026-09-07 | Nested worktrees out of the doc map |
| v0.15.2 | 2026-10-05 | `CLAUDE.md` imports `AGENTS.md` |
| v0.15.3 | 2026-11-21 | Per-file sync conflict acceptance; stable doc maps |
| v0.16.0 | 2026-11-21 | `doctor` reports what needs action; doc map opt-in |

---

## v0.9.0 — Core-concept stabilization ✅ (released 2026-05-31)

**Goal**: make the existing product easier to understand and trust
before widening its ecosystem surface in v1.0+.

This tranche responds to a maintainer review after v0.8.5: aikata's
value is reducing the human cost of maintaining shared project context
for humans and AI coding agents. Plausible future extensions must not
turn it into an all-purpose template platform whose value is hard to
explain. [ADR 0028](./docs/adr/0028-prioritize-core-concept-stabilization.md)
records the priority rule; the
[v0.9.0 design note](./docs/decisions/v0.9-core-concept-stabilization.md)
records the evidence, target landing point, and follow-up questions.

- [x] **Live-document convergence** ✅ (shipped in v0.9.1) — aligned
      README (status + ADR index), SPEC (`enable`/`new`), ROADMAP
      (v0.8.2/v0.8.3 released), adoption docs, and dogfood config with
      the shipped surface; added a narrow `repolint` check asserting the
      README ADR index covers every `docs/adr/` file (no `doctor`
      behavior change).
- [x] **Default standard-scope audit** ✅ (Q-DESIGN-12 resolved) —
      every default-scaffolded file is confirmed to have a
      distinct role in the shared-context model. `docs/prompts.md`, an
      empty reusable-prompt skeleton, is moved to an opt-in
      `aikata enable prompts` / `--with-prompts` capability (v0.9.2,
      [ADR 0034](./docs/adr/0034-reusable-prompts-opt-in-capability.md));
      the removal is sync-visible but non-destructive (ADR 0019).
- [x] **`doctor` scope follow-up ADR** ✅ — the direction is settled by
      [ADR 0033](./docs/adr/0033-doctor-default-scope-direction.md): a
      managed-document default with an explicit broader audit mode, with
      `doctor.exclude` kept as an escape hatch. The behavior change is
      deliberately deferred to its own scoped step with before/after
      coverage proof, preserving a coherent story for adopted and
      pre-manifest projects (Q-DOCTOR-02 resolved, direction only).
- [x] **Stack-brief simplification** ✅ (Q-DESIGN-13 resolved) — the
      Flutter / TypeScript briefs gain a code-free canonical layout
      convention (v0.9.0,
      [ADR 0029](./docs/adr/0029-stack-brief-layout-convention.md)) and
      are trimmed to standard-aligned guardrails (v0.9.1,
      [ADR 0030](./docs/adr/0030-trim-stack-briefs-to-standard-guardrails.md));
      aikata generates no stack code.
- [x] **v1.0 backlog pruning** ✅ — external stack repositories,
      third-party skill management, new workflow domains, and broad
      native-wrapper proliferation are confirmed **off the critical
      path**: they stay demand-driven, recorded in the v1.0 / v1.x
      sections and "Out-of-scope, indefinitely" rather than as v0.9.x
      commitments. None is pulled forward without concrete dogfooding
      evidence (consistent with "Out of v0.9.0 intentionally" below).

Out of v0.9.0 intentionally:

- Distribution-channel publication remains the separate v0.9.3 / v0.9.4 /
  v0.9.9 lines (ADR 0032).
- New built-in stacks, workflow domains, multi-stack composition,
  external stack repositories, and third-party skill management stay
  deferred unless concrete demand justifies them.

---

## v0.9.9 — Native package-manager channels (deferred until demand)

Third and lowest-priority channel-publication line (ADR 0032 D3). The
convenience-only package-manager channels and their dependent self-update
branches. These fill no open install gap — `curl … | sh` (v0.2.1, with
SHA-256 verification) already covers no-Go install, and `go install`
covers Go users — so they are deferred until concrete demand (a user
asking for `brew install aikata` / `npx aikata`) justifies the standing
maintenance cost. They need out-of-repo maintainer action and were
previously numbered v0.8.x; moved back one minor when the v0.8.x security
& governance hardening line was inserted ahead of them (ADR 0022).

- [ ] **Homebrew tap** (`shigindo-inc/tap/aikata`) published from
      the release workflow. Requires creating the
      `shigindo-inc/homebrew-tap` GitHub repo and adding the
      `HOMEBREW_TAP_GITHUB_TOKEN` secret to this repo.
- [ ] **npm wrapper** for `npx aikata` distribution. Requires npm
      org credentials (`shigindo-inc` scope) configured as
      `NPM_TOKEN`. v0.6.0 shipped `internal/install`'s Source enum
      pre-populated with `npm`; the wrapper just needs to publish.
- [ ] **`homebrew` / `npm` branches of `aikata update --apply`** —
      added to the v0.9.4 self-update surface once items above are real
      enough to test in CI.

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
- [ ] **Plugin / skill distribution beyond Claude and Codex** — publish aikata in
      native distribution shapes where they are stable enough to support:
      Cursor custom modes or rule packs, Gemini CLI extensions, and a VS
      Code extension that wraps the CLI.
      Per-tool scope is driven by H1 dogfooding evidence; the Claude
      plugin (v0.6) and Codex skill plugin (v0.9.6) define the surface
      shape only where the platform concepts line up.
- [ ] Third-party skill / plugin marketplace interop policy. ADR 0015
      resolves first-party wrapper distribution; this remaining item is
      only about whether aikata should ever scaffold manifests for
      curated third-party team skill sets. See
      [Q-ECOSYSTEM-04](./docs/decisions/open-questions.md#q-ecosystem-04--external-skill--plugin-marketplace-interop).

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
- Full cross-channel `aikata update` behavior after v0.4.x: native
  installs can self-update; Homebrew, npm, Go, and OS package-manager
  installs are delegated to their owning package manager or shown as
  actionable commands.

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
| v0.3.0 | ✅ | ✅ | ✅ | — | — | — | — | — |
| v0.3.1 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.3.2 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.4.0 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.4.1 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.4.2 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.5.0 | ✅ | ✅ | ✅ | minimal | — | — | — | — |
| v0.6.0 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.6.1 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.6.2 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.6.3 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.7.x | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.8.x | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.9.0 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.9.2 | ✅ | ✅ | ✅ | minimal | scaffold (manual) | — | — | — |
| v0.9.3 | ✅ | ✅ | ✅ | minimal + universal | marketplace (ready) | — | — | `npx skills add` |
| v0.9.6 | ✅ | ✅ | ✅ | minimal + universal | marketplace (ready) | — | — | Codex plugin |
| v0.9.9 | ✅ | ✅ | ✅ | minimal + universal | marketplace (ready) | `npx aikata` | tap | Codex plugin + `npx skills add` |
| v0.10.0 | ✅ | ✅ | ✅ | `aikata-cli` + `aikata-context` | same marketplace plugin | inherit | inherit | Codex plugin + `npx skills add` |
| v1.0 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Codex / Cursor / Gemini / VS Code |

Plugin / skill scope evolution:

- **v0.3.1** — "Claude knows when to shell out to aikata." One SKILL.md,
  no commands, no agents.
- **v0.6** — adds `/aikata-init`, `/aikata-generate`, `/aikata-doctor`
  slash commands and is installable as a single plugin.
- **v0.7.x** — no new distribution channel; schema / adoption hardening
  only.
- **v0.8.x** — no new distribution channel; security & governance
  hardening of the aikata repository only (ADR 0022).
- **v0.9.0** — stabilizes the core concept and generated-document
  surface. It does not add a distribution channel.
- **v0.9.2** — adds opt-in brand-exploration authoring artifacts. It
  does not add a distribution channel.
- **v0.9.3** — first channel-publication line (ADR 0032): the first-party
  universal skill package for `npx skills add ... --agent universal` plus
  Claude Code marketplace *readiness* (the listing submission stays gated
  on upstream availability + maintainer action). The package wraps the
  aikata CLI; it does not install arbitrary third-party skills.
- **v0.9.4** — adds the native `aikata update --apply` self-update
  *mechanism* for the channels that already exist (install-script /
  go-install / github-release). It adds no new install channel, so it has
  no cadence-table row.
- **v0.9.5** — adds `/aikata-new` and `/aikata-enable` to the existing
  Claude Code plugin and documents `enable` / `new` in both first-party
  skills. It extends the *content* of existing surfaces, not the set of
  install channels, so it has no cadence-table row.
- **v0.9.6** — adds minimal Codex App metadata and a first-party Codex
  skill plugin installable from the repository's self-hosted marketplace.
  The plugin stays byte-identical to the universal skill content and adds
  no MCP server or app integration (ADR 0036).
- **v0.9.7** — tightens adoption mutation boundaries (managed-surface
  `doctor` default + `--all-markdown`, a minimal `.gitignore` block, the
  opt-in `env` capability, and the `.aikata-proposed/` fallback,
  ADR 0037). It changes core behaviour, not the set of install channels,
  so it has no cadence-table row.
- **v0.9.9** — adds the convenience package-manager channels (Homebrew
  tap, `npx aikata`) and the brew / npm branches of `aikata update
  --apply`. Deferred as lowest priority because `curl … | sh` and
  `go install` already cover the install gap (ADR 0032).
- **v0.10.0** — splits the first-party skill responsibility into
  `aikata-cli` and `aikata-context` while keeping one marketplace entry and
  one plugin install. The new context skill turns the canonical-document,
  working-memory, ADR, long-term-memory, verification, and handoff model into
  an agent-usable daily operating loop. Because the maintainer is the only
  existing user, the old single-skill layout is replaced without a
  compatibility layer.
- **v1.0** — extends native wrappers into Cursor, Gemini CLI, and a thin
  VS Code wrapper where each platform has a stable native extension
  surface. Per-tool feature parity is not promised; the promise is "you
  can discover and invoke aikata from your tool's native surface."
  Installing arbitrary third-party skills remains an ecosystem question,
  not part of the core CLI contract yet.

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
