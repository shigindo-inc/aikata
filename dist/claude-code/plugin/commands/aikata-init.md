---
description: Scaffold a new aikata project (AGENTS.md / SPEC.md / ARCHITECTURE.md / generated AI-tool configs).
---

# /aikata-init

Run `aikata init` to scaffold a new aikata project in the current
working directory. The user invoked this slash command, so they
want the scaffold to happen; do not stop to confirm intent.

## Step 1 — verify aikata is installed

If `aikata` is not on `PATH`, surface the three install paths before
falling back to manual scaffolding:

- `curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh` (Linux / macOS)
- `go install github.com/shigindo-inc/aikata/cmd/aikata@latest` (any platform with Go 1.21+)
- Manual download from `https://github.com/shigindo-inc/aikata/releases/latest`

## Step 2 — gather inputs

Ask the user (or infer from the conversation):

1. **Project name** — single token, used as `{{.ProjectName}}` everywhere.
2. **Preset** — `standard` (default) / `minimal` / `flutter` / `typescript`.
3. **Language** — `en` (default) or `ja`.
4. **AI tools** — comma-separated list from `claude | cursor | codex` (default `claude`).
5. **Opt-ins** — `--with-memory`, `--with-ui`, `--with-api`,
   `--with-tdd`, `--with-changelog`, `--monorepo`. Ask only when the
   user did not signal preferences already.

## Step 3 — invoke aikata

Prefer the non-interactive form so the result is deterministic:

```bash
aikata init <name> \
  --preset <preset> \
  --lang <lang> \
  --ai-tools <tools> \
  [--with-memory] [--with-ui] ... \
  --no-interactive
```

When the target directory already has files, run with `--dry-run`
first, show the user the plan, and only proceed with `--force` after
they explicitly approve.

## Step 4 — verify

After scaffolding, run:

```bash
aikata doctor
```

A clean (zero errors, zero warnings) report means the project is
ready. Surface any errors back to the user immediately.

## Notes

- aikata writes `.aikata/manifest.yaml` at init time (v0.5+).
  Future `aikata sync` runs use it as the merge ancestor — do not
  edit by hand.
- For `--monorepo` projects, also point the user at
  `docs/monorepo.md` and `apps/_example/AGENTS.md` so they know the
  per-app rule pattern.
