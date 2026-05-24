---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# ADR 0008 - Use `.aikata/` for aikata-owned Configuration

- **Status**: Accepted
- **Date**: 2026-05-24
- **Deciders**: aikata maintainers
- **Related**: ADR 0003 (Do-No-Harm Policy)

## Context

v0.2 stores aikata's project configuration at `.ai/aikata.yaml`.
The original Phase 1 design introduced `.ai/` as a general home for
AI-tool-facing artifacts and the aikata config file, but it did not record
why that generic directory name was preferred over an aikata-owned namespace.

In practice, generated tool-facing artifacts already live in each tool's
native location, such as `CLAUDE.md` and `.cursor/rules/`. The remaining
config under `.ai/` is therefore aikata-owned state, not a shared AI-tool
interchange directory.

The name `.ai/` is also broad enough that a future AI tool, convention, or
workspace manager could reasonably claim it for unrelated state. That would
create avoidable ambiguity in user projects.

## Decision

Starting in the v0.3.x line, new aikata projects should store aikata-owned
configuration under `.aikata/`, with the primary config path:

```text
.aikata/aikata.yaml
```

The migration should be backward compatible:

- New `aikata init` output writes `.aikata/aikata.yaml`.
- Commands that read config first check `.aikata/aikata.yaml`, then fall
  back to the legacy `.ai/aikata.yaml`.
- Reading the legacy path emits a deprecation warning with the new path.
- A migration helper should move `.ai/aikata.yaml` to `.aikata/aikata.yaml`
  without touching tool-native generated artifacts.
- The legacy read fallback remains through at least v0.x; v1.0 decides
  whether to remove it.

This ADR records the direction only. The code and template changes are
scheduled for v0.3.x, not for the documentation-only change that introduced
this ADR.

## Consequences

**Positive**:

- Gives aikata a clear namespace for its own durable config.
- Reduces collision risk with future AI-tool conventions that might use
  `.ai/`.
- Aligns with the existing pattern used by tool-native directories such as
  `.cursor/` and aikata's own transient `.aikata-proposed/` directory.

**Negative**:

- Requires a path migration for existing v0.1 and v0.2 projects.
- Temporarily increases documentation and implementation complexity because
  both paths must be understood.
- Any user scripts that read `.ai/aikata.yaml` directly need to update.

## Alternatives Considered

- **Keep `.ai/` permanently**: rejected because it is a broad shared name
  for state that is now specifically aikata-owned.
- **Use a root `aikata.yaml` file**: rejected because it adds another
  visible root file and weakens top-level minimalism.
- **Use `.config/aikata.yaml`**: rejected because many project ecosystems
  already use `.config/` for application-specific local state, and the
  ownership is less obvious than `.aikata/`.
