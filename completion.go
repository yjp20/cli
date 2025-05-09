package cli

import (
	"context"
	"embed"
	"fmt"
	"sort"
)

const (
	completionCommandName = "completion"

	// This flag is supposed to only be used by the completion script itself to generate completions on the fly.
	completionFlag = "--generate-shell-completion"
)

type renderCompletion func(cmd *Command) (string, error)

var (
	//go:embed autocomplete
	autoCompleteFS embed.FS

	shellCompletions = map[string]renderCompletion{
		"bash": func(c *Command) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/bash_autocomplete")
			return fmt.Sprintf(string(b), c.Name), err
		},
		"zsh": func(c *Command) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/zsh_autocomplete")
			return fmt.Sprintf(string(b), c.Name), err
		},
		"fish": func(c *Command) (string, error) {
			return c.ToFishCompletion()
		},
		"pwsh": func(c *Command) (string, error) {
			b, err := autoCompleteFS.ReadFile("autocomplete/powershell_autocomplete.ps1")
			return string(b), err
		},
	}
)

func buildCompletionCommand(rootCmd *Command) *Command {
	return &Command{
		Name:   completionCommandName,
		Hidden: true,
		Action: func(ctx context.Context, cmd *Command) error {
			return printShellCompletion(ctx, cmd, rootCmd)
		},
	}
}

func printShellCompletion(_ context.Context, cmd *Command, rootCmd *Command) error {
	var shells []string
	for k := range shellCompletions {
		shells = append(shells, k)
	}

	sort.Strings(shells)

	if cmd.Args().Len() == 0 {
		return Exit(fmt.Sprintf("no shell provided for completion command. available shells are %+v", shells), 1)
	}
	s := cmd.Args().First()

	renderCompletion, ok := shellCompletions[s]
	if !ok {
		return Exit(fmt.Sprintf("unknown shell %s, available shells are %+v", s, shells), 1)
	}

	completionScript, err := renderCompletion(rootCmd)
	if err != nil {
		return Exit(err, 1)
	}

	_, err = cmd.Writer.Write([]byte(completionScript))
	if err != nil {
		return Exit(err, 1)
	}

	return nil
}
