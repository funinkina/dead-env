package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/funinkina/deadenv/internal/keychain"
	"github.com/funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var promptUnsetConfirm = tui.PromptConfirm

func NewUnsetCommand() *cli.Command {
	return &cli.Command{
		Name:      "unset",
		Usage:     "Remove a key from a profile",
		ArgsUsage: "<profile> <KEY>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Remove key without confirmation",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			profileName := strings.TrimSpace(cmd.Args().First())
			if profileName == "" {
				return fmt.Errorf("profile name is required")
			}

			keyName := strings.TrimSpace(cmd.Args().Get(1))
			if keyName == "" {
				return fmt.Errorf("key name is required")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			if !cmd.Bool("yes") {
				ok, err := promptUnsetConfirm(fmt.Sprintf("Remove key %q from profile %q?", keyName, profileName))
				if err != nil {
					return fmt.Errorf("confirming key removal: %w", err)
				}
				if !ok {
					_, _ = fmt.Fprintln(commandWriter(cmd), "Removal cancelled.")
					return nil
				}
			}

			if err := service.UnsetKey(profileName, keyName); err != nil {
				if errors.Is(err, keychain.ErrKeyNotFound) {
					return fmt.Errorf("key %q not found in profile %q", keyName, profileName)
				}
				return fmt.Errorf("unsetting key %q from profile %q: %w", keyName, profileName, err)
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Key %q removed from profile %q.\n", keyName, profileName)
			return nil
		},
	}
}
