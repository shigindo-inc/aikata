---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-11
audience: [human, agent]
---

# Design — Doc Map (`docmap`)

> Detailed design for the doc-cartography artifact decided in
> [ADR 0044](../adr/0044-doc-map-derived-artifact.md). For **why**, read the
> ADR; this note records the **how** at design depth so the implementation
> phases can be planned and executed independently.

---

## 1. Problem & responsibility

aikata has no artifact describing the *document set itself* — inventory,
cross-references, freshness, managed/external split. The hand-curated
Navigation Matrix in `AGENTS.md` §3 carries judgement but rots and is not a
clean index.

This feature adds **doc-cartography**, a responsibility distinct from:

- project mission / overview (owned by `README.md` / `SPEC.md` / `AGENTS.md`),
- task→file judgement (owned by the Navigation Matrix),
- source-code structure (out of scope; no source reads),
- runtime context selection / token counting (out of scope; scaffold-time
  tool only).

The map carries only **mechanically derivable facts** about documents.

---

## 2. Artifacts

One scan → two renderings, both under `.aikata/` (machine zone):

- **`.aikata/docmap.yaml`** — data layer / single source of the map's truth.
- **`.aikata/docmap.md`** — readable view (tree + Mermaid link-graph +
  one-line summaries).

Optional additional renderers (`txt`, `json`, `mmd`) are config-gated and
never mandatory.

### 2.1 `docmap.yaml` schema

```yaml
version: 1
generated: 2026-06-11            # date only; Clock-injected for golden stability
docs:
  - path: SPEC.md
    title: "SPEC — What & Why"
    summary: "Describes what aikata does and why."   # best-effort, may be ""
    status: draft                # from frontmatter when present
    updated: 2026-06-01          # frontmatter `updated`, else file mtime
    managed: true                # in ManagedIncludeGlobs(targetDir)
    links:                       # doc→doc references that resolve inside the tracked set
      - ARCHITECTURE.md
      - ROADMAP.md
      - GLOSSARY.md
```

- Edges are carried per-document in `links`; there is no separate `edges`
  section (avoid redundancy).
- The map excludes its own outputs (`docmap.yaml` / `docmap.md`) and the
  generated AI-tool artifacts from the tracked set (self-reference noise).
- `docs` is sorted by `path` for deterministic, diff-stable output.

### 2.2 `docmap.md` rendering

Sections, derived entirely from `docmap.yaml`:

1. **Tree** — the tracked documents as a directory tree, each leaf showing
   `title · status · updated`, with a marker for `managed: false`
   (external) documents.
2. **Relationship graph** — a Mermaid graph of `doc → doc` links. When the
   node count exceeds a degrade threshold (provisional ~40), fall back to a
   flat adjacency list instead of an unreadable diagram (the threshold is
   [Q-DOCMAP-03](./open-questions.md#q-docmap)).
3. **Index** — `path → one-line summary`, the compression payoff: reading
   the map alone conveys the gist of every document without opening each.

---

## 3. Summary extraction (best-effort, graceful degradation)

In order, first non-empty wins:

1. optional `summary:` frontmatter key (never required;
   [Q-DOCMAP-02](./open-questions.md#q-docmap)),
2. leading `>` blockquote (e.g. SPEC/ARCHITECTURE open with
   `> This document describes …`),
3. first non-heading paragraph after the H1,
4. the H1 text,
5. the filename.

This works on documents with no aikata frontmatter, so a repository that
adopts aikata later needs **no document refactor** to be mapped.

---

## 4. Configuration (`.aikata/aikata.yaml`)

```yaml
docmap:
  formats: [yaml, md]        # renderings to emit (default). txt/json/mmd addable.
  targets: ["**/*.md"]       # documents to catalog (default: all Markdown)
  exclude: [".aikata/**"]    # additional skips; generated AI-tool paths always excluded
```

- Matching uses the existing `internal/glob` `Match` / `MatchAny`.
- `managed` is computed via `internal/doctor/scope.go`'s
  `ManagedIncludeGlobs(targetDir)`.
- Default `targets` breadth (all Markdown vs. managed-only first) is
  [Q-DOCMAP-01](./open-questions.md#q-docmap); leading position is all
  Markdown with the `managed` flag distinguishing them.

---

## 5. Triggers & isolation

- `aikata map` — explicit regeneration.
- `init` / `fill` / `enable` / `sync` / `generate` — run the same rebuild
  as a **final, isolated step**, decoupled from per-tool `generate`
  provider failures (a failing provider must not leave the map stale, and
  vice versa). The map step reports its own status.
- `aikata doctor` — freshness check: rebuild in memory, compare the hash to
  the on-disk `docmap.yaml`; mismatch → `warning`; `--fix` regenerates.

`docmap.*` are not manifest-tracked and not subject to `sync`'s 3-way
merge — freshness is guaranteed by the doctor check, not by merge
(ADR 0044 D6).

---

## 6. Reused infrastructure

| Need | Existing | Location |
|---|---|---|
| Frontmatter parse | `parseFrontmatter`, `frontmatterValue` | `internal/doctor/checks.go` |
| Link extraction / resolution | `linkRE` + `checkLinks` resolution logic | `internal/doctor/checks.go` |
| Managed-surface flag | `ManagedIncludeGlobs`, `canonicalDocNames`, `managedDocGlobs` | `internal/doctor/scope.go` |
| Glob match | `Match`, `MatchAny` | `internal/glob/glob.go` |
| Config read/write | `Load`, `Save`, `AikataYaml` | `internal/config/` |
| Freshness hash | `HashContent` | `internal/config/manifest.go` |
| Atomic write | `writeAtomic` | `internal/config/atomic.go` |
| Template render | `Render`, `Helpers` | `internal/templates/render.go` |
| Cobra subcommand | `new<Verb>Cmd` + register in `newRootCmd` | `internal/cli/root.go` |
| CLI exit codes | `&ExitError{Code, Err}` | `internal/cli/errors.go` |
| Golden tests | `testdata/golden/<preset>/`, fixed clock | `internal/doctor/doctor_test.go` |

### Targeted refactor (DRY, scoped to this feature)

`checkLinks` is hard-coded to `AGENTS.md`. Extract frontmatter parsing and
link extraction into a small shared package (e.g. `internal/docmeta`) used
by both `doctor` and `docmap`, so link parsing cannot drift between them.
No unrelated refactoring.

---

## 7. Implementation phases

Each phase is one PR (~≤400 lines of meaningful change), with golden tests
and a fixed clock.

- **P1** — `internal/docmap` scan + `docmap.yaml` rendering; `internal/docmeta`
  extraction.
- **P2** — `docmap.md` rendering (tree + Mermaid graph + index; adjacency-list
  degrade).
- **P3** — `aikata map` command + isolated rebuild hook in `init` / `fill` /
  `enable` / `sync` / `generate`.
- **P4** — `doctor` freshness check (`HashContent` compare; `--fix` regenerate;
  `--json` issue code).
- **P5** — `docmap` config (`formats` / `targets` / `exclude`) + optional
  `txt` / `json` / `mmd` renderers.

---

## 8. Testing

- Golden input projects under `testdata/golden/` produce deterministic
  `docmap.yaml` / `docmap.md` (fixed `generated:` via injected clock).
- Cases: managed + external docs, missing frontmatter, no links, link to a
  document outside the tracked set (dropped), large graph (degrade path),
  self-exclusion of `docmap.*`.
- Doctor: assert the freshness `warning` fires on drift and clears after
  `--fix`.
- Isolation: assert a failing AI-tool provider in `generate` still rebuilds
  the map (and vice versa).

---

## 9. Non-functional

- Stays within ARCHITECTURE §11 budgets (the scan reads document files only:
  frontmatter + first lines + link regex).
- No network I/O. No source-code reads. `text/template` output only.

---

## 10. Open questions

Tracked under [Q-DOCMAP](./open-questions.md#q-docmap): default `targets`
breadth, optional `summary:` frontmatter recognition, and the Mermaid
degrade threshold. None blocks the design.
