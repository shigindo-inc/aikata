// Package generate produces per-AI-tool configuration files from canonical
// sources (AGENTS.md + SPEC.md + ARCHITECTURE.md + GLOSSARY.md, etc.).
//
// Each supported tool implements a small Provider interface (planned),
// allowing Claude, Cursor, Codex, Gemini, Copilot, and Windsurf to be added
// independently without changes elsewhere.
//
// Generated artifacts are explicitly lossy and disposable: the canonical
// markdown remains authoritative (see ADR 0002).
package generate
