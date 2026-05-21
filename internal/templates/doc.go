// Package templates wraps the embedded template filesystem and renders
// individual templates with the standard library text/template engine plus
// a small set of helper functions (now, joinSlash, kebab, ...).
//
// Templates are embedded at compile time via //go:embed; the binary needs
// no on-disk template directory.
//
// See ARCHITECTURE.md §5.
package templates
