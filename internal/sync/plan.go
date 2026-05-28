package sync

import (
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
)

// withFlags mirrors the optional-component knobs that scaffold.Options
// carries. We keep them in one struct so derivePlan returns a single
// value (LoD) instead of six parameters.
type withFlags struct {
	WithMemory    bool
	WithUI        bool
	WithAPI       bool
	WithTDD       bool
	WithChangelog bool
	WithMonorepo  bool
}

// overrides carries one-off CLI scope overrides into derivePlan. Each
// pointer field is nil when the user did not pass the matching CLI
// flag. Per ADR 0013, overrides have the highest priority in the
// scope-derivation hierarchy (CLI > manifest > aikata.yaml > defaults)
// and are intentionally **not** written back to either config or the
// manifest — they are a transient lens over one sync invocation.
type overrides struct {
	Preset       *string
	Lang         *string
	Stacks       *[]string
	WithMonorepo *bool
}

// derivePlan picks the preset / lang / opt-in component flags / stack
// list that sync should re-render with. The hierarchy is:
//
//  1. CLI overrides (sticky for one invocation only).
//  2. The manifest (init-time + post-init `aikata add` history, per
//     ADR 0014).
//  3. `.aikata/aikata.yaml` preferences (`components.*`, `stacks`,
//     legacy `features.monorepo` / `features.tdd`, `project.lang`).
//  4. Defaults (preset=standard, lang=en).
//
// In manifest-present mode the manifest is the primary source for
// preset / lang / `--with-*` flags, but aikata.yaml's `stacks` and
// component flags are OR-merged on top — the manifest records "files
// that exist", aikata.yaml records the user's intent, and both
// signals together cover post-init opt-ins the manifest may not have
// grown yet (especially for legacy projects).
//
// In rebaseline mode (manifest absent) the manifest contributes
// nothing and aikata.yaml plus defaults carry the load. Per ADR 0011
// D4, rebaseline writes a manifest using the **upstream rendering**
// as the ancestor, so honouring aikata.yaml here is how the user's
// preferences make it into that initial render.
//
// Schema-v2 (`components.*`) and the legacy `features.{tdd,monorepo}`
// keys are both honoured during the v0.x compatibility window so
// projects that have not yet been migrated by a write-side path do
// not regress. The migrator in [config.MigrateAikataYaml] keeps
// rewriting them into the new shape whenever a writer (`aikata
// generate` / `aikata sync` / `aikata doctor --fix`) runs.
func derivePlan(ancestor config.Manifest, cfg config.AikataYaml, manifestPresent bool, ov overrides) (
	preset, lang string, flags withFlags, stacks []string,
) {
	if manifestPresent {
		preset = ancestor.Preset
		lang = ancestor.Lang
		flags = inferFlags(ancestor)
	} else {
		preset = "standard"
		lang = cfg.Project.Lang
		if lang == "" {
			lang = "en"
		}
	}

	// aikata.yaml: pull stacks and the component flags scaffold
	// honours. Schema-v2 `components.*` is the canonical source;
	// pre-v2 `features.{tdd,monorepo}` is still OR-merged so projects
	// that have not been touched by a writer since v0.6.x continue to
	// pick up their existing intent. Other Features keys
	// (`obsidian_hints`) are deliberately ignored — they have no
	// scaffold-side effect.
	stacks = append([]string(nil), cfg.Stacks...)
	if cfg.Components.Memory {
		flags.WithMemory = true
	}
	if cfg.Components.UI {
		flags.WithUI = true
	}
	if cfg.Components.API {
		flags.WithAPI = true
	}
	if cfg.Components.TDD {
		flags.WithTDD = true
	}
	if cfg.Components.Changelog {
		flags.WithChangelog = true
	}
	if cfg.Components.Monorepo {
		flags.WithMonorepo = true
	}
	if cfg.Features != nil {
		if cfg.Features["monorepo"] {
			flags.WithMonorepo = true
		}
		if cfg.Features["tdd"] {
			flags.WithTDD = true
		}
	}

	// CLI overrides (last-write-wins). nil-checked so absent flags
	// don't accidentally clear values inferred above.
	if ov.Preset != nil {
		preset = *ov.Preset
	}
	if ov.Lang != nil {
		lang = *ov.Lang
	}
	if ov.Stacks != nil {
		stacks = append([]string(nil), (*ov.Stacks)...)
	}
	if ov.WithMonorepo != nil {
		flags.WithMonorepo = *ov.WithMonorepo
	}
	return
}

// inferFlags maps the manifest's recorded file paths back to the
// `--with-*` flags the user originally passed. Detection rules are
// kept narrow: each flag is owned by exactly one canonical filename
// (or path prefix) so a false positive is unlikely.
//
// monorepo detection looks for either the `docs/monorepo.md`
// explainer or any `apps/...` entry; both are written by the
// monorepo component's renderer.
func inferFlags(m config.Manifest) withFlags {
	var f withFlags
	for _, file := range m.Files {
		switch {
		case strings.HasPrefix(file.Path, "docs/memory/"):
			f.WithMemory = true
		case file.Path == "UI.md":
			f.WithUI = true
		case file.Path == "API.md":
			f.WithAPI = true
		case file.Path == "docs/testing.md":
			f.WithTDD = true
		case file.Path == "CHANGELOG.md":
			f.WithChangelog = true
		case file.Path == "docs/monorepo.md", strings.HasPrefix(file.Path, "apps/"):
			f.WithMonorepo = true
		}
	}
	return f
}
