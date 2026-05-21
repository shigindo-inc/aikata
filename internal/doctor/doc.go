// Package doctor verifies consistency of an aikata project.
//
// Implements aikata doctor. Read-only by default; the --fix flag enables
// safe auto-corrections (e.g. refreshing a stale `updated:` frontmatter
// after a known edit).
//
// Exit code 3 on errors, per ARCHITECTURE.md §7.
package doctor
