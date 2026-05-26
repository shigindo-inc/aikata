---
description: Run aikata's read-only consistency checks against the current project.
---

# /aikata-doctor

Run `aikata doctor` to audit an aikata-managed project for the eight
read-only consistency checks aikata ships. Use this whenever:

- The user wants confidence the project is healthy.
- A CI build failed on doctor and the user wants to debug locally.
- Before any `aikata generate` / `aikata sync` to confirm the input
  shape is valid.

## Step 1 — plain run

```bash
aikata doctor
```

Exit codes:

- `0` — clean (zero error-level findings).
- `3` — at least one error-level finding.

Warning- and info-level findings stay at exit 0 unless `--strict` is
passed. The v0.5+ CI dogfood gate uses `aikata doctor --strict`; for
local exploration the default behaviour is friendlier.

## Step 2 — JSON for tooling

```bash
aikata doctor --json
```

Emits the versioned `{version: 1, kind: "doctor", issues: [...],
summary: {...}}` envelope. Use this when chaining doctor's output
into another tool or when copying findings into a bug report.

## Step 3 — auto-fix the trivial subset

```bash
aikata doctor --fix --dry-run   # preview
aikata doctor --fix             # apply
```

`--fix` repairs:

- Missing frontmatter blocks (scaffolds them).
- Missing required frontmatter keys (appends them).
- Stale `updated:` values (bumps to today).

Everything else (broken markdown links, missing files, ADR
numbering gaps) requires manual repair — surface the diagnostic and
the file:line to the user.

## Notes

- Doctor is read-only by default; only `--fix` writes.
- `--strict` is the right flag when running doctor in CI.
- ADR-numbering issues are reported at info level; they signal
  that ADR files exist but were skipped by `internal/adr.Scan`
  (slug rule violation). Renaming the file is the fix.
