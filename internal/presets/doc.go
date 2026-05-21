// Package presets defines the built-in preset registry and the rules for
// combining presets with feature flags.
//
// A preset is a named bundle of templates plus default feature flags.
// Built-in presets ship as embedded directories under templates/presets/.
//
// Pure logic, no I/O — file operations belong in internal/scaffold.
package presets
