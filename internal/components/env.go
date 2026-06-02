package components

// envSingleton is the registered Component instance for
// `aikata enable env`. It scaffolds `.env.example` — a placeholder
// listing of the project's environment variables (ADR 0037). The file
// was default-scaffolded by the standard / stack scopes through
// v0.9.6; v0.9.7 moves it behind an opt-in capability so a repository
// with no runtime environment variables (a documentation or Agent
// Skills repo, say) no longer carries env residue (ADR 0003
// zero-residue; Q-INTEROP-05). When enabled it pairs with the
// `.env` ignore lines in the scaffolded `.gitignore`.
var envSingleton = singleFile{
	name:       "env",
	desc:       "Environment-variable template at .env.example.",
	targetPath: ".env.example",
	tmplBase:   "components/env",
	tmplName:   ".env.example.tmpl",
}

// Env is the singleton registered in the components registry.
var Env Component = envSingleton

// RenderEnv is the init-time entry point used by scaffold.Run when
// `--with-env` is set. The shape mirrors components.RenderPrompts so
// scaffold dispatches to every optional single-file component through
// one shared table.
func RenderEnv(p SingleFileParams) (map[string]string, error) {
	return envSingleton.render(p.Lang, p.ProjectName, p.Clock)
}
