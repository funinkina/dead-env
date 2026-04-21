package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	shellBash = "bash"
	shellZsh  = "zsh"
	shellFish = "fish"
)

func NewInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Print shell hook snippet for deadenv use <profile>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "shell",
				Usage: "Shell type: bash|zsh|fish",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			shellName, err := resolveShellName(cmd.String("shell"), os.Getenv("SHELL"))
			if err != nil {
				return err
			}

			snippet, err := shellHookSnippet(shellName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprint(commandWriter(cmd), snippet)
			return nil
		},
	}
}

func resolveShellName(flagValue, envShell string) (string, error) {
	if shellName := strings.ToLower(strings.TrimSpace(flagValue)); shellName != "" {
		return validateShellName(shellName)
	}

	auto := strings.ToLower(strings.TrimSpace(filepath.Base(envShell)))
	if auto == "" {
		return "", fmt.Errorf("shell is required (use --shell=bash|zsh|fish)")
	}

	return validateShellName(auto)
}

func validateShellName(shellName string) (string, error) {
	switch shellName {
	case shellBash, shellZsh, shellFish:
		return shellName, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (expected bash, zsh, or fish)", shellName)
	}
}

func shellHookSnippet(shellName string) (string, error) {
	switch shellName {
	case shellBash, shellZsh:
		return fmt.Sprintf(`# deadenv shell hook — paste into ~/.%src
deadenv() {
  if [ "$1" = "use" ]; then
    eval "$(command deadenv export "$2")"
  else
    command deadenv "$@"
  fi
}
`, shellName), nil
	case shellFish:
		return `# deadenv shell hook — paste into ~/.config/fish/config.fish
function deadenv
  if test "$argv[1]" = "use"
    command deadenv export "$argv[2]" --format=fish | source
  else
    command deadenv $argv
  end
end
`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shellName)
	}
}
