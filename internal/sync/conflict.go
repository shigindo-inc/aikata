package sync

import (
	"bytes"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/shigindo-inc/aikata/internal/config"
	"github.com/shigindo-inc/aikata/internal/managed"
)

const (
	conflictStart    = "<<<<<<< current (your edits)"
	conflictAncestor = "||||||| ancestor (.aikata/manifest.yaml)"
	conflictEnd      = ">>>>>>> upstream"
)

func hasSyncConflictMarkers(body []byte) bool {
	for line := range bytes.Lines(body) {
		switch strings.TrimRight(string(line), "\r\n") {
		case conflictStart, conflictAncestor, conflictEnd:
			return true
		}
	}
	return false
}

func acceptConflicts(opts Options, manifest config.Manifest, hashes map[string]string, upstream map[string]string, isOwned func(string) bool) (RunResult, error) {
	var result RunResult
	accepted := make(map[string]string)
	for _, raw := range opts.AcceptConflicts {
		path := filepath.ToSlash(filepath.Clean(raw))
		body, exists := upstream[path]
		if !exists || isOwned(path) || managed.IsAppendPath(path) {
			return RunResult{}, fmt.Errorf("cannot accept %q: it is not a mergeable template path", raw)
		}
		if _, seen := accepted[path]; seen {
			continue
		}
		file, marked, err := classifyAndMerge(path, hashes, upstream, opts.Root)
		if err != nil {
			return RunResult{}, err
		}
		if file.Status == StatusConflict && marked == nil {
			return RunResult{}, fmt.Errorf("cannot accept %q: remove the remaining sync conflict markers first", raw)
		}
		if file.Status != StatusConflict {
			return RunResult{}, fmt.Errorf("cannot accept %q: no divergent resolution for the selected templates; run ordinary sync for clean paths", raw)
		}
		accepted[path] = config.HashContent([]byte(body))
		result.Files = append(result.Files, FileResult{Path: path, Status: StatusResolutionAccepted})
		result.Applied++
	}

	if !opts.DryRun {
		maps.Copy(hashes, accepted)
		manifest.Files = make([]config.ManifestFile, 0, len(hashes))
		for _, path := range slices.Sorted(maps.Keys(hashes)) {
			manifest.Files = append(manifest.Files, config.ManifestFile{Path: path, SHA256: hashes[path]})
		}
		if err := config.SaveManifest(opts.Root, manifest); err != nil {
			return RunResult{}, fmt.Errorf("record resolved conflicts: %w", err)
		}
	}
	return result, nil
}
