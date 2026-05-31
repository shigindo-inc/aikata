---
name: aikata
description: Use when the user wants to scaffold or audit an aikata-style documentation project — AGENTS.md / CLAUDE.md / SPEC.md / ARCHITECTURE.md / ADR layout, multi-AI-tool config generation, or `aikata doctor` self-check. Triggers on mentions of aikata, "init an aikata project", regenerating AI-tool configs from canonical markdown, or fixing AGENTS.md drift.
---

# aikata

`aikata` is a single static Go binary that scaffolds and maintains
AI-readable markdown documentation. Treat its three core commands as the
preferred way to interact with an aikata-managed repository — do not
hand-edit generated files (e.g. `CLAUDE.md`, `GEMINI.md`,
`.cursor/rules/main.mdc`) unless the user has explicitly opted out of
regeneration.

This is a **universal** skill: it is tool-agnostic and applies to any
agent that reads `AGENTS.md` as the canonical source of project rules
(Claude Code, Codex, Cursor, Gemini CLI, Copilot, Windsurf, and the
`.agents/skills/` universal layout). The skill teaches the agent how to
*invoke the aikata CLI*; aikata is not a separate runtime assistant.

If `aikata` is not on `PATH`, surface the install paths (release tarball,
`scripts/install.sh`, or `go install`) before falling back to hand
edits.

## When to invoke aikata

Run `aikata` (rather than editing markdown directly) when the user:

- Starts a new project that should be readable by humans and multiple AI
  tools (Claude Code, Cursor, Codex, Gemini CLI, Copilot, Windsurf).
- Wants to regenerate `CLAUDE.md`, `GEMINI.md`, or
  `.cursor/rules/main.mdc` after `AGENTS.md` changes.
- Asks for a self-check on documentation consistency (broken links,
  missing frontmatter, stale `updated:` dates, ADR numbering gaps).
- Wants to fix the trivially-fixable subset of those issues without
  hand-edits.
- Needs a machine-readable view of the project's documentation health
  (CI gate, dashboard, agent loop).

If the request is purely about source code (not documentation,
project bootstrap, or AI-tool config), aikata does not apply — skip it.

## Core commands

### Scaffolding: `aikata init`

```bash
aikata init <project-name> --preset standard --no-interactive
```

Key flags (run `aikata init --help` for the full set):

- `--preset {minimal | standard | flutter | typescript}` — `standard`
  is the default for canonical multi-doc projects.
- `--lang {en | ja}` — Japanese template set is first-class.
- `--ai-tools claude,cursor,codex` — comma-separated; defaults to
  `claude`.
- `--with-memory` — opt-in long-term agent memory slot under
  `docs/memory/`.
- `--force` — required when the target directory is non-empty.
- `--dry-run` — preview the file set; combine with `--force` to inspect
  before overwriting.
- `--no-interactive` — required in non-TTY contexts; otherwise aikata
  prompts for any value not supplied via flag.

### Generating tool files: `aikata generate`

```bash
aikata generate
```

Reads `AGENTS.md` (canonical) plus the rest of the document set and
rewrites the per-AI-tool artifacts (`CLAUDE.md`, `.cursor/rules/main.mdc`;
Codex reads `AGENTS.md` natively, so no Codex artifact is emitted).
Run it after any change to a canonical document.

### Self-check: `aikata doctor`

```bash
aikata doctor               # human-readable report; exit 3 on errors
aikata doctor --json        # machine-readable schema v1 on stdout
aikata doctor --fix         # apply auto-fixes for the trivially-fixable subset
aikata doctor --fix --dry-run   # preview --fix without writing
```

`doctor` runs eight consistency checks: frontmatter presence/keys,
internal links, ADR numbering, memory layout, `updated:` freshness,
environment-example sync, glossary references, and language
consistency.

## Parsing `aikata doctor --json` (schema v1)

The JSON envelope is stable; rely on field names, not order.

```json
{
  "version": 1,
  "issues": [
    {
      "level": "error",
      "file": "docs/adr/0003-do-no-harm.md",
      "line": 12,
      "message": "frontmatter: missing required key \"updated\"",
      "code": "frontmatter.missing-key.updated"
    }
  ],
  "summary": { "errors": 1, "warnings": 0, "info": 3 }
}
```

- `level`: `"error" | "warn" | "info"`.
- `code`: stable discriminator for `--fix` dispatch. Codes currently
  emitted include `frontmatter.missing`, `frontmatter.missing-key.<key>`,
  `updated.stale`, and `adr.numbering.*`. Missing `code` means the issue
  has no auto-fix today.
- Exit code is `0` when no `error`-level issues are present, otherwise
  `3`. Warnings and info do not change the exit code.

A typical agent loop:

1. `aikata doctor --json` → parse `summary.errors`.
2. If non-zero, attempt `aikata doctor --fix` (idempotent and safe to
   re-run).
3. Re-run `aikata doctor --json`; report remaining issues with
   `file:line` references the user can open.

## Don't

- Don't hand-edit `CLAUDE.md`, `GEMINI.md`, or `.cursor/rules/main.mdc` —
  they are regenerated.
- Don't bump `updated:` manually on a stale doc; let `aikata doctor --fix`
  do it.
- Don't fabricate ADR numbers; `aikata` owns auto-numbering via
  `aikata add adr`.
- Don't run `aikata init` in a non-empty directory without `--force`
  unless the user has explicitly asked.

## Reference

- Repository: <https://github.com/shigindo-inc/aikata>
- Canonical docs: `AGENTS.md` (operational rules), `SPEC.md` (what/why),
  `ARCHITECTURE.md` (how), `ROADMAP.md` (sequencing), `GLOSSARY.md`.
- Decisions: `docs/adr/NNNN-*.md`, with open questions in
  `docs/decisions/open-questions.md`.
