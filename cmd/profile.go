package cmd

import (
	"context"
	"fmt"
	"strings"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/profile"
	"funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var (
	loadPairsFromFile   = profile.FromFile
	loadPairsFromEditor = profile.FromEditor
	promptConfirm       = tui.PromptConfirm
	printPairSummary    = tui.PrintPairSummary
)

func NewProfileCommand() *cli.Command {
	return &cli.Command{
		Name:  "profile",
		Usage: "Manage profiles",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowSubcommandHelp(cmd)
		},
		Commands: []*cli.Command{
			newProfileNewCommand(),
		},
	}
}

func newProfileNewCommand() *cli.Command {
	return &cli.Command{
		Name:      "new",
		Usage:     "Create a profile",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "from",
				Usage: "Path to an env-format file used to create the profile",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			profileName := strings.TrimSpace(cmd.Args().First())
			if profileName == "" {
				return fmt.Errorf("profile name is required")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			fromPath := strings.TrimSpace(cmd.String("from"))

			var (
				pairs    []envPair.EnvPair
				pairsErr error
			)

			if fromPath != "" {
				pairs, pairsErr = loadPairsFromFile(fromPath)
			} else {
				pairs, pairsErr = loadPairsFromEditor(profile.EditorTemplate(profileName))
			}

			if pairsErr != nil {
				return pairsErr
			}

			if len(pairs) == 0 {
				return profile.ErrEmptyContent
			}

			if err := printPairSummary(commandWriter(cmd), pairs); err != nil {
				return fmt.Errorf("printing key summary: %w", err)
			}

			ok, err := promptConfirm("Create profile with these keys?")
			if err != nil {
				return fmt.Errorf("confirming profile creation: %w", err)
			}
			if !ok {
				return profile.ErrCancelled
			}

			if err := service.Create(profileName, pairs); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q created with %d keys.\n", profileName, len(pairs))
			return nil
		},
	}
}
