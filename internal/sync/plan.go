package sync

import (
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
)

// withFlags mirrors the optional-component knobs that scaffold.Options
// carries. We keep them in one struct so derivePlan returns a single
// value (LoD) instead of five parameters.
type withFlags struct {
	WithMemory    bool
	WithUI        bool
	WithAPI       bool
	WithTDD       bool
	WithChangelog bool
}

// derivePlan picks the preset / lang / opt-in component flags that
// sync should re-render with. In normal mode it trusts the manifest
// (preset + lang) and infers the flags from the manifest's file list
// (presence-based). In rebaseline mode (manifest absent) it falls
// back to `.aikata/aikata.yaml`'s project.lang and the safest preset
// default; the user can re-run with explicit overrides in a v0.5.x
// follow-up.
func derivePlan(ancestor config.Manifest, cfg config.AikataYaml, manifestPresent bool) (preset, lang string, flags withFlags) {
	if manifestPresent {
		preset = ancestor.Preset
		lang = ancestor.Lang
		flags = inferFlags(ancestor)
		return
	}
	preset = "standard"
	lang = cfg.Project.Lang
	if lang == "" {
		lang = "en"
	}
	// Without a manifest we have no signal for the opt-ins; leave
	// them false. The rebaseline run will record whatever is on
	// disk under the standard preset's surface, and the next real
	// sync will use that snapshot as the ancestor.
	return
}

// inferFlags maps the manifest's recorded file paths back to the
// `--with-*` flags the user originally passed. Detection rules are
// kept narrow: each flag is owned by exactly one canonical filename
// so a false positive is unlikely.
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
		}
	}
	return f
}
