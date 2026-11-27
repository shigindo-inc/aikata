---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# ADR 0052 — ADR inventory: the CLI lists, a skill judges

- **Status**: Accepted (2026-11-27)
- **Date**: 2026-11-27
- **Related**: ADR 0001 (ADR format), ADR 0039 (documentation hygiene;
  ADR bodies are not rewritten), ADR 0045 (no `aikata compact`),
  ADR 0049 (doctor reports what needs action), ADR 0050 (doc map off by
  default)

## Context

Repositories that use aikata keep adding ADRs. The largest downstream
repository has 189 ADRs (4.6 MB). Nobody can tell from the file list
which decisions still hold, and the agent rules file (`AGENTS.md`,
24 KB, loaded into every session) restates ADR rules by hand: it
mentions ADRs 29 times. The maintainers want to:

1. inventory the ADRs: effective, superseded, partially superseded,
   unreferenced, not yet implemented;
2. condense the decisions that are still effective;
3. propose moving rules out of ADRs into `AGENTS.md`, `REVIEW.md`, or a
   lint rule;
4. keep the policy while shrinking what loads into every session.

aikata calls no LLM API (SPEC §2.2, SPEC §8), and ADR 0045 deferred an
`aikata compact` command for that reason. Steps 2 and 3 need judgment,
so the CLI can only supply facts for them.

### What the repositories write today

Measured on 2026-11-27 over the 12 local repositories that have
`.aikata/aikata.yaml`:

| Style | Repositories | ADRs |
|---|---|---|
| Frontmatter `status: draft`, decision in the body `**Status**` line (aikata's template) | 5 (aikata itself: 48) | 61 |
| Decision words in frontmatter *and* in the body line | 7 | 255 |

In the 189-ADR repository:

- The body line exists in 182 ADRs (175 use `**状態**`, 7 `**Status**`);
  7 carry only frontmatter.
- Frontmatter values: `accepted` 106, `active` 56, `superseded` 16,
  `draft` 9, `proposed` 2.
- A frontmatter decision word and the body line disagree in 5 ADRs:
  `accepted` vs. superseded (1), `superseded` vs. `active` (1),
  `accepted` vs. 「役目を終えた」 (retired, 1), `proposed` vs. `active` (2).
- 17 ADRs state partial supersession only as prose inside the body
  line, for example 「active（一部失効。… は ADR 0115 で失効した。… は残る）」.
  Every one of them links the superseding ADR in that line.
- 1 ADR is referenced by no other file; 3 are referenced only by other
  ADRs.
- 176 ADRs carry a one-paragraph 「要点」 summary and 166 a 「決定」 line
  (a repository convention aikata does not define).

Frontmatter `status` is required on every Markdown file and means the
document lifecycle everywhere outside ADRs
(`internal/doctor/checks.go:21`, `internal/doctor/fix.go:101`,
`internal/docmap/docmap.go:122`). `aikata doctor` (v0.16.0) takes the
ADR body line first and falls back to frontmatter
(`internal/doctor/checks.go:141`), so for ADRs the same key means the
lifecycle or the decision depending on whether a body line exists. Its
`--json` output holds only issues, so no tool can read a parsed status
from aikata today.

> **確度**: 実測 — every count above, from scripts over the working
> trees on 2026-11-27 (`git grep` for references; first body status line
> per file; frontmatter `status:`). The 「partial」 count is a keyword
> match (一部 / のみ / 置き換え / 取り下げ / 失効) on the status line and
> may miss phrasings. 二次情報 — that adr-tools (`adr list`,
> `adr generate graph`) and log4brains (a static site from MADR status)
> list ADRs but do not model partial supersession or rule extraction;
> from memory of their docs, not re-read. 推測 — that nobody in the 7
> frontmatter-style repositories reads frontmatter `status` as the
> decision with another tool. 未調査 — repositories outside this
> machine. The decision does not depend on the 二次情報 item: aikata's
> doctor already parses statuses and needs one shared parser either way.
> If the 推測 is wrong, the migration in Decision 2 breaks that tool, so
> the CHANGELOG entry names the change.

## Decision

1. **The body status line is the only decision status.** ADRs get no
   exception to the lifecycle meaning of frontmatter `status`. `doctor`
   warns (`adr.decision-in-frontmatter`) when an ADR's frontmatter
   `status` holds a decision word: `proposed`, `accepted`, `active`,
   `superseded`, `deprecated`, `rejected`.
2. **aikata ships the migration, and the fallback goes in the same
   release.** `aikata doctor --fix` (with `--dry-run` first) rewrites a
   frontmatter decision word to the lifecycle value `stable`, and adds a
   body status line built from the frontmatter word to an ADR that has
   none. When the frontmatter word and the body line's class differ, or
   the body line's class is unknown, `doctor` reports
   `adr.status-conflict` instead, and the fix leaves that file for a
   person to settle (fail-closed). From that release on, `doctor` no
   longer reads frontmatter for the decision. Run on a copy of the
   189-ADR repository, the fix rewrote 174 frontmatter values (7 of them
   also gained a body line) and left 6 files to a person; the sixth is
   an ADR whose status line starts 「一部 active」.
3. **Parsing lives in `internal/adr`.** That package already holds the
   ADR numbering rules as their single source
   (`internal/adr/numbering.go:1-6`). The status parser moves there from
   `internal/doctor/checks.go:131-151`, and `doctor` calls it. The parser
   maps a status to a class by its first word: `pending` (proposed),
   `effective` (accepted / active), `superseded`, `retired` (deprecated /
   rejected), or `unknown`. It also returns every ADR number linked from
   the status line. It does not decide "partially superseded": an
   `effective` ADR whose status line links another ADR is reported with
   those links, and the skill reads the prose.
4. **`aikata list adrs --json`** is a new read-only command, named
   verb-first like `aikata new adr`. Per ADR it emits number, title,
   path, raw status, class, linked ADR numbers, inbound references
   (files outside the ADR that link it or name it as `ADR NNNN`), and
   bytes. Link extraction reuses `docmeta.ExtractLinks`. `doctor --json`
   stays a list of issues (ADR 0049).
5. **A plugin skill, `adr-inventory`, does the judgment.** It runs
   `aikata list adrs --json` and `aikata doctor --json`, then, with the
   agent, writes:
   - an inventory table of the five groups (effective, superseded,
     partially superseded, unreferenced, not yet implemented). "Partially
     superseded" and "not yet implemented" are judgments from the ADR
     text and the code, and the table marks them so;
   - a condensed list of effective decisions, one line each, into the
     repository's existing architecture document (for example a "Current
     decisions" section of `ARCHITECTURE.md`), not into a new
     always-loaded file;
   - rule-extraction proposals: for each rule that `AGENTS.md` restates
     from an ADR, whether it belongs in `AGENTS.md`, `REVIEW.md`, or a
     lint rule. The skill proposes; a person applies.
   The skill never edits an ADR's decision text. A decision that changed
   gets a new superseding ADR (ADR 0039).
6. **Out of scope now:** a CLI command that writes condensed text, and
   detection of partial supersession in the CLI. Either needs its own ADR
   and evidence that the skill keeps redoing the same computation.

## Measuring success

On the 189-ADR repository, before vs. after the migration and one skill
run:

- ADRs that `aikata list adrs` classes as `unknown`: target 0.
- `adr.decision-in-frontmatter` warnings: 174 → 0, and
  `adr.status-conflict`: 6 → 0, settled by a person.
- `AGENTS.md` bytes (baseline 24,269) and ADR mentions in it (baseline
  29): both go down, and every removed rule appears in `REVIEW.md`, a
  lint rule, or the condensed list.
- No ADR decision text changes in the diff.

## Consequences

- One parser serves `doctor` and `list adrs`; a change to status phrasing
  touches one package.
- Every repository that writes decision words in frontmatter (7 of the
  12 measured) sees warnings after upgrading until it runs the fix. The
  CHANGELOG lists this as a breaking change.
- The skill depends on agent judgment, so two runs can word the
  condensed list differently. The inventory numbers do not vary.

## Alternatives Considered

- **Keep both sources and warn only on disagreement.** No rewrite is
  needed, but the same key keeps two meanings for ADRs forever, and the
  fallback decides which meaning applies by whether a body line exists.
  The 5 disagreements came from that duplication.
- **Frontmatter `status` as the decision.** It gives ADRs a meaning for
  that key that no other document has, and a single scalar cannot hold
  the 17 partial-supersession notes.
- **Add the inventory to `doctor --json`.** doctor reports what needs
  action (ADR 0049); a listing of healthy ADRs is a different job.
- **Add ADR fields to the doc map.** The map is off by default
  (ADR 0050) and is a committed file, which works against shrinking
  context.
- **No CLI change; the skill reads files.** The agent would parse
  statuses itself, a second rule source that drifts from doctor's.
- **`aikata compact` / LLM condensation in the CLI.** Ruled out by SPEC
  and ADR 0045.
