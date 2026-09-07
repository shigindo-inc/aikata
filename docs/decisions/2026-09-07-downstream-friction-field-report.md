---
project: aikata
status: draft
version: 0.0.1
updated: 2026-09-07
audience: [human, agent]
---

# Field report — friction observed in a downstream project

> Measured findings from one real aikata-managed project (`itteco`: a
> Flutter + Cloud Functions app, 47 ADRs, ~200 docs, several LLM agents
> working in parallel). Not a decision note. Each finding below is a
> candidate issue with the measurement that produced it, so a maintainer
> can judge disposition without re-deriving the evidence.
>
> Measured on 2026-09-07 against aikata CLI v0.14.0 unless stated.
> The upstream research note that produced these lives in the downstream
> repo at `docs/research/2026-09-06-aikata-adoption-and-session-friction.md`.

---

## Why this report exists

The project adopted aikata fully — `AGENTS.md`, `SPEC.md`,
`ARCHITECTURE.md`, `GLOSSARY.md`, 47 ADRs, `docs/memory/`,
`docs/tasks/current.md`, `docmap`. The owner still reported that the
benefit felt thin, and the measurements below say why: the slots exist,
but three of them decay in ways aikata does not currently detect, and
one axis the project needs has no slot at all.

None of this is a request to make aikata a session runtime. Parallel
session coordination was explicitly judged a non-problem by the project
owner. What follows is about the document set only.

---

## F1 — No slot for implementation status (highest impact)

**Symptom.** An agent starting fresh reads an ADR marked
`状態: active`, reads the code, finds them different, and reports a
contradiction. There is no rule saying which one wins, so it re-digs
through GitHub issues, commit messages and PR bodies every time.

**Measurement.**

| | |
|---|---|
| ADRs in the project | 47 |
| Marked `active` | 39 |
| Mentioning implementation state at all | 8 |
| Rule for reading ADR-vs-code divergence | none found |

`active` means "not withdrawn". It says nothing about whether the
decision has been implemented. Faced with an `active` ADR and code that
differs, a reader has three possible readings — the ADR is the target
and code has not caught up, the code moved on and the ADR is stale, or
it is a bug — and nothing in the layout picks a default.

One ADR in that project writes its status as
`active（決定は確定、実装は Phase 2）`. That is a maintainer hand-patching
a field the schema does not offer.

**Candidate directions.** Not proposing a design, only bounding it.
Adding a mutable implementation field to ADRs conflicts with
"immutable once Accepted" (ADR 0001) — withdrawal happens once, but
implementation state changes repeatedly, so 47 documents would acquire
ongoing update duty. A separate ledger keyed by ADR keeps ADRs
immutable and puts the update duty in one place. Either way the
cheaper half is a documented default for reading divergence, which
costs a paragraph in the generated `AGENTS.md` template and removes
two of the three readings.

---

## F2 — `map` indexes gitignored worktrees

**Symptom.** `docmap` fills with entries from worktrees that are not
part of the project, so maintainers stop regenerating it. A stale
docmap sends agents back to reading the full `AGENTS.md` "read first"
list instead of searching the index.

**Measurement.**

| | |
|---|---|
| Worktrees attached to the project checkout | 12 |
| Of those, outside the project tree | 8 (7 under `~/.codex/worktrees/`, 1 in a scratchpad) |
| `docmap.md` last regenerated | 2026-08-30 (8 days before measurement) |
| `.aikata/docmap.md` churn since 2026-08-08 | 1,467 lines, 6th largest of all Markdown in the repo |

Reproduced live while writing this report. Running `aikata generate`
after a one-paragraph edit to `AGENTS.md` rewrote `.aikata/docmap.md`
by **+3,997 / −149 lines**, of which **2,009 added lines reference
worktree paths** — roughly half the insertion. The map is rebuilt as a
side effect of `generate`, not only by `map`, so a maintainer who edits
a canonical document and regenerates gets a four-thousand-line diff
they did not ask for.

That is the mechanism behind the stale docmap: the workaround is "clean
up worktrees first", the cost of doing so exceeds the benefit of a
fresh index, and so the index is left to rot.

---

## F3 — `docs/tasks/current.md` becomes a diary, and shrinking it does not hold

**Symptom.** The short-term working-state file grows into a completion
log. `SPEC §4.1` says it is not a backlog, and the `track-context`
skill says it is not an archive, but nothing detects the drift.

**Measurement.** In the downstream project:

- Commit `ee8cd606` (2026-08-18) rewrote the file `+52 / −136`, and
  recorded the reason in the file itself: short-term working memory
  that becomes an archive stops working as a tool for remembering what
  you were just doing.
- Three weeks later the file was back to 526 lines.
- The head of the file was stale too: its "last measured" line said
  2026-09-04 and its "next step" pointed at work that had already
  merged to `main` on 2026-09-05.

That last point is the interesting one. The obvious fix — keep the file
short — does not address it. What rotted was the summary at the top,
not the log at the bottom, so a thinner file would have carried the
same wrong statement in fewer lines.

**Candidate direction.** `doctor` cannot know whether the content is
true, but it can see cheap proxies: file length against a threshold,
and the age of a `最終実測` / `updated` marker against the working
tree's last commit. A warning is enough; this does not need to fail.

---

## F4 — `sync` cannot preserve customized preset files

Files rendered from presets and then edited by the project converge
back to the template on every `sync`. Observed repeatedly on
`README.md`, `.gitignore` and `docs/tasks/current.md`. Recorded as a
standing annoyance in the downstream project's agent memory; it
recurs each release. No issue exists upstream, so it is filed here for
triage rather than described in detail.

---

## F5 — `layout.md` documents capabilities without saying when they arrived

**Symptom.** A reader on an older CLI cannot tell "not yet" from
"never".

**What happened.** During this investigation, `docs/layout.md` and
ADR 0047 documented `enable modeling` (`docs/usecases.md` +
`docs/domain.md`). The installed CLI was v0.14.0, whose
`list capabilities` returns eleven identifiers and does not include
`modeling`. The investigator concluded the capability did not exist and
corrected a colleague who had recommended it. That was wrong: the
capability shipped in v0.15.0, released 2026-09-06, and the colleague
was right.

**Why it belongs in this report.** This is F1 in aikata's own
repository — a document describing a decided target state, an
implementation that had not caught up in the reader's environment, and
no marker to tell them apart. The generic fix and the specific one
point the same way: annotate the capability rows in `layout.md` with
the version that introduced them, and have `doctor` note when the CLI
is older than the project's documents assume.

---

## What was deliberately left out

- **Session coordination.** Busy/idle tracking, worktree launching and
  inter-session messaging are handled outside aikata in that project,
  and the owner judged parallel work to be causing little harm. Folding
  it into aikata was considered and rejected on the downstream side.
- **Anything requiring aikata changes to unblock the project.** The
  three highest-priority downstream fixes — a divergence-reading rule,
  an implementation-status ledger, and routing decided target states to
  ADR or issues rather than `current.md` — are all doable with the
  current tool. They are listed here only so the maintainer knows what
  the project is about to do locally, and can pull any of it upstream
  later if it generalises.

## Not verified

- Whether F1 and F3 are deliberate scope exclusions rather than gaps.
- Whether the 12 worktrees are in fact the source of the docmap
  entries (counts measured, attribution inferred).
- How many of the 47 ADRs are actually unimplemented, which is what
  determines the cost of any ledger.
