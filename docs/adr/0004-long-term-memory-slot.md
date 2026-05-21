---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# ADR 0004 — Long-term Memory Slot under `docs/memory/`

- **Status**: Accepted
- **Date**: 2026-05-21
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0003 (Do-No-Harm Policy);
  [`docs/decisions/open-questions.md`](../decisions/open-questions.md) Q-DESIGN-07

## Context

aikata so far recognized two agent-facing artifacts:

| Artifact | Lifetime | Author | Purpose |
|---|---|---|---|
| `AGENTS.md` | Long | Human (canonical) | **Rules** — invariant operating constraints |
| `docs/tasks/current.md` | Short | Agent (frequent) | **Working memory** — current task state |

These two together leave a gap. Real cooperation with LLM agents produces
a third class of artifact: **long-term memory** — accumulated facts about
the user, validated preferences, ongoing project context, and stable
references to external systems. This material:

- is **not** an invariant rule (so it does not belong in `AGENTS.md`);
- is **not** ephemeral working state (so it does not belong in
  `docs/tasks/current.md`);
- must survive across sessions and across agent restarts;
- must be readable by humans and by every agent without translation.

The industry has already converged on a pattern for this:

- **Claude Code** uses project-level `CLAUDE.md` plus a session-scoped
  memory under `.claude/projects/<hash>/memory/` with type-tagged files
  (`user_*.md`, `feedback_*.md`, `project_*.md`, `reference_*.md`).
- **Cursor** writes long-form rules under `.cursor/rules/` (mixing rules
  and memory).
- **superpowers** plugin standardizes a `memory/` directory with the
  same four types listed above.

Without an explicit slot in the aikata scaffold, agents either (a) write
preferences into `AGENTS.md` (polluting rules with mutable facts) or
(b) lose them between sessions (forcing the user to re-state them).
Both outcomes degrade the human-LLM cooperation aikata exists to enable.

## Decision

Add a **dedicated long-term memory slot** to the aikata scaffold:

### 1. Location and shape

```
docs/memory/
├── README.md
├── user.md
├── feedback.md
├── project.md
└── reference.md
```

- Flat layout. One file per memory type. No subdirectories before v1.x.
- Filenames are fixed and singular. Custom types are not supported in v1.0.

### 2. Memory types

| Type | Holds | Example |
|---|---|---|
| `user` | Profile, role, knowledge, preferences of the **user** | "Native Japanese speaker", "Flutter-leaning OSS developer" |
| `feedback` | Continuing **instructions** from the user (corrections + validated approaches) | "No AI signatures in commits", "Run `make test && make lint` before commit" |
| `project` | **Ongoing context** of the project that is not derivable from code or git log | "v0.1 MVP excludes Flutter preset", "Phase 2 dependency budget: cobra only" |
| `reference` | Pointers to **external systems** | GitHub repo URL, agents.md spec link, dashboard URLs |

These four types match Claude's auto-memory taxonomy and the superpowers
plugin's convention. We deliberately do **not** invent aikata-specific
names: agents trained on the existing pattern transfer directly.

### 3. File format

Every memory file carries the standard frontmatter **plus** a new key
`memory_type`:

```yaml
---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
memory_type: user|feedback|project|reference
---
```

Body: a flat bullet list. New entries append at the bottom. Old entries
are not deleted; they are marked **(superseded YYYY-MM-DD)** in place.
This makes diffs auditable and lets `aikata doctor` flag stale entries
without losing history.

### 4. Boundary with `AGENTS.md`

| Question | Goes to |
|---|---|
| "Must this always be true regardless of context?" | `AGENTS.md` |
| "Did the user prefer X in the past, and is X likely still preferred?" | `docs/memory/feedback.md` |
| "Is this a fact about who the user is?" | `docs/memory/user.md` |
| "Is this a one-off, in-flight work state?" | `docs/tasks/current.md` |

When an `AGENTS.md` rule and a `memory/feedback.md` entry conflict,
`AGENTS.md` wins (rules trump preferences).

### 5. Do-No-Harm compliance

- **Default off in user projects.** `aikata init --preset standard`
  does **not** create `docs/memory/`. The directory appears only when
  `aikata init --with-memory` is passed (flag implemented in v0.2).
- **Zero residue when absent.** `AGENTS.md` references to memory are
  conditional ("if present"), so a project without `docs/memory/`
  produces no broken links.
- **Inert to outsiders.** Memory files are plain markdown; users without
  aikata still read them normally.
- **aikata's own repository** seeds `docs/memory/` immediately as
  **dogfooding**. This is not the default behavior for downstream
  projects; it is a one-off bootstrap because:
  - The aikata maintainers have already accumulated memory worth
    persisting (no-AI-signature rule, real-name copyright, etc.).
  - It surfaces memory-format edge cases before v0.2 implementation.

### 6. Scope of this ADR

This ADR establishes the **slot** (γ). It does **not** decide how
`aikata generate` will project memory into tool-specific memory
channels (Claude `.claude/memory/`, Cursor long-form rules, etc.); that
is option (δ) and is recorded as Q-DESIGN-07 in `open-questions.md`
pending real-world experience with the slot.

## Consequences

**Positive**:

- Removes the impulse to bloat `AGENTS.md` with mutable preferences.
- Gives agents a deterministic place to write learnings, reducing the
  "asks the same question twice" failure mode.
- Aligns aikata with the emerging cross-tool memory pattern, easing
  Claude / Cursor / superpowers interop later.
- The `dogfooding seed` immediately makes the aikata repo more useful
  to agents working on it (they can read `feedback.md` to learn the
  no-AI-signature rule rather than rediscovering it from commit log).

**Negative**:

- Adds 6 files (5 memory + 1 ADR) to the repository in one commit.
  Mitigated by the value of dogfooding and the fact that all files are
  plain markdown.
- Memory drift risk: entries may become stale faster than `AGENTS.md`.
  Mitigation: each entry has a date suffix; `aikata doctor` will check
  staleness in a later release.
- Boundary calls (rule vs. preference) require judgment. Mitigation:
  Section 4's decision matrix; precedent collected in `docs/memory/`
  itself.

## Alternatives Considered

- **(a) Do nothing.** Rejected: pushes mutable facts into `AGENTS.md`
  or loses them entirely; misses an emerging cross-tool convention.
- **(b) Single `docs/memory.md` file.** Rejected: mixes four concerns
  (user / feedback / project / reference) that have different lifetimes
  and different audiences; doctor checks become harder.
- **(c) Five files with custom names** (`who.md`, `corrections.md`, …).
  Rejected: foregoes interop with existing tools' taxonomies.
- **(d) Generate-only**, no canonical slot — author memory in tool-
  specific formats and merge on read. Rejected: violates the canonical-
  source principle (ADR 0002) and forces aikata to know every tool's
  memory format up front.

## Verification

- `aikata doctor` (v0.2+) will check:
  - every memory file has the required `memory_type` frontmatter key;
  - `memory_type` matches the filename;
  - no entry older than 365 days lacks a `(superseded …)` mark.
- The `aikata init --with-memory` golden test (v0.2) will assert that
  the five files are produced with valid frontmatter.
- This repository's own `docs/memory/` serves as a manual reference
  fixture until the golden test exists.
