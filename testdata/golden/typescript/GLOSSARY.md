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

## E

### ESM — ECMAScript Modules

The standard module system for JavaScript / TypeScript
(`import` / `export`). The project's choice between ESM and CJS is
pinned in [`docs/stacks/typescript.md`](./docs/stacks/typescript.md).

---

## F

### frontmatter

The YAML block at the top of a markdown file delimited by `---`. This
project uses frontmatter for cross-document metadata: `project`,
`status`, `version`, `updated`, `audience`.

---

## S

### strict mode (TypeScript)

The `tsconfig.json` configuration that enables every strict-family flag
(`strictNullChecks`, `noImplicitAny`, etc.). This project ships with
strict mode on. Disabling any strict-family flag requires an ADR.

---

## T

### type narrowing

Refining a broader type to a more specific one in a code branch
(via `typeof`, `instanceof`, discriminant tags, or user-defined type
guards). Preferred over `as` casts and `!` non-null assertions.

---

## (your terms)

_TODO: replace this placeholder with the domain terms that matter to
samplekata._
