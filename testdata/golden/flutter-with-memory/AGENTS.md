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
1. [`docs/stacks/flutter.md`](./docs/stacks/flutter.md) — stack-specific rules
1. [`docs/memory/`](./docs/memory/) — long-term context (user
   preferences, project notes, references); read at least
   [`feedback.md`](./docs/memory/feedback.md) before non-trivial work
1. [`docs/tasks/current.md`](./docs/tasks/current.md) — short-term working state

## 3. Navigation matrix

| Task type | Files to read first |
|---|---|
| Implement a new widget / screen | `SPEC.md`, `docs/stacks/flutter.md` |
| Add or modify state management | `ARCHITECTURE.md`, `docs/stacks/flutter.md` |
| Fix a bug | `docs/tasks/current.md`, related tests under `test/` |
| Record a design decision | new file under `docs/adr/`, follow [ADR 0001](./docs/adr/0001-record-architecture-decisions.md) |
| Update terminology | `GLOSSARY.md`, then grep for outdated forms |
| Recall user / project context | `docs/memory/{user,feedback,project,reference}.md` |
| Diagnose a hairy problem | `docs/troubleshooting.md` first |

## 4. Hard rules

- **Never commit secrets.** Reference patterns in `.env.example`.
- **Update [`docs/tasks/current.md`](./docs/tasks/current.md)** when you
  start and finish work.
- **Add tests** for new code: widget tests under `test/` mirror `lib/`.
  Run `flutter test` before declaring work complete.
- **Run `flutter analyze`** before commit. Zero warnings is the bar.
- **Respect null safety** — no `!` to silence non-null assertions
  without a comment justifying it.
- **Use [Conventional Commits](https://www.conventionalcommits.org/)** —
  types: feat, fix, docs, style, refactor, test, chore, perf, ci, build.
- **No AI signatures in commits** — describe the change, not the tooling.
- **Record design decisions as ADRs** under [`docs/adr/`](./docs/adr/) —
  follow the format defined in
  [ADR 0001](./docs/adr/0001-record-architecture-decisions.md).
- **Update [GLOSSARY.md](./GLOSSARY.md)** when introducing new domain
  terms.

## 5. Context slots

Keep these slots separate:

| Slot | Lifetime | Use for |
|---|---|---|
| `AGENTS.md` | Long, invariant | Rules that must always hold |
| `docs/memory/` | Long, mutable | Durable facts, preferences, and references |
| `docs/tasks/current.md` | Short, rewriteable | Current in-flight work state |
| `docs/adr/` | Permanent | Design decisions and their rationale |

Do not use `docs/tasks/current.md` as a backlog or task archive. Move
durable information out of it before handoff.

## 6. When stuck

In order of preference:

1. Check [`docs/troubleshooting.md`](./docs/troubleshooting.md) — the
   problem may already be documented.
2. Check [`docs/adr/`](./docs/adr/) for past architectural decisions on
   the topic.
3. Check [`docs/stacks/flutter.md`](./docs/stacks/flutter.md) for
   stack-specific guidance.
4. Write the question down — add it to `docs/tasks/current.md` and
   surface it to the maintainer rather than silently guessing on an
   architecturally relevant point.
