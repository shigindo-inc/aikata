// Package config reads and writes aikata's project configuration file.
//
// Starting in v0.3.2 the canonical path is .aikata/aikata.yaml
// (ADR 0008); .ai/aikata.yaml remains a read-only fallback so v0.2
// and v0.3.0 / v0.3.1 projects continue to load. Resolve() picks the
// right path for a given project root and reports whether the legacy
// fallback was used so callers can surface a deprecation warning.
//
// Defines the project-level configuration schema (version: 1) used by
// aikata init, aikata add, and aikata generate. Migration logic for
// future schema versions lives here.
//
// See ARCHITECTURE.md §4.
package config
