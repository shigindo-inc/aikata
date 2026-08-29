package doctor

import "github.com/shigindo-inc/aikata/internal/config"

// canonicalDocNames are the top-level Markdown documents aikata
// scaffolds and therefore owns the frontmatter contract for. They form
// part of the default managed-surface walk (ADR 0033 direction;
// ADR 0037 D1). Optional single-file components (UI.md, API.md,
// CHANGELOG.md) are listed so they are validated whenever a project
// has enabled them.
var canonicalDocNames = []string{
	"README.md",
	"AGENTS.md",
	"SPEC.md",
	"ARCHITECTURE.md",
	"GLOSSARY.md",
	"ROADMAP.md",
	"CHANGELOG.md",
	"UI.md",
	"API.md",
}

// managedDocGlobs are the docs/ subtrees and individual documents
// aikata scaffolds. Expressed as internal/glob patterns where `**`
// matches zero or more complete path segments.
var managedDocGlobs = []string{
	"docs/adr/**",
	"docs/memory/**",
	"docs/stacks/**",
	"docs/tasks/**",
	"docs/workflows/**",
	"docs/design/**",
	"docs/troubleshooting.md",
	"docs/prompts.md",
	// The opt-in modeling pair. Listed individually (not as a subtree)
	// because they are two fixed files, and validated whenever a
	// project has enabled the capability.
	"docs/usecases.md",
	"docs/domain.md",
	"docs/monorepo.md",
	// Monorepo per-app instructions follow aikata's AGENTS.md
	// frontmatter contract even though they are user-managed (the
	// monorepo layout seeds apps/_example/AGENTS.md). Keep them in the
	// default walk so per-app docs are not silently unvalidated.
	"apps/**/AGENTS.md",
}

// ManagedIncludeGlobs returns the include-glob set that scopes the
// default `aikata doctor` walk to the document surface aikata manages
// (ADR 0033 direction; ADR 0037 D1 behaviour). It unions:
//
//   - the canonical top-level document names;
//   - the known managed docs/ subtrees;
//   - every path recorded in `.aikata/manifest.yaml`, when present,
//
// so adopted and pre-manifest projects keep coherent validation rather
// than silently losing it. The result is passed as Options.Includes;
// walkMarkdown then validates only Markdown files matching one of these
// globs. A malformed or absent manifest falls back to the static set.
//
// The broad-audit mode (`aikata doctor --all-markdown`) does not call
// this — it leaves Options.Includes empty so the walk covers every
// Markdown file as before.
func ManagedIncludeGlobs(targetDir string) []string {
	globs := make([]string, 0, len(canonicalDocNames)+len(managedDocGlobs)+8)
	globs = append(globs, canonicalDocNames...)
	globs = append(globs, managedDocGlobs...)

	m, err := config.LoadManifest(targetDir)
	if err != nil {
		// fs.ErrNotExist is the common adopted/pre-manifest case; a
		// malformed manifest must likewise not crash doctor. Either way,
		// fall back to the static managed set.
		return globs
	}
	for _, f := range m.Files {
		globs = append(globs, f.Path)
	}
	return globs
}
