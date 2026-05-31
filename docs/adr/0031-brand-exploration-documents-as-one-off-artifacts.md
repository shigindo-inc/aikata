---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-31
audience: [human, agent]
---

# ADR 0031 - Brand exploration documents as one-off artifacts

- **Status**: Accepted
- **Date**: 2026-05-31
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy), ADR 0007 (no generic
  `DESIGN.md`), ADR 0017 (post-init command taxonomy), ADR 0019
  (`sync` missing-file repair semantics), ADR 0028 (prioritize
  core-concept stabilization)

## Context

Mobile-app dogfooding surfaced two recurring documents that are useful
to humans and AI agents during brand exploration:

- app-icon concept exploration;
- mascot-character idea exploration.

These documents are more specific than the optional `UI.md` component.
They record product context for external image-generation LLMs, brand
constraints, candidate directions, comparison criteria, reusable
self-contained prompts, negative prompts, and follow-up work after a
direction is selected.

They should not be default files in `standard` or the future `extended`
scope:

- CLI, backend, library, and many web projects do not need either file.
- Even mobile apps may need an icon document but no mascot document.
- "Generate by default; delete if unnecessary" is not a valid opt-out
  under aikata's maintenance model. A managed file deleted by the user
  participates in `sync` missing-file repair (ADR 0019), so deletion
  does not express a durable opt-out.
- `extended` is the operational-readiness pack, not a product-branding
  preset.

The post-init taxonomy in ADR 0017 distinguishes durable capabilities
(`aikata enable`) from one-off authoring scaffolds (`aikata new`). Brand
exploration documents are stamped once, then rewritten heavily for the
project. They do not represent a durable aikata capability.

## Decision

### D1 - Add two short `new` artifact commands in v0.9.2

Add:

```bash
aikata new app-icon
aikata new mascot
```

They stamp:

```text
docs/design/app-icon-concepts.md
docs/design/mascot-character-ideas.md
```

The CLI identifiers intentionally stay shorter than the generated
filenames. `app-icon` is specific enough to distinguish the artifact
from a favicon, logo, or in-product icon library. `mascot` is sufficient
without repeating `character-ideas`.

### D2 - Keep both artifacts opt-in and project-owned after stamping

Neither artifact is emitted by `aikata init --scope standard`, the future
`extended` scope, `--with-ui`, or any stack selection. Neither adds a
config flag or an interactive init question.

Like `aikata new adr`, each command is a one-off authoring action. The
generated file is not registered as a durable capability and is not
added to `.aikata/manifest.yaml`. After creation, the project owns its
content; `aikata sync` must not restore or merge it.

### D3 - Ship focused bilingual starter templates

Both artifacts ship `en` and `ja` templates with the standard five-key
frontmatter. The templates remain concise starter structures, not
pre-filled brand strategies.

The app-icon template includes:

- product context for an external LLM;
- brand and technical constraints;
- concept candidates and a comparison matrix;
- self-contained image-generation prompts and negative prompts;
- follow-up work after selection.

The mascot template includes:

- product context for an external LLM;
- mascot role, tone, and constraints;
- candidate characters and a comparison matrix;
- self-contained image-generation prompts and negative prompts;
- intended product surfaces and follow-up work after selection.

The templates make no external API calls. Users copy prompts into their
chosen image-generation tool manually.

### D4 - Defer broader branding artifacts until evidence exists

Do not add a branding hierarchy or commands such as `new logo`,
`new brand-guide`, or `new design app-icon` in this increment. Add a
new short artifact identifier only when a repeated authoring need is
demonstrated.

## Consequences

**Positive**:

- App projects gain reusable brand-exploration scaffolds without making
  non-app projects pay for them.
- The command surface stays consistent with `aikata new adr`: short
  artifact names for one-off documents.
- Generated prompts remain useful to external LLMs that cannot read the
  repository.
- `standard`, `extended`, and `UI.md` keep their existing responsibilities.

**Negative**:

- Two more artifact names appear under `aikata list artifacts`.
- The generated files intentionally do not receive upstream improvements
  through `sync`; downstream edits are expected to diverge immediately.
- Users who want broader brand documentation still author it manually.

## Alternatives Considered

- **Emit both files in `standard` or `extended`.** Rejected: many projects
  do not need them, and deletion is not a durable opt-out under ADR 0019.
- **Add `--with-branding` / `aikata enable branding`.** Rejected: there is
  no durable capability to maintain, and icon / mascot needs are
  independently optional.
- **Use `aikata new app-icon-concepts` and
  `aikata new mascot-character-ideas`.** Rejected: the filenames are
  descriptive, but the CLI identifiers are unnecessarily long.
- **Fold the content into `UI.md`.** Rejected: the exploration documents
  have a distinct lifecycle and can grow independently; `UI.md` stays a
  concise UI / UX guideline surface.
