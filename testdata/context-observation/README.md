---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# Recurring question observation fixture

This synthetic fixture checks the proposed local counting contract for
[#178](https://github.com/shigindo-inc/aikata/issues/178). The Python
reference reads only the explicitly supplied fixture and returns JSON on
stdout. It has no collector, transcript integration, background process
or external service. It requires Python 3.8 or newer and the standard library.

## Run

From the repository root:

```sh
python3 -B -m unittest discover -s testdata/context-observation -v
python3 -B testdata/context-observation/aggregate.py testdata/context-observation/synthetic.json
```

`expected.json` is the sample output checked by the tests. An invalid
envelope, inconsistent event replay or unsupported field returns exit 2.

## Reading the fixture

- `phase_exposed` counts that a work reached a phase, even with no search.
  One work can be exposed in multiple phases. A question without matching
  exposure is visible in the missing-exposure count and is omitted from
  the corresponding rate. Works without events are not observable.
- `question_observed` counts once per work, phase and question. Exact
  event replays and later observations of that same combination do not
  increase the count. The rolled-up view counts a work once and uses its
  earliest eligible question week. Across phases, any observed repeat
  counts as repeat. A no-repeat total requires coverage of all recorded
  positive phases for that work.
- `outcome_observed` is a label rather than a count of tool calls.
  Missing outcomes are separate. A planned verification reread is
  labeled no repeat search when the source was already known and the
  reread verifies an assertion.
  A zero denominator is unavailable. Denominators of 1–9 keep counts but
  suppress the rate in this experiment.
- The fixture has 10 implementation exposures, two eligible works
  with the sync question and one sync question with missing exposure.
  One eligible work has a known repeat and the other has no outcome.
  The topic total is 3 of 10 because review adds one distinct work
  and a second appearance of an already counted work.

## Manual audit design

Sample candidate-positive, candidate-negative, and unclassified work,
including no-search work. Within each stratum, record sample size, whether
the question was present, and false merges separately from missed
equivalents. Avoid pooling strata with different sampling fractions.

Example labels:

| Audited pair or case | Expected interpretation |
|---|---|
| “Can I accept one conflicting file?” / “他を残して、このファイルだけ受理できる？” | Same decision if both refer to per-file sync conflict acceptance under the same applicability conditions. |
| Accept one resolved conflict / discard the file and restart sync | Separate decisions even when both consult one ADR. |
| Reopen the source to verify a test assertion | Planned verification, not a repeat search for the same evidence. |

The synthetic labels exercise reporting. They do not measure real
detection precision, recall, time savings or the effect of guidance.
The next stage needs explicit enrollment and prospectively labeled tasks,
with sample denominators and remaining unknowns reported.

## Start an opt-in manual window

Agree on the participating work, a time window, and one question family
before assigning arms. Copy the envelope below to a local file outside
the checkout. Set `enrolled` to true only after agreement. Use the
synthetic fixture as a shape example for events and audit labels, never
as observations.

```json
{
  "schema_version": 1,
  "enrolled": false,
  "questions": ["q-accept-sync-conflict"],
  "events": [],
  "audit": []
}
```

Retain the same opaque work ID when the task is resumed. Add a phase
exposure even when no search occurs. If the question comes up, record at
most one question event per phase and a later outcome if observable.
Leave unknown outcomes absent. Audit a sample of candidate-positive,
candidate-negative and unclassified work before interpreting the counts.

Run the same aggregator with the local file as its only argument.
A window with no observations returns unavailable rates. Review the
aggregate and consent scope before sharing it. Keep the raw observation
file local and follow the retention terms agreed for the window.
