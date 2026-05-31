package components

// promptsSingleton is the registered Component instance for
// `aikata enable prompts`. It scaffolds docs/prompts.md — a small,
// project-local library of reusable prompts (ADR 0034). The file was a
// default-scaffolded document through v0.9.1; v0.9.2 moves it behind an
// opt-in capability so the default `standard` scope stays lean and an
// empty prompt library no longer imposes a maintenance cost on every
// project (Q-DESIGN-12).
var promptsSingleton = singleFile{
	name:       "prompts",
	desc:       "Reusable prompt library at docs/prompts.md.",
	targetPath: "docs/prompts.md",
	tmplBase:   "components/prompts",
	tmplName:   "prompts.md.tmpl",
}

// Prompts is the singleton registered in the components registry.
var Prompts Component = promptsSingleton

// RenderPrompts is the init-time entry point used by scaffold.Run when
// `--with-prompts` is set. The shape mirrors components.RenderUI so
// scaffold dispatches to every optional single-file component through
// one shared table.
func RenderPrompts(p SingleFileParams) (map[string]string, error) {
	return promptsSingleton.render(p.Lang, p.ProjectName, p.Clock)
}
