// Package config reads and writes aikata's project configuration file.
//
// The canonical path is .aikata/aikata.yaml (ADR 0008). v0.7.4
// removed the pre-v0.3.2 .ai/aikata.yaml read fallback.
//
// Defines the project-level configuration schema (version: 1) used by
// aikata init, aikata add, and aikata generate. Migration logic for
// future schema versions lives here.
//
// See ARCHITECTURE.md §4.
package config
