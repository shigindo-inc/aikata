---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-02
audience: [human, agent]
---

# ADR 0039 - Documentation hygiene & context budget

- **Status**: Accepted
- **Date**: 2026-06-02
- **Deciders**: aikata maintainers
- **Related**: ADR 0001 (Record Architecture Decisions — ADR bodies are
  immutable), ADR 0002 (`AGENTS.md` is canonical), ADR 0006 (locale /
  Japanese documentation policy), ADR 0028 (prioritize core-concept
  stabilization). Supersedes the informal "kept as a pointer; removable
  in a future cleanup" habit in `docs/decisions/open-questions.md`.

## Context

aikata *is* its documentation — the product thesis is that a coherent
set of operational Markdown lets any human or AI agent pick the project
up from `README.md` alone (Phase 1, ADR 0002). That same surface is the
first-read context an AI agent loads every session. By v0.9.7 the repo's
docs had grown past ~12k lines, and the growth was uneven:

- `docs/decisions/open-questions.md` carried eight fully-resolved entries
  kept "as a pointer," even though its own header says a resolved
  question's entry is *removed* once the resolution moves to an ADR. The
  pointers duplicated the ADR record and bloated every session's context.
- `ROADMAP.md` (1,443 lines) was ~80% released-version checklists —
  valuable history, but not "direction," which is the file's stated job.
- Non-ADR design notes under `docs/decisions/` stayed at full length
  after the work they scoped had shipped and been recorded in a binding
  ADR.

There was no recorded policy for *which* method applies to *which* file
class, so each cleanup was an ad-hoc judgment call and the default was to
leave stale content in place. The result: resolved/stale entries entered
the AI's first-read context every session, raising the chance of
acting on a superseded position.

## Decision

Adopt a per-file-class hygiene rubric, pruned **every release**, and
record it here as the binding policy. The method differs by class
because the files serve different readers and lifetimes.

| Class | Files | Method |
|---|---|---|
| **First-read context** | `AGENTS.md`, `SPEC.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `docs/decisions/open-questions.md` | Keep lean; prune resolved/stale content every release. A resolved open-question is **removed** (the ADR is the durable record), not kept as a pointer. |
| **Reference / history** | `ROADMAP.md`, `CHANGELOG.md` | Archive the released past to `docs/*-archive.md` (**move, not delete**); keep current + future detail plus a one-line pointer and a compact summary in the live file. |
| **Immutable** | `docs/adr/**` | Never edit an accepted ADR body (ADR 0001); only the README ADR index and cross-links may change. A reversed decision gets a *new* superseding ADR. |
| **Design notes** | `docs/decisions/*.md` (non-ADR) | Once the work ships and is recorded in a binding ADR, condense the note to a short pointer stub to that ADR. |
| **Memory** | `docs/memory/**` | Governed by the memory discipline (ADR 0004); no mechanical pruning. |

**Guardrails.**

- **Move, not delete.** Archiving relocates content to a sibling
  `docs/*-archive.md`; git history is not a substitute for a reachable
  document. Honours the no-silent-deletes discipline.
- **Frontmatter on archives.** Any new `docs/*-archive.md` carries the
  five-key frontmatter so `aikata doctor --all-markdown --strict` stays
  green; relative links are rewritten for the new file location.
- **Reviewed in the diff.** Pruning a resolved open-question or
  condensing a design note is a deliberate, diff-reviewed change — never
  a silent truncation. The PR diff shows content removed *here* and
  preserved *there* (ADR / archive).
- **Info-level GLOSSARY orphans are acceptable.** Terms not referenced by
  another doc are core vocabulary, not waste; `doctor` reports them at
  info level and they are left in place.

## Consequences

- Resolved entries leave the first-read surface promptly: this release
  removes eight resolved open-questions (Q-DESIGN-04/-09/-12/-13,
  Q-ECOSYSTEM-03, Q-DOCTOR-01, Q-INTEROP-04, Q-INTEROP-05), taking
  `open-questions.md` from 476 to ~340 lines. The ADRs remain the record.
- `ROADMAP.md` drops from 1,443 to ~545 lines; the released Phase 1 –
  v0.8.5 detail moves verbatim to `docs/roadmap-archive.md` behind a
  compact summary table.
- A condensed design note points to its binding ADR instead of
  duplicating it.
- The policy is now explicit, so future releases prune by rule rather
  than by ad-hoc judgment. `CONTRIBUTING.md` links here so release
  authors apply it as part of the release ritual.
- `CHANGELOG.md` is intentionally **not** archived in this pass: it is
  not first-read AI context and Keep-a-Changelog favours completeness.
  The rubric permits archiving it later (pre-v0.9.0 → `docs/changelog-
  archive.md`) if it ever becomes a context burden.

## Alternatives considered

- **Keep resolved entries as pointers.** The prior habit. Rejected: it
  duplicates the ADR, and the pointers accreted until the file was
  mostly resolved items — the exact context bloat this ADR targets.
- **Delete history outright instead of archiving.** Rejected: violates
  no-silent-deletes; "it's in git" is not a reachable document for a
  human or agent reading the repo.
- **One uniform method for all docs.** Rejected: immutable ADRs, living
  first-read context, and append-only changelogs have genuinely
  different lifetimes; a single rule would damage one to suit another.
