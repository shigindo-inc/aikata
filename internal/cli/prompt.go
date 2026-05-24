package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// promptResult holds the values `aikata init` collects interactively.
// Defaults flow in from the cobra flag values so the user can accept
// them with an empty Enter. v0.3 extended the set with Lang and
// AITools to bring the prompt back to flag parity. See SPEC.md §4.1.
type promptResult struct {
	Name       string
	Preset     string
	WithMemory bool
	Lang       string
	AITools    []string
}

// promptSkip lets the caller (init.go) tell runPrompt which fields
// the user already pinned via an explicit flag. Skipped questions
// keep the supplied default unchanged and produce no UI. The Name
// field has no Skip — it is required and the prompt asks only when
// the default is empty.
type promptSkip struct {
	Preset     bool
	WithMemory bool
	Lang       bool
	AITools    bool
}

// validAITools enumerates the tool identifiers init accepts in v0.3.
// Kept narrow on purpose; new tools opt in by adding a generator
// under internal/generate/ and updating this list.
var validAITools = map[string]struct{}{
	"claude": {}, "cursor": {}, "codex": {},
}

// runPrompt asks the user for the values needed by `aikata init`. The
// flow is deliberately bufio-based so aikata stays Go-1.21 compatible
// and free of bubbletea-class dependencies (see ARCHITECTURE.md §10).
// Defaults are shown in brackets; pressing Enter keeps the default.
// Questions whose corresponding flag was explicitly supplied are
// silently skipped via the `skip` argument.
func runPrompt(r io.Reader, w io.Writer, defaults promptResult, skip promptSkip) (promptResult, error) {
	sc := bufio.NewScanner(r)
	result := defaults

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

	if !skip.Preset {
		prompt := fmt.Sprintf("Preset (standard | minimal | flutter | typescript) [%s]: ", result.Preset)
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
		case "typescript", "4":
			result.Preset = "typescript"
		default:
			return result, fmt.Errorf("prompt: unknown preset %q (expected standard | minimal | flutter | typescript)", got)
		}
	}

	if !skip.Lang {
		defLang := result.Lang
		if defLang == "" {
			defLang = "en"
		}
		prompt := fmt.Sprintf("Document language (en | ja) [%s]: ", defLang)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		switch got {
		case "":
			result.Lang = defLang
		case "en", "ja":
			result.Lang = got
		default:
			return result, fmt.Errorf("prompt: unknown language %q (expected en | ja)", got)
		}
	}

	if !skip.AITools {
		defStr := strings.Join(result.AITools, ",")
		if defStr == "" {
			defStr = "claude"
		}
		prompt := fmt.Sprintf("AI tools (comma-separated; claude | cursor | codex) [%s]: ", defStr)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		raw := got
		if raw == "" {
			raw = defStr
		}
		parsed, err := parseAITools(raw)
		if err != nil {
			return result, err
		}
		result.AITools = parsed
	}

	if !skip.WithMemory {
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

// parseAITools normalizes a comma-separated tool list into a sorted,
// de-duplicated slice with every entry validated against
// validAITools. Empty inputs yield an error so the caller can re-ask;
// runPrompt substitutes the default before calling this so end users
// don't see that case.
func parseAITools(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("prompt: ai-tools list cannot be empty")
	}
	seen := map[string]struct{}{}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		tool := strings.ToLower(strings.TrimSpace(part))
		if tool == "" {
			continue
		}
		if _, ok := validAITools[tool]; !ok {
			return nil, fmt.Errorf("prompt: unknown ai-tool %q (expected claude | cursor | codex)", tool)
		}
		if _, dup := seen[tool]; dup {
			continue
		}
		seen[tool] = struct{}{}
		out = append(out, tool)
	}
	if len(out) == 0 {
		return nil, errors.New("prompt: ai-tools list resolved to empty after trimming")
	}
	return out, nil
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
