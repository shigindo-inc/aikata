// Package scaffold generates project file sets from presets.
//
// Implements aikata init and aikata add. Writes files atomically: a staging
// directory is materialized first, then renamed into place so a partial
// failure leaves the target directory untouched.
//
// Knows about preset names and template rendering; does not know about
// specific AI tools (those live in internal/generate).
//
// See ARCHITECTURE.md §2.1.
package scaffold
