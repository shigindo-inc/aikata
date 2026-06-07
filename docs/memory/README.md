---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# `docs/memory/` — Long-term Agent Memory

This directory holds **long-term memory** for agents collaborating on
this repository. It is the third class of agent-facing artifact, distinct
from rules and short-term working state.

| Artifact | Lifetime | Author | Purpose |
|---|---|---|---|
| [`AGENTS.md`](../../AGENTS.md) | Long, invariant | Human (canonical) | **Rules** — what must always be true |
| `docs/memory/*` | Long, mutable | Agent + Human | **Memory** — what we have learned |
| [`docs/tasks/current.md`](../tasks/current.md) (planned v0.1) | Short | Agent (frequent) | **Working state** — current in-flight work state |

For the full design rationale, read
[ADR 0004 — Long-term Memory Slot](../adr/0004-long-term-memory-slot.md).

---

## Files in this directory

| File | Holds |
|---|---|
| [`user.md`](./user.md) | Profile / role / knowledge / preferences of the **user** |
| [`feedback.md`](./feedback.md) | Continuing **instructions** from the user (corrections and validated approaches) |
| [`project.md`](./project.md) | **Ongoing context** of the project not derivable from code or git log |
| [`reference.md`](./reference.md) | Pointers to **external systems** (URLs, dashboards, channels) |

These four types match the convention used by Claude Code auto-memory
and the superpowers plugin. Do not invent new types in v1.0.

---

## Write rules

1. **Frontmatter is mandatory.** Every memory file must have the
   standard frontmatter plus `memory_type: user|feedback|project|reference`
   matching the filename.

2. **Append, do not rewrite.** New entries go at the bottom of the
   relevant file as a top-level bullet. Do not reorder or reflow
   existing entries.

3. **Date every entry.** Begin each entry with `[YYYY-MM-DD]` so the
   age is grep-friendly and `aikata doctor` can flag stale items.

4. **Supersede in place.** When an entry is no longer accurate, do not
   delete it. Append `**(superseded YYYY-MM-DD: <reason>)**` at the
   end of the line. Keeps history auditable.

5. **One claim per bullet.** Compound entries split poorly when
   superseded.

6. **Cross-link liberally.** If a memory item depends on an ADR, an
   open question, or a rule in `AGENTS.md`, link to it inline.

---

## Boundary with `AGENTS.md`

When in doubt, ask:

> "Must this always be true regardless of context?"

- **Yes** → it belongs in `AGENTS.md` as a Hard Rule.
- **No, but it is a strong preference of the user** → `feedback.md`.
- **It describes who the user is** → `user.md`.
- **It is a one-off, in-flight work state** → `docs/tasks/current.md`.

If `AGENTS.md` and a memory entry conflict, **`AGENTS.md` wins**.
Memory is preference; rules are invariants.

---

## Why this exists

Without this slot, agents would either bloat `AGENTS.md` with mutable
facts (corrupting the rule set) or forget across sessions (forcing the
user to repeat themselves). Both outcomes degrade human-LLM
collaboration. See ADR 0004 §Context.

---

## What this is **not**

- **Not a chat log.** Do not paste session transcripts here. Distill.
- **Not a TODO list or task archive.** Use `docs/tasks/current.md`
  (planned v0.1) only for short-lived in-flight work state.
- **Not a substitute for ADRs.** Architectural decisions still go to
  `docs/adr/`. Memory captures preferences, not decisions.
- **Not present in user projects by default.** This directory is
  generated only when the user passes `aikata init --with-memory`
  (planned v0.2). The aikata repo seeds it directly as dogfooding.
