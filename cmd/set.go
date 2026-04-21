package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var promptSetValue = tui.PromptHidden

func NewSetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set a key in a profile",
		ArgsUsage: "<profile> <KEY> [VALUE]",
		Action: func(_ context.Context, cmd *cli.Command) error {
			profileName := strings.TrimSpace(cmd.Args().First())
			if profileName == "" {
				return fmt.Errorf("profile name is required")
			}

			keyName := strings.TrimSpace(cmd.Args().Get(1))
			if keyName == "" {
				return fmt.Errorf("key name is required")
			}

			value := strings.TrimSpace(cmd.Args().Get(2))

			if value == "" {
				var err error
				value, err = promptSetValue(fmt.Sprintf("Value for %s", keyName))
				if err != nil {
					return fmt.Errorf("reading value: %w", err)
				}
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			if err := service.SetKey(profileName, keyName, value); err != nil {
				return fmt.Errorf("setting key %q in profile %q: %w", keyName, profileName, err)
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Key %q set in profile %q.\n", keyName, profileName)
			return nil
		},
	}
}
