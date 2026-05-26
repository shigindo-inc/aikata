---
description: Regenerate per-AI-tool config files (CLAUDE.md, .cursor/rules/main.mdc) from AGENTS.md.
---

# /aikata-generate

Run `aikata generate` to project the canonical `AGENTS.md` into
per-AI-tool config files. This is the safe way to keep `CLAUDE.md`
and `.cursor/rules/main.mdc` in sync after editing the source rules.

## When to run

- The user edited `AGENTS.md` and wants the per-tool files refreshed.
- `aikata doctor` reports generated artifacts are out of date.
- The project just gained or removed an AI-tool entry in
  `.aikata/aikata.yaml`.

## Step 1 — verify the project shape

```bash
aikata doctor
```

Resolve any error-level findings first; `aikata generate` will not
clobber a misconfigured project, but the result will be just as
misconfigured as the input.

## Step 2 — invoke generate

```bash
aikata generate
```

Output covers every tool listed in `.aikata/aikata.yaml`'s `ai_tools`
key. Today: `claude` writes `CLAUDE.md`, `cursor` writes
`.cursor/rules/main.mdc`, `codex` is a no-op (reads `AGENTS.md`
directly).

## Step 3 — review the diff

```bash
git diff CLAUDE.md .cursor/rules/main.mdc
```

The output should be a faithful projection of the `AGENTS.md`
changes. If something looks unexpected, the divergence is probably
in `AGENTS.md` itself rather than the generator.

## Notes

- Do not hand-edit `CLAUDE.md` or `.cursor/rules/main.mdc`. They are
  generated artifacts; the next `aikata generate` will overwrite
  manual edits.
- Codex reads `AGENTS.md` directly per ADR 0005, so the codex tool
  is intentionally a no-op for `generate`.
