---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# Glossary

Terminology used across samplekata documentation and source code.
Entries marked **(domain)** are project-specific concepts; the rest are
industry-standard terms whose interpretation is fixed here to reduce
ambiguity for both humans and LLMs.

> **Why this file matters**: pinning vocabulary in one place reduces
> translation drift in LLM output and helps reviewers spot terminology
> mismatches.

---

## A

### ADR — Architecture Decision Record

A short markdown document that captures a single architectural decision,
its context, and its consequences. Stored under `docs/adr/` with the
file name pattern `NNNN-title.md`. Format follows
[`docs/adr/0001-record-architecture-decisions.md`](./docs/adr/0001-record-architecture-decisions.md).

### agent

An LLM-driven coding assistant (Claude Code, Cursor, Codex, Gemini CLI,
Copilot, Windsurf, …) that reads this project's documentation and
produces or edits code.

---

## C

### canonical source

The **single source of truth** for a piece of information. When a
generated artifact diverges from its canonical source, the canonical
source wins.

### Conventional Commits

Commit message convention (`<type>(<scope>): <subject>`). Mandatory in
this project; see [AGENTS.md](./AGENTS.md).

---

## F

### frontmatter

The YAML block at the top of a markdown file delimited by `---`. This
project uses frontmatter for cross-document metadata: `project`,
`status`, `version`, `updated`, `audience`.

---

## (your terms)

_TODO: replace this placeholder with the domain terms that matter to
samplekata._
