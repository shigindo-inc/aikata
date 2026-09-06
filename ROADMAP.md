---
project: aikata
status: draft
version: 0.4.0
updated: 2026-08-14
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

## Released milestones (Phase 1 – v0.8.5) — summary

The detailed per-version checklists for the released Phase 1 – v0.8.5 line
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

The released v0.9.x line is detailed below alongside upcoming work.

---

## v0.9.0 — Core-concept stabilization (pending)

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

## v0.9.2 — Scope discipline ✅ (released 2026-06-01)

Two scope-discipline changes that keep the default `standard` scaffold
lean: add opt-in brand-exploration artifacts, and move the empty
reusable-prompt library off the default surface. Both follow the v0.9.0
core-concept stabilization principle (ADR 0028) — the default scaffold
carries only documents with a distinct, non-latent role; convenience
documents are opt-in.

**Brand-exploration artifacts** ([ADR 0031](./docs/adr/0031-brand-exploration-documents-as-one-off-artifacts.md)).
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

**Reusable-prompts opt-in** ([ADR 0034](./docs/adr/0034-reusable-prompts-opt-in-capability.md),
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
[ADR 0032](./docs/adr/0032-split-channel-publication-by-distribution-value.md)
split out of the former single v0.9.9 line. v0.9.3 ships the distribution
surface that matches aikata's core identity — agent-facing shared-context
tooling, discoverable where its users already work (ADR 0028) — and is the
**prioritized** line. Numeric order is direction, not ship order: v0.9.3
is independent of the still-unshipped v0.9.2 brand-exploration line (the
v0.8.3-before-v0.8.2 precedent).

- [x] **Universal `npx skills add` package** — first-party aikata usage
      guidance at `dist/universal-skill/SKILL.md` per
      [ADR 0015](./docs/adr/0015-first-party-skill-plugin-distribution.md).
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
recorded in [ADR 0035](./docs/adr/0035-native-self-update-safety.md).

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
[ADR 0036](./docs/adr/0036-codex-native-distribution.md).

---

## v0.9.7 — Adoption mutation boundaries ✅ (released 2026-06-02)

Dogfooding `aikata init` + `aikata doctor --fix` against the existing
`shigindo-flutter-skills` Agent Skills repository exposed an ownership
boundary failure: broad Markdown validation modified third-party
`skills/**` contracts and project-owned prose, while the scaffolded
`.gitignore` block carried stack and editor policies beyond aikata's
shared-context-document purpose. The public `.aikata-proposed/` adoption
fallback was also documented but unimplemented.

[ADR 0037](./docs/adr/0037-tighten-adoption-mutation-boundaries.md) is
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

[ADR 0038](./docs/adr/0038-unify-gitignore-managed-block-across-init-and-sync.md)
and [ADR 0039](./docs/adr/0039-documentation-hygiene-and-context-budget.md)
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

## v0.9.9 — Native package-manager channels (pending)

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
[ADR 0040](./docs/adr/0040-collaboration-operation-skill-split.md).

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
[ADR 0039](./docs/adr/0039-documentation-hygiene-and-context-budget.md). The
artifact is **doc-cartography**, a responsibility distinct from project
mission (`README.md` / `SPEC.md` / `AGENTS.md`) and from the hand-curated
Navigation Matrix (`AGENTS.md` §3, which keeps human judgement). It is built
from documents only — no source code is read — so the stack-agnostic core is
preserved. The decision is
[ADR 0044](./docs/adr/0044-doc-map-derived-artifact.md); the design note is
[`docs/decisions/docmap-design.md`](./docs/decisions/docmap-design.md).

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
[Q-DOCMAP](./docs/decisions/open-questions.md#q-docmap).

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
  ([ADR 0043](./docs/adr/0043-command-wrapper-skill-surface-and-simple-skill-names.md) D5).
- The aikata CLI stays **observation-only**: the doc map *reports*
  off-structure documents; the CLI never moves user files. The agent
  performs moves under the skill, **dry-run by default**, and only after
  explicit per-move (or batched) user confirmation, preferring `git mv` to
  preserve history.
- No silent restructuring, no overwrite, no deletion — relocation
  suggestions only, applied on approval.

- [x] **ADR — migration mutation boundary** ✅ — [ADR 0046](./docs/adr/0046-structure-migration-assistant-boundary.md)
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
[ADR 0047](./docs/adr/0047-modeling-capability-and-per-feature-design-loop-boundary.md).

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
