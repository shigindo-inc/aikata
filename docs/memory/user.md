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

- [2026-05-21] Native language: Japanese. Documentation default is
  English (`--lang en`) for OSS reach, but agent ↔ user conversation
  is in Japanese.
- [2026-05-21] Comfortable reading English technical prose; prefers
  Japanese for high-context discussions (design intent, trade-offs).

## Stack background

- [2026-05-21] Flutter / Dart developer leaning. The aikata `flutter`
  preset is a personal driver, not a hypothetical. Treat Flutter as a
  v0.2 first-class target rather than a nice-to-have.
- [2026-05-21] Go is the chosen implementation language for aikata
  itself (ADR-equivalent in ARCHITECTURE.md §1). The user has less Go
  fluency than Dart; prefer idiomatic Go patterns and explain
  non-obvious Go idioms when they appear in PRs.

## Tooling preferences

- [2026-05-21] Uses Claude Code (this conversation), Obsidian
  (personal vault, separate repo), and command-line git. The aikata
  Do-No-Harm Policy (ADR 0003) reflects the user's strong feelings
  about not letting Obsidian-isms leak into non-Obsidian projects.

## Decision style

- [2026-05-21] Prefers explicit early decisions over deferred ones
  when the cost of changing later is high (see Phase 1 Q&A on
  language / license / module path).
- [2026-05-21] Welcomes the assistant flagging concerns proactively
  ("懸念があれば質問しつつ対応したい"). Do not silently make
  architecturally relevant calls.
