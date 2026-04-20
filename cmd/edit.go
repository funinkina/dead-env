package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
				_, _ = fmt.Fprintln(commandWriter(cmd), "No changes detected — profile unchanged.")
				return nil
			}

			if errors.Is(err, profile.ErrEmptyContent) {
				_, _ = fmt.Fprintln(commandWriter(cmd), "Edit cancelled: no variables found. Profile unchanged.")
				return nil
			}

			if errors.Is(err, profile.ErrCancelled) {
				_, _ = fmt.Fprintln(commandWriter(cmd), "Edit cancelled. Profile unchanged.")
				return nil
			}

			var partialErr *profile.PartialApplyError
			if errors.As(err, &partialErr) {
				return formatPartialApplyError(partialErr)
			}

			if errors.Is(err, profile.ErrEditorFailed) {
				return fmt.Errorf("failed to open editor; check $VISUAL or $EDITOR: %w", err)
			}

			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q updated.\n", profileName)
			return nil
		},
	}
}

func formatPartialApplyError(err *profile.PartialApplyError) error {
	if err == nil {
		return profile.ErrApplyChanges
	}

	succeeded := append([]string(nil), err.Succeeded...)
	sort.Strings(succeeded)

	failed := err.FailedKeys()

	var b strings.Builder
	b.WriteString("partial update failed")

	if len(succeeded) > 0 {
		b.WriteString("; succeeded keys: ")
		b.WriteString(strings.Join(succeeded, ", "))
	}

	if len(failed) > 0 {
		b.WriteString("; failed keys: ")
		for i, key := range failed {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(key)
			if keyErr := err.Failed[key]; keyErr != nil {
				b.WriteString(" (")
				b.WriteString(keyErr.Error())
				b.WriteString(")")
			}
		}
	}

	return fmt.Errorf("%w: %s", profile.ErrApplyChanges, b.String())
}
