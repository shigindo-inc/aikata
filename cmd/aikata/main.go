// Command aikata is the entry point for the aikata CLI.
//
// All command logic lives in internal/cli; this file exists only to wire
// the OS process to the cobra command tree and translate the cobra exit
// state into a process exit code.
package main

import (
	"errors"
	"os"

	"github.com/shigindo-inc/aikata/internal/cli"
)

// version is the aikata binary version. Overridden at release time via
// -ldflags "-X main.version=v0.1.0".
var version = "0.0.1-dev"

func main() {
	err := cli.Execute(version)
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
