---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# TypeScript — Stack-Specific Rules

> Conventions and rules that apply specifically because
> samplekata is a TypeScript project. These are *additive* to the
> invariants in [AGENTS.md](../../AGENTS.md); when they conflict,
> AGENTS.md wins.

---

## 1. tsconfig

- `strict: true` (and the full strict family stays on). Disabling any
  strict-family flag requires an ADR.
- `target` and `module` chosen to match the runtime (see §2 / §3).
- _Recommended, tune per project: `noUncheckedIndexedAccess` and
  `exactOptionalPropertyTypes` for stricter indexing / optionality;
  `incremental: true` with a gitignored `.tsbuildinfo` for fast
  rebuilds._

## 2. Module format (ESM vs CJS)

- _TODO: pin the choice here and link the ADR that recorded it._
- File extensions:
  - ESM: explicit `.js` / `.ts` extensions in import specifiers when
    `moduleResolution: nodenext`.
  - CJS: no extension is fine; do not mix.
- Do not ship both flavors from one package without an ADR justifying
  the dual-publish.

## 3. Runtime target

- _TODO: pin Node.js LTS / Bun / Deno / browser._
- `package.json` `engines` enforces the minimum.

## 4. Package manager

- _TODO: pin npm / pnpm / yarn / bun. Commit the lockfile. CI runs
  `<pm> ci` (or equivalent reproducible install)._
- Do not mix package managers in one project.

## 5. Lint

- ESLint with `@typescript-eslint` recommended-type-checked config at
  minimum. Project-specific tightenings go in `eslint.config.*` with a
  one-line justification comment.
- `eslint .` must report **zero warnings** before commit. CI enforces
  this.
- Use `prettier` (or `dprint`) for formatting; do not hand-tune
  whitespace.

## 6. Type discipline

- **No `any`** without an inline comment justifying the escape hatch.
  Prefer `unknown` and narrow.
- **Avoid `as` casts.** Prefer user-defined type guards or
  `satisfies`.
- **Avoid `!` (non-null assertion)** unless the call site cannot be
  refactored to narrow. Each use needs a comment.
- Use `import type` for type-only imports so the emit stays tree-shake
  friendly.

## 7. Test runner

- _TODO: pin **vitest** or **jest** and link the ADR. Mixing runners is
  not allowed without an ADR._
- Tests under `test/` mirror `src/` one-to-one.
- Run the test command before declaring work complete.

## 8. Errors

- Subclass `Error` for new error categories (not plain strings or
  object literals); document which errors a public surface can throw.
- Use `cause` (ES2022) to chain errors instead of dropping the
  original stack.

## 9. Async

- `Promise` is the unit of asynchrony. Prefer `async`/`await` over
  raw `.then` chains.
- Do not float promises in non-test code — either `await` them or
  intentionally fire-and-forget with a comment.
- Reject early; do not swallow errors with empty `catch` blocks.

## 10. Project layout — where things live

Declaring a home for shared values up front prevents a common
AI-collaboration drift: string and numeric literals duplicate across
modules because no single source of truth was agreed. Recommended homes
(adjust to the project, but pick one and keep it consistent):

- `src/constants/` — shared constants (keys, routes, limits, feature
  flags). A literal used in more than one place goes here, not inline.
- `src/config/` — runtime configuration and environment wiring (read
  once, typed, validated at the boundary).
- For a frontend project: a single design-token / theme location (for
  example `src/theme/`) for colours, spacing, and typography, so UI
  values are not hard-coded per component.

aikata does not generate these files. The convention is shared as a
document so a human and an AI agent place new code in the same spot; the
agent creates the directory or file on demand when the project first
needs it, following this section.

_TODO: pin the layout for this project's shape (backend service vs
frontend app vs library) and record it in an ADR if it diverges. Link
the design-token source of truth if a package owns it._

## 11. When to revise this file

- A new dependency that significantly changes how state, IO, or build
  works → add a section.
- A team-wide preference that does not fit a single ADR → add it here.
- A rule that turned out to be wrong → remove it and explain in the
  commit message.
