package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func NewHistoryCommand() *cli.Command {
	return &cli.Command{
		Name:      "history",
		Usage:     "Show profile change history",
		ArgsUsage: "<profile>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "key",
				Usage: "Filter history to a specific key",
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

			keyFilter := strings.TrimSpace(cmd.String("key"))

			entries, err := service.GetHistory(profileName, keyFilter)
			if err != nil {
				return fmt.Errorf("fetching history for profile %q: %w", profileName, err)
			}

			if len(entries) == 0 {
				_, _ = fmt.Fprintf(commandWriter(cmd), "No history found for profile %q.\n", profileName)
				return nil
			}

			var header string
			if keyFilter != "" {
				header = fmt.Sprintf("Profile: %s (key: %s)\n", profileName, keyFilter)
			} else {
				header = fmt.Sprintf("Profile: %s\n", profileName)
			}
			_, _ = fmt.Fprintln(commandWriter(cmd), header)
			_, _ = fmt.Fprintln(commandWriter(cmd), strings.Repeat("─", 48))

			for _, entry := range entries {
				timestamp := entry.Timestamp.Format("2006-01-02 15:04")
				shortHash := ""
				if entry.ValueHash != "" && len(entry.ValueHash) >= 4 {
					shortHash = fmt.Sprintf(" (hash: %s...)", entry.ValueHash[:4])
				}
				_, _ = fmt.Fprintf(
					commandWriter(cmd),
					"%s  %-10s %-15s%s\n",
					timestamp,
					entry.Operation,
					entry.Key,
					shortHash,
				)
			}

			return nil
		},
	}
}
