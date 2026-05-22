---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# Flutter — Stack-Specific Rules

> Conventions and rules that apply specifically because
> samplekata is a Flutter project. These are *additive* to the
> invariants in [AGENTS.md](../../AGENTS.md); when they conflict,
> AGENTS.md wins.

---

## 1. Code style and lints

- `analysis_options.yaml` extends `package:flutter_lints/flutter.yaml`
  at minimum. Project-specific tightenings go in the same file with a
  one-line justification comment.
- `flutter analyze` must report **zero warnings** before commit. CI
  enforces this.
- Use `dart format` (default 80-col line length). Do not hand-tune
  whitespace.

## 2. Null safety

- The project ships with sound null safety (Dart 3+).
- Avoid `!` (the bang operator). Each use must have an inline comment
  explaining why the value is invariantly non-null at that point.
- Prefer `late` over `?` when a field is non-null after `initState` or
  the constructor body.

## 3. State management

_TODO: name the chosen approach (Provider / Riverpod / Bloc / GetX /
plain `setState`) and link the ADR that recorded the choice. Do not
mix two state-management libraries without an ADR justifying it._

## 4. Widget authoring

- **Const constructors everywhere they apply.** The linter (`prefer_const_constructors`)
  is on; do not silence it.
- Split widgets at the screen boundary first, then by repeated subtree.
  Prefer extracting widgets to passing massive `build` methods.
- `BuildContext` across `await`: always check `if (!mounted) return;`
  before using the context after the await resolves.
- Avoid `setState` calls outside `StatefulWidget`s. State logic that
  needs to leave the widget belongs in §3's chosen state manager.

## 5. Async, isolates, and main-thread budget

- Use `compute` (or a dedicated isolate) for any CPU-bound work that
  would otherwise stall the UI for more than one frame (~16 ms).
- Long-running tasks must show a non-blocking progress indicator within
  a frame of starting.

## 6. Build_runner / codegen

- Generated files live next to their sources (`foo.g.dart`,
  `foo.freezed.dart`).
- Generated files are **committed** so contributors do not need to run
  build_runner before they can read the code. CI runs
  `dart run build_runner build --delete-conflicting-outputs` and fails
  if the working tree is dirty afterwards.

## 7. Testing

- **Unit tests** under `test/`, mirroring `lib/` one-to-one.
- **Widget tests** for every reusable widget (the kind another file
  imports).
- **Golden tests** for any UI that should not visually regress. Goldens
  are platform-sensitive — run them on the same CI image, do not check
  in goldens captured locally if your local platform differs.
- **Integration tests** under `integration_test/` for flows that cross
  more than one screen.
- Run `flutter test` (and `flutter test integration_test/` when
  present) before declaring work complete.

## 8. Platform channels and native code

- Each platform channel must have a single Dart-side caller and a
  matching native handler; mismatched channels are silent failures.
- Document every channel in `ARCHITECTURE.md` with: channel name, both
  ends (file paths), and message schema.

## 9. Accessibility

- Every interactive widget exposes a `Semantics` label (explicitly or
  via the default).
- Color is never the sole carrier of information.
- Text scales: do not set `MediaQuery.textScaler = TextScaler.noScaling`;
  let the OS-level setting through.

## 10. When to revise this file

- A new dependency that significantly changes how state, navigation, or
  IO works → add a section.
- A team-wide preference that does not fit a single ADR → add it here.
- A rule that turned out to be wrong → remove it and explain in the
  commit message.
