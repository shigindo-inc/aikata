---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-02
audience: [human, agent]
---

# ADR 0037 - Tighten adoption mutation boundaries in v0.9.7

- **Status**: Accepted
- **Date**: 2026-06-02
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0018 (managed-block
  append for project-owned files), ADR 0021 (`doctor.exclude`), ADR 0025
  (`sync.own`), ADR 0033 (`doctor` managed-surface direction)

## Context

Dogfooding adoption against the existing
`shigindo-flutter-skills` repository exposed four ways aikata reaches
beyond the document surface it can safely own.

First, `aikata doctor --fix` still walks almost every Markdown file
under the repository root. The repository contains Agent Skills whose
`skills/**/SKILL.md` files use the separate `name` + `description`
frontmatter contract, plus supporting `skills/**/references/*.md`
files and a project-owned `CONTRIBUTING.md`. Because none were excluded,
`doctor --fix` added aikata's five-key frontmatter to 18 files it does
not manage. ADR 0021 documented the same class of failure and shipped
`doctor.exclude`; ADR 0033 subsequently chose a managed-surface default
as the intended end state but deliberately deferred the behavior flip.
This report is the concrete follow-up trigger.

Second, the scaffolded `.gitignore` managed block is non-destructive at
the file level but too broad at the rule level. The Flutter variant adds
AI-tool paths for tools the project did not select, stack build ignores,
editor ignores, and broad local-file patterns. A documentation tool
should not silently decide whether a downstream project commits IDE
settings, lockfiles, Gradle wrappers, or other stack artifacts.

Third, `.env.example` is emitted by every standard / stack scaffold even
when the repository has no runtime environment variables. Its placeholder
examples are useful for an application but residue in a documentation or
Agent Skills repository. This conflicts with ADR 0003's default-off and
zero-residue rules for optional features.

Fourth, the documented `.aikata-proposed/` adoption fallback does not
exist in the implementation. SPEC, ROADMAP, `.gitignore`, and the
adoption guide say that `aikata init` in a non-empty directory without
`--force` writes proposals there. `internal/scaffold.Run` instead returns
`ErrTargetDirNotEmpty` before rendering anything.

The common failure is ownership drift: aikata is taking or advertising
responsibility for project concerns outside its shared-context document
contract.

## Decision

The four boundary fixes below were confirmed with the maintainer and
ship together in v0.9.7.

### D1 - Ship ADR 0033's managed-surface `doctor` default

The default `doctor` Markdown walk (the `frontmatter` / `updated` /
`glossary` checks) validates the union of:

- canonical top-level document names known to aikata (`README.md`,
  `AGENTS.md`, `SPEC.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`,
  `CHANGELOG.md`, `UI.md`, `API.md`);
- known aikata-owned document subtrees (`docs/adr/`, `docs/memory/`,
  `docs/stacks/`, `docs/tasks/`, `docs/workflows/`, `docs/design/`, plus
  `docs/troubleshooting.md`, `docs/prompts.md`, `docs/monorepo.md`);
- Markdown paths recorded in `.aikata/manifest.yaml`, when present.

It does **not** validate arbitrary repository Markdown by default.
`skills/**`, project-owned `CONTRIBUTING.md`, vendored docs, and
third-party contract files remain untouched unless they enter the
managed union explicitly. The union is computed by
`doctor.ManagedIncludeGlobs` and threaded through `Options.Includes`; an
empty include set is the broad walk.

The confirmed broad-audit opt-in is **`aikata doctor --all-markdown`**: a
discoverable one-shot flag rather than a persisted `doctor.scope` setting
introduced before a real need appears. `doctor.exclude` (ADR 0021)
remains additive under both modes, and `doctor --fix` uses the same scope
as the report it fixes. aikata's own repository has no manifest and ships
many docs outside the managed surface, so its CI gate runs
`aikata doctor --all-markdown --strict`.

### D2 - Reduce the `.gitignore` managed block to aikata-owned residue

Keep the ADR 0018 managed-block writer. The confirmed choice is to remove
stack-native ignores rather than retain a smaller reviewed subset: stack
selection adds collaboration guidance, not source-project generation
policy. The block body now carries only rules aikata can justify:

- `/.aikata-proposed/` (the D4 fallback scratch directory);
- generated AI-tool artifact paths;
- a **minimal, always-on secret baseline** (`.env`, `.env.local`).

The secret baseline is intentionally **unconditional** — not tied to the
env capability (D3). The failure modes are asymmetric: a missing `.env`
ignore can leak credentials into irreversible git history, while an
ignore rule for an absent `.env` costs nothing. aikata already ships
"never commit secrets — reference `.env.example`" in every generated
`AGENTS.md`, so shipping the matching enforcement is the coherent choice
and is the one near-universal rule aikata can defend owning. Coupling it
to `.env.example` would also leave exactly the projects that have a
`.env` but no committed example file unprotected. The baseline is kept
narrow: the broader `*.local` convention is dropped (it covers
non-secret files such as `settings.local.json` and is closer to
editor/stack policy).

Unconditional future-tool paths beyond the artifact list, editor / OS
rules, coverage rules, `*.local`, and the Flutter / TypeScript build
rules are removed from the templates, along with the stale comment
recommending the deleted `docs.generate_gitignore: false` setting. The
`minimal` scope still ships no `.gitignore` at all.

`.aikata/` is **not** ignored: its config and manifest are project state
users should commit so `sync` has a stable baseline.

### D3 - Move `.env.example` off the default scaffold (opt-in `env` capability)

The confirmed choice is an explicit capability, since application
repositories have a legitimate recurring need. `.env.example` becomes the
opt-in `env` component (`aikata enable env` / `aikata init --with-env`),
mirroring the v0.9.2 reusable-prompts capability (ADR 0034): a
`components.env` schema-v2 flag, a `components/env/<lang>/.env.example.tmpl`
template, manifest tracking, and `sync` preservation. Only the **example
file** is opt-in; the `.gitignore` secret baseline that protects `.env`
itself is unconditional (D2). The canonical
`AGENTS.md` / `README.md` templates no longer hard-link to
`./.env.example` (which would dangle when the file is absent); they
mention the `.env.example` pattern as inline code instead, matching the
`minimal` scope. The existing `doctor` env cross-reference check already
no-ops when the file is absent. Removing it from the default scope is
sync-visible but non-destructive (ADR 0019).

### D4 - Implement the documented `.aikata-proposed/` fallback

The confirmed choice is to implement the advertised fallback rather than
delete it from the contract, because adopting existing repositories is
exactly the path that needs stronger protection. When the target
directory is non-empty and `--force` is absent, `scaffold.Run` renders
the full scaffold under `.aikata-proposed/` and returns success with an
actionable notice instead of erroring. A non-empty `.aikata-proposed/` is
refused with `ErrProposalExists` rather than overwritten. `--dry-run`
previews the proposal path without writing. This closes the prior
SPEC / implementation gap and justifies keeping the ignore rule.

### D5 - Carry migration and verification proof in the implementation PR

The implementation PR must:

- add regression coverage proving default `doctor --fix` leaves
  `skills/**` and project-owned `CONTRIBUTING.md` byte-identical;
- prove broad-audit mode still reports those paths when explicitly
  selected;
- add golden coverage for the reduced `.gitignore` bodies and optional
  `.env.example` residue;
- test `.aikata-proposed/` creation, collision refusal, and zero writes
  outside the proposal tree;
- align SPEC, ARCHITECTURE, adoption docs, templates, and generated
  fixtures with the accepted decisions;
- append migration notes for projects that already carry the broader
  managed block.

## Consequences

### Positive

- `doctor --fix` stops mutating third-party Markdown contracts by
  default.
- The `.gitignore` writer remains visibly non-destructive while its
  owned block becomes small enough to defend rule by rule.
- Standard scaffolds stop carrying environment-variable residue into
  repositories that do not need it.
- Existing-repository adoption gains the proposal workflow already
  promised by the public docs.

### Negative

- Narrowing `doctor` changes what an unflagged CI gate validates.
  ADR 0033 accepted this direction but requires broad-mode
  discoverability and before / after proof.
- Existing projects retain previously appended `.gitignore` lines until
  a managed-block refresh (`init --force` or a future sync integration).
- Making `.env.example` optional changes golden fixtures and sync-visible
  upstream output. ADR 0019's no-silent-delete rule means existing files
  remain on disk.
- `.aikata-proposed/` adds a second render destination and collision path
  to scaffold.

## Alternatives Considered

- **Keep broad `doctor` and auto-add `skills/**` exclusions.** Rejected:
  it treats one third-party layout as special and leaves the ownership
  inversion intact for the next format.
- **Use manifest-only `doctor` scope.** Rejected by ADR 0033: adopted and
  pre-manifest projects would silently lose useful validation.
- **Keep the broad `.gitignore` as a convenience.** Rejected: convenience
  does not justify silently selecting project policy unrelated to
  shared-context documentation.
- **Remove `.gitignore` management entirely.** Rejected: generated
  artifacts and proposal scratch files still need an explicit default,
  and ADR 0018 already provides a narrow non-destructive mechanism.
- **Remove `.aikata-proposed/` from the docs instead of implementing it.**
  Viable but not recommended: explicit proposal output gives adopters a
  safer path to inspect generated documents before accepting them.

