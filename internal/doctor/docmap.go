package doctor

import (
	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/docmap"
)

// checkDocMap verifies that `.aikata/docmap.yaml` is current — the
// freshness guarantee for the doc map, which is deliberately not
// manifest-tracked nor merged by sync (ADR 0044 D6). It rebuilds the map
// in memory and compares; drift or a missing map is reported as a
// warning (never an error), and `aikata doctor --fix` regenerates it.
//
// The check only applies to an established aikata project: a directory
// with no `.aikata/aikata.yaml` is not expected to carry a doc map, so
// the check is a no-op there (it does not nag arbitrary repositories).
// The managed/external flag reuses the same ManagedIncludeGlobs the rest
// of doctor uses, so the rebuild matches what `aikata map` writes.
func checkDocMap(opts Options) ([]Issue, error) {
	if _, err := config.Resolve(opts.TargetDir); err != nil {
		return nil, nil
	}
	bopts := docmap.OptionsFor(opts.TargetDir, ManagedIncludeGlobs(opts.TargetDir))
	fresh, missing, err := docmap.Fresh(bopts)
	if err != nil {
		return nil, err
	}
	switch {
	case missing:
		return []Issue{{
			Level: LevelWarning, File: ".aikata/docmap.yaml",
			Code:    "docmap.missing",
			Message: "doc map not generated; run `aikata map` or `aikata doctor --fix`",
		}}, nil
	case !fresh:
		return []Issue{{
			Level: LevelWarning, File: ".aikata/docmap.yaml",
			Code:    "docmap.stale",
			Message: "doc map is out of date; run `aikata map` or `aikata doctor --fix`",
		}}, nil
	}
	return nil, nil
}
