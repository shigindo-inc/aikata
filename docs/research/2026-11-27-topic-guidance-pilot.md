---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# Source-linked guidance route replay

> **Summary**: Two staged source-linked cards preserved all required evidence and reduced bytes compared with a recorded extra navigation lookup. Both routes opened the same source sections. The cards add bytes when the mandatory material already points directly to the sections.
> **Read when**: Planning the field comparison in [#179](https://github.com/shigindo-inc/aikata/issues/179).
> **Status**: Complete for route replay, open for field validation.
> **Confidence**: High for fixture results. Field impact is unmeasured.
> **Adoption**: Deferred pending evidence from independent, opted-in work.

## Question

Can a short, source-linked guide reduce discovery payload while keeping the original decision clauses available?

## Result

| Staged question | Required sections, both routes | Extra navigation bytes, current → guided | Guide bytes if the extra lookup is already unnecessary |
|---|---:|---:|---:|
| Accept one resolved sync conflict | 2/2 | 817 → 274 | +274 |
| Determine the scope of the former skill-command rule | 2/2 | 1,746 → 536 | +536 |

## Evidence and reproduction

The two questions, excerpts required for correctness, and both routes are pinned in [cases.json](../../testdata/context-guidance/cases.json). The replay confirms that the current navigation command actually surfaces the expected references, then checks required fragments in the opened sections.

```sh
python3 -B testdata/context-guidance/compare.py --check
```

Both arms open the same number of files and sections and include the same fixture-defined shared reads. The comparison counts UTF-8 bytes, not model tokens, time, or user effort. It uses two staged questions and does not establish recurrence. A guide may be unnecessary if existing instructions, navigation or search are already sufficient.

## Opt-in field procedure

1. A maintainer and participants agree on the repository, one question family and an observation window. Use a local file outside the checkout. Freeze the applicability rule, required source sections and review criteria before seeing outcomes. Do not install a collector.
2. Allocate random work IDs for independent work and keep them for resumes and child work. Record stage exposure even when no search happens. With consent, sample candidate-positive, candidate-negative and unclassified work. Inspect whether cases are equivalent and distinguish intentional verification from repeated discovery.
3. Use the existing route for one arm and a temporary source-linked card for the other. Keep required reads and task types comparable. Use different independent work in each arm to avoid learning the answer by repeating a task. Assign by a predeclared alternating order within each task type and review evidence recovery without the route label where practical.
4. Store only documented codes, numeric measurements and opaque IDs. Verify the structured observations using the [observation fixture contract](../../testdata/context-observation/README.md). Keep measurements and consent local. Report only reviewed aggregates with denominators, missing results and exclusions.
5. Report required evidence, wrong-state claims, avoidable re-search and time to evidence before drawing an adoption conclusion. If only a small sample is available, report counts and limitations. If the same issue reappears across independent work but evidence is incomplete or conflicting, improve the canonical source before introducing an additional guide.

## Related

[Delivery plan](../plans/2026-11-27-task-scoped-document-retrieval.md) · [Observation issue #178](https://github.com/shigindo-inc/aikata/issues/178)
