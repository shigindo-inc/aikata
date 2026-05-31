package components

// appIconSingleton is the registered Component instance for
// `aikata new app-icon`. It stamps a one-off app-icon concept-
// exploration document under docs/design/ (ADR 0031). The CLI
// identifier stays shorter than the filename: `app-icon` is specific
// enough to distinguish the artifact from a favicon, logo, or
// in-product icon library.
var appIconSingleton = oneOffArtifact{
	name:       "app-icon",
	desc:       "App-icon concept exploration at docs/design/app-icon-concepts.md.",
	targetPath: "docs/design/app-icon-concepts.md",
	tmplBase:   "components/app-icon",
	tmplName:   "app-icon.md.tmpl",
}

// AppIcon is the singleton registered in the components registry.
var AppIcon Component = appIconSingleton
