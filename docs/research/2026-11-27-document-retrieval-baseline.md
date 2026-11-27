---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# JA/EN section retrieval baseline

> **Summary**: The reproducible navigation-and-ripgrep baseline recovered every required section for 7 of 13 positive held-out questions and returned no candidates for all three negative questions. Japanese paraphrases frequently missed English evidence.
> **Read when**: Comparing a retrieval candidate against [#176](https://github.com/shigindo-inc/aikata/issues/176).
> **Status**: Observed
> **Confidence**: High for the replay, low for generalization.
> **Adoption**: Use as an experimental control.

## Question

How much required evidence can a small, reproducible version of the existing documentation navigation and ripgrep workflow recover under the proposed lookup budget?

## Findings

| Split | Positive questions with all evidence | Required sections | Correct empty | Historical section without successor | Output bytes |
|---|---:|---:|---:|---:|---:|
| Tuning | 2/7 | 4/10 | 1/1 | 0 | 3,549 |
| Held-out | 7/13 | 12/20 | 3/3 | 0 | 7,459 |

Held-out complete evidence was 1/6 for Japanese positive questions and 6/7 for English. Two held-out Japanese paraphrases returned no results (h03 and h05). The Japanese questions about the doctor scope and partial supersession each recovered only part of the required evidence (h07 and h09). The corresponding English partial-supersession question recovered both sections (h08). An English multi-source question missed both labeled sections (h06), showing that language alone does not explain the misses.

## Evidence

The [fixture](../../testdata/context-retrieval/cases.json) contains 24 annotated questions (8 tuning, 16 held-out, 12 Japanese and 12 English). Gold sections use exact path and heading pairs. Fixtures include exact IDs, paraphrases, navigation, partial supersession, multi-source evidence and explicit negative queries. Labels were written and checked against the pinned text before measuring results.

Run from the repository root with Python 3.9+, Git and ripgrep:

```sh
python3 testdata/context-retrieval/baseline.py
python3 testdata/context-retrieval/baseline.py --json
```

The [runner](../../testdata/context-retrieval/baseline.py) extracts 72 public Markdown documents from commit `58844055d98b986f607b1f2c4d9987cefac30487` into a temporary directory. It replays hand-selected fixed query terms with ripgrep, visits the navigation entrypoints first and ranks remaining grep matches by number of terms, then path and line. For a sole explicit ADR number it opens that file's Decision heading. It caps results at five candidates and 2,400 UTF-8 bytes per question. It emits per-case candidates, timing, byte counts, and misses. Two runs produced identical candidates and scores after excluding latency.

## Counterevidence and limitations

This is a scripted, repeatable proxy for the manual route. The terms are authored per question and can include English identifiers in Japanese questions. A human can issue follow-up searches, and an adaptive agent may do better than this replay. The split separates query IDs, with some overlapping topics and evidence. It is unsuitable as evidence of generalization to unseen topics after tuning on overlapping examples.

The bytes count the serialized candidate array, including its pointers, excluding corpus setup and common mandatory reads. This run measured search time after setup, with medians near 8 ms on the development machine. It measured evidence recovery, not whether a human or model actually read the full section or made the correct decision. A historical section returned without its successor is a risk proxy. A zero count does not establish that supersession was understood. Actual token consumption, end-to-end resolution time and user quality remain unmeasured.

## References

[Evaluation work item](https://github.com/shigindo-inc/aikata/issues/176) · [Draft design](../spec/task-scoped-document-retrieval.md) · [Delivery plan](../plans/2026-11-27-task-scoped-document-retrieval.md)
