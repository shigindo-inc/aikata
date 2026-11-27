---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-21
audience: [human, agent]
---

# ADR 0051 — Effective context check for nested repositories

- **Status**: Accepted (2026-11-27)
- **Date**: 2026-11-21
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0049 (doctor
  reports what needs action)

## Context

aikata generates one wrapper per agent from `AGENTS.md`, but it never
looks at what an agent actually loads. In a downstream setup, a parent
directory is one git repository and each `apps/*` directory is a
separate git repository inside it. A Claude Code session in a child
loads the child's rules (about 22 KB) plus the parent's `AGENTS.md`
(about 8.7 KB). The parent covers that by hand: one prose section says
which of its sections apply to the children.

How the tools load instruction files (checked against vendor docs on
2026-11-21):

- **Claude Code** loads `CLAUDE.md` and `CLAUDE.local.md` from the working
  directory and **every directory above it**, root first. The walk does not
  stop at a git boundary. `@imports` resolve relative to the importing
  file, up to four hops. The `claudeMdExcludes` setting skips files by
  absolute-path glob.
- **Codex** reads `AGENTS.md` from the project root (normally the git root)
  down to the working directory, and stops adding files at
  `project_doc_max_bytes` (32 KiB by default). That it never reads above
  the git root is an inference; the docs do not say.
- **Cursor**: whether rules from directories above the opened workspace
  apply is not documented.

So the parent's rules leak into child sessions only through Claude Code,
and Codex can silently drop rules once a repository's files pass 32 KiB.

> **確度**: 実測 — the child's `.claude/settings.json` and
> `settings.local.json` exist, and the parent's `AGENTS.md` is about
> 8.7 KB. 二次情報 — the Claude Code loading rules are quoted verbatim from
> code.claude.com/docs/en/memory, but through a subagent. The Codex rules
> came from a summarizing fetch of the vendor page. 推測 — that Codex never
> reads above the git root. 未調査 — Cursor's handling of parent
> directories, and real behavior checked with `/context all`. Before
> implementing, check the Claude Code rows against `/context all` in the
> child repository. The decision itself (report only, no settings writes)
> does not depend on the 推測 or 未調査 items; they only change what the
> Codex and Cursor rows of the report say.

## Decision (proposed)

1. Add `aikata doctor --context`. It reports, for the current directory
   and each enabled AI tool, the instruction files that load at session
   start, in load order: path, bytes, and whether the file lies outside
   the current git repository. It reports bytes, not tokens, because
   aikata does no token counting (SPEC §8). It also reports:
   - paragraphs repeated verbatim across those files (info);
   - files loaded from outside the repository (info, with the
     `claudeMdExcludes` pattern that would skip them). The exclusion
     drops a whole file, so when the parent also holds rules the children
     need, the report says to move those rules into a file the child
     imports before excluding;
   - Codex input over `project_doc_max_bytes` (warning, because the tail
     is dropped without notice).
2. Scope is files that live in the repository and its ancestor
   directories. User-level files (`~/.claude`, `~/.codex`), hooks, plugin
   and skill injections, and transcripts are out of scope; the separate
   transcript-measurement tool covers them.
3. The check reads files only. It does not write `.claude/settings*.json`.
   Those files are hand-maintained, and an exclusion pattern holds an
   absolute path, which must not be committed.

## Consequences

- A child repository can see what its sessions inherit and how to cut it
  off; the parent's rules that children do need then move into a file the
  child imports explicitly.
- aikata gains knowledge of each tool's loading rules. They change
  between vendor releases, so the rules live in one table per tool, each
  row citing the vendor doc it came from.

## Alternatives Considered

- Let `aikata generate` write `claudeMdExcludes` into the child's
  settings. It needs an absolute path and edits a file people maintain by
  hand.
- An `inherits:` key in `AGENTS.md` frontmatter. No agent reads it, so it
  would only document intent; the exclusion has to happen in each tool's
  own setting.
- Cover `~/.claude` too. It ties aikata to one vendor's private layout.
