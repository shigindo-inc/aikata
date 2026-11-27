---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-27
audience: [human, agent]
---

# Task-scoped document retrieval and recurring questions

> **Summary**: Prototype an explicit read-only lookup of authoritative Markdown sections. Evaluate local recurring-question counts separately before adding topic guidance or committing to a CLI.
> **Read when**: Designing or evaluating [#176](https://github.com/shigindo-inc/aikata/issues/176) through [#179](https://github.com/shigindo-inc/aikata/issues/179).
> **Status**: Draft
> **Open**: Public command and schema, search weights, bilingual aliases, pilot thresholds and release scope.

## Decisions in this draft

- Keep mandatory reads in AGENTS.md and other applicable instructions independent of retrieval and byte budgets.
- Keep ADR filenames and the original decision text. Results and topic guides link to evidence. Similarity and ordinary links do not establish supersession.
- Start with version-pinned benchmark fixtures and manual observation. The experiment adds no always-on collection, model call, repository write, or generated index dependency.

## Design

### Retrieval experiment

Request fields are a nonblank query, optional document type or current/history hint, and a result/byte budget. Compare an initial limit of five candidates and 2,400 UTF-8 bytes with the manual baseline. Treat both values as experiment settings. The caller chooses when to open a section or full file.

Each candidate carries repository-relative path, revision (or working-tree identifier), heading ancestry, inclusive 1-based line range, and match reason. ADR decision class is optional and comes from the ADR parser where available. An explicit ID includes historical candidates regardless of ranking hints. A duplicate ID returns ambiguity. A stale locator requires a new lookup. A missing or ambiguous result is explicit, with no inferred decision.

Use an in-memory scan of the explicitly selected repository surface. Reuse managed-document scope and document parsing where their contracts fit. Check exclusions and repository boundaries before opening files. Reject explicit out-of-scope, nonregular or symlink paths and surface malformed configuration. Search exact identifiers first, then weighted title and heading terms, short literal matches and a stable path/line tie-break. Evaluate JA/EN aliases with held-out examples. Expand at most one hop during the experiment, and keep status relationships distinct from ordinary references. Reserve output space for provenance and truncation state. An impossibly small budget returns a diagnostic.

Share ADR status parsing with the upstream inventory. The inventory's repository-wide reference count is unnecessary for a lookup. The doc map's frontmatter status and path-only links cannot represent ADR decision state or section ranges. Extract portable helpers only where tests demonstrate parity. Test CRLF, fenced headings, repeated headings, boundaries, stable ranking, budget and supersession. Isolate any change to existing parser behavior from parity refactoring.

### Observation experiment

Use explicit local enrollment and begin with manual events. A versioned event contains opaque observation-window work and event IDs, a week, phase, controlled question ID when known, and an enumerated event/outcome. Phase exposure has no question ID. IDs stay stable through retries, resumes and child work. Reject unsupported schema, invalid fields and inconsistent duplicates. Exact retries are idempotent. Keep only numeric durations/counters if justified by a declared metric. Do not accept raw prompts, transcript text, absolute paths or session identifiers.

Count a question at most once per work/phase/question. Topic totals re-union across phases. Occurrence rate uses distinct works observed passing the phase, including no-search work. Repeat-search rate uses question-positive works with observed outcomes. Show missing exposure, unknown outcomes and sparse samples separately. A legitimate verification reread is not evidence of avoidable friction. Review samples from candidate-positive, candidate-negative and unclassified work. Report false merges separately from missed equivalents, with denominators per stratum.

For example, two attempts in the same work and phase count as one occurrence. If two works expose the phase and one asks the question, occurrence is 1/2. If that work's outcome is unobserved, repeat-search rate is unavailable. Check these cases in synthetic fixtures before using real opt-in observations. kansoku can consume a local aggregate later. It is not a prerequisite for the manual trial.

### Guidance pilot

Nominate recurring questions across independent works and weeks. Confirm avoidable friction and stable source material before publishing a short guide in an existing location. Compare matched tasks against the baseline with required evidence fixed in advance. Inspect evidence recall and mistaken supersession before comparing rereads, time to evidence and task-level input by cache category. Revisit guidance with poor results or excessive upkeep. Record ship, revise or stop in [#175](https://github.com/shigindo-inc/aikata/issues/175).

## Constraints and premises

| Confidence | Premise |
|---|---|
| Verified | Workspace is v0.15.3. Public v0.16.0 defaults doc-map maintenance off. Upstream main at `8cdf2be` contains ADR inventory and startup context diagnostics. Neither supplies section lookup. |
| Secondary | kansoku's prior Read count is a within-transcript repeated-path proxy. It does not measure recurrence across work items or aikata-specific savings. |
| Hypothesis | Bounded section lookup and source-linked guidance can reduce discovery friction without losing required evidence. |
| Unmeasured | Held-out JA/EN quality, classification accuracy, end-to-end token and time savings, and maintenance effort. These determine the ship decision. |

Before stabilizing a public command or adding runtime selection, revisit [SPEC §8](../../SPEC.md#8-out-of-scope-examples-for-clarity) and [ADR 0045](../adr/0045-documentation-value-model.md). Follow [Do-No-Harm](../adr/0003-do-no-harm-policy.md) for optional features.

## Related

[Delivery plan](../plans/2026-11-27-task-scoped-document-retrieval.md) · [Upstream comparison](https://github.com/shigindo-inc/aikata/compare/v0.16.0...8cdf2be55b7f394339db82113203d37696c2723d) · [Startup diagnostic follow-up](https://github.com/shigindo-inc/aikata/issues/174)
