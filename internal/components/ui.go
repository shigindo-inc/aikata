package components

import "github.com/shigindo-inc/aikata/internal/templates"

// uiSingleton is the registered Component instance for `aikata add ui`.
// Scope is UI / UX / product-design guidance only — generic
// architecture decisions stay in `ARCHITECTURE.md` or an ADR, and
// `DESIGN.md` is intentionally never introduced (ADR 0007).
var uiSingleton = singleFile{
	name:       "ui",
	desc:       "UI / UX / product-design guidelines at UI.md.",
	targetPath: "UI.md",
	tmplBase:   "components/ui",
	tmplName:   "ui.md.tmpl",
}

// UI is the singleton registered in the components registry.
var UI Component = uiSingleton

// RenderUI is the init-time entry point used by scaffold.Run when
// `--with-ui` is set. The shape mirrors components.RenderMemory so
// scaffold dispatches to every optional component through one shared
// table.
func RenderUI(p SingleFileParams) (map[string]string, error) {
	return uiSingleton.render(p.Lang, p.ProjectName, p.Clock)
}

// SingleFileParams carries the inputs every Render<Name> shim needs
// from scaffold.Run. Kept as a reduced shape of scaffold.Options so
// the two callers (init-time and add-time) share one rendering core.
type SingleFileParams struct {
	Lang        string
	ProjectName string
	Clock       templates.Clock
}
