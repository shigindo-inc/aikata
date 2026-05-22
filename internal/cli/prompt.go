package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// promptResult holds the three values `aikata init` collects
// interactively. Defaults flow in from the cobra flag values so the
// user can accept them with an empty Enter.
type promptResult struct {
	Name       string
	Preset     string
	WithMemory bool
}

// runPrompt asks the user for the values needed by `aikata init`.
// Already-set defaults are shown in brackets; pressing Enter keeps
// them. Each question accepts a short string; the function does not
// try to be a curses-style UI — it's deliberately bufio-based so
// aikata stays Go-1.21 compatible and free of bubbletea-class
// dependencies (see ARCHITECTURE.md §10 / Task 6 design note).
func runPrompt(r io.Reader, w io.Writer, defaults promptResult) (promptResult, error) {
	sc := bufio.NewScanner(r)
	result := defaults

	// Project name — required.
	if result.Name == "" {
		got, err := readLine(sc, w, "Project name: ")
		if err != nil {
			return result, err
		}
		if got == "" {
			return result, errors.New("prompt: project name cannot be empty")
		}
		result.Name = got
	}

	// Preset — Enter keeps the existing default.
	{
		prompt := fmt.Sprintf("Preset (standard | minimal | flutter) [%s]: ", result.Preset)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		switch got {
		case "":
			// keep default
		case "standard", "1":
			result.Preset = "standard"
		case "minimal", "2":
			result.Preset = "minimal"
		case "flutter", "3":
			result.Preset = "flutter"
		default:
			return result, fmt.Errorf("prompt: unknown preset %q (expected standard | minimal | flutter)", got)
		}
	}

	// With-memory — defaults to N when withMemory==false, Y otherwise.
	{
		defStr := "N"
		if result.WithMemory {
			defStr = "Y"
		}
		prompt := fmt.Sprintf("Include long-term memory slot under docs/memory/? [y/N, default %s]: ", defStr)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		switch strings.ToLower(got) {
		case "":
			// keep default
		case "y", "yes":
			result.WithMemory = true
		case "n", "no":
			result.WithMemory = false
		default:
			return result, fmt.Errorf("prompt: unknown yes/no value %q", got)
		}
	}

	return result, nil
}

// readLine prints the prompt to w and returns the user's trimmed
// input. EOF before any input is reported as an error so callers can
// distinguish "user pressed Enter" (empty string) from "stdin closed".
func readLine(sc *bufio.Scanner, w io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", fmt.Errorf("prompt: write: %w", err)
	}
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", fmt.Errorf("prompt: read: %w", err)
		}
		return "", fmt.Errorf("prompt: unexpected EOF")
	}
	return strings.TrimSpace(sc.Text()), nil
}

// isTTYFunc is a package-level variable so tests can swap in a
// deterministic detector. The default reads from os.Stdin's mode.
var isTTYFunc = func() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
