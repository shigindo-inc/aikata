---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
memory_type: user
---

# Memory — `user`

Profile, role, knowledge, and persistent preferences of the user. See
[`README.md`](./README.md) for write rules.

---

## Identity

- [2026-05-21] Primary maintainer: **Satoshi Minami**. GitHub
  organization: `shigindo-inc`. Repository:
  `https://github.com/shigindo-inc/aikata`.
- [2026-05-21] Working under the org name `shigindo-inc`; the
  unsuffixed form `shigindo` is **not** the current org and would
  require a GitHub rename to use. Module path is therefore
  `github.com/shigindo-inc/aikata`.

## Language

- [2026-05-21] Documentation default is English (`--lang en`) for
  OSS reach. Agent ↔ maintainer conversation typically happens in
  Japanese; surface trade-offs and design rationale in the language
  the conversation is already using rather than translating
  preemptively.

## Stack background

- [2026-05-21] Go (1.21+) is the implementation language for aikata
  itself — see [ARCHITECTURE.md §1](../../ARCHITECTURE.md#1-implementation-language)
  for the rationale. Prefer idiomatic Go patterns; explain
  non-obvious idioms inline rather than assuming familiarity.

## Tooling preferences

- [2026-05-21] Editors and AI tools used by the maintainer vary;
  aikata itself stays editor-agnostic. The Do-No-Harm Policy
  (ADR 0003) — including the "no Obsidian-isms leak into
  non-Obsidian projects" rule — reflects this stance.

## Decision style

- [2026-05-21] Prefers explicit early decisions over deferred ones
  when the cost of changing later is high (see Phase 1 Q&A on
  language / license / module path).
- [2026-05-21] Welcomes the assistant flagging concerns proactively
  ("懸念があれば質問しつつ対応したい"). Do not silently make
  architecturally relevant calls.
