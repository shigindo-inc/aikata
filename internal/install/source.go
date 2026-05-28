// Package install records and reads metadata about how the running
// aikata binary was installed. v0.6 establishes the surface so v0.8.x
// or later releases that ship native `aikata update` (self-update)
// can pick the right upgrade path per install channel (ADR 0009).
//
// Detection priority (highest first):
//
//  1. Build-time ldflag override. GoReleaser builds bake
//     `installSource=github-release` so binaries from
//     `https://github.com/.../releases/...` self-identify even when
//     no sibling metadata file is present.
//  2. Sibling metadata file
//     (`<install-dir>/aikata.install-source`). `scripts/install.sh`
//     writes this; future Homebrew / npm wrappers will write
//     `homebrew` and `npm` to the same path.
//  3. Module version from runtime/debug.ReadBuildInfo. A
//     non-pseudo-version with a recognizable module path → `go-install`.
//  4. Fallback → `unknown`. Honest about the absence of signal
//     instead of guessing.
//
// The package never modifies the metadata file, only reads it. Writes
// are owned by the install channel (install.sh / Homebrew formula /
// npm wrapper). This keeps aikata itself from inventing install
// claims it cannot verify.
package install

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Source enumerates the install channels aikata recognizes today.
// Stored as a string so the value flows through JSON envelopes
// unchanged.
type Source string

const (
	// SourceUnknown is the fallback when no channel can be identified.
	SourceUnknown Source = "unknown"
	// SourceGitHubRelease covers binaries shipped via GoReleaser to a
	// GitHub Release (the canonical pre-built channel).
	SourceGitHubRelease Source = "github-release"
	// SourceInstallScript covers `scripts/install.sh` installs, which
	// write the metadata file next to the placed binary.
	SourceInstallScript Source = "install-script"
	// SourceGoInstall covers `go install
	// github.com/shigindo-inc/aikata/cmd/aikata@<version>`.
	SourceGoInstall Source = "go-install"
	// SourceHomebrew covers `brew install shigindo-inc/tap/aikata`
	// (v0.8.x+ Homebrew tap; placeholder until the formula ships).
	SourceHomebrew Source = "homebrew"
	// SourceNpm covers `npx aikata` / `npm i -g aikata` (v0.8.x+ npm
	// wrapper; placeholder until the wrapper ships).
	SourceNpm Source = "npm"
)

// MetadataFilename is the leaf name of the sibling file that records
// the channel that placed the binary. install.sh writes this; aikata
// reads it.
const MetadataFilename = "aikata.install-source"

// ldflagSource is overridden at build time via
// `-X github.com/shigindo-inc/aikata/internal/install.ldflagSource=...`
// for binaries shipped through channels that own their build (today:
// GoReleaser). Empty means "fall through to runtime detection".
var ldflagSource string

// Detect returns the best-effort Source for the running binary. The
// function never returns an error; an unrecognized environment yields
// SourceUnknown. Callers may pass a nil exePath to use os.Executable;
// non-nil values are used as the lookup root (for tests).
func Detect(exePath string) Source {
	if s := normalizeSource(ldflagSource); s != "" {
		return s
	}
	if s := readMetadataNear(exePath); s != "" {
		return s
	}
	if s := detectFromBuildInfo(); s != "" {
		return s
	}
	return SourceUnknown
}

// normalizeSource maps a raw string (from ldflags or the metadata
// file) to a known Source. Unrecognized inputs return the empty
// string so callers can fall through to the next detection layer.
func normalizeSource(raw string) Source {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "":
		return ""
	case string(SourceGitHubRelease):
		return SourceGitHubRelease
	case string(SourceInstallScript):
		return SourceInstallScript
	case string(SourceGoInstall):
		return SourceGoInstall
	case string(SourceHomebrew):
		return SourceHomebrew
	case string(SourceNpm):
		return SourceNpm
	default:
		return ""
	}
}

// readMetadataNear looks for `aikata.install-source` next to the
// running binary. Returns the empty Source when the file is absent or
// unreadable; missing metadata is not an error.
func readMetadataNear(exePath string) Source {
	if exePath == "" {
		got, err := os.Executable()
		if err != nil {
			return ""
		}
		exePath = got
	}
	dir := filepath.Dir(exePath)
	body, err := os.ReadFile(filepath.Join(dir, MetadataFilename))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ""
		}
		return ""
	}
	first := firstLine(string(body))
	return normalizeSource(first)
}

// detectFromBuildInfo inspects runtime/debug.ReadBuildInfo. A tagged
// `go install` produces a module Version that matches `vX.Y.Z`; a
// pseudo-version indicates a dirty / untagged checkout. We classify
// only the tagged case as `go-install`; pseudo-versions stay
// SourceUnknown so the v0.4.2 dev-build path in internal/release
// still owns that detection lane.
func detectFromBuildInfo() Source {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return ""
	}
	if info.Main.Path == "" {
		return ""
	}
	v := info.Main.Version
	if v == "" || v == "(devel)" {
		return ""
	}
	if strings.Contains(v, "-") || strings.Contains(v, "+") {
		// Pseudo-version or dirty marker. The internal/release
		// package treats those as dev-build for the update check.
		return ""
	}
	return SourceGoInstall
}

// firstLine returns the contents of s up to (but not including) the
// first newline. Used so callers can extend the metadata file with
// additional `key=value` lines without breaking the single-line
// "source" convention.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
