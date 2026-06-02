---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# ARCHITECTURE — How

> This document explains **how** samplekata is built. For the
> **what / why**, read [SPEC.md](./SPEC.md). Individual decisions live
> under [`docs/adr/`](./docs/adr/). For TypeScript-specific conventions
> (tsconfig strictness, lint rules, test runner choice, ESM / CJS
> stance), read [`docs/stacks/typescript.md`](./docs/stacks/typescript.md).

---

## 1. Implementation language & runtime

- **TypeScript** 5.x compiled with `tsc` (and / or a bundler — captured
  in an ADR if non-trivial).
- Target runtime: _TODO — Node.js LTS | Bun | Deno | browser._
  Minimum runtime version pinned in `package.json` `engines`.
- Module format: _TODO — ESM | CJS._ Captured in
  `docs/stacks/typescript.md`.

## 2. Repository layout

```
samplekata/
├── README.md
├── AGENTS.md
├── SPEC.md
├── ARCHITECTURE.md
├── GLOSSARY.md
├── .gitignore
├── .aikata/
│   └── aikata.yaml
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/
│   │   └── typescript.md
│   ├── tasks/
│   │   └── current.md
│   └── troubleshooting.md
├── src/
│   └── index.ts             # entry point
├── test/                    # mirrors src/
├── package.json
├── tsconfig.json
└── (optional) eslint.config.* | .eslintrc.*
```

_TODO: extend with the source-code tree once it exists._

## 3. Major components

_TODO: list the major folders under `src/` (one per domain or layer)
and describe each in a short paragraph._

## 4. Data flow

_TODO: describe how data moves through the system. State / cache /
event-bus choices belong in an ADR._

## 5. Dependencies

_TODO: list npm dependencies with a one-line justification each.
Distinguish runtime from devDependencies. Avoid micro-dependencies
that wrap one stdlib function._

## 6. Error handling & logging

_TODO: describe the error-wrapping convention (custom Error subclasses
/ Result type / etc.), exit codes for CLI surfaces, log levels, and
the logger of record._

## 7. Testing strategy

- **Unit tests** under `test/`, mirroring `src/` one-to-one.
- **Integration tests** for I/O-touching code.
- CI runs `tsc --noEmit`, `eslint .`, then the test command (`vitest`
  or `jest`) captured in `docs/stacks/typescript.md`.

## 8. Distribution & releases

_TODO: describe how artifacts are packaged and delivered
(npm publish / Docker image / serverless deploy / etc.)._
