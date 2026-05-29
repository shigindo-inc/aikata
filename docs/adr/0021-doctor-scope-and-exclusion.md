---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-29
audience: [human, agent]
---

# ADR 0021 - Doctor scope and exclusion semantics

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0015 (first-party
  skill / plugin distribution), ADR 0016 (aikata.yaml schema v2),
  Q-DOCTOR-01

## Context

`aikata doctor`'s `checkFrontmatter` enforces the aikata
five-key frontmatter (`project`, `status`, `version`, `updated`,
`audience`) on **every** `.md` file under the target directory.
The only escape hatches today are two hardcoded sets in
`internal/doctor/checks.go`: `skippedDirs` (directory names like
`.git`, `dist`, `node_modules`, `.cursor`, …) and `skippedFiles`
(file names like `CLAUDE.md`, `GEMINI.md`). Users have no way to
extend or customise either list.

That works fine for greenfield aikata projects. It breaks for users
who layer aikata onto a repository that also contains files governed
by a **different markdown frontmatter spec**, in particular:

- Claude Code plugin layouts at `plugins/<name>/.claude-plugin/`,
  with `plugins/<name>/skills/<name>/SKILL.md` whose frontmatter is
  defined by Anthropic's spec as `name` + `description` only.
- The same shape under `plugins/<name>/agents/<name>.md`.
- Future Codex / Gemini-CLI plugin layouts likely to follow the
  same per-tool frontmatter contract.

A real report from `personal-skills` (aikata v0.6.1) surfaced
62 spurious `frontmatter missing required key` errors against a
single Claude Code plugin tree, blocking the user from putting
`aikata doctor` on a CI gate. The user's escape options were all
unsatisfying:

1. Add aikata's five keys on top of Claude Code's two. The combined
   block parses today but couples two specs that may diverge.
2. Suppress doctor errors in CI (defeats the purpose of the gate).
3. Reshape the project to hide the plugin tree (a workaround, not
   a fix).

aikata itself does not hit this on its own repo only because
`dist/` is in the hardcoded `skippedDirs` — `dist/claude-code/skill/SKILL.md`
is exempted **by name coincidence**, not by an exclusion contract.
That is a blind spot: aikata exempts its own first-party plugin
artefact but provides no equivalent for users.

The Do-No-Harm Policy (ADR 0003) frames the principle:
aikata-managed files belong to aikata; user-owned files belong to
the user. The same principle extends to linting — aikata should
verify its own contract and stay out of subtrees governed by a
third-party contract.

## Decision

`aikata doctor` gains a configurable exclusion list driven from
`.aikata/aikata.yaml`. A new optional top-level `doctor:` block
holds an `exclude:` glob list. Paths matching any of those globs
are skipped at the markdown-walk layer and therefore exempted from
`checkFrontmatter`, `checkUpdated`, and `checkGlossary`
consistently.

### Schema (additive to schema v2)

```yaml
doctor:
  exclude:
    - "plugins/**"
    - "**/.claude-plugin/**"
    - "**/SKILL.md"
```

- `doctor:` is **optional**. Omitting it (or writing
  `doctor: {}` / `doctor.exclude: []`) preserves the pre-v0.7.3
  behaviour exactly.
- The block is declared with `omitempty` on the Go struct so
  golden-test fixtures stay byte-identical when the value is unset.
- Schema migration: the v1 → v2 migrator in
  `internal/config/schema_migrate.go` is **unchanged**. The new
  block is purely additive; v1 projects keep migrating to v2
  without ever materialising a `doctor:` block.
- Configuration-file shape: kept under `doctor:` (mirrors the
  command name and `docs:` / `components:` siblings), not under
  the user-suggested `lint:` (aikata has no `lint` command).

### Glob shape

The matcher lives in `internal/doctor/glob.go` and supports three
tokens, evaluated against the slash-form relative path passed to
`walkMarkdown`:

| Token | Meaning |
|---|---|
| `*` | Zero or more characters within a single path segment. |
| `**` | Zero or more **complete** path segments (recursive). |
| literal | Exact byte match. |

`?` and character classes are intentionally out of scope; the
patterns we want (`plugins/**`, `**/SKILL.md`,
`**/.claude-plugin/**`, `vendor/legacy.md`) are all expressible
with the three tokens above and the implementation stays well under
50 lines.

### Composition with `skippedDirs` / `skippedFiles`

User-supplied `doctor.exclude` is **additive** with the hardcoded
sets. Both layers are consulted; matching either causes the file
to be skipped. The hardcoded sets stay in place as the
backward-compatibility baseline. Promoting any of them to a
configurable default is out of scope here.

### No built-in defaults

aikata ships with `Doctor.Exclude == nil`. The ADR documents
recommended snippets for Claude Code plugin layouts so users can
copy them in, but the binary itself takes no implicit position on
third-party plugin specs. This keeps the Do-No-Harm contract
clean: aikata only excludes paths the user has explicitly named.

### Wiring

- `internal/config/aikata_yaml.go` gains a `Doctor` struct and an
  `AikataYaml.Doctor` field tagged `yaml:"doctor,omitempty"`.
- `internal/doctor/doctor.go` gains `Options.Excludes []string`.
  Existing call sites that build `doctor.Options` with only
  `TargetDir` continue to compile and observe identical behaviour
  (zero-value `nil` slice = no excludes).
- `internal/cli/doctor.go` loads the config via
  `config.LoadMigrated` after resolving `TargetDir`. A config-not-
  found result is treated as "no aikata project" and yields an
  empty exclude list (the existing behaviour for non-aikata trees
  is preserved).
- `walkMarkdown` filters relative paths through `MatchAny` before
  invoking its callback, so `checkFrontmatter` / `checkUpdated` /
  `checkGlossary` (the only three call sites) all honour the
  exclusion automatically.

## Consequences

### Positive

- Users layering aikata over a Claude Code plugin tree can put
  `aikata doctor --strict` on a CI gate without hand-modifying
  third-party SKILL.md frontmatter.
- The exclusion is declarative and lives in the project's own
  config — no per-CI environment-variable hack.
- The matcher is small enough to read at a glance, so future
  glob requests (e.g. `?` support) are tractable additions.
- Centralising the filter in `walkMarkdown` means any future
  walk-based check picks up exclusion semantics for free.

### Negative

- The exclude list is yet another thing users have to maintain
  when their repository layout changes. The mitigation is the
  recommended snippets in this ADR and the adoption guide.
- The matcher is a small custom implementation, not a battle-tested
  glob library. The scope is narrow (three tokens) and unit-tested,
  but it is one more chunk of code aikata owns. Trade-off accepted
  in favour of `go.mod` minimalism (see Alternatives Considered).
- `doctor:` becomes a new top-level config block. Schema growth is
  manageable for now; if a fourth or fifth doctor-tunable lands we
  may want to group them differently.

### Out of scope (deferred)

- **`doctor.frontmatter_required_paths`** (reverse include
  specification). The feedback document proposed this alongside
  `exclude`. We are punting until a real user reports the
  symmetric problem (wanting only `docs/**` and the top-level docs
  to be checked, ignoring everything else).
- **Auto-detection of known plugin layouts**
  (proposal B from the feedback). Coupling aikata to evolving
  third-party plugin specs is premature; revisit once Claude Code
  / Codex / Gemini-CLI plugin shapes stabilise (see ADR 0015).
- **Severity downgrade** (proposal C from the feedback) for files
  outside an aikata-managed scope. Severity tuning interacts with
  `--strict` semantics and deserves its own ADR if pursued.
- **`aikata sync` exclusion**. sync is manifest-driven and does
  not currently surface a comparable noise issue; will revisit
  if a user reports it.

## Alternatives Considered

- **`github.com/bmatcuk/doublestar/v4`** as the glob matcher.
  Industry-standard, well-tested, but adds a third direct dependency
  to a project whose Hard Rule 11 is "minimise external
  dependencies". The patterns aikata needs are expressible in ~40
  LOC of straight string manipulation; we accept the maintenance
  cost in exchange for keeping `go.mod` lean.
- **`path/filepath.Match` only.** The Go standard library matcher
  does not support `**`. Users would have to enumerate each
  intermediate path level (`plugins/*/skills/*/SKILL.md`,
  `plugins/*/agents/*.md`, ...), which is exactly the brittle
  enumeration the feedback document warned against.
- **`lint:` block instead of `doctor:`.** The feedback document
  proposed `lint:` as the YAML key. We picked `doctor:` because
  aikata has no `lint` command and the existing schema convention
  is "responsibility area" naming (`docs:`, `components:`,
  `features:`).
- **Auto-detect Claude Code plugin layouts** (`.claude-plugin/`,
  `SKILL.md`) and apply a frontmatter-spec switch. Tempting because
  it requires zero config from the user, but it pins aikata to a
  moving target. Deferred behind ADR 0015's "first-party wrappers,
  not a new agent" line until the third-party spec stabilises.
- **Severity downgrade** (treat missing frontmatter as `info`
  outside the aikata-managed scope). Avoids new schema but doesn't
  give the user a way to say "this subtree isn't mine". An exclude
  list does, so we prefer it.

## Implementation map

- `internal/config/aikata_yaml.go` — `Doctor` struct + field.
- `internal/config/aikata_yaml_test.go` — round-trip test that
  preserves `doctor.exclude` across `Marshal` / `Unmarshal`.
- `internal/doctor/glob.go` — `Match(pattern, path)` and
  `MatchAny(patterns, path)`; ~40 LOC.
- `internal/doctor/glob_test.go` — table-driven matcher tests
  including the `plugins/**`, `**/SKILL.md`, literal, and
  no-match cases.
- `internal/doctor/doctor.go` — `Options.Excludes` field +
  godoc.
- `internal/doctor/checks.go` — `walkMarkdown` consults
  `MatchAny` before invoking its callback.
- `internal/doctor/doctor_test.go` — new exclude-effective test
  exercising `checkFrontmatter`, `checkUpdated`, and
  `checkGlossary`.
- `internal/cli/doctor.go` — load config via `LoadMigrated`,
  thread `Doctor.Exclude` into `Options.Excludes`.
- `internal/cli/doctor_test.go` — integration test verifying
  config-driven exclusion takes effect end-to-end.

## Recommended snippet (Claude Code plugin layouts)

For repositories that adopt the Claude Code plugin convention
under `plugins/`:

```yaml
doctor:
  exclude:
    - "plugins/**"
    - "**/.claude-plugin/**"
```

If the user prefers a narrower allow-list, the filename-only form
is also supported:

```yaml
doctor:
  exclude:
    - "**/SKILL.md"
```

Both shapes are equivalent for the common case of "skip Claude
Code plugin frontmatter". The adoption guide cross-links this
section.
