---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
memory_type: reference
---

# Memory — `reference`

Pointers to external systems and authoritative external references
relevant to this project. Reading these is often required to resolve a
question; bookmark them here so they are one click away. See
[`README.md`](./README.md) for write rules.

---

## Repository

- [2026-05-21] GitHub repo:
  [`shigindo-inc/aikata`](https://github.com/shigindo-inc/aikata).
- [2026-05-21] Local working trees vary by contributor; do not record
  absolute paths here. Use `$REPO_ROOT` in commands when documenting
  shell snippets.
  **(superseded 2026-05-21: removed the maintainer's absolute path
  before the OSS-readiness scrub of Task 3A)**

## Internal canonical entry points

- [2026-05-21] [`../../README.md`](../../README.md) — human entry
  point.
- [2026-05-21] [`../../AGENTS.md`](../../AGENTS.md) — agent rules
  (canonical).
- [2026-05-21] [`../../SPEC.md`](../../SPEC.md) — what / why.
- [2026-05-21] [`../../ARCHITECTURE.md`](../../ARCHITECTURE.md) — how.
- [2026-05-21] [`../../ROADMAP.md`](../../ROADMAP.md) — when.
- [2026-05-21] [`../adr/`](../adr/) — Architecture Decision Records.
- [2026-05-21] [`../decisions/open-questions.md`](../decisions/open-questions.md)
  — undecided items.
- [2026-05-21] [`../origin/`](../origin/) — historical planning docs
  (do not edit).

## External specs and prior art

- [2026-05-21] [agents.md open spec](https://agents.md/) — canonical
  source format aikata bets on (ADR 0002).
- [2026-05-21] [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
  — CHANGELOG.md format.
- [2026-05-21] [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
  — versioning policy.
- [2026-05-21] [Conventional Commits](https://www.conventionalcommits.org/)
  — commit-message convention.
- [2026-05-21] [MADR](https://adr.github.io/madr/) — ADR template
  inspiration (we use a simplified Nygard/MADR hybrid, see ADR 0001).

## Competitor / related projects

- [2026-05-21] [Goldziher/ai-rulez](https://github.com/Goldziher/ai-rulez)
  — 19+ AI-tool support; aikata's main competitor in concept.
- [2026-05-21] [block/ai-rules](https://github.com/block/ai-rules) —
  simpler than ai-rulez; monorepo support.
- [2026-05-21] [intellectronica/ruler](https://github.com/intellectronica/ruler)
  — rule centralization.
- [2026-05-21] agentsmesh — newcomer attempting full-config
  unification.

## Go ecosystem references (for Phase 2+)

- [2026-05-21] [spf13/cobra](https://github.com/spf13/cobra) — CLI
  framework (Phase 2 dependency).
- [2026-05-21] [charmbracelet/huh](https://github.com/charmbracelet/huh)
  — interactive prompts (Task 6).
  **(superseded 2026-05-22: not adopted in v0.1. huh v1+ requires
  Go 1.23 and pulls in ~30 indirect deps; bufio-based prompt used
  instead. See CHANGELOG `aikata init --with-memory` / Task 6 note.)**
- [2026-05-21] [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
  — terminal styling (Task 6).
  **(superseded 2026-05-22: not adopted in v0.1 — same reason as huh.)**
- [2026-05-21] [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3)
  — YAML parser (Task 5).
- [2026-05-21] [goreleaser/goreleaser](https://goreleaser.com/) —
  binary release (Task 8).
