package docmap

import (
	"bytes"
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// Fresh reports whether the on-disk `.aikata/docmap.yaml` still matches a
// fresh in-memory rebuild from opts. It is the comparison behind the
// doctor freshness check (ADR 0044 D6): the map is not manifest-tracked
// and not merged by sync, so this hash compare is what guarantees it
// stays current.
//
//   - missing is true when no map file exists yet (never generated).
//   - fresh is true when the on-disk bytes equal the rebuild.
//
// Only structural drift counts: the rebuild's `generated:` date is
// aligned to the on-disk value before comparing, so the map does not read
// as stale merely because time has passed since it was written. A map
// file that fails to parse is treated as not fresh (drift), so `--fix`
// regenerates it.
func Fresh(opts Options) (fresh bool, missing bool, err error) {
	onDisk, err := os.ReadFile(YAMLPath(opts.TargetDir))
	if errors.Is(err, fs.ErrNotExist) {
		return false, true, nil
	}
	if err != nil {
		return false, false, err
	}
	m, err := Build(opts)
	if err != nil {
		return false, false, err
	}
	var prev Map
	if e := yaml.Unmarshal(onDisk, &prev); e == nil && prev.Generated != "" {
		m.Generated = prev.Generated
	}
	rebuilt, err := MarshalYAML(m)
	if err != nil {
		return false, false, err
	}
	return bytes.Equal(rebuilt, onDisk), false, nil
}
