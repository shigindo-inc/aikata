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
		return linkedVersion
	}
	if ok && info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	if linkedVersion != "" {
		return linkedVersion
	}
	return devVersion
}
