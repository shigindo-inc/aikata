// Command aikata is the entry point for the aikata CLI.
//
// All command logic lives in internal/cli; this file exists only to wire
// the OS process to the cobra command tree and translate the cobra exit
// state into a process exit code.
package main

import (
	"errors"
	"os"
	"runtime/debug"
	"strings"

	"github.com/shigindo-inc/aikata/internal/cli"
)

// version is the aikata binary version. Overridden at release time via
// -ldflags "-X main.version=v0.1.0".
const devVersion = "0.0.1-dev"

var version = devVersion

func main() {
	err := cli.Execute(effectiveVersion(version))
	if err == nil {
		return
	}
	// cobra has already printed the user-facing error message to
	// stderr. Map the error to an exit code per ARCHITECTURE.md §7.2:
	// unwrap an *ExitError if one is present, otherwise fall back to 1.
	var ee *cli.ExitError
	if errors.As(err, &ee) {
		os.Exit(ee.Code)
	}
	os.Exit(1)
}

func effectiveVersion(linkedVersion string) string {
	info, ok := debug.ReadBuildInfo()
	return resolveVersion(linkedVersion, info, ok)
}

func resolveVersion(linkedVersion string, info *debug.BuildInfo, ok bool) string {
	if linkedVersion != "" && linkedVersion != devVersion {
		return normalizeVersion(linkedVersion)
	}
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return normalizeVersion(info.Main.Version)
	}
	if linkedVersion != "" {
		return linkedVersion
	}
	return devVersion
}

// normalizeVersion forces release versions to a single canonical
// "vX.Y.Z" shape. `go install` already returns the v-prefixed module
// version, but GoReleaser ldflags inject the bare semver string and
// custom build invocations may pass either form; normalizing at the
// boundary keeps `aikata --version` output identical across channels.
// The dev sentinel and empty strings are returned unchanged.
func normalizeVersion(s string) string {
	if s == "" || s == devVersion || strings.HasPrefix(s, "v") {
		return s
	}
	return "v" + s
}
