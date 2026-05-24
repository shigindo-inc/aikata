package components

var apiSingleton = singleFile{
	name:       "api",
	desc:       "API interface specification at API.md.",
	targetPath: "API.md",
	tmplBase:   "components/api",
	tmplName:   "api.md.tmpl",
}

// API is the singleton registered in the components registry.
var API Component = apiSingleton

// RenderAPI is the init-time entry point used by scaffold.Run when
// `--with-api` is set.
func RenderAPI(p SingleFileParams) (map[string]string, error) {
	return apiSingleton.render(p.Lang, p.ProjectName, p.Clock)
}
