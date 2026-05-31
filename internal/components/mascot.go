package components

// mascotSingleton is the registered Component instance for
// `aikata new mascot`. It stamps a one-off mascot-character idea-
// exploration document under docs/design/ (ADR 0031). `mascot` is
// sufficient as a CLI identifier without repeating `character-ideas`.
var mascotSingleton = oneOffArtifact{
	name:       "mascot",
	desc:       "Mascot-character idea exploration at docs/design/mascot-character-ideas.md.",
	targetPath: "docs/design/mascot-character-ideas.md",
	tmplBase:   "components/mascot",
	tmplName:   "mascot.md.tmpl",
}

// Mascot is the singleton registered in the components registry.
var Mascot Component = mascotSingleton
