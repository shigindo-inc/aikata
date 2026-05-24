package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shigindo-inc/aikata/internal/scaffold"
)

// newInitCmd builds the `aikata init` subcommand. The scaffolding-time
// surface is `--preset`, `--name`, `--lang`, `--ai-tools`, and the
// optional-component flags (`--with-memory`, `--with-ui`, `--with-api`,
// `--with-tdd`, `--with-changelog`). The post-init counterpart lives
// behind `aikata add <component>`.
func newInitCmd() *cobra.Command {
	var (
		preset        string
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
		aiToolsCSV    string
	)

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a new aikata project in the current directory",
		Long: "Generate a coherent set of markdown documents for a new project.\n\n" +
			"In non-interactive mode (--no-interactive, or when stdin is not\n" +
			"a TTY), the project name must be supplied either as the positional\n" +
			"argument or via --name. The target directory is always the current\n" +
			"working directory.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && name == "" {
				name = args[0]
			}

			aiTools, err := parseAITools(aiToolsCSV)
			if err != nil {
				return &ExitError{Code: 2, Err: err}
			}

			interactive := !noInteractive && isTTYFunc()
			if interactive {
				skip := promptSkip{
					Preset:        cmd.Flags().Changed("preset"),
					WithMemory:    cmd.Flags().Changed("with-memory"),
					WithUI:        cmd.Flags().Changed("with-ui"),
					WithAPI:       cmd.Flags().Changed("with-api"),
					WithTDD:       cmd.Flags().Changed("with-tdd"),
					WithChangelog: cmd.Flags().Changed("with-changelog"),
					Lang:          cmd.Flags().Changed("lang"),
					AITools:       cmd.Flags().Changed("ai-tools"),
				}
				result, err := runPrompt(cmd.InOrStdin(), cmd.OutOrStdout(), promptResult{
					Name:          name,
					Preset:        preset,
					WithMemory:    withMemory,
					WithUI:        withUI,
					WithAPI:       withAPI,
					WithTDD:       withTDD,
					WithChangelog: withChangelog,
					Lang:          lang,
					AITools:       aiTools,
				}, skip)
				if err != nil {
					return &ExitError{Code: 2, Err: err}
				}
				name = result.Name
				preset = result.Preset
				withMemory = result.WithMemory
				withUI = result.WithUI
				withAPI = result.WithAPI
				withTDD = result.WithTDD
				withChangelog = result.WithChangelog
				lang = result.Lang
				aiTools = result.AITools
			}

			if name == "" {
				return &ExitError{
					Code: 2,
					Err:  errors.New("project name is required: pass it as the positional arg, --name, or via the interactive prompt"),
				}
			}

			target, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("init: getwd: %w", err)
			}

			opts := scaffold.Options{
				ProjectName:   name,
				Preset:        preset,
				TargetDir:     target,
				Lang:          lang,
				Force:         force,
				DryRun:        dryRun,
				WithMemory:    withMemory,
				WithUI:        withUI,
				WithAPI:       withAPI,
				WithTDD:       withTDD,
				WithChangelog: withChangelog,
				Stacks:        stacksForPreset(preset),
				AITools:       aiTools,
				Stdout:        cmd.OutOrStdout(),
			}

			if err := scaffold.Run(opts); err != nil {
				if errors.Is(err, scaffold.ErrTargetDirNotEmpty) {
					return &ExitError{Code: 2, Err: err}
				}
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&preset, "preset", "standard", "preset name (minimal | standard | flutter | typescript)")
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
	cmd.Flags().StringVar(&aiToolsCSV, "ai-tools", "claude", "comma-separated AI tools to enable in .aikata/aikata.yaml (claude | cursor | codex)")

	return cmd
}

// stacksForPreset returns the stack identifiers implied by a preset
// name. Stack-flavored presets like `flutter` set their own stack
// implicitly; `minimal` / `standard` ship without one.
func stacksForPreset(preset string) []string {
	switch preset {
	case "flutter":
		return []string{"flutter"}
	case "typescript":
		return []string{"typescript"}
	default:
		return nil
	}
}
