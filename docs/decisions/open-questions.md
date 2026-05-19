---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
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
- **Open part**: how to express **Claude-only** features (Skills, hooks,
  fast mode) without bloating the canonical `AGENTS.md`.
- **Leading position**: per-tool **extension blocks** under
  `templates/ai_tools/<tool>/extensions/`, concatenated after the
  canonical content during `aikata generate`.
- **Unblocks decision**: real `aikata generate` usage in v0.1.

### Q-DESIGN-02 — Commit generated artifacts or `.gitignore` them?

- **Status**: Resolved for v0.1 by
  [ADR 0003 / Consequences](../adr/0003-do-no-harm-policy.md).
  - aikata's own repo: **commit**.
  - User projects via `aikata init`: **gitignore by default**,
    `--no-gitignore-generated` opts out.
- Kept here only to document that the user-project default itself
  remains revisitable based on community feedback.

### Q-DESIGN-03 — Multilingual document structure

- Two-file duplication (`SPEC.md` + `SPEC.ja.md`) is high-maintenance.
- Two candidates:
  - **(a) Translation-table**: canonical `SPEC.md` in one language plus
    a sidecar `i18n/spec.ja.json` of (paragraph hash → translation).
  - **(b) On-demand LLM translation**: ship one language; the user (or
    a future `aikata draft`) translates on demand.
- **Leading**: (b) for v0.x; revisit at v1.x.
- **Unblocks**: a real bilingual user trying both.

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
