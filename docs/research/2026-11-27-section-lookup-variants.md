---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# Section lookup variants

> **Summary**: Authored phrases helped the second benchmark, but did not solve difficult multi-source or near-neighbor questions in a separate confirmation set. Keep automatic free-form retrieval experimental and require a separate evidence judgment.
> **Read when**: Reviewing the next step for [#177](https://github.com/shigindo-inc/aikata/issues/177).
> **Status**: Reproduced offline.
> **Confidence**: High for the recorded runs, unmeasured for field quality or savings.
> **Adoption**: No public CLI change.

## Question

Can Japanese vocabulary, abstention, or explicit successor references improve a bounded lookup without losing required evidence?

## Result

The second fixture has eight Japanese and eight English questions. Four are intentionally negative. The phrases supplied to the scripted navigation baseline were authored with each question. The first five lookup arms used raw queries. For the diagnostic arms, both workflows were given the same authored phrases.

| Second fixture | Complete positive cases | Required sections | Correctly empty negatives |
|---|---:|---:|---:|
| Scripted navigation + ripgrep | 4/12 | 6/17 | 4/4 |
| Raw query, original lookup | 5/12 | 8/17 | 1/4 |
| Raw query, abstention | 5/12 | 8/17 | 2/4 |
| Raw query, literal Japanese aliases | 5/12 | 8/17 | 0/4 |
| Raw query, explicit successor expansion | 5/12 | 8/17 | 1/4 |
| Raw query, all three changes | 5/12 | 8/17 | 1/4 |
| Authored phrases, ranked as terms | 10/12 | 15/17 | 2/4 |
| Authored phrases, require at least one complete phrase | 10/12 | 15/17 | 4/4 |

### Independent confirmation

The [confirmation fixture](../../testdata/context-retrieval/confirmation.json)
contains 12 new questions, six with multiple required sections and six
about unsupported behavior with similar documentation. Its 13 gold
sections do not overlap with either earlier set.

| Confirmation | Complete positive cases | Required sections | Empty unsupported requests |
|---|---:|---:|---:|
| Scripted navigation + ripgrep | 2/6 | 6/13 | 0/6 |
| Raw query, original or combined lookup | 0/6 | 2/13 | 0/6 |
| Authored phrases, ranked as terms | 0/6 | 5/13 | 0/6 |
| Authored phrases, require at least one complete phrase | 2/6 | 8/13 | 0/6 |

For unsupported behavior, a nearby source can help establish the boundary.
An empty result measures lookup abstention. A returned candidate does
not imply that the feature exists, and this fixture did not measure
whether a downstream answer would make that unsupported claim. Future
evaluation should require the applicable boundary evidence and score the
answer separately. Keep exact missing-ID tests as empty-result checks.

## Evidence and reproduction

[Fresh questions](../../testdata/context-retrieval/fresh-heldout.json) were annotated independently before scoring. The alias set uses literal phrases from the original tuning queries. The abstention rule distinguishes unresolved explicit ADR IDs and weak matches. Successor expansion only uses headings that explicitly name an ADR clause they supersede. The candidate still needs original source verification.

```sh
python3 -B testdata/context-retrieval/baseline.py --fixture testdata/context-retrieval/fresh-heldout.json
python3 -B testdata/context-lookup/compare.py --fixture testdata/context-retrieval/fresh-heldout.json --variant original --variant abstain --variant vocabulary --variant successor --variant combined
python3 -B testdata/context-lookup/compare.py --fixture testdata/context-retrieval/fresh-heldout.json --input baseline-terms --variant original
python3 -B testdata/context-lookup/compare.py --fixture testdata/context-retrieval/fresh-heldout.json --input baseline-terms --require-literal --variant original
python3 -B testdata/context-retrieval/baseline.py --fixture testdata/context-retrieval/confirmation.json
python3 -B testdata/context-lookup/compare.py --fixture testdata/context-retrieval/confirmation.json --input baseline-terms --require-literal --variant original
```

## Limits

The literal guard was introduced after examining false candidates in the second fixture. Its result in that fixture is a diagnostic, not untouched validation. Authored phrases are stronger inputs than raw questions and are not automatically inferred. A phrase hit only nominates evidence. It cannot establish applicability, especially for similar questions about behavior the product lacks. The runs measure source recovery rather than answer correctness, time to complete a task, or model tokens.
