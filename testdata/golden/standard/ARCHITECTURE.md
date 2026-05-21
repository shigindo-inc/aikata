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
> under [`docs/adr/`](./docs/adr/).

---

## 1. Implementation language & runtime

_TODO: name the language, runtime, and minimum supported versions._

## 2. Repository layout

```
samplekata/
├── README.md
├── AGENTS.md
├── SPEC.md
├── ARCHITECTURE.md
├── GLOSSARY.md
├── .env.example
├── .gitignore
├── .ai/
│   └── aikata.yaml
└── docs/
    ├── adr/
    │   └── 0001-record-architecture-decisions.md
    ├── tasks/
    │   └── current.md
    ├── troubleshooting.md
    └── prompts.md
```

_TODO: extend with the source-code tree once it exists._

## 3. Major components

_TODO: list and describe the modules / services / packages. One
short paragraph each._

## 4. Data flow

_TODO: describe how data moves through the system._

## 5. Dependencies

_TODO: list external dependencies. Justify each (avoid heavy or
unmaintained ones)._

## 6. Error handling & logging

_TODO: describe the error-wrapping convention, exit codes, and log
levels._

## 7. Testing strategy

_TODO: describe unit / integration / golden test layers and the CI
gate._

## 8. Distribution & releases

_TODO: describe how artifacts are packaged and delivered._
