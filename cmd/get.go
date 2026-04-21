package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

func NewGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a key value from a profile",
		ArgsUsage: "<profile> <KEY>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "reveal",
				Usage: "Show plaintext value",
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

			value, err := service.GetKey(profileName, keyName)
			if err != nil {
				if errors.Is(err, keychain.ErrKeyNotFound) {
					return fmt.Errorf("key %q not found in profile %q. Set it with: deadenv set %s %s", keyName, profileName, profileName, keyName)
				}
				return fmt.Errorf("getting key %q from profile %q: %w", keyName, profileName, err)
			}

			reveal := cmd.Bool("reveal")
			if reveal {
				_, _ = fmt.Fprintln(commandWriter(cmd), value)
			} else {
				_, _ = fmt.Fprintln(commandWriter(cmd), tui.MaskValue(value))
			}

			return nil
		},
	}
}