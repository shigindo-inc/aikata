package generate

// CodexProvider is registered for symmetry with claude/cursor but emits
// zero files. OpenAI Codex (CLI and cloud) reads AGENTS.md directly per
// its 2026 spec — root + subdirectory walk plus an optional
// AGENTS.override.md — so the canonical file already satisfies it.
// See ADR 0005 for the pass-through rationale; the cli layer reports
// the no-op to stderr so users who enable `codex` in .ai/aikata.yaml
// understand nothing was written.
type CodexProvider struct{}

// Name implements Provider.
func (CodexProvider) Name() string { return "codex" }

// Files implements Provider. Always returns an empty map.
func (CodexProvider) Files(_ Context) (map[string]string, error) {
	return map[string]string{}, nil
}
