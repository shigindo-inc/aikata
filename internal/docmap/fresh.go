package docmap

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ExistingFiles lists the map files of any format present under
// `.aikata/`, by file name.
func ExistingFiles(root string) ([]string, error) {
	var found []string
	for _, r := range renderers {
		_, err := os.Stat(pathFor(root, r.file))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		found = append(found, r.file)
	}
	return found, nil
}

// FirstDrift reports the first missing or stale enabled map file. An empty
// filename means every enabled view matches the current document map.
func FirstDrift(opts Options) (file string, missing bool, err error) {
	m, err := Build(opts)
	if err != nil {
		return "", false, err
	}
	if err := preserveGeneration(opts.TargetDir, &m); err != nil {
		return "", false, err
	}

	for _, r := range selectedRenderers(opts.Formats) {
		current, err := os.ReadFile(pathFor(opts.TargetDir, r.file))
		if errors.Is(err, fs.ErrNotExist) {
			return r.file, true, nil
		}
		if err != nil {
			return "", false, err
		}
		expected, err := r.body(m)
		if err != nil {
			return "", false, err
		}
		if !bytes.Equal(expected, current) {
			return r.file, false, nil
		}
	}
	return "", false, nil
}

// preserveGeneration keeps the previous date when the map data is identical.
// Both generation and diagnosis use it, so time alone never creates drift.
func preserveGeneration(root string, m *Map) error {
	current, err := os.ReadFile(YAMLPath(root))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var prev Map
	if err := yaml.Unmarshal(current, &prev); err != nil {
		return nil // Malformed generated output can be regenerated.
	}
	if _, err := time.Parse("2006-01-02", prev.Generated); err != nil {
		return nil
	}
	candidate := *m
	candidate.Generated = prev.Generated
	expected, err := MarshalYAML(candidate)
	if err != nil {
		return err
	}
	if bytes.Equal(expected, current) {
		m.Generated = prev.Generated
	}
	return nil
}
