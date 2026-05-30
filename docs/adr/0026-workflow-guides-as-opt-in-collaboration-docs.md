---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-30
audience: [human, agent]
---

# ADR 0026 - Workflow guides as opt-in collaboration documents

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: aikata maintainers
- **Related**: ADR 0002 (`AGENTS.md` is canonical), ADR 0003
  (Do-No-Harm Policy), ADR 0016 (`.aikata/aikata.yaml` schema v2),
  ADR 0017 (post-init command taxonomy), ADR 0024 (scope / stack axes)

## Context

aikata exists to make a project readable by humans and AI coding agents.
The existing document set already captures product requirements
(`SPEC.md`), technical structure (`ARCHITECTURE.md`), terminology
(`GLOSSARY.md`), design decisions (`docs/adr/`), long-term memory
(`docs/memory/`), and current task state (`docs/tasks/current.md`).

One recurring collaboration gap sits between these slots: project
workflow. In small teams and solo projects, agents need stable answers
to questions such as:

- which branch model to use;
- how to name branches and commits;
- when pull requests are required;
- which review and self-merge rules apply;
- how releases are tagged;
- which CI checks must be green before merge.

Those rules are operationally important but do not fit cleanly into the
existing files:

- `AGENTS.md` should stay compact and invariant. Embedding a complete
  Git workflow there would make the canonical instruction file heavy.
- `CONTRIBUTING.md` is human-contributor-facing and belongs to the
  future `extended` / governance pack. It should not be required merely
  so an agent knows how to open a branch or PR.
- `docs/tasks/current.md` is ephemeral working memory, not a durable
  workflow rulebook.
- `docs/memory/feedback.md` records validated preferences over time,
  but it is not the right place for a first-class project policy that
  every collaborator should read.

A concrete motivating example is a small-team GitHub Flow policy: short-
lived branches, Conventional Commits, small PRs, squash-only merges,
SemVer release tags, trusted-committer self-merge when CI is green, and
optional GitHub Repository Rulesets / CODEOWNERS enforcement. The
policy is useful to agents, but parts of it are personal or environment-
specific (trusted account names, vault paths, GitHub plan constraints,
and helper-command names). Shipping that whole policy as a default
template would violate the Do-No-Harm Policy.

## Decision

Add a new optional document category: **workflow guides**.

### D1 - Location

Workflow guides live under:

```text
docs/workflows/
```

The first planned built-in guide is:

```text
docs/workflows/git.md
```

The location keeps the project root minimal, avoids overloading
`AGENTS.md`, and leaves `CONTRIBUTING.md` free for contributor-facing
governance material.

### D2 - Canonical relationship to `AGENTS.md`

`AGENTS.md` remains the canonical source for agent behaviour. It should
not inline full workflow guide content. Instead, when a workflow guide is
enabled, `AGENTS.md` gets only a short conditional pointer, for example:

```md
## Workflow

If present, follow `docs/workflows/git.md` for branch, commit, PR,
review, merge, release, and CI gate rules.
```

The workflow guide itself is the canonical source for the detailed
workflow policy.

### D3 - CLI shape

The user-facing command shape is:

```bash
aikata enable workflow git
```

This is intentionally broader than `enable git-workflow`. Git is the
first workflow domain, not the only possible one. The same shape can
later support release, deployment, incident, or review workflows without
creating one top-level capability per policy.

The durable config record should be a list, not a boolean:

```yaml
workflows:
  - git
```

This avoids growing `components.git_workflow`, `components.release_workflow`,
and similar one-off fields. The existing `components:` block remains for
single-file / template-scope capabilities; workflow guides are their own
orthogonal axis, similar in spirit to `stacks:` and `ai_tools:`.

### D4 - Initial Git workflow scope

The v0.8.4 implementation should generate only the documentation guide:

```text
docs/workflows/git.md
```

The built-in content should cover the portable policy:

- GitHub Flow with short-lived branches from `main`;
- Conventional Commits;
- small pull requests;
- squash-only merge policy;
- SemVer release tags;
- hotfix / mobile release branch conventions where applicable;
- CI gates and required checks as policy.

The built-in guide must not hard-code personal account names, local vault
paths, private helper command names, or a specific GitHub paid-plan
assumption. Those belong in a personal template, a downstream project
edit, or a later custom-template import mechanism.

### D5 - GitHub enforcement artifacts are deferred

The v0.8.4 feature does not generate `.github/` enforcement files:

- `.github/CODEOWNERS`
- `.github/pull_request_template.md`
- `.github/rulesets/*.json`
- `.github/workflows/*.yml`

Those artifacts move aikata from "document-centered collaboration
scaffold" toward "GitHub repository setup tool". They can still be
valuable, but they need a separate opt-in design because they are more
environment-specific and more likely to collide with existing repo
configuration.

A future command may extend the same domain, for example:

```bash
aikata enable workflow git --with-github-files
```

or a separate capability may own GitHub enforcement artifacts. That
decision is deliberately out of scope for v0.8.4.

### D6 - Custom import is deferred

Importing a local or remote workflow source such as:

```bash
aikata import workflow --from <path>
```

is deferred. It is useful for personal vault-to-repo workflows and team
template repositories, but it raises questions about trust, template
variables, sync ownership, and conflict behaviour. v0.8.4 should first
establish the built-in document slot and command shape.

## Consequences

### Positive

- Agents get a deterministic place to read operational collaboration
  rules without bloating `AGENTS.md`.
- The design stays aligned with top-level minimalism: no new root file.
- The first feature is useful immediately for Git-heavy AI
  collaboration, while the command and config shape remain broad enough
  for non-Git workflow guides later.
- The Do-No-Harm Policy is preserved: the feature is default-off,
  isolated under `docs/workflows/`, and inert to projects that do not
  enable it.

### Negative

- Adds another optional document category that users and agents must
  understand.
- Introduces a config axis (`workflows:`) outside `components:`; this is
  cleaner long-term but requires schema / loader work rather than a
  one-boolean component addition.
- The first release stops at documentation, so users still need to set
  GitHub branch protection, rulesets, CODEOWNERS, and PR templates by
  hand or through another tool.

## Alternatives Considered

- **Inline the Git workflow into `AGENTS.md`.** Rejected: bloats the
  canonical instruction file and makes non-Git projects pay for a Git-
  specific policy.
- **Add `docs/git-workflow.md`.** Rejected: less extensible than
  `docs/workflows/git.md` and creates a pattern that does not generalize
  well to multiple workflow domains.
- **Use `CONTRIBUTING.md`.** Rejected for v0.8.4: contributor-facing
  governance remains part of the future `extended` pack. AI-facing
  operational workflow should be available without enabling the whole
  governance surface.
- **Add `aikata enable git-workflow`.** Rejected: the name is easy to
  understand but too narrow for a category that should grow to other
  workflow domains.
- **Generate GitHub enforcement files immediately.** Rejected: useful
  but too environment-specific for the first step and likely to collide
  with existing `.github/` state.

## Verification

The v0.8.4 implementation should include:

- golden tests for enabling `workflow git`;
- config persistence tests for `workflows: [git]`;
- sync / manifest coverage showing `docs/workflows/git.md` participates
  like other aikata-managed documents;
- doctor coverage for frontmatter and links in the new guide;
- a minimal / default golden assertion that `docs/workflows/` and
  workflow references are absent unless the guide is enabled.
