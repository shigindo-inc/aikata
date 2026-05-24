---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0006 — Locale and Japanese Documentation Policy

- **Status**: Accepted
- **Date**: 2026-05-24
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm), Q-DESIGN-03

## Context

aikata is intentionally identifiable as a Japanese OSS project, and
Japanese-speaking users are a first-class audience. At the same time, the
repository is public OSS: top-level project documents must remain easy for
non-Japanese contributors, package consumers, and search engines to read.

The project already supports generated Japanese project documents via
`aikata init --lang ja`, with parallel `en/` and `ja/` template trees and
golden tests. That support is part of the product surface. The separate
question is how aikata's **own repository documentation** should serve
Japanese users without creating a second canonical documentation set.

Full duplication such as `README.md` plus `README.ja.md`, `SPEC.md` plus
`SPEC.ja.md`, and so on would make every product decision carry a
translation synchronization cost. A mixed bilingual single-file mode
would also complicate both human reading and agent context, and is already
deferred to v1.x in the roadmap.

## Decision

We adopt **English canonical repository documentation with a Japanese
access layer**.

Concretely:

1. **Repository canonical documents stay English.** Top-level documents
   such as `README.md`, `SPEC.md`, `ARCHITECTURE.md`, `ROADMAP.md`,
   `GLOSSARY.md`, and `AGENTS.md` remain English-first unless a future
   ADR supersedes this policy.
2. **Japanese generated templates are first-class.** `--lang ja` means
   the generated project document language is Japanese. The Japanese
   template trees must keep the same file sets as English templates, and
   golden tests protect that parity.
3. **Japanese repo docs are an access layer, not a mirror.** The aikata
   repository may include focused `.ja.md` documents under `docs/` for
   Japanese users, but they summarize and route readers to the English
   canonical documents rather than duplicating them wholesale.
4. **`--lang` only means document language.** CLI prompt language,
   diagnostics, support language, and future translation helpers are
   separate concerns. They must not be inferred from `project.lang`
   without a new decision.
5. **Bilingual same-file mode remains deferred.** A future mode that
   places Japanese-for-humans and English-for-LLMs in the same canonical
   document is still a v1.x design topic.

## Consequences

**Positive**:

- Non-Japanese contributors see one canonical English documentation set.
- Japanese users get an explicit entry point and can generate Japanese
  project documents without adopting bilingual maintenance.
- Translation drift is limited to product templates, where parity is
  mechanically testable, and small access documents, where drift is low
  impact.
- The policy satisfies ADR 0003: English remains the default, and Japanese
  support is explicit rather than imposed on users who do not need it.

**Negative**:

- Japanese readers still need to consult English canonical documents for
  detailed design and contribution rules.
- The Japanese access layer can lag behind canonical docs. Mitigation:
  keep access docs short and navigation-oriented.
- The policy does not solve localized CLI messages. That remains a
  separate design decision.

## Alternatives Considered

- **Full Japanese mirror for every canonical document**: rejected for
  v0.x. It would maximize Japanese accessibility but create high
  synchronization cost across every product and architecture change.
- **English-only repository documentation**: rejected. It underserves the
  project's Japanese OSS identity and makes `--lang ja` harder to
  discover for the users who benefit most.
- **Bilingual paragraphs in every canonical document**: rejected for now.
  It increases document length and agent context cost, and it blurs which
  language is authoritative when the two drift.
