---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-28
audience: [human, agent]
memory_type: feedback
---

# Memory — `feedback`

Continuing instructions from the user: corrections delivered in past
sessions and approaches the user has explicitly validated. See
[`README.md`](./README.md) for write rules. **When this file and
[`../../AGENTS.md`](../../AGENTS.md) conflict, AGENTS.md wins.**

---

## Commit & PR hygiene

- [2026-05-21] **No AI signatures in commits.** Do not append
  `Co-Authored-By: Claude` or similar lines. The commit describes the
  change, not the tooling. (Recorded as Hard Rule in
  [AGENTS.md §4-6](../../AGENTS.md#4-hard-rules); restated here so
  agents see it on first read of memory.)
- [2026-05-21] **Conventional Commits required.** Type ∈
  {feat, fix, docs, style, refactor, test, chore, perf, ci, build}.
  Scope is the affected package or document. The user reviewed and
  accepted the commit messages produced in Phase 1 (`docs:`,
  `docs(agents):`, etc.) without modification — that style is
  validated.
- [2026-05-21] **`main` direct commits stop at Phase 1.** From Phase 2
  onward, every change goes through a `feat/*` branch and a PR. The
  user confirmed B4 option (c).

## License / copyright

- [2026-05-21] **LICENSE Copyright holder is the real name "Satoshi
  Minami"**, not the placeholder "aikata contributors" and not the
  org name. If the org switches to a different umbrella later, append
  ("Satoshi Minami and aikata contributors") rather than replace.

## Documentation style

- [2026-05-21] The user explicitly requested **bilingual operation**:
  agent ↔ user conversation in **Japanese**, generated documents in
  **English** by default. Do not switch the documents to Japanese
  without an explicit `--lang ja` request.
- [2026-05-21] When asking questions, prefer **structured tables /
  numbered options** over free-form prose. The user has consistently
  responded to A/B/C-style choices faster.
- [2026-05-21] **Frontmatter is required on every markdown** in the
  repo: `project`, `status`, `version`, `updated`, `audience` (plus
  `memory_type` under `docs/memory/`). Validated by the Phase 1
  acceptance checklist.

## Design discipline

- [2026-05-21] **Top-level minimalism — 8 non-hidden files max.** Any
  9th file (e.g. `CLAUDE.md` wrapper) requires an ADR. The user
  approved this discipline by accepting ADR 0002's exception note for
  the Claude wrapper.
- [2026-05-21] **Do-No-Harm Policy is binding** (ADR 0003). Every new
  optional feature must demonstrate compliance in its PR description.
- [2026-05-21] **Agents must flag concerns before substantive work**:
  "重要な検討事項とか何か懸念等あればそれについてもユーザーに質問しつつ
  対応したい" — translate to: pause and surface trade-offs, do not
  silently choose.
- [2026-05-28] **Do not silently delete aikata files/folders.** Even
  when a command narrows scope or a file is no longer part of the
  current generated surface, aikata must not remove it without an
  explicit, previewable cleanup operation.
- [2026-05-28] **No compatibility aliases for pre-v1 CLI cleanup unless
  explicitly requested.** The user prefers a smaller, clearer command
  surface over preserving pre-release `aikata add` spellings.

## Scope discipline

- [2026-05-21] **YAGNI on optional sub-structure**: `docs/memory/`
  stays flat (no subdirectories) in v1.x. Avoid premature
  hierarchies. The user accepted ADR 0004 §1 verbatim.
- [2026-05-21] **`aikata generate` for AI-tool memory channels (option δ)
  is deferred**. The user said "これはどういうものか理解してないので
  おいおいかも". Do not start work on it without re-confirmation.

## Process tools

- [2026-05-21] Use `TaskCreate` / `TaskUpdate` to track multi-step
  work, marking each step `in_progress` before doing it and
  `completed` immediately after. The user did not push back on this
  pattern in Phase 1.
- [2026-05-21] Use the **plan mode** workflow for non-trivial work:
  plan file → user approval (`ExitPlanMode`) → implementation. The
  user has invoked plan mode multiple times by themselves; treat it
  as the default for any new design decision.
