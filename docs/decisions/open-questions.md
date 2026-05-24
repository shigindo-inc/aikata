---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
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
  - User projects via `aikata init`: **gitignore by default**,
    `--no-gitignore-generated` opts out.
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

### Q-DESIGN-04 — Preset composition (`--preset flutter --preset oss`)

- Should presets compose, layer, or be mutually exclusive?
- **Leading**: presets are **feature-flag bundles**. Two presets compose
  by set-union of flags + last-write-wins on template overrides.
- **Unblocks**: a second preset (`flutter`) and a meaningful use case
  for composition.

### Q-DESIGN-05 — Ownership of `docs/tasks/current.md`

- Agents rewrite it constantly. Humans also edit. Concurrent edits
  produce conflicts whose resolution is non-trivial.
- **Leading**: format the file as a thin sectioned outline (Status /
  Next / Notes); document a "human edits go to Notes" convention; let
  agents rewrite Status and Next freely.
- **Unblocks**: real dogfooding evidence.

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

### Q-INTEROP-03 — Adopting a user's existing `CLAUDE.md`

- A user with a hand-written `CLAUDE.md` running `aikata init` should
  not lose their work.
- **Leading**: `aikata init` in a non-empty dir emits proposals to
  `.aikata-proposed/` (already in [SPEC §4.1](../../SPEC.md#41-aikata-init-name));
  an `aikata adopt <file>` command (v0.4?) could parse the user's file
  into the canonical `AGENTS.md` skeleton.

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

### Q-ECOSYSTEM-03 — License

- **Resolved**: MIT, recorded in
  [`LICENSE`](../../LICENSE) and [SPEC.md](../../SPEC.md).
- Kept as a back-reference; will be removed in a future cleanup.

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
