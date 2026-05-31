---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# samplekata

> One-line description of samplekata (a Flutter project) goes here.

This project was scaffolded with
[aikata](https://github.com/shigindo-inc/aikata) — the `flutter` preset.

## Read first

| Read for… | Document |
|---|---|
| What & Why | [SPEC.md](./SPEC.md) |
| How (technical) | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| Terminology | [GLOSSARY.md](./GLOSSARY.md) |
| Agent rules | [AGENTS.md](./AGENTS.md) |
| Stack-specific rules | [docs/stacks/flutter.md](./docs/stacks/flutter.md) |

### Decisions & design

- [`docs/adr/`](./docs/adr/) — Architecture Decision Records.

### Operational notes (frequently changed)

- [`docs/tasks/current.md`](./docs/tasks/current.md) — current in-flight work.
- [`docs/troubleshooting.md`](./docs/troubleshooting.md) — known pitfalls.

## Quickstart

```bash
flutter pub get
flutter analyze
flutter test
flutter run
```

## Configuration

aikata stores its own settings under [`.aikata/aikata.yaml`](./.aikata/aikata.yaml).
Environment variables expected by samplekata are documented in
[`.env.example`](./.env.example).

## License

(unspecified — add a `LICENSE` file before publishing)
