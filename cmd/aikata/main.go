// Command aikata is the entry point for the aikata CLI.
//
// All command logic lives in internal/cli; this file exists only to wire
// the OS process to the cobra command tree and translate the cobra exit
// state into a process exit code.
package main

import (
	"fmt"
	"os"

	"github.com/shigindo-inc/aikata/internal/cli"
)

// version is the aikata binary version. Overridden at release time via
// -ldflags "-X main.version=v0.1.0".
var version = "0.0.1-dev"

func main() {
	if err := cli.Execute(version); err != nil {
		// cobra has already printed the user-facing error message to
		// stderr. Translate to a non-zero exit code per
		// ARCHITECTURE.md §7.2. Generic errors map to exit code 1;
		// finer-grained codes are emitted by cli.Execute itself via
		// os.Exit when needed.
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
