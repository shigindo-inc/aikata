---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-17
audience: [human, agent]
---

# Design — CLI/skill responsibility split & watched external doc homes

> Design note (rationale source) for two decisions that ship together:
> a foundational **CLI ↔ skill division-of-labor principle** and its first
> application, **lifecycle observation of foreign-owned ("watched")
> document homes**. The durable records are two ADRs in `docs/adr/`
> (ADR-A principle, ADR-B feature); this note carries the reasoning they
> cite. It mirrors how
> [`v0.9-core-concept-stabilization.md`](./v0.9-core-concept-stabilization.md)
> backs ADR 0028/0033/0034.

---

## 1. Why this exists

Two motivations surfaced from dogfooding aikata alongside other ecosystem
skills (notably `superpowers`, which generates `spec`/`plans` documents,
and `remember`):

1. **aikata's positioning is "bounded-context core in a wider generative-AI
   ecosystem", not "single ecosystem".** It should interoperate cleanly via
   its published contract (`AGENTS.md` + `docs/` layout + skill surface),
   not absorb other skills. (Confirmed against
   [`v0.9-core-concept-stabilization.md`](./v0.9-core-concept-stabilization.md)
   §2/§5: "not a plugin marketplace"; "third-party skill management" is a
   **deferred non-goal**, not a feature to build.)
2. **doc-hygiene is already aikata's core** (`doctor`, archive, ADR 0039).
   Other skills' docs (e.g. `superpowers` plans) go **stale** and nobody
   curates them. aikata can extend its lifecycle *observation* to those
   docs — **without owning, generating, or relocating them blindly** —
   when the user explicitly registers them.

The hard constraint learned the hard way: ADR 0021/0033. When `doctor`
applied aikata's frontmatter contract to a foreign `SKILL.md`, it produced
**62 spurious errors**. So aikata must never apply its own document
contract to foreign docs, and must never validate them by default.

This note settles two coupled decisions (packaged as B from brainstorming:
one design → two ADRs).

---

## 2. ADR-A — CLI ↔ skill responsibility principle (foundational)

The division of labor between the aikata Go CLI and the first-party skills
has been implicit, scattered across ADR 0046 (CLI observation-only, moves
in skill), ADR 0043 D5 (judgement → skill, not a verb), ADR 0037
(mutation boundaries), ADR 0003 (Do-No-Harm). ADR-A states it as one
cross-cutting principle.

| Aspect | aikata CLI (Go binary; deterministic, stack-agnostic, offline) | skill (LLM-driven; judgement) |
|---|---|---|
| Role | **Emits facts** — observation / signal / inventory | **Interprets facts and proposes** |
| Mutation | Only what it owns (canonical scaffold, managed blocks, generated artifacts). Never rewrites foreign/user content beyond managed-append | Judgement-requiring mutation (e.g. `git mv`) **only behind a confirm gate** |
| Network | **None** (no network, no LLM API, no source parse) | May call `gh` etc. (e.g. PR-state interpretation lives here) |
| Foreign docs | **Observe + contract-free warning signals only** | Flexible inspection / confirmation / **recommendation**; actual moves delegated to `migrate-structure` |

**The seam (one sentence):** *the CLI emits offline facts; the skill
interprets them into proposals; the human approves any mutation.*

**Reach of the principle:** every present and future aikata capability is
classified by two tests — *fact or judgement?* and *offline or network?* —
which uniquely place the behavior on the CLI or skill side. ADR-B is the
first dogfooding instance.

---

## 3. ADR-B — Watched external doc homes (the feature)

### 3.1 Data model — `aikata.yaml`, new top-level `watched:`

```yaml
watched:
  - path: "plans/**"     # foreign-owned glob (e.g. superpowers' plans/)
    stale_after: "30d"   # optional; defaults to the global default (§3.4)
```

- **Absent / empty list = feature off.** Do-No-Harm is satisfied by
  absence — nothing is scanned unless the user opts in by listing a path.
  No separate enable flag (mirrors the "empty `doctor.exclude` = no-op"
  idiom).
- **Two fields only** (`path`, optional `stale_after`). `owner` and
  per-entry `checks` overrides are deferred (§5).
- Additive optional section → no schema major bump (within ADR 0016
  schema v2).

### 3.2 CLI side — facts only (C from brainstorming)

1. **`aikata map`** tags each watched doc with **`watched: true`** in the
   doc map. `watched` is the subset of `external` (foreign-owned) docs the
   user opted to observe — a third state alongside `managed` / `external`.
   `map` keeps its descriptive-inventory role (ADR 0044).
2. **`aikata doctor --all-markdown`** (the existing broad-audit mode, the
   explicit non-default scope of ADR 0033 — *not* the default walk, *not*
   `--strict` CI) runs **contract-free lifecycle checks** on watched docs:
   - **age-staleness** — git last-commit date older than `stale_after`
     (git commit date, *not* filesystem mtime, which checkout resets);
   - **broken relative link** — a relative link pointing at a nonexistent
     file;
   - **never** applies aikata's frontmatter contract to watched docs
     (prevents the 62-error regression).
   - Emits machine-readable signals via `--json`.

### 3.3 skill side — interpretation & proposal

- Consumes `map` (*what to watch*) and `doctor --all-markdown --json`
  (*what is stale*) and **recommends** archive / cleanup to the human.
- Any relocation is **delegated to `migrate-structure`'s
  observe → propose → confirm-move gate** (ADR 0046). No new mutation path
  is introduced. Prefer in-place archive over relocation to avoid breaking
  a foreign skill's path coupling.
- Flexible judgement the CLI deliberately omits (e.g. read PR state with
  `gh` to argue "this plan is obsolete") lives here.

### 3.4 Data flow

```text
aikata.yaml watched[]
  → aikata map                     (watched tag — inventory)
  → aikata doctor --all-markdown --json   (age / broken-link signals)
  → skill interprets & recommends
  → human approves
  → migrate-structure git mv       (in-place archive preferred)
```

### 3.5 Defaults

- `stale_after` global default when omitted: **90 days** (conservative;
  favors avoiding nuisance warnings to users of the foreign skill while
  still catching truly abandoned docs). Overridable per entry.
- Watched checks run **only** under `--all-markdown`; never in default
  `doctor` or `doctor --strict`, so the meaning of an existing green run
  is unchanged (ADR 0033 two-sided risk).

---

## 4. Non-goals (explicit)

- ❌ CLI moving / archiving / rewriting a watched doc (CLI has **no**
  mutation path over watched docs).
- ❌ Applying the frontmatter contract to watched docs.
- ❌ Network/API-dependent judgement (PR/issue state) in the CLI.
- ❌ Orphan / unreferenced-doc detection in v1 (noisy for foreign docs;
  deferred pending evidence).
- ❌ Blind scan of other skills' docs (only registered `path`s are read).
- ❌ aikata "adopting" or managing a foreign skill itself (the original
  "manager" model; remains the `v0.9` §2/§5 non-goal "plugin
  marketplace").

---

## 5. Deferred — open questions (to record in `open-questions.md`)

- **Q-WATCH-01** — per-entry `owner` / `checks` override fields: add when
  real use needs finer granularity.
- **Q-WATCH-02** — orphan detection for watched docs: revisit after
  measuring noise.
- **Q-WATCH-03** — an `enable`-style convenience verb to register watched
  homes: add if hand-editing `aikata.yaml` becomes painful.
- **Q-WATCH-04** — depth of skill-side PR-state (`gh`) interpretation:
  its own later spec.

---

## 6. Deliverables and where they land on the managed surface

The durable information must live on aikata's managed document surface,
not only in this note:

| Artifact | Location | Managed surface |
|---|---|---|
| ADR-A (CLI/skill principle) | `docs/adr/NNNN-cli-skill-responsibility-split.md` | ✅ canonical (doctor) |
| ADR-B (watched external homes) | `docs/adr/NNNN-watched-external-doc-homes.md` | ✅ canonical (doctor) |
| `watched` glossary term | `GLOSSARY.md` ("watched (external) document", 3rd state after managed / external) | ✅ canonical |
| Layout triad note | `docs/layout.md` §5 (add watched concept) | ✅ managed |
| Open questions | `docs/decisions/open-questions.md` (Q-WATCH-01..04) | repo-internal |
| This design note | `docs/decisions/` (rationale source) | repo-internal |

Implementation surfaces (`internal/config` schema, `internal/docmap`,
`internal/doctor` scope/checks, the skill content) follow from the
implementation plan; this note is the spec the plan is built from.
