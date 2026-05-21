// Package cli defines the cobra command tree for aikata.
//
// Responsibilities:
//   - Define every CLI subcommand (init, add, doctor, generate, update, list).
//   - Parse flags and translate them into typed requests for lower layers.
//   - Render user-facing messages and exit codes.
//
// Non-responsibilities:
//   - No file I/O. Delegate to internal/scaffold, internal/generate, etc.
//   - No business rules. The cli package is the seam between the OS and aikata.
//
// See ARCHITECTURE.md §2.1 for the package boundary contract.
package cli
