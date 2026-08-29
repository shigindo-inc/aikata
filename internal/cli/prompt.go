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
// them with an empty Enter. The optional-component fields are kept in
// sync with the `--with-*` flag set.
type promptResult struct {
	Name          string
	Scope         string
	Stack         string
	WithMemory    bool
	WithUI        bool
	WithAPI       bool
	WithTDD       bool
	WithChangelog bool
	WithMonorepo  bool
	WithPrompts   bool
	WithModeling  bool
	WithEnv       bool
	Lang          string
	AITools       []string
}

// promptSkip lets the caller (init.go) tell runPrompt which fields
// the user already pinned via an explicit flag. Skipped questions
// keep the supplied default unchanged and produce no UI. The Name
// field has no Skip — it is required and the prompt asks only when
// the default is empty.
type promptSkip struct {
	Scope         bool
	Stack         bool
	WithMemory    bool
	WithUI        bool
	WithAPI       bool
	WithTDD       bool
	WithChangelog bool
	WithMonorepo  bool
	WithPrompts   bool
	WithModeling  bool
	WithEnv       bool
	Lang          bool
	AITools       bool
}

// validAITools enumerates the tool identifiers init accepts in v0.3.
// Kept narrow on purpose; new tools opt in by adding a generator
// under internal/generate/ and updating this list.
var validAITools = map[string]struct{}{
	"claude": {}, "cursor": {}, "codex": {},
}

// runPrompt asks the user for the values needed by `aikata init`. The
// flow is deliberately bufio-based to keep aikata free of
// bubbletea-class dependencies (see ARCHITECTURE.md §10).
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

	if !skip.Scope {
		defScope := result.Scope
		if defScope == "" {
			defScope = "standard"
		}
		prompt := fmt.Sprintf("Scope (standard | minimal) [%s]: ", defScope)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		switch got {
		case "":
			result.Scope = defScope
		case "standard", "1":
			result.Scope = "standard"
		case "minimal", "2":
			result.Scope = "minimal"
		default:
			return result, fmt.Errorf("prompt: unknown scope %q (expected standard | minimal)", got)
		}
	}

	if !skip.Stack {
		label := result.Stack
		if label == "" {
			label = "none"
		}
		prompt := fmt.Sprintf("Stack (none | flutter | typescript) [%s]: ", label)
		got, err := readLine(sc, w, prompt)
		if err != nil {
			return result, err
		}
		switch got {
		case "":
			// keep default
		case "none":
			result.Stack = ""
		case "flutter", "typescript":
			result.Stack = got
		default:
			return result, fmt.Errorf("prompt: unknown stack %q (expected none | flutter | typescript)", got)
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

	// Optional-component questions. The order matches scaffold.Run's
	// dispatch table so the user-visible sequence and the underlying
	// flag set stay in lockstep.
	optionalQs := []struct {
		skip   bool
		target *bool
		prompt string
	}{
		{skip.WithMemory, &result.WithMemory, "Include long-term memory slot under docs/memory/?"},
		{skip.WithUI, &result.WithUI, "Include UI.md (UI / UX guidelines)?"},
		{skip.WithAPI, &result.WithAPI, "Include API.md (API interface spec)?"},
		{skip.WithTDD, &result.WithTDD, "Include docs/testing.md (test strategy)?"},
		{skip.WithChangelog, &result.WithChangelog, "Include CHANGELOG.md (release notes)?"},
		{skip.WithMonorepo, &result.WithMonorepo, "Configure this project as a monorepo (apps/ + nested AGENTS.md)?"},
		{skip.WithPrompts, &result.WithPrompts, "Include docs/prompts.md (reusable prompt library)?"},
		{skip.WithModeling, &result.WithModeling, "Include docs/usecases.md + docs/domain.md (use cases and domain model)?"},
		{skip.WithEnv, &result.WithEnv, "Include .env.example (environment-variable template)?"},
	}
	for _, q := range optionalQs {
		if q.skip {
			continue
		}
		if err := askYesNo(sc, w, q.prompt, q.target); err != nil {
			return result, err
		}
	}

	return result, nil
}

// askYesNo asks a y/N question, accepts blank / y / yes / n / no
// (case-insensitive), and updates *target in place. Blank input keeps
// the current value of *target as the default. Unknown input returns
// an error so the caller can surface it to the user.
func askYesNo(sc *bufio.Scanner, w io.Writer, question string, target *bool) error {
	defStr := "N"
	if *target {
		defStr = "Y"
	}
	got, err := readLine(sc, w, fmt.Sprintf("%s [y/N, default %s]: ", question, defStr))
	if err != nil {
		return err
	}
	switch strings.ToLower(got) {
	case "":
		// keep default
	case "y", "yes":
		*target = true
	case "n", "no":
		*target = false
	default:
		return fmt.Errorf("prompt: unknown yes/no value %q", got)
	}
	return nil
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
