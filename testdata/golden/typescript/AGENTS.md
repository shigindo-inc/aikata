---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: agent
---

# Agent Instructions for samplekata

## 1. Project overview

See [SPEC.md](./SPEC.md) for the what and why. Technical structure lives
in [ARCHITECTURE.md](./ARCHITECTURE.md). Domain vocabulary is pinned in
[GLOSSARY.md](./GLOSSARY.md).

## 2. Before you start

Read in this order (markdown auto-numbers; the `1.` prefix is intentional):

1. [README.md](./README.md) — overview
1. **This file (AGENTS.md)** — operating rules
1. [SPEC.md](./SPEC.md) — requirements
1. [ARCHITECTURE.md](./ARCHITECTURE.md) — technical structure
1. [GLOSSARY.md](./GLOSSARY.md) — terminology
1. [`docs/stacks/typescript.md`](./docs/stacks/typescript.md) — stack-specific rules
1. [`docs/tasks/current.md`](./docs/tasks/current.md) — current working memory

## 3. Navigation matrix

| Task type | Files to read first |
|---|---|
| Implement a new module / endpoint | `SPEC.md`, `docs/stacks/typescript.md` |
| Modify the public API surface | `SPEC.md`, `ARCHITECTURE.md`, related tests |
| Fix a bug | `docs/tasks/current.md`, related tests |
| Record a design decision | new file under `docs/adr/`, follow [ADR 0001](./docs/adr/0001-record-architecture-decisions.md) |
| Update terminology | `GLOSSARY.md`, then grep for outdated forms |
| Diagnose a hairy problem | `docs/troubleshooting.md` first |

## 4. Hard rules

- **Never commit secrets.** Reference patterns in [`.env.example`](./.env.example).
- **Update [`docs/tasks/current.md`](./docs/tasks/current.md)** when you
  start and finish work.
- **Add tests** for new code: tests mirror `src/` one-to-one. Run the
  test suite (`vitest` / `jest` — captured in
  `docs/stacks/typescript.md`) before declaring work complete.
- **Type-check passes** before commit (`tsc --noEmit` or equivalent).
  Zero errors is the bar.
- **Lint passes with zero warnings** (`eslint .` or equivalent).
- **No `any`** without an inline comment justifying the escape hatch.
  Prefer `unknown` and narrow.
- **Use [Conventional Commits](https://www.conventionalcommits.org/)** —
  types: feat, fix, docs, style, refactor, test, chore, perf, ci, build.
- **No AI signatures in commits** — describe the change, not the tooling.
- **Record design decisions as ADRs** under [`docs/adr/`](./docs/adr/) —
  follow the format defined in
  [ADR 0001](./docs/adr/0001-record-architecture-decisions.md).
- **Update [GLOSSARY.md](./GLOSSARY.md)** when introducing new domain
  terms.

## 5. When stuck

In order of preference:

1. Check [`docs/troubleshooting.md`](./docs/troubleshooting.md) — the
   problem may already be documented.
2. Check [`docs/adr/`](./docs/adr/) for past architectural decisions on
   the topic.
3. Check [`docs/stacks/typescript.md`](./docs/stacks/typescript.md) for
   stack-specific guidance.
4. Write the question down — add it to `docs/tasks/current.md` and
   surface it to the maintainer rather than silently guessing on an
   architecturally relevant point.
