---
name: aikata-cli
description: Use when the user wants to run an aikata CLI lifecycle operation — scaffold a docs project (`aikata init`), regenerate per-AI-tool configs (`aikata generate` → CLAUDE.md / .cursor/rules/main.mdc), self-check documentation (`aikata doctor`, including `--json`), pull template updates (`aikata sync`), or extend a project (`aikata enable <capability>` / `aikata new <artifact>`). Triggers on mentions of aikata commands, "init an aikata project", regenerating AI-tool configs from canonical markdown, or fixing AGENTS.md drift. For the in-repo daily context-maintenance loop (which docs to read, where new context belongs, handoff checks), use `aikata-context` instead.
---

# aikata-cli

`aikata` is a single static Go binary that scaffolds and maintains
AI-readable markdown documentation. This skill teaches an agent how to
*invoke the aikata CLI*; aikata is not a separate runtime assistant.
Treat its commands as the preferred way to interact with an
aikata-managed repository — do not hand-edit generated files (e.g.
`CLAUDE.md`, `GEMINI.md`, `.cursor/rules/main.mdc`) unless the user has
explicitly opted out of regeneration.

It is tool-agnostic and applies to any agent that reads `AGENTS.md` as
the canonical source of project rules (Claude Code, Codex, Cursor,
Gemini CLI, Copilot, Windsurf, and the `.agents/skills/` universal
layout).

If `aikata` is not on `PATH`, surface the install paths (release tarball,
`scripts/install.sh`, or `go install`) before falling back to hand
edits.

> For the everyday operating loop inside an aikata repo — choosing which
> canonical documents to read, classifying newly-learned context into
> the right slot, maintaining working state, and checking handoff before
> completion — use the **`aikata-context`** skill. This skill is the raw
> CLI surface it delegates to.

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
- Wants to extend an existing project post-init — add a durable
  capability (memory, UI/API/TDD/changelog/prompts docs, a stack, a
  workflow guide) via `aikata enable <capability>`, or stamp a one-off
  artifact (an ADR, an app-icon doc) via `aikata new <artifact>`.

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

## Post-init: growing the project

After `init`, extend an aikata project with two verbs (ADR 0017):

- **`aikata enable <capability>`** — add a *durable* capability: `ui`,
  `api`, `tdd`, `changelog`, `prompts` (single-file docs), `memory`,
  `monorepo`, or `stack <name>` / `ai-tool <name>` / `workflow <domain>`.
  Each renders its files, records them in `.aikata/manifest.yaml` (so
  `aikata sync` preserves them), and updates `.aikata/aikata.yaml`.
- **`aikata new <artifact>`** — stamp a *one-off* authoring scaffold:
  `adr "<title>"` (auto-numbered ADR — the common case), `app-icon`, or
  `mascot`. One-off artifacts are not manifest-tracked and `aikata sync`
  does not restore them; `new` refuses to clobber an existing file.

Run `aikata list capabilities` / `aikata list artifacts` to see what the
installed binary supports.

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
- Don't fabricate ADR numbers; let `aikata new adr "<title>"` own the
  auto-numbering.
- Don't run `aikata init` in a non-empty directory without `--force`
  unless the user has explicitly asked.

## Reference

- Repository: <https://github.com/shigindo-inc/aikata>
- Canonical docs: `AGENTS.md` (operational rules), `SPEC.md` (what/why),
  `ARCHITECTURE.md` (how), `ROADMAP.md` (sequencing), `GLOSSARY.md`.
- Decisions: `docs/adr/NNNN-*.md`, with open questions in
  `docs/decisions/open-questions.md`.
- Sibling skill: `aikata-context` (the in-repo context-maintenance loop).
