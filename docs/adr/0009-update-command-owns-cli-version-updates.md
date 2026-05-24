---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0009 - Reserve `aikata update` for CLI Version Updates

- **Status**: Accepted
- **Date**: 2026-05-24
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0008
  (`.aikata/` config namespace)

## Context

The initial spec used `aikata update` for a Copier-style project template
diff-merge: detecting drift between bundled templates and a user's
canonical project documents, then interactively applying hunks without
overwriting user edits.

That meaning is precise from an implementation perspective, but it
conflicts with a common CLI expectation: `update` updates the installed
tool. Claude Code's `claude update` reinforces this convention for the
same audience aikata targets.

aikata also has two different update domains:

- **CLI version update**: replace or upgrade the installed aikata binary.
- **Project template sync**: merge newer aikata templates into an
  existing project while preserving local edits.

Using `update` for the second domain makes the first harder to explain,
especially now that aikata has a native install script and plans Homebrew
and npm distribution channels.

## Decision

Reserve `aikata update` for updating the aikata CLI itself.

Rename the planned project template diff-merge command to:

```text
aikata sync
```

The resulting command boundaries are:

- `aikata update` updates or checks the installed aikata CLI version.
- `aikata sync` reviews and merges bundled template changes into the
  current project.
- `aikata generate` regenerates tool-specific artifacts from canonical
  sources. It may warn that `aikata sync` is available, but it must not
  rewrite canonical project documents as a hidden side effect.

`aikata update` must respect the install source:

- Native / installer-managed binaries may self-update after release asset
  and checksum verification.
- Homebrew, npm, Go, and OS package-manager installs remain owned by
  their package manager. `aikata update` may print, or behind explicit
  opt-in run, the relevant manager command; it must not overwrite those
  managed binaries directly.
- Unknown install sources get actionable manual instructions.

## Consequences

**Positive**:

- Aligns aikata with the CLI convention reinforced by Claude Code.
- Keeps `generate` narrowly scoped to derived AI-tool artifacts.
- Makes the template merge operation clearer: syncing project documents
  is different from updating the installed binary.
- Leaves room for a safe native updater while preserving package-manager
  ownership.

**Negative**:

- Renames the roadmap's largest planned feature before implementation,
  so old planning notes may mention `aikata update` for template merges.
- Adds one more command name (`sync`) that users must learn.
- Requires install-source metadata before `aikata update` can do more
  than print guidance safely.

## Alternatives Considered

- **Keep `aikata update` for template merges**: rejected because it
  conflicts with user expectation from adjacent tools and blocks a clear
  CLI self-update command.
- **Use `aikata upgrade` for CLI updates**: rejected for now because
  Claude Code uses `update`, and aikata intentionally targets users who
  already work with Claude Code.
- **Let `aikata generate` update templates automatically**: rejected
  because generated AI-tool artifacts are disposable, while canonical
  project documents contain user edits and require explicit review.
