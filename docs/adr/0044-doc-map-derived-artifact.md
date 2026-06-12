---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-11
audience: [human, agent]
---

# ADR 0044 - Doc map as a mandatory derived artifact

- **Status**: Accepted
- **Date**: 2026-06-11
- **Deciders**: aikata maintainers
- **Related**: ADR 0039 (documentation hygiene & context budget — this ADR
  is a direct extension of its context-budget thesis), ADR 0003
  (Do-No-Harm), ADR 0002 (`AGENTS.md` canonical), ADR 0033 / ADR 0037
  (doctor managed-surface scope, reused here), ADR 0014 (manifest
  tracking — this ADR deliberately opts the artifact out).

## Context

aikata makes the **document** the unit of truth. Yet there is currently no
artifact that describes the *document set itself* — its inventory, its
cross-references, and its freshness. The closest thing,
[`AGENTS.md`](../../AGENTS.md) §3's **Navigation Matrix**, is a
hand-curated `task → files to read` table. It carries human **judgement**,
which is valuable, but it has three weaknesses as a map of the docs:

1. It is embedded inside `AGENTS.md`, so it is rarely referenced as a
   standalone object.
2. It is hand-written, so it does not follow files as they are added or
   removed — it rots, the exact failure mode aikata exists to prevent.
3. It mixes document references with code-path references, so it is not a
   clean document index.

A separate need surfaced during dogfooding: when an agent (or a human)
opens an unfamiliar aikata repo, orienting across the canonical set plus
`docs/**` costs several first reads before any real work starts. A small,
always-current map of *what documents exist, how they relate, and how
stale they are* would cut that orientation cost — squarely the
context-budget goal of [ADR 0039](./0039-documentation-hygiene-and-context-budget.md).

The risk to guard against is scope creep. aikata must not become a
codebase-analysis tool (AST / dependency graphs) or a documentation-site
generator, and it must keep its **stack-agnostic core** (it reads no
source code). The decision below keeps the feature strictly inside the
document surface.

## Decision

### D1 — A doc-cartography artifact, distinct in responsibility

Introduce a derived artifact whose single responsibility is
**doc-cartography**: the inventory of documents, their mutual references,
their freshness, and a managed/external distinction. It explicitly does
**not** restate the project's purpose or "big picture" (that is the job of
`README.md` / `SPEC.md` / `AGENTS.md`) and does **not** re-curate the
Navigation Matrix. The Navigation Matrix keeps the human **judgement**
("for this task, read X"); the doc map carries only **mechanically
derivable facts**. They are complementary, not competing, so they cannot
drift into two sources of the same truth.

### D2 — Mandatory, machine-zone placement

The artifact lives under `.aikata/` (the aikata-owned machine zone, next
to `aikata.yaml` / `manifest.yaml`) and is **mandatory** — not an opt-in
capability:

- `.aikata/docmap.yaml` — the structured data layer (single source of the
  map's truth).
- `.aikata/docmap.md` — the human/agent-readable view.

Because it is a pure derived artifact in the machine zone, mandatory
inclusion imposes **no maintenance burden** and does not consume a
top-level-minimalism file slot, so it is compatible with the Do-No-Harm
Policy (D8). A short pointer to `.aikata/docmap.md` is added to the
canonical Read order so the hidden location stays discoverable.

### D3 — One scan, two renderings

A single scan of the document surface produces both files. `docmap.yaml`
is the data layer (on-brand YAML, diff-friendly, machine-parseable);
`docmap.md` renders that same data as a directory tree plus a Mermaid
link-graph plus one-line summaries. Both audiences (agent and human) read
either file, so the split is **precise vs. readable**, not
agent-vs-human. `txt` / `json` / `mmd` are optional additional renderers,
configurable, never mandatory (`txt` is a strict subset of `md`).

### D4 — Derived from documents only; stack-agnostic preserved

The map is built from the document surface aikata already understands:
frontmatter metadata, the Markdown link graph, the filesystem tree, and
the manifest. It reads **no source code** and runs **no language-specific
analysis**, so the stack-agnostic core is preserved. The catalog spans
both aikata-**managed** documents (per
[`internal/doctor/scope.go`](../../ARCHITECTURE.md)'s
`ManagedIncludeGlobs`) and **external** user documents, each tagged with a
`managed` flag. Per-document summaries are extracted **best-effort** with
graceful degradation (optional `summary:` frontmatter → leading `>`
blockquote → first body paragraph → H1 → filename), so a repository
adopting aikata after the fact needs **no refactor of its existing
documents** to be mapped.

### D5 — Triggers: explicit verb plus isolated rebuild step

Refreshing the map follows trigger model **C**:

- `aikata map` regenerates it on demand.
- Every command that mutates the document surface (`init`, `fill`,
  `enable`, `sync`, `generate`) runs the same map rebuild as a **final,
  isolated step**.

The rebuild is **decoupled from per-tool `generate` failures**: a failing
AI-tool provider must never leave the map stale, and vice versa. The map
step reports its own status independently.

### D6 — Not manifest-tracked; freshness guaranteed by doctor

`docmap.yaml` / `docmap.md` are **not** recorded in `.aikata/manifest.yaml`
and are **not** subject to `aikata sync`'s 3-way merge (ADR 0014 / 0025).
They are aikata-owned derived state, regenerated rather than merged. To
catch drift when a document changed without a rebuild, `aikata doctor`
gains a **freshness check**: it rebuilds the map in memory and compares the
hash to the on-disk `docmap.yaml`; a mismatch is a `warning`, fixable with
`aikata doctor --fix` (which regenerates it).

### D7 — Configurable surface and formats

`.aikata/aikata.yaml` gains a `docmap` section: `formats` (default
`[yaml, md]`), `targets` (default `["**/*.md"]`), and `exclude` (default
includes `.aikata/**` and generated AI-tool artifacts, so the map never
catalogs itself). The matcher is the existing
[`internal/glob`](../../ARCHITECTURE.md) `MatchAny`. Users can widen or
narrow the tracked set after the fact.

### D8 — Do-No-Harm

The artifact is aikata-owned derived state with zero maintenance
obligation, it does not touch the visible top-level surface, and it reads
no source code. A downstream project that ignores `.aikata/docmap.*` pays
nothing; nothing about the feature penalizes non-adopters (ADR 0003).

## Consequences

**Positive**:

- An always-current, low-cost map of the document set: one read orients an
  agent or human to *what exists, how it links, and how fresh it is*.
- The Navigation Matrix can, over time, shrink to pure judgement because
  the inventory now lives in a derived artifact.
- Reuses existing infrastructure (frontmatter parser, link walker,
  managed-surface scope, glob matcher, atomic write, content hashing), so
  the net new surface is small.

**Negative**:

- A new mandatory pair of files in every aikata project's `.aikata/`. They
  are committed (like the manifest), so they appear in diffs on every
  document change. Accepted: the diff *is* the value (it shows how the map
  moved).
- `doctor` does slightly more work (an in-memory rebuild for the freshness
  check). Bounded by the existing per-file document walk; well within the
  §11 performance budget.
- Coupling every surface-mutating command to a rebuild step adds a code
  path that must stay isolated from other failures (D5); enforced by
  tests.

## Alternatives Considered

- **Opt-in capability under `docs/`** (e.g. `enable docmap` →
  `docs/map.md`). Rejected by the maintainer: the map is machine-derived
  state, not a hand-maintained canonical document, so it belongs in the
  machine zone and should always be present, not behind a flag.
- **Regenerate only inside `aikata generate`.** Rejected: per-tool
  provider errors could leave the map stale, the exact failure the
  maintainer wanted to avoid (D5).
- **A single Markdown file only** (no YAML data layer). Rejected: a
  structured data layer is the clean source for the renderings and for any
  future programmatic consumer; YAML is on-brand and diff-friendly.
- **Read source code to build a structural/dependency map.** Rejected: a
  different tool category that would break the stack-agnostic core and the
  "no source reads" boundary (SPEC §8).
- **Generate an actual documentation website / sitemap.** Rejected: a
  Markdown tree satisfies the "sitemap-like" intent without making aikata
  a static-site generator.

## References

- ADR 0039 (documentation hygiene & context budget — the thesis this
  extends).
- ADR 0003 (Do-No-Harm Policy — D8 compliance).
- ADR 0033 / 0037 (doctor managed-surface scope — reused for the
  `managed` flag).
- ADR 0014 / 0025 (manifest tracking & sync preservation — deliberately
  opted out in D6).
- Design note:
  [`docs/decisions/docmap-design.md`](../decisions/docmap-design.md).
- Open questions:
  [Q-DOCMAP](../decisions/open-questions.md#q-docmap).
