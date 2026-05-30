---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# Testing

Test strategy and tooling for samplekata. Keep this file scoped
to *what is tested at which layer and how* — the runbook for adding
a new test, not a tutorial on testing in general.

## Why this matters for AI collaboration

Tests are the cheapest objective signal that a change actually works.
For AI agents the economics shift further: writing tests is no longer
the slow part, while a fast red-green loop lets an agent iterate and
self-verify without a human in the loop. Treat a green run — not "looks
right" — as the bar for "done".

> **Recommendation (opt-in):** for code with clear inputs and outputs,
> consider test-first — write the failing test, then the code. It pins
> the intended behaviour before the implementation drifts. This is a
> recommendation, not a rule; skip it where it does not fit (spikes,
> exploratory work, throwaway prototypes).

## Strategy

TODO — name the test layers (unit / integration / contract / e2e),
state what each is responsible for, and where the boundaries are.

## Tooling

TODO — list the test runners, fixtures, golden-file harness, CI
configuration, and any project-specific assertions the team relies on.
