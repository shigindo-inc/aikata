---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: agent
---

# Agent Instructions for aikata Development

> **You are reading the canonical source** for agent behavior in this
> repository. Tool-specific files (`CLAUDE.md`, `.cursor/rules/`, …),
> when present, are **generated** from this file. If they disagree, this
> file wins. See [ADR 0002](./docs/adr/0002-agents-md-as-canonical.md).

> Human contributors: start from
> [`CONTRIBUTING.md`](./CONTRIBUTING.md) for the friendlier quick-start
> and PR checklist. This file is the operational source of truth and
> the input to `aikata generate`; CONTRIBUTING.md is a summary.

---

## 1. Project overview

**aikata** is a lightweight CLI tool that scaffolds AI-readable markdown
documents for projects.

- **What & Why** — read [SPEC.md](./SPEC.md).
- **How** — read [ARCHITECTURE.md](./ARCHITECTURE.md).
- **When** — read [ROADMAP.md](./ROADMAP.md).
- **Terminology** — read [GLOSSARY.md](./GLOSSARY.md).
- **Open design questions** — read
  [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md).

The initial design rationale lives in the canonical operational docs
listed above; pre-v0.1 planning notes are preserved in git history
(see commit `ea48abf`).

---

## 2. Before you start

Read in this order:

1. [README.md](./README.md) — project overview.
2. **This file (AGENTS.md)** — operating rules.
3. [SPEC.md](./SPEC.md) — requirements.
4. [ARCHITECTURE.md](./ARCHITECTURE.md) — technical structure.
5. [GLOSSARY.md](./GLOSSARY.md) — terminology.
6. [`docs/memory/`](./docs/memory/) — long-term memory (user
   preferences, project context, references). Read at least
   [`feedback.md`](./docs/memory/feedback.md) before non-trivial work.
7. `docs/tasks/current.md` — current working memory _(not yet present;
   introduced with the standard preset in v0.1. Until then, surface
   in-flight work via PR descriptions and commit messages)_.

For the design rationale behind any specific decision, consult the
matching ADR under [`docs/adr/`](./docs/adr/).

---

## 3. Navigation matrix

| Task type | Files to read first |
|---|---|
| Add a new preset | `ARCHITECTURE.md`, `internal/presets/`, `templates/presets/` |
| Modify CLI behavior | `SPEC.md` §4, `internal/cli/`, `cmd/aikata/main.go` |
| Add AI-tool support | `ARCHITECTURE.md` §3, `internal/generate/`, `templates/ai_tools/` |
| Fix a bug | `docs/tasks/current.md` (when present), related `*_test.go` |
| Change error/exit codes | `ARCHITECTURE.md` §7, every `internal/cli/*.go` |
| Update terminology | `GLOSSARY.md`, then run a grep for outdated forms |
| Record a design decision | new file under `docs/adr/`, follow [ADR 0001](./docs/adr/0001-record-architecture-decisions.md) |
| Cut a release | [`CONTRIBUTING.md` § Release flow](./CONTRIBUTING.md#release-flow-for-maintainers), [`ARCHITECTURE.md` §6.5](./ARCHITECTURE.md#65-versioning--the-release-ritual) |
| Update planning notes | `docs/decisions/open-questions.md` |
| Recall user / project context | `docs/memory/{user,feedback,project,reference}.md` (see [ADR 0004](./docs/adr/0004-long-term-memory-slot.md)) |
| Touch documents only | the relevant top-level `.md` + `docs/` |

---

## 4. Hard rules

These are not negotiable. Violating any of them blocks a PR.

1. **Read before editing.** Before non-trivial changes, ground
   yourself in [SPEC.md](./SPEC.md), [ARCHITECTURE.md](./ARCHITECTURE.md),
   and the relevant ADR under [`docs/adr/`](./docs/adr/).
2. **Update `docs/tasks/current.md`** when you start and finish work,
   **once that file exists** (it is introduced by the standard preset in
   v0.1). Until then, the PR description and the commit log serve as the
   working-memory of record — keep them current as you progress.
3. **Never commit secrets.** Reference `.env.example` instead. `.env`
   itself is gitignored.
4. **Add tests for new code.** No exceptions in `internal/scaffold`,
   `internal/doctor`, `internal/generate`. Golden tests required for new
   presets.
5. **Use Conventional Commits.** Type ∈ {feat, fix, docs, style,
   refactor, test, chore, perf, ci, build}. Scope is the affected
   package or document.
6. **No AI signatures in commits.** Do not add `Co-Authored-By: Claude`
   or similar lines. Commit messages describe the change, not the
   author tooling.
7. **Run before commit** (once the Go project is initialized):
   `make test && make lint`.
8. **OS-neutral paths.** Always use `filepath.Join`. No hard-coded `/`
   or `\\` in path construction.
9. **No `panic`** in non-test code paths. Return wrapped errors:
   `fmt.Errorf("doing X: %w", err)`.
10. **Top-level minimalism.** Do not add a 9th non-hidden top-level
    file without an ADR justifying it.
11. **Do-No-Harm Policy.** Any optional feature must satisfy
    [ADR 0003](./docs/adr/0003-do-no-harm-policy.md). Demonstrate
    compliance in your PR description.
12. **Canonical source rule.** Edit `AGENTS.md`, **not** generated
    tool-specific files. Generated files will be overwritten by
    `aikata generate`.

---

## 4a. Rules vs. memory vs. working state

Three slots, three lifetimes. **Do not blur them.**

- **Rules** (this file, `AGENTS.md`): invariant constraints — always
  true regardless of context. Edits go through a PR + ADR.
- **Long-term memory** ([`docs/memory/`](./docs/memory/)): mutable
  facts and preferences accumulated over time. Append entries with a
  `[YYYY-MM-DD]` prefix; supersede in place, do not delete. See
  [ADR 0004](./docs/adr/0004-long-term-memory-slot.md).
- **Working memory** (`docs/tasks/current.md`, planned v0.1): current
  in-flight task state. Rewrite freely.

When a memory entry and a Hard Rule disagree, **the Hard Rule wins**.

---

## 5. Code style

- `panic` forbidden in production paths.
- Exported functions / types require godoc comments.
- Wrap errors with context using `%w`.
- Keep external dependencies minimal; new ones need a CHANGELOG entry
  and a rationale (see [ARCHITECTURE.md §10](./ARCHITECTURE.md#10-dependencies)).
- Prefer `log/slog` (structured) over the deprecated `log` package.

---

## 6. Document style

- Frontmatter keys (required on every `.md`): `project`, `status`,
  `version`, `updated`, `audience`.
- `audience: agent` only on AGENTS.md; everything else is
  `[human, agent]`.
- Default document language is **English**. Japanese variants (when
  added in v0.2) live alongside the English version with a `.ja.md`
  suffix.
- Top-level `.md` files target ~200 lines max. Move detail into
  `docs/` if you overflow.

---

## 7. Commit hygiene

- One logical change per commit.
- One feature per PR; split if the diff exceeds ~400 lines of meaningful
  change.
- A change that creates or modifies an ADR must reference the ADR
  filename in the commit message.

---

## 8. When stuck

In order of preference:

1. Check [`docs/decisions/open-questions.md`](./docs/decisions/open-questions.md)
   — your question may already be tracked.
2. Check [`docs/adr/`](./docs/adr/) for past decisions on the topic.
3. Skim related ADRs and `git log` for the design intent behind the
   surrounding code.
4. If still unclear, **write the question down**: add it to
   `open-questions.md` and ask the human maintainer. Do not silently
   guess on an architecturally relevant point.

---

## 9. What this file is not

- Not a substitute for `SPEC.md` / `ARCHITECTURE.md`. Keep `AGENTS.md`
  focused on **how to operate** in this repo, not on requirements or
  design.
- Not the place for Claude-specific or Cursor-specific tips. Those
  belong under `templates/ai_tools/<tool>/extensions/` (planned v0.1).
- Not a stable API. Rules may evolve; treat this file as living
  documentation under the same review process as code.
