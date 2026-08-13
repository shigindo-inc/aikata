package cli

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// newInitCmd builds the `aikata init` subcommand. The scaffolding-time
// surface is two orthogonal axes — `--scope` (documentation breadth)
// and `--stack` (target technology) — plus `--lang`, `--ai-tools`, and
// the optional-component flags (`--with-memory`, `--with-ui`,
// `--with-api`, `--with-tdd`, `--with-changelog`, `--monorepo`,
// `--with-modeling`).
//
// `--preset` survives as a deprecated alias for `--scope` / `--stack`
// (ADR 0024) and is removed in v1.0. The post-init counterpart lives
// behind `aikata enable <capability>`.
func newInitCmd() *cobra.Command {
	var (
		preset        string
		scope         string
		stacks        []string
		name          string
		noInteractive bool
		force         bool
		dryRun        bool
		lang          string
		withMemory    bool
		withUI        bool
		withAPI       bool
		withTDD       bool
		withChangelog bool
		withMonorepo  bool
		withPrompts   bool
		withModeling  bool
		withEnv       bool
		aiToolsCSV    string
	)

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a new aikata project in the current directory",
		Long: "Generate a coherent set of markdown documents for a new project.\n\n" +
			"Two orthogonal axes select what is scaffolded: --scope sets the\n" +
			"documentation breadth (minimal | standard) and --stack adds a\n" +
			"target-technology brief (flutter | typescript). --preset is a\n" +
			"deprecated alias for the two and is removed in v1.0.\n\n" +
			"In non-interactive mode (--no-interactive, or when stdin is not\n" +
			"a TTY), the project name must be supplied either as the positional\n" +
			"argument or via --name. The target directory is always the current\n" +
			"working directory.",
		Args:     cobra.MaximumNArgs(1),
		PostRunE: docMapPostRun, // isolated doc-map rebuild (ADR 0044 D5)
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && name == "" {
				name = args[0]
			}

			aiTools, err := parseAITools(aiToolsCSV)
			if err != nil {
				return &ExitError{Code: 2, Err: err}
			}

			presetSet := cmd.Flags().Changed("preset")
			scopeSet := cmd.Flags().Changed("scope")
			stackSet := cmd.Flags().Changed("stack")

			// --preset is a deprecated alias and must not be mixed with the
			// axes it stands in for: there is no precedence rule to apply.
			if presetSet && (scopeSet || stackSet) {
				return &ExitError{Code: 2, Err: errors.New(
					"--preset cannot be combined with --scope/--stack; --preset is a deprecated alias for them")}
			}

			// Translate the deprecated alias to the native axes up front so
			// the rest of the flow (prompt skip, resolution) is axis-only.
			if presetSet {
				s, st, perr := presetToScopeStack(preset)
				if perr != nil {
					return &ExitError{Code: 2, Err: perr}
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"warning: --preset is deprecated and will be removed in v1.0; use --scope %s%s instead\n",
					s, stackFlagHint(st))
				scope = s
				stacks = st
			}

			interactive := !noInteractive && isTTYFunc()
			if interactive {
				skip := promptSkip{
					Scope:         scopeSet || presetSet,
					Stack:         stackSet || presetSet,
					WithMemory:    cmd.Flags().Changed("with-memory"),
					WithUI:        cmd.Flags().Changed("with-ui"),
					WithAPI:       cmd.Flags().Changed("with-api"),
					WithTDD:       cmd.Flags().Changed("with-tdd"),
					WithChangelog: cmd.Flags().Changed("with-changelog"),
					WithMonorepo:  cmd.Flags().Changed("monorepo"),
					WithPrompts:   cmd.Flags().Changed("with-prompts"),
					WithModeling:  cmd.Flags().Changed("with-modeling"),
					WithEnv:       cmd.Flags().Changed("with-env"),
					Lang:          cmd.Flags().Changed("lang"),
					AITools:       cmd.Flags().Changed("ai-tools"),
				}
				defStack := ""
				if len(stacks) > 0 {
					defStack = stacks[0]
				}
				result, err := runPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptResult{
					Name:          name,
					Scope:         scope,
					Stack:         defStack,
					WithMemory:    withMemory,
					WithUI:        withUI,
					WithAPI:       withAPI,
					WithTDD:       withTDD,
					WithChangelog: withChangelog,
					WithMonorepo:  withMonorepo,
					WithPrompts:   withPrompts,
					WithModeling:  withModeling,
					WithEnv:       withEnv,
					Lang:          lang,
					AITools:       aiTools,
				}, skip)
				if err != nil {
					return &ExitError{Code: 2, Err: err}
				}
				name = result.Name
				scope = result.Scope
				if !stackSet && !presetSet {
					stacks = stacksFromPromptStack(result.Stack)
				}
				withMemory = result.WithMemory
				withUI = result.WithUI
				withAPI = result.WithAPI
				withTDD = result.WithTDD
				withChangelog = result.WithChangelog
				withMonorepo = result.WithMonorepo
				withPrompts = result.WithPrompts
				withModeling = result.WithModeling
				withEnv = result.WithEnv
				lang = result.Lang
				aiTools = result.AITools
			}

			if name == "" {
				return &ExitError{
					Code: 2,
					Err:  errors.New("project name is required: pass it as the positional arg, --name, or via the interactive prompt"),
				}
			}

			// Resolve the (scope, stacks) axes to the concrete template tree
			// scaffold renders. v0.8.2 ships only the combinations that have
			// a template tree; everything else is an explicit error rather
			// than a half-wired fallback (ADR 0024 Scope boundary).
			tree, normStacks, terr := resolveTemplateTree(scope, stacks)
			if terr != nil {
				return &ExitError{Code: 2, Err: terr}
			}

			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("init: getwd: %w", err)
			}

			opts := scaffold.Options{
				ProjectName:   name,
				Preset:        tree,
				TargetDir:     target,
				Lang:          lang,
				Force:         force,
				DryRun:        dryRun,
				WithMemory:    withMemory,
				WithUI:        withUI,
				WithAPI:       withAPI,
				WithTDD:       withTDD,
				WithChangelog: withChangelog,
				WithMonorepo:  withMonorepo,
				WithPrompts:   withPrompts,
				WithModeling:  withModeling,
				WithEnv:       withEnv,
				Stacks:        normStacks,
				AITools:       aiTools,
				Stdout:        cmd.OutOrStdout(),
			}

			if err := scaffold.Run(opts); err != nil {
				if errors.Is(err, scaffold.ErrProposalExists) {
					return &ExitError{Code: 2, Err: err}
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "standard", "documentation scope (minimal | standard)")
	cmd.Flags().StringSliceVar(&stacks, "stack", nil, "target stack; repeatable or comma-separated (flutter | typescript). Empty = stack-agnostic")
	cmd.Flags().StringVar(&preset, "preset", "", "DEPRECATED alias for --scope/--stack (minimal | standard | flutter | typescript); removed in v1.0")
	cmd.Flags().StringVar(&name, "name", "", "project name; overrides the positional argument when both are given")
	cmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "skip interactive prompts")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files in a non-empty target directory")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the plan without writing files")
	cmd.Flags().StringVar(&lang, "lang", "en", "document language (en | ja)")
	cmd.Flags().BoolVar(&withMemory, "with-memory", false, "include long-term agent memory under docs/memory/ (ADR 0004)")
	cmd.Flags().BoolVar(&withUI, "with-ui", false, "include UI / UX guidelines at UI.md")
	cmd.Flags().BoolVar(&withAPI, "with-api", false, "include API interface spec at API.md")
	cmd.Flags().BoolVar(&withTDD, "with-tdd", false, "include test strategy at docs/testing.md")
	cmd.Flags().BoolVar(&withChangelog, "with-changelog", false, "include release notes at CHANGELOG.md")
	cmd.Flags().BoolVar(&withMonorepo, "monorepo", false, "configure as a monorepo with nested apps/<name>/AGENTS.md files (v0.6+)")
	cmd.Flags().BoolVar(&withPrompts, "with-prompts", false, "include a reusable-prompt library at docs/prompts.md (ADR 0034)")
	cmd.Flags().BoolVar(&withModeling, "with-modeling", false, "include a use-case ledger and domain model at docs/usecases.md + docs/domain.md")
	cmd.Flags().BoolVar(&withEnv, "with-env", false, "include an environment-variable template at .env.example (ADR 0037)")
	cmd.Flags().StringVar(&aiToolsCSV, "ai-tools", "claude", "comma-separated AI tools to enable in .aikata/aikata.yaml (claude | cursor | codex)")

	return cmd
}

// presetToScopeStack maps a deprecated `--preset` value to the
// equivalent (scope, stacks) pair. ADR 0024 fixes this table: the
// stack-flavored presets are `--scope standard` plus a single stack.
func presetToScopeStack(preset string) (scope string, stacks []string, err error) {
	switch preset {
	case "minimal":
		return "minimal", nil, nil
	case "standard":
		return "standard", nil, nil
	case "flutter":
		return "standard", []string{"flutter"}, nil
	case "typescript":
		return "standard", []string{"typescript"}, nil
	default:
		return "", nil, fmt.Errorf("unknown preset %q (expected minimal | standard | flutter | typescript)", preset)
	}
}

// resolveTemplateTree maps the (scope, stacks) axes to the concrete
// template tree under presets/<tree>/ and returns the normalized stack
// list to persist. Only the four combinations that have a template
// tree today are accepted; the rest return an explicit "not yet
// supported" error (ADR 0024 Scope boundary) so a user never gets a
// silently half-wired scaffold.
func resolveTemplateTree(scope string, stacks []string) (tree string, normStacks []string, err error) {
	normStacks = normalizeStacks(stacks)
	switch scope {
	case "minimal":
		if len(normStacks) > 0 {
			return "", nil, fmt.Errorf(
				"scope %q with a stack is not yet supported; use --scope standard (tracked in ROADMAP v0.8.2)", scope)
		}
		return "minimal", nil, nil
	case "standard":
		switch len(normStacks) {
		case 0:
			return "standard", nil, nil
		case 1:
			switch normStacks[0] {
			case "flutter", "typescript":
				return normStacks[0], normStacks, nil
			default:
				return "", nil, fmt.Errorf("unknown stack %q (expected flutter | typescript)", normStacks[0])
			}
		default:
			return "", nil, fmt.Errorf(
				"multiple stacks are not yet supported (got %s); v0.8.2 accepts a single stack with --scope standard",
				strings.Join(normStacks, ", "))
		}
	case "extended":
		return "", nil, errors.New(`scope "extended" is reserved for v1.0 and is not selectable yet`)
	case "":
		return "", nil, errors.New("scope is required (minimal | standard)")
	default:
		return "", nil, fmt.Errorf("unknown scope %q (expected minimal | standard)", scope)
	}
}

// normalizeStacks lower-cases, trims, drops empties, and de-duplicates
// a raw stack list (which may arrive comma-split from cobra) while
// preserving first-seen order.
func normalizeStacks(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// stacksFromPromptStack converts the single-stack prompt answer ("" =
// none) into the slice shape the resolver and Options.Stacks expect.
func stacksFromPromptStack(stack string) []string {
	stack = strings.TrimSpace(stack)
	if stack == "" {
		return nil
	}
	return []string{stack}
}

// stackFlagHint renders the trailing `--stack <name>` fragment for the
// deprecation notice, or "" when the alias carried no stack.
func stackFlagHint(stacks []string) string {
	if len(stacks) == 0 {
		return ""
	}
	sorted := append([]string(nil), stacks...)
	sort.Strings(sorted)
	return " --stack " + strings.Join(sorted, ",")
}
