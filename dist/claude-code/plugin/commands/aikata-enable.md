---
description: Enable a durable aikata capability (memory, ui, api, tdd, changelog, prompts, monorepo, stack, ai-tool, workflow) in the current project.
---

# /aikata-enable

Run `aikata enable <capability> [args]` to add a durable capability to the
current aikata project. The user invoked this slash command, so act on it.

Arguments the user passed: `$ARGUMENTS`

## Step 1 — verify aikata is installed

If `aikata` is not on `PATH`, surface the three install paths first:

- `curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh` (Linux / macOS)
- `go install github.com/shigindo-inc/aikata/cmd/aikata@latest` (any platform with Go 1.21+)
- Manual download from `https://github.com/shigindo-inc/aikata/releases/latest`

## Step 2 — resolve the capability

Available capabilities (run `aikata list capabilities` to confirm what
this binary supports):

- Single-file docs: `ui` → `UI.md`, `api` → `API.md`, `tdd` →
  `docs/testing.md`, `changelog` → `CHANGELOG.md`, `prompts` →
  `docs/prompts.md`.
- `memory` → the long-term agent memory slot under `docs/memory/`.
- `monorepo` → nested `apps/<name>/AGENTS.md` layout.
- `stack <name>` (e.g. `stack flutter`), `ai-tool <name>` (e.g.
  `ai-tool cursor`), `workflow <domain>` (e.g. `workflow git`).

If `$ARGUMENTS` names a valid capability, use it as-is. If it is empty or
ambiguous, ask which capability before running.

## Step 3 — invoke

```bash
aikata enable $ARGUMENTS
```

For example: `aikata enable ui` or `aikata enable workflow git`.

## Step 4 — verify

```bash
aikata doctor
```

A clean report means the new capability integrated correctly.

## Notes

- `enable` records a **durable** capability: it renders the files,
  records them in `.aikata/manifest.yaml` (so `aikata sync` preserves
  them), and flips the matching schema-v2 `components.*` flag or appends
  to `stacks` / `ai_tools` / `workflows` in `.aikata/aikata.yaml`
  (ADR 0017). It is idempotent — re-running is safe.
- For one-off authoring scaffolds (an ADR, an app-icon doc) use
  `/aikata-new` instead.
