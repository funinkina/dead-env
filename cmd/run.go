package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"funinkina/deadenv/internal/profile"

	"github.com/urfave/cli/v3"
)

var runProfileCommand = func(service *profile.ProfileService, profileName, commandName string, args []string, opts profile.RunOptions) error {
	return service.Run(profileName, commandName, args, opts)
}

func NewRunCommand() *cli.Command {
	return &cli.Command{
		Name:            "run",
		Usage:           "Run a command with profile variables injected",
		ArgsUsage:       "<profile> -- <command> [args...]",
		SkipFlagParsing: true,
		Action: func(_ context.Context, cmd *cli.Command) error {
			profileName, commandName, commandArgs, err := parseRunArgs(cmd.Args().Slice())
			if err != nil {
				return err
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			return runProfileCommand(service, profileName, commandName, commandArgs, profile.RunOptions{
				Stdin:  os.Stdin,
				Stdout: commandWriter(cmd),
				Stderr: commandErrWriter(cmd),
			})
		},
	}
}

func parseRunArgs(args []string) (string, string, []string, error) {
	if len(args) == 0 {
		return "", "", nil, fmt.Errorf("profile name is required")
	}

	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}

	if separator == -1 {
		return "", "", nil, fmt.Errorf("missing -- separator (usage: deadenv run <profile> -- <command> [args...])")
	}

	if separator != 1 {
		return "", "", nil, fmt.Errorf("run command expects exactly one profile argument before --")
	}

	if len(args) <= separator+1 {
		return "", "", nil, fmt.Errorf("command is required after --")
	}

	profileName := strings.TrimSpace(args[0])
	if profileName == "" {
		return "", "", nil, fmt.Errorf("profile name is required")
	}

	commandName := strings.TrimSpace(args[separator+1])
	if commandName == "" {
		return "", "", nil, fmt.Errorf("command is required after --")
	}

	commandArgs := append([]string(nil), args[separator+2:]...)

	return profileName, commandName, commandArgs, nil
}
