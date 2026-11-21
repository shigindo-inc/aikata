package doctor

import (
	"errors"
	"io/fs"
	"path"
	"strings"

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
	if !bopts.Enabled {
		return checkUnmaintainedDocMap(opts.TargetDir)
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

// checkUnmaintainedDocMap flags map files left on disk while the map is off.
// Nothing rebuilds them, so they drift from the documents without notice.
// Deleting them is the owner's call, so there is no auto-fix.
func checkUnmaintainedDocMap(root string) ([]Issue, error) {
	files, err := docmap.ExistingFiles(root)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	return []Issue{{
		Level: LevelWarning,
		File:  path.Join(config.PrimaryDir, files[0]),
		Code:  "docmap.unmaintained",
		Message: "doc map files (" + strings.Join(files, ", ") + ") are no longer rebuilt; " +
			"set `docmap.enabled: true` in .aikata/aikata.yaml to keep them, or delete them",
	}}, nil
}
