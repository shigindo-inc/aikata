---
project: aikata
status: draft
version: 0.0.1
updated: 2026-11-21
audience: [human, agent]
---

# ADR 0049 — Doctor reports what needs action

- **Status**: Accepted
- **Date**: 2026-11-21
- **Related**: ADR 0039 (its "info-level GLOSSARY orphans" bullet is
  superseded here), ADR 0033 (doctor scope)

## Context

A downstream project with 182 ADRs ran `doctor` and got 15 errors, 85
warnings, and 112 infos. 84 of the warnings were "ADR missing **Status**
line" on ADRs that do declare a status, either as a Japanese
`**状態**:` line or only in the frontmatter `status:` key. 100 of the infos
were GLOSSARY terms no other document mentions. Meanwhile four duplicate
ADR numbers, the one finding that needed a fix, were reported as info.
With that much noise nobody read the report, and no gate ran it.

## Decision

1. The ADR status is read from the body metadata line, accepting
   `**Status**` or `**状態**` with an ASCII or full-width colon. If there is
   no such line, the frontmatter `status` value is used. The body line
   wins because aikata's own ADRs use frontmatter `status` for the
   document lifecycle (`draft`) and the body line for the decision.
2. A duplicate ADR number is an error. "ADR NNNN" in any reference becomes
   ambiguous, so one of the files must be renumbered. A gap in the
   numbering stays info: it is usually a retired number.
3. The GLOSSARY "term is not referenced" check runs only with
   `doctor --verbose`. Unreferenced terms are usually core vocabulary
   (ADR 0039), so reporting each by default buried actionable findings.

## Consequences

- Projects with duplicate ADR numbers start exiting 3 from `doctor`. That
  is intended; renumbering is the fix.
- The default report gets shorter; `--verbose` restores the glossary
  advisory.
- `--json` consumers see new `code` values: `adr.missing-status` and
  `adr.duplicate-number`.

## Alternatives Considered

- Make the status key configurable. More surface than the two spellings
  actually seen in use.
- Keep the glossary check at info by default and rely on `doctor.exclude`.
  Excluding the file also disables its frontmatter checks.
- Raise numbering gaps to warning. Retired numbers are legitimate and
  would leave a warning nobody can clear.
