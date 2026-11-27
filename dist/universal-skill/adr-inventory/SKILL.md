---
name: adr-inventory
user-invocable: false
description: Use when a repository has accumulated many ADRs and the user wants to take stock of them — which decisions still hold, which are superseded or only partly superseded, which nobody refers to, which are not implemented yet — or wants the effective decisions condensed, or wants rules that AGENTS.md restates from ADRs moved into REVIEW.md or a lint rule so less loads into every session. Triggers on "inventory our ADRs", "which ADRs are still active?", "summarize the current decisions", "our AGENTS.md is too long, move ADR rules out", "ADR の棚卸し", "効いている決定をまとめて". Reads facts from `aikata list adrs --json` and `aikata doctor --json`; proposes, and writes only after the user approves; never edits an ADR's decision text. For the doctor migration itself use manage-docs; for the daily context loop use track-context.
---

# adr-inventory

This skill takes stock of a repository's ADRs and proposes how to keep their
policy while loading less of it into every session (ADR 0052). The work is
split on purpose:

- **aikata computes the facts.** `aikata list adrs --json` gives, per ADR,
  the number, title, path, the raw status line, its class, the ADRs the
  status line names, how many Markdown files refer to it, and its size.
  `aikata doctor --json` reports ADRs whose status needs attention.
- **You make the judgments** with the user: what "partially superseded"
  means for a given ADR, whether a decision is implemented, how to word a
  one-line summary, where a rule belongs.

You are in an aikata-managed repository when the root has an `AGENTS.md`
and/or a `.aikata/aikata.yaml`, and the ADRs live under `docs/adr/`.

## The contract (do not deviate)

1. **Never parse ADR status yourself.** Use the `status`, `class` and
   `links` fields from `aikata list adrs --json`. If they look wrong, the
   fix belongs in aikata (one parser for doctor and the list), not in your
   own reading.
2. **Never edit an ADR's decision text.** A decision that no longer holds
   gets a new ADR that supersedes it (`aikata new adr`), written with the
   user.
3. **Propose, then write.** Every file change — the condensed list, a
   moved rule, an `AGENTS.md` cut — is shown as a plan first and applied
   only after the user approves it.
4. **Mark what you judged.** In every table, say which cells came from
   aikata's output and which are your judgment.

## Step 1 — Settle the status data

Run `aikata doctor --json` and look at the `adr.*` codes:

| Code | What to do |
|---|---|
| `adr.decision-in-frontmatter` | The decision sits in frontmatter `status`. Offer `aikata doctor --fix --dry-run`, then `aikata doctor --fix`. It moves the decision to the body status line and sets frontmatter `status` to `stable`. |
| `adr.status-conflict` | Frontmatter and the status line disagree. Show both values and ask the user which is right; `--fix` leaves these files alone. |
| `adr.missing-status` | No decision anywhere. Ask the user for the status and add a `- **Status**: …` line. |
| `adr.duplicate-number` | Two ADRs share a number, so every `ADR NNNN` reference is ambiguous. Report it; renumbering is the user's call. |

Continue with the inventory even if some of these remain, and list them in
the report.

## Step 2 — Inventory

Run `aikata list adrs --json` and build one table with these groups:

| Group | Source |
|---|---|
| Effective | `class: effective` with an empty `links` list |
| Partially superseded | `class: effective` whose `links` is not empty: read the status line prose and say which parts were replaced and by which ADR (judged) |
| Superseded / retired | `class: superseded` or `retired` |
| Pending | `class: pending` |
| Needs a person | `class: unknown` or `missing` |
| Unreferenced | `doc_refs: 0`. aikata counts Markdown only, so search the code for the ADR number before calling it unreferenced (judged) |
| Not yet implemented | Judged from the ADR text and the code; quote the evidence |

Give counts per group and the total bytes per group, so the user sees
where the weight is.

## Step 3 — Condense the effective decisions

Propose one line per effective or partially superseded ADR: the decision
in one sentence, with a link to the ADR. When the repository already keeps
a summary line in each ADR (for example a 「要点」 or 「決定」 blockquote),
start from that text instead of rewriting it. For a partially superseded
ADR, state only what still holds.

Put the list in the repository's existing architecture document (for
example a "Current decisions" section of `ARCHITECTURE.md`). Do not create
a new file that loads into every session. Ask before writing.

## Step 4 — Propose where each rule belongs

Read `AGENTS.md` (and any other always-loaded instruction file) and list
every rule it restates from an ADR. For each, propose one home:

| Home | When |
|---|---|
| Stays in `AGENTS.md` | An agent must know it before touching any code |
| `REVIEW.md` | It is checked when reviewing a change |
| A lint rule or test | A machine can check it; name the check you would add |
| Only the ADR | It is background; the condensed list links to it |

Show the table and the resulting `AGENTS.md` size before and after. Apply
nothing until the user picks which rows to take.

## Step 5 — Report

Summarize:

- the group counts from Step 2 and the `adr.*` codes still open;
- `AGENTS.md` bytes and ADR mentions, before and after;
- every rule removed from `AGENTS.md` and where it now lives;
- the ADRs that need a person, with the question for each.

## What this skill never does

- Rewrite an ADR's decision text or delete an ADR.
- Write `AGENTS.md`, `REVIEW.md`, lint rules or the architecture document
  without approval.
- Classify an ADR's status by reading the file instead of aikata's output.
