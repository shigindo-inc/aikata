---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-20
audience: agent
---

# Agent Instructions for aikata Development

> **You are reading the canonical source** for agent behavior in this
> repository. Tool-specific files (`CLAUDE.md`, `.cursor/rules/`, …),
> when present, are **generated** from this file. If they disagree, this
> file wins. See [ADR 0002](./docs/adr/0002-agents-md-as-canonical.md).

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

The historical planning record (frozen, do not edit) is under
[`docs/origin/`](./docs/origin/).

---

## 2. Before you start

Read in this order:

1. [README.md](./README.md) — project overview.
2. **This file (AGENTS.md)** — operating rules.
3. [SPEC.md](./SPEC.md) — requirements.
4. [ARCHITECTURE.md](./ARCHITECTURE.md) — technical structure.
5. [GLOSSARY.md](./GLOSSARY.md) — terminology.
6. `docs/tasks/current.md` — current working memory _(not yet present;
   introduced with the standard preset in v0.1. Until then, surface
   in-flight work via PR descriptions and commit messages)_.

For full context on non-trivial changes, additionally read
[`docs/origin/initial-design.md`](./docs/origin/initial-design.md) and
[`docs/origin/initial-setup.md`](./docs/origin/initial-setup.md).

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
| Update planning notes | `docs/decisions/open-questions.md` |
| Touch documents only | the relevant top-level `.md` + `docs/` |

---

## 4. Hard rules

These are not negotiable. Violating any of them blocks a PR.

1. **Read before editing.** Before non-trivial changes, read
   [`docs/origin/initial-design.md`](./docs/origin/initial-design.md). It
   has full context not duplicated in the operational docs.
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
3. Search the origin docs (`docs/origin/`) for design intent.
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
