package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"funinkina/deadenv/internal/profile"

	"github.com/urfave/cli/v3"
)

var runEdit = func(service *profile.ProfileService, profileName string, opts profile.EditOptions) error {
	return service.Edit(profileName, opts)
}

func NewEditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a profile in your preferred editor",
		ArgsUsage: "<profile>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Apply changes without confirmation",
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

			err = runEdit(service, profileName, profile.EditOptions{
				Yes: cmd.Bool("yes"),
				Out: commandWriter(cmd),
			})

			if errors.Is(err, profile.ErrNoChanges) {
				_, _ = fmt.Fprintln(commandWriter(cmd), "No changes detected.")
				return nil
			}

			if errors.Is(err, profile.ErrEmptyContent) {
				_, _ = fmt.Fprintln(commandWriter(cmd), "Edit cancelled: no variables found.")
				return nil
			}

			if errors.Is(err, profile.ErrCancelled) {
				_, _ = fmt.Fprintln(commandWriter(cmd), "Edit cancelled.")
				return nil
			}

			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(commandWriter(cmd), "Profile updated.")
			return nil
		},
	}
}
