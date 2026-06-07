---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-02
audience: [human, agent]
---

# Open Questions

Decisions that are **not yet** made. When one of these is resolved, the
resolution moves to an ADR under [`docs/adr/`](../adr/) and the entry
here is removed.

Each entry has: an ID, the question, today's leading position (if any),
what unblocks a decision, and the latest update date.

---

## Q-DESIGN

### Q-DESIGN-01 — How do `AGENTS.md` and tool-specific files relate at v1.0+?

- **Status**: Partially answered by [ADR 0002](../adr/0002-agents-md-as-canonical.md).
  Task 7 shipped the first generator (Claude) — `CLAUDE.md` is now a
  thin generated wrapper around `AGENTS.md`. The wrapper is
  intentionally minimal in v0.1.
- **Open part**: how to express **Claude-only** features (Skills,
  hooks, fast-mode references) without bloating the canonical
  `AGENTS.md`.
- **Leading position**: per-tool **extension blocks** under
  `templates/ai_tools/<tool>/extensions/`, concatenated after the
  canonical content during `aikata generate`. v0.2 candidate.
- **Unblocks decision**: a real Claude-only feature that needs to
  ship in this repo (none yet — dogfooding hasn't surfaced one).

### Q-DESIGN-02 — Commit generated artifacts or `.gitignore` them?

- **Status**: Resolved for v0.1 by
  [ADR 0003 / Consequences](../adr/0003-do-no-harm-policy.md).
  - aikata's own repo: **commit**.
  - User projects via `aikata init`: generated AI-tool artifacts are
    **gitignored by default** via the ADR 0018 managed-append block.
    A project that wants to commit them removes the lines from the
    aikata-owned block or lists `.gitignore` under `sync.own`
    ([ADR 0025](../adr/0025-sync-divergent-file-preservation.md)); the
    once-planned `--no-gitignore-generated` flag was never shipped and
    the inert `docs.generate_gitignore` field was removed in v0.8.3.
- Kept here only to document that the user-project default itself
  remains revisitable based on community feedback.

### Q-DESIGN-03 — Multilingual document structure

- **Status**: Partially answered by
  [ADR 0006](../adr/0006-locale-and-japanese-documentation-policy.md).
  aikata's own repository keeps English canonical docs and adds a
  focused Japanese access layer under `docs/`; generated project
  templates remain first-class for both `en` and `ja`.
- **Open part**: whether v1.x should add a bilingual same-file mode
  (Japanese for humans + English for LLMs in one canonical document),
  a translation-table sidecar, or an on-demand translation workflow.
- **Leading**: keep bilingual same-file mode deferred through v0.x.
  Validate the current split with real Japanese users first:
  English repo docs, `docs/*.ja.md` access docs, and `--lang ja`
  generated templates.
- **Unblocks decision**: at least one real project that needs both
  Japanese human-facing prose and English LLM-facing prose in the same
  canonical file.
- **Updated**: 2026-05-24.

### Q-DESIGN-05 — Ownership of `docs/tasks/current.md`

- Agents rewrite it constantly. Humans also edit. Concurrent edits
  produce conflicts whose resolution is non-trivial.
- **Leading**: format the file as a thin sectioned outline (Status /
  Next / Notes); document a "human edits go to Notes" convention; let
  agents rewrite Status and Next freely.
- **Partial mechanism**: a project that has fully taken over the file
  can list it under `sync.own` so `aikata sync` never overwrites or
  conflict-marks it ([ADR 0025](../adr/0025-sync-divergent-file-preservation.md)
  D2). That settles the *sync* side; the open part remains the
  human-vs-agent in-file editing convention above.
- **2026-06-07 direction**: keep `docs/tasks/current.md` as the single
  mandatory entry point for v0.11.1. Clarify that `docs/tasks/` is
  short-lived working state, not a backlog or task archive.
- **Deferred option**: allow opt-in per-work files such as
  `docs/tasks/<slug>.md` for long-running or parallel work, with
  `current.md` remaining the required index / active pointer. Avoid
  `next.md` or numbered `next_XX.md` because they obscure which file an
  agent must read first.
- **Unblocks**: real dogfooding evidence.
- **Updated**: 2026-06-07.

### Q-DESIGN-07 — Memory generate-projection across AI-tool memory channels

- **Status**: Open. Scope (γ) — the canonical `docs/memory/` slot —
  is resolved by [ADR 0004](../adr/0004-long-term-memory-slot.md) and
  **shipped as `aikata init --with-memory` in v0.1** (Task 5A). This
  entry captures the still-**deferred scope (δ)**: projecting memory
  into tool-specific memory channels via `aikata generate`.
- **Question**: How should `aikata generate` mirror
  `docs/memory/{user,feedback,project,reference}.md` into:
  - Claude Code's `.claude/memory/` typed files (one `<type>_<slug>.md`
    per entry), or appended sections in `CLAUDE.md`?
  - Cursor's `.cursor/rules/long-term/` (or wherever Cursor consolidates
    persistent context, which is in flux)?
  - Codex / Gemini CLI / Copilot / Windsurf — what do they accept as
    long-term memory, if anything?
- **Why deferred**: Each target tool's memory mechanism is moving fast
  (Q1 2026). Locking a generate format too early creates churn. Better
  to live with the canonical slot for one release cycle and observe.
- **Unblocks**: Investigation owner (likely the v0.3 owner) reads one
  or two real session logs from each tool, drafts a generate strategy
  in a new ADR, and ships behind `aikata generate --memory` in v0.4
  if the strategy proves stable.
- **Updated**: 2026-05-21.

### Q-DESIGN-08 — Cursor `.cursor/rules/` glob-scoped detailed generation

- **Status**: Open. Scope (basic) — single `.cursor/rules/main.mdc`
  wrapper with `alwaysApply: true` — is resolved by
  [ADR 0005](../adr/0005-cursor-codex-pass-through.md) and **shipped
  in v0.2** (Task 9). This entry captures the still-**deferred scope**:
  splitting rules into multiple `.mdc` files with `globs` for
  per-file-pattern scoping.
- **Question**: Should `aikata generate` emit one `.mdc` per rule
  category (e.g. `testing.mdc` for test files, `frontend.mdc` for
  UI code, etc.), each with appropriate `globs:` frontmatter, instead
  of the single `main.mdc` wrapper?
- **Why deferred**: We have no field experience yet with the wrapper.
  Splitting upfront risks designing the wrong taxonomy. Better to
  ship the wrapper, observe how users (and dogfooding) actually want
  to scope rules, then split.
- **Unblocks**: Two or three real-world cases where the
  single-wrapper approach loses information that `globs` could
  preserve.
- **Updated**: 2026-05-22.

### Q-DESIGN-06 — Stack-agnostic core boundary

- The Flutter preset must not require core changes. Where does the
  boundary live: pure templates? embedded Go code? a plugin interface?
- **Leading**: pure templates + a small per-stack YAML manifest (no
  code). Plugin interface deferred to v1.x.
- **Unblocks**: writing the Flutter preset.

### Q-DESIGN-10 — Post-init command taxonomy

- **Status**: Partially resolved by
  [ADR 0017](../adr/0017-post-init-command-taxonomy.md) in v0.7.1.
  `aikata enable <capability>` (memory, ui, api, tdd, changelog,
  monorepo, stack `<name>`, ai-tool `<name>`) and
  `aikata new <artifact>` (adr) ship; the pre-v0.7.1 `aikata add`
  parent is removed without aliases. `aikata expand <tier>` is
  intentionally deferred until `extended` exists or a real project
  surfaces a minimal → standard growth need.
- **Open part**: the `aikata expand` semantics for projects that
  were init'd with `aikata init --preset minimal` (which writes no
  `.aikata/aikata.yaml` today). When `extended` becomes a real tier
  target, the verb gains a second use case and the open semantics
  are easier to settle.
- **Unblocks**: v1.0 surface freeze once `expand` is decided one way
  or the other.
- **Updated**: 2026-05-29.

_(Q-DESIGN-11 — verification gate in the standard template — was
resolved and moved to
[ADR 0027](../adr/0027-verification-expectation-in-generated-templates.md),
Accepted 2026-05-31.)_

---

## Q-PROMPT

### Q-PROMPT-01 — Optional-feature and OSS-intent prompts

- **Status**: v0.3 brought the interactive `aikata init` prompt to
  parity with the supported non-interactive flags (project name,
  preset, language, AI tools, long-term memory). The optional-feature
  questions (UI / API / TDD / changelog) and the OSS-intent question
  remain unasked.
- **Open part**: those questions should land in lockstep with their
  matching non-interactive flags (`--with-ui`, `--with-api`,
  `--with-tdd`, `--with-changelog`, `--preset extended`) so the prompt
  never drifts away from the flag surface again.
- **Leading**: add each question in the same change that introduces
  its flag, with `--no-interactive` and explicit-flag detection
  skipping it. Target v0.4 (see [ROADMAP.md](../../ROADMAP.md)).
- **Unblocks**: v0.4 authoring-ergonomics work (the new `--with-*`
  flags and `aikata add` second wave).

---

## Q-NAME

### Q-NAME-01 — npm / PyPI / domain availability

- **Status**: GitHub + npm + PyPI **availability confirmed** at the
  time of this writing. No known competing OSS named `aikata`.
- **Open part**: domain (`aikata.dev` / `aikata.io`) registration and
  long-term branding (logo, pronunciation guidance for non-Japanese
  speakers).
- **Unblocks**: v0.4 distribution work.

### Q-NAME-02 — Pronunciation / romanization for non-Japanese users

- Two-syllable "aikata" (アイカタ); options for guidance text in
  `README.md`.
- **Leading**: a short `Pronunciation` paragraph in `README.md`
  ("ai-kata, /aɪˈkɑːtə/").
- **Unblocks**: nothing — purely an editorial task.

---

## Q-INTEROP

### Q-INTEROP-01 — Importing ai-rulez configs

- Should aikata read an existing `ai-rulez.yaml` and translate it into
  an aikata project?
- **Leading**: **no**, in v0.x. Different concept (rules vs documents)
  → translation would be lossy in the wrong direction. Document this
  position in the FAQ instead.

### Q-INTEROP-02 — `agents.md` open-spec conformance level

- Strict (every claim in the spec) or pragmatic (cover the parts users
  actually need)?
- **Leading**: pragmatic for v0.1; aim for strict conformance by v1.0.

### Q-INTEROP-03 — Adopting a user's existing repository

- A user with hand-written canonical docs (a bespoke `AGENTS.md`, etc.)
  should be able to bring a repo under aikata without losing their work.
- **Resolved** by [ADR 0042](../adr/0042-fill-command-for-canonical-document-completion.md):
  the option-free `aikata fill` writes only the **missing** canonical
  documents and never overwrites existing files, adopting an unmanaged
  repo (default scope `standard`) or topping up a managed one. `aikata
  init` in a non-empty dir still emits the full proposal to
  `.aikata-proposed/` (SPEC §4.1) for users who prefer a manual merge.
- **Dropped**: the broader `aikata adopt <file>` idea that would *parse* a
  hand-written `CLAUDE.md` into the `AGENTS.md` skeleton. It was unproven
  demand; write-missing covers the real pain, and a parse step can be
  revisited if a concrete need appears.

---

## Q-ECOSYSTEM

### Q-ECOSYSTEM-01 — External preset distribution

- Pattern: `aikata add stack github.com/user/preset-foo`?
- **Leading**: yes for v1.0. Use Go module-style references + a
  signing / trust check.
- **Unblocks**: v1.0 release planning.

### Q-ECOSYSTEM-02 — Community template review

- Who curates community templates? What is the security model?
- **Leading**: a `aikata.dev/templates` index page, no automatic
  trust — users must explicitly fetch by full path. No central
  approval.

### Q-ECOSYSTEM-04 — External skill / plugin marketplace interop

- **Status**: Partially resolved by
  [ADR 0015](../adr/0015-first-party-skill-plugin-distribution.md).
  As of 2026-05-28 the surrounding ecosystem has several overlapping
  distribution shapes:
  - `npx skills add ...` from the open agent skills ecosystem installs
    `SKILL.md` packages into many agent-specific locations, including
    shared `.agents/skills/` layouts for "universal" agents.
  - Claude Code has first-party Skills, Plugins, and plugin marketplaces
    (`.claude/skills/`, plugin-bundled skills, and marketplace install
    flows).
  - Codex has Skills plus plugin manifests under `.codex-plugin/` and
    marketplace catalogs under `.agents/plugins/marketplace.json`.
  - Gemini CLI has extensions via `gemini-extension.json`, bundling
    commands, MCP servers, hooks, sub-agents, themes, and agent skills.
- **Resolved part**: aikata will ship first-party wrappers that teach
  agents how to use the aikata CLI. `npx skills add ... --agent
  universal` support shipped in v0.9.3, with source artifacts under
  `dist/universal-skill/` (the prioritized first line of the
  channel-publication split,
  [ADR 0032](../adr/0032-split-channel-publication-by-distribution-value.md)).
  Codex native skill-plugin distribution ships in v0.9.6 under
  `dist/codex/plugin/`, refining ADR 0015's original v1.0 deferral now
  that the required Codex shape is stable
  ([ADR 0036](../adr/0036-codex-native-distribution.md)). aikata will not
  be distributed as an "aikata agent" personality.
- **Open part**: should aikata ever scaffold / manage curated
  third-party skill and plugin manifests for teams?
- **Leading**: no third-party skill package management in v0.x. For
  third-party skills, document recommended commands or manifest
  locations, but avoid installing remote code until a trust, pinning,
  update, and removal model is captured in a future ADR.
- **Unblocks**: v1.0 native wrapper distribution beyond Claude and
  Codex, any
  future `aikata add skill-source ...` or team marketplace manifest
  feature, and memory projection decisions that might depend on native
  skill/plugin packaging.
- **Updated**: 2026-06-01.

---

## Q-DOCTOR

### Q-DOCTOR-02 — Should `doctor` default to aikata-managed documents?

- **Status**: **Resolved (direction only; behavior deferred)**. The
  direction is settled by
  [ADR 0033](../adr/0033-doctor-default-scope-direction.md): the default
  walk *should* validate primarily the document surface aikata manages,
  with an explicit broader audit mode for additional Markdown paths, and
  `doctor.exclude` ([ADR 0021](../adr/0021-doctor-scope-and-exclusion.md))
  is kept as an escape hatch. The **behavior change is deliberately
  deferred** to its own scoped step with before/after coverage proof, so
  pre-v1.0 adopted and pre-manifest projects do not lose useful validation
  silently and the binding `--strict` CI gate is not narrowed without
  warning (ADR 0033 D2 / D3).
- **Follow-up to**:
  [ADR 0021](../adr/0021-doctor-scope-and-exclusion.md) and the v0.9
  core-concept stabilization review
  ([ADR 0028](../adr/0028-prioritize-core-concept-stabilization.md)).
- **Resolved**: 2026-06-01. Kept as a pointer; the behavior-change step
  inherits the preconditions in ADR 0033 D3.

---

## Q-HYPOTHESES (to validate via dogfooding)

These map to [SPEC.md §7](../../SPEC.md#7-hypotheses-to-validate).

| ID | Hypothesis | Success criterion |
|---|---|---|
| H1 | Multi-AI-tool users prefer aikata over ai-rulez | ≥ 10 self-reported converts in v0.1–v0.3 window |
| H2 | Document-centered > rules-centered scaffolds | ≥ 5 external PRs / issues praising the doc-first model |
| H3 | Flutter devs want this | Flutter preset gets used in ≥ 3 public repos by v0.3 |
| H4 | Japanese OSS identity is a strength | Japanese-language issues / posts referencing aikata by v0.3 |

Each will be revisited at the v0.3 → v0.4 transition.

---

## Q-DIFFERENTIATION (continuous)

These are not blockers but recurring concerns to keep in view.

- ai-rulez ships a "lite mode" → reinforce the **document-centered**
  framing rather than competing on feature count.
- The `agents.md` spec gains universal adoption → aikata's value
  becomes the **scaffolding + presets**, not the file format.
- AI-coding fashion churn → keep the canonical layer stable, push
  churn into `internal/generate`.
