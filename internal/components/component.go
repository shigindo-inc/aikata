package components

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"time"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/templates"
)

// Component is the contract every entry in the registry implements.
// The interface is deliberately narrow: a name + metadata for
// discoverability, and Add for the authoring operation. Argument
// validation lives inside Add so each component can shape its own
// errors without forcing a one-size-fits-all spec.
type Component interface {
	Name() string
	Description() string
	Status() string
	// Add executes the component's authoring operation. Contract:
	//   - failure must leave the filesystem unchanged (all-or-nothing)
	//   - idempotent when feasible (return nil + notice on Stderr when
	//     there is nothing to do)
	//   - dry-run support is the caller's responsibility via ctx.DryRun
	Add(ctx AddContext) error
}

// Status values used by Component.Status. StatusActive marks a
// component the registry can apply today; StatusReserved holds a name
// for a future release.
const (
	StatusActive   = "active"
	StatusReserved = "reserved"
)

// AddContext carries the inputs every Component.Add invocation needs.
// Required fields:
//
//   - TargetDir — absolute project root
//   - ProjectName — value of project.name in .aikata/aikata.yaml
//
// Optional fields:
//
//   - Lang — "en" / "ja"; empty defaults to "en"
//   - Clock — time source for template helpers; nil = time.Now
//   - Args — positional arguments after `aikata add <name>`
//   - DryRun — print plan to Stdout without writing
//   - Stdout / Stderr — message sinks; nil falls back to os.Stdout / os.Stderr
type AddContext struct {
	TargetDir   string
	ProjectName string
	Lang        string
	Clock       templates.Clock
	Args        []string
	DryRun      bool
	Stdout      io.Writer
	Stderr      io.Writer
}

// Now returns the timestamp for this Add invocation, derived from
// AddContext.Clock or time.Now when Clock is nil.
func (c AddContext) Now() time.Time {
	if c.Clock != nil {
		return c.Clock()
	}
	return time.Now()
}

// RecordInManifest registers the given (path → content) entries in
// `.aikata/manifest.yaml` so a subsequent `aikata sync` recognises
// them as aikata-managed templates (see ADR 0014). Hashes recorded are
// the SHA-256 of the freshly-rendered template content — which makes
// any on-disk customisation register as `user-only-edit` on the next
// sync rather than `upstream-added`.
//
// Contract:
//
//   - Loads the existing manifest if present and merges; absent →
//     seeds a fresh manifest with version=1, preset="standard", lang
//     from `.aikata/aikata.yaml`. The chosen defaults are conservative
//     and match what the rebaseline path does for legacy projects.
//   - Empty `rendered` map is a no-op (returns nil without I/O).
//   - Idempotent: re-calling with the same rendered map produces the
//     same manifest bytes (entries are keyed by path; later writes
//     overwrite earlier hashes).
//   - Component implementations call this immediately after their
//     file-write step (whether writeIfMissing or direct os.WriteFile),
//     regardless of whether the on-disk file was newly written or
//     skipped — the manifest tracks the template ancestor, not what's
//     on disk.
//   - Components that author no files (ai_tool, adr) must not call
//     this. Stack only calls it when it actually emitted the stack
//     guide.
func RecordInManifest(targetDir string, rendered map[string]string) error {
	if len(rendered) == 0 {
		return nil
	}
	if targetDir == "" {
		return fmt.Errorf("components: RecordInManifest: target directory is required")
	}

	m, lerr := config.LoadManifest(targetDir)
	if lerr != nil {
		if !errors.Is(lerr, fs.ErrNotExist) {
			return fmt.Errorf("components: RecordInManifest: load manifest: %w", lerr)
		}
		// Seed a fresh manifest. preset is not stored in aikata.yaml
		// today, so default to "standard" — the same default the sync
		// rebaseline path uses for legacy projects. lang comes from
		// aikata.yaml; fall back to "en" if absent.
		//
		// When aikata.yaml is also absent we infer that the project
		// was never init'd by aikata. Recording a manifest in that
		// case would silently mint .aikata/manifest.yaml in someone
		// else's project, which violates Do-No-Harm (ADR 0003); we
		// no-op instead. The file write has already happened by the
		// caller, so nothing is left dangling.
		cfg, _, cerr := config.Load(targetDir)
		if cerr != nil {
			if errors.Is(cerr, fs.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("components: RecordInManifest: load config: %w", cerr)
		}
		lang := cfg.Project.Lang
		if lang == "" {
			lang = "en"
		}
		m = config.Manifest{
			Version: config.ManifestVersion,
			Preset:  "standard",
			Lang:    lang,
		}
	}

	byPath := make(map[string]string, len(m.Files)+len(rendered))
	for _, f := range m.Files {
		byPath[f.Path] = f.SHA256
	}
	for path, body := range rendered {
		byPath[path] = config.HashContent([]byte(body))
	}

	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	files := make([]config.ManifestFile, 0, len(paths))
	for _, p := range paths {
		files = append(files, config.ManifestFile{Path: p, SHA256: byPath[p]})
	}
	m.Files = files

	if err := config.SaveManifest(targetDir, m); err != nil {
		return fmt.Errorf("components: RecordInManifest: %w", err)
	}
	return nil
}
