---
description: Stamp a one-off aikata authoring artifact (ADR, app-icon, mascot) in the current project.
---

# /aikata-new

Run `aikata new <artifact> [args]` to stamp a one-off authoring scaffold.
The user invoked this slash command, so they want the artifact created;
do not stop to confirm intent.

Arguments the user passed: `$ARGUMENTS`

## Step 1 — verify aikata is installed

If `aikata` is not on `PATH`, surface the three install paths first:

- `curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh` (Linux / macOS)
- `go install github.com/shigindo-inc/aikata/cmd/aikata@latest` (any platform with Go 1.21+)
- Manual download from `https://github.com/shigindo-inc/aikata/releases/latest`

## Step 2 — resolve the artifact

Available artifacts (run `aikata list artifacts` to confirm what this
binary supports):

- `adr "<title>"` → `docs/adr/NNNN-<slug>.md` (auto-numbered) — **the
  common case**: record a decision the user just made.
- `app-icon` → `docs/design/app-icon-concepts.md` — brand icon
  exploration (bilingual starter, image-generation prompts).
- `mascot` → `docs/design/mascot-character-ideas.md` — mascot
  exploration.

If `$ARGUMENTS` already names a valid artifact (and, for `adr`, a title),
use it as-is. If it is empty or ambiguous, ask which artifact — and for
`adr`, the decision title — before running.

## Step 3 — invoke

```bash
aikata new $ARGUMENTS
```

For example: `aikata new adr "Use Go modules"`.

## Notes

- One-off artifacts are **not** recorded in `.aikata/manifest.yaml` and
  `aikata sync` never restores or merges them — after stamping, the
  project owns the file (ADR 0031).
- `aikata new` **refuses to clobber** an existing file. If the target
  already exists, surface that to the user rather than overwriting.
- `new` stamps one-off artifacts; for durable project capabilities
  (memory, UI guide, a stack, a workflow guide) use `/aikata-enable`.
