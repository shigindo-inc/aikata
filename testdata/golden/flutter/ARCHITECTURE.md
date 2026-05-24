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
> under [`docs/adr/`](./docs/adr/). For Flutter-specific conventions
> (lints, state management choice, build_runner usage, null-safety
> stance), read [`docs/stacks/flutter.md`](./docs/stacks/flutter.md).

---

## 1. Implementation language & runtime

- **Dart** 3.x with **Flutter** 3.x (channel: _TODO — stable | beta_).
- Minimum SDK constraints pinned in `pubspec.yaml`.

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
├── .aikata/
│   └── aikata.yaml
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/
│   │   └── flutter.md
│   ├── tasks/
│   │   └── current.md
│   ├── troubleshooting.md
│   └── prompts.md
├── lib/
│   └── main.dart           # app entry; widget tree starts here
├── test/
│   └── …                   # mirrors lib/ one-to-one
├── pubspec.yaml
├── analysis_options.yaml   # lints (see docs/stacks/flutter.md)
└── (platform dirs: ios/, android/, web/, macos/, linux/, windows/)
```

_TODO: extend with the source-code tree once it exists._

## 3. Major components

_TODO: list the major feature folders under `lib/` (one per screen or
domain) and describe each in a short paragraph._

## 4. Data flow

_TODO: describe how data moves through the app. State management
choice (Provider / Riverpod / Bloc / etc.) is captured in an ADR; the
rationale lives in `docs/stacks/flutter.md`._

## 5. Dependencies

_TODO: list pub.dev packages with a one-line justification each.
Prefer published-by-flutter.dev or popular long-maintained packages.
Discourage thin wrappers around the stdlib._

## 6. Error handling & logging

_TODO: describe the error-wrapping convention (Result / Either /
exceptions), exit / crash policy, and the logger of record._

## 7. Testing strategy

- **Unit tests** under `test/`, mirroring `lib/` one-to-one.
- **Widget tests** for any custom widget exposed in `lib/`.
- **Golden tests** for any pixel-critical UI (use `flutter_test`'s
  `matchesGoldenFile`).
- **Integration tests** under `integration_test/` for cross-screen
  flows.
- CI runs `flutter analyze` then `flutter test` (and
  `flutter test --machine integration_test/` when present).

## 8. Distribution & releases

_TODO: describe how artifacts are packaged and delivered
(TestFlight / Google Play internal / Web hosting / etc.)._
