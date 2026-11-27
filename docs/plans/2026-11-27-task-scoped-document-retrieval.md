---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# Task-scoped retrieval delivery plan

> **Summary**: Offline comparisons and the source-linked guidance replay are complete. Keep automatic retrieval and maintained guides experimental pending opted-in work.
> **Read when**: Picking up or reporting progress on [#175](https://github.com/shigindo-inc/aikata/issues/175).
> **Status**: Active
> **Progress**: 8/10
> **Next action**: Run an opt-in window for one question family, then compare an explicit phrase-based lookup and a temporary guide on independent work.
> **Blockers**: Real recurrence and guidance impact need observed, opted-in work. Current retrieval results do not justify a public CLI.

## Goal

Deliver a well-grounded decision about whether bounded section lookup and recurring-question guidance should ship, with quality and cost evidence. The [draft design](../spec/task-scoped-document-retrieval.md) owns the contracts. The parent issue owns public status.

## Phases

- [x] Reconcile related work with upstream main, released behavior and open issues.
- [x] Publish the parent and four linked work items.
- [x] Draft retrieval, observation and guidance contracts with acceptance criteria.
- [x] [#176](https://github.com/shigindo-inc/aikata/issues/176) Pin the 72-document corpus, label 24 JA/EN queries and run a scripted manual-route baseline. [Benchmark notes](../research/2026-11-27-document-retrieval-baseline.md).
- [x] [#178](https://github.com/shigindo-inc/aikata/issues/178) Build synthetic observation fixtures, expected counts and independent-work deduplication checks.
- [x] [#177](https://github.com/shigindo-inc/aikata/issues/177) Run a standalone bounded lookup against #176. Keep changes outside the product while the prototype misses the quality bar.
- [x] [#177](https://github.com/shigindo-inc/aikata/issues/177) Compare vocabulary, abstention and successor expansion separately on a fresh fixture. [Variant results](../research/2026-11-27-section-lookup-variants.md).
- [x] [#179](https://github.com/shigindo-inc/aikata/issues/179) Replay two staged navigation routes and document the field procedure. [Guidance replay](../research/2026-11-27-topic-guidance-pilot.md).
- [ ] [#178](https://github.com/shigindo-inc/aikata/issues/178) Audit candidate-positive, candidate-negative and unknown cases from explicitly opted-in work. The synthetic test does not estimate field precision.
- [ ] [#179](https://github.com/shigindo-inc/aikata/issues/179) Pilot one or two guides using #176/#178. Match work by task type and pin required sources before each comparison. Record ship, revise or stop in #175.

### First comparison

| Held-out outcome | Scripted navigation + ripgrep | Offline lookup |
|---|---:|---:|
| All required evidence in positive cases | 7/13 | 4/13 |
| Required sections recovered | 12/20 | 9/20 |
| Correctly empty negative cases | 3/3 | 0/3 |
| Historical section surfaced without its required complement | 0 | 2 |

Both use the same revision, candidate cap and maximum output size. The baseline uses author-supplied search terms and an exact-ID navigation shortcut. The lookup uses raw query text, so this is an operational comparison and does not isolate the ranking algorithm. Neither run generated answers or measured model tokens or task-level savings. Treat the observed misses as development data before designing a fresh held-out set.

Reproduce from the repository root:

```sh
python3 -B testdata/context-retrieval/baseline.py
python3 -B testdata/context-lookup/compare.py
python3 -B -m unittest discover -s testdata/context-lookup
python3 -B -m unittest discover -s testdata/context-observation
```

The lookup keeps source revision and line ranges and checks duplicate IDs. Miss review found empty results for JA-only paraphrases (h03, h05, h07, h15), and neighboring candidates for an absent ID embedded in a sentence (h11). In two partial-supersession cases (h08, h09), the old clause surfaced without its successor. Try vocabulary, abstention and explicit successor handling independently before proposing a shared-parser change.

### Pilot decision

Automatic free-form lookup is not ready to ship. In a separate hard
confirmation set, authored phrases with a literal guard recovered 8/13
required sections compared with 6/13 for the scripted baseline. Both
returned related material in all six unsupported-behavior cases. Test
applicability separately from candidate retrieval in the field pilot.

The staged guides preserved all required sections in both tasks and
saved 543 and 1,210 navigation bytes against an extra search. If the
existing instructions already provide the needed links, the guides add
bytes. Nominate guides only after observing repeat discovery in
independent work, then confirm benefits against the normal workflow.

## Verification

- Check child issue states against #175 when updating this plan.
- Hold required evidence and comparison budgets fixed. Publish misses and unknowns, including a Japanese query without an adequate match.
- Exercise multi-session deduplication, missing exposure, missing outcome, legitimate rereads and sparse denominators with synthetic records.
- On implementation, confirm the existing map, doctor and inventory behavior against current upstream and run the repository's affected tests and CI checks.

## Risks

The working tree was v0.15.3 when this plan began. Reconcile implementation with current upstream main before changing shared parsers. Optional collection should be tested outside tracked outputs. Keep physical ADR reorganization, mandatory metadata, background injection and embedding search as separate decisions if a measured limitation warrants them.
