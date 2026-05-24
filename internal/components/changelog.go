package components

var changelogSingleton = singleFile{
	name:       "changelog",
	desc:       "Release notes at CHANGELOG.md (Keep a Changelog format).",
	targetPath: "CHANGELOG.md",
	tmplBase:   "components/changelog",
	tmplName:   "changelog.md.tmpl",
}

// Changelog is the singleton registered in the components registry.
var Changelog Component = changelogSingleton

// RenderChangelog is the init-time entry point used by scaffold.Run
// when `--with-changelog` is set.
func RenderChangelog(p SingleFileParams) (map[string]string, error) {
	return changelogSingleton.render(p.Lang, p.ProjectName, p.Clock)
}
