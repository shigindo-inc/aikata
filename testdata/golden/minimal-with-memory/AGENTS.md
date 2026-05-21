---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: agent
---

# Agent Instructions for samplekata

## 1. Project overview

See [SPEC.md](./SPEC.md) for the what and why.

## 2. Before you start

Read in this order:

1. [README.md](./README.md) — overview
2. **This file (AGENTS.md)** — operating rules
3. [SPEC.md](./SPEC.md) — requirements
4. [`docs/memory/`](./docs/memory/) — long-term context (user
   preferences, project notes); read
   [`feedback.md`](./docs/memory/feedback.md) before non-trivial work

## 3. Hard rules

- Never commit secrets. Reference `.env.example` patterns instead.
- Update [SPEC.md](./SPEC.md) when requirements change.
- Use [Conventional Commits](https://www.conventionalcommits.org/).
- No AI signatures in commits.

## 4. When stuck

Document the question in the PR description or commit body and surface
it to the maintainer rather than silently guessing on an architectural
point.
