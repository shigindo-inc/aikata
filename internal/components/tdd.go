package components

var tddSingleton = singleFile{
	name:       "tdd",
	desc:       "Test strategy at docs/testing.md.",
	targetPath: "docs/testing.md",
	tmplBase:   "components/tdd",
	tmplName:   "tdd.md.tmpl",
}

// TDD is the singleton registered in the components registry.
var TDD Component = tddSingleton

// RenderTDD is the init-time entry point used by scaffold.Run when
// `--with-tdd` is set.
func RenderTDD(p SingleFileParams) (map[string]string, error) {
	return tddSingleton.render(p.Lang, p.ProjectName, p.Clock)
}
