package doctor

import (
	"errors"
	"io/fs"
	"path"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/docmap"
)

// checkDocMap compares the configured map files for established projects.
func checkDocMap(opts Options) ([]Issue, error) {
	if _, err := config.Resolve(opts.TargetDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	bopts, err := docmap.OptionsFor(opts.TargetDir, ManagedIncludeGlobs(opts.TargetDir))
	if err != nil {
		return nil, err
	}
	file, missing, err := docmap.FirstDrift(bopts)
	if err != nil || file == "" {
		return nil, err
	}
	code, message := "docmap.stale", "doc map is out of date"
	if missing {
		code, message = "docmap.missing", "doc map file not generated"
	}
	return []Issue{{
		Level:   LevelWarning,
		File:    path.Join(config.PrimaryDir, file),
		Code:    code,
		Message: message + "; run `aikata map` or `aikata doctor --fix`",
	}}, nil
}
