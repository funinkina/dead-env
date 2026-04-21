package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/profile"
	"github.com/funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var (
	loadPairsFromFile   = profile.FromFile
	loadPairsFromEditor = profile.FromEditor
	listProfiles        = func(service *profile.ProfileService) ([]string, error) { return service.ListProfiles() }
	getProfilePairs     = func(service *profile.ProfileService, profileName string) ([]envPair.EnvPair, error) {
		return service.GetPairs(profileName)
	}
	deleteProfile = func(service *profile.ProfileService, profileName string) error {
		return service.Delete(profileName)
	}
	renameProfile = func(service *profile.ProfileService, oldName, newName string) error {
		return service.Rename(oldName, newName)
	}
	copyProfile = func(service *profile.ProfileService, srcName, dstName string) error {
		return service.Copy(srcName, dstName)
	}
	promptConfirm    = tui.PromptConfirm
	printPairSummary = tui.PrintPairSummary
)

func NewProfileCommand() *cli.Command {
	return &cli.Command{
		Name:  "profile",
		Usage: "Manage profiles",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowSubcommandHelp(cmd)
		},
		Commands: []*cli.Command{
			newProfileListCommand(),
			newProfileNewCommand(),
			newProfileShowCommand(),
			newProfileDeleteCommand(),
			newProfileRenameCommand(),
			newProfileCopyCommand(),
		},
	}
}

func newProfileShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Show profile keys with masked values by default",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "reveal",
				Usage: "Reveal plaintext values",
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

			pairs, err := getProfilePairs(service, profileName)
			if err != nil {
				return fmt.Errorf("showing profile %q: %w", profileName, err)
			}

			if len(pairs) == 0 {
				_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q has no keys.\n", profileName)
				return nil
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile: %s\n", profileName)
			_, _ = fmt.Fprintln(commandWriter(cmd), strings.Repeat("─", 40))

			reveal := cmd.Bool("reveal")
			for _, pair := range pairs {
				value := profileDisplayValue(pair.Value, reveal)
				_, _ = fmt.Fprintf(commandWriter(cmd), "  %s  %s\n", pair.Key, value)
			}

			return nil
		},
	}
}

func profileDisplayValue(value string, reveal bool) string {
	if reveal {
		return value
	}

	return tui.MaskValue(value)
}

func newProfileDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a profile and all its keys",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Delete profile without confirmation",
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

			if !cmd.Bool("yes") {
				ok, err := promptConfirm(fmt.Sprintf("Delete profile %q and all keys?", profileName))
				if err != nil {
					return fmt.Errorf("confirming profile deletion: %w", err)
				}

				if !ok {
					_, _ = fmt.Fprintln(commandWriter(cmd), "Deletion cancelled.")
					return nil
				}
			}

			if err := deleteProfile(service, profileName); err != nil {
				return fmt.Errorf("deleting profile %q: %w", profileName, err)
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q deleted.\n", profileName)
			return nil
		},
	}
}

func newProfileRenameCommand() *cli.Command {
	return &cli.Command{
		Name:      "rename",
		Usage:     "Rename a profile",
		ArgsUsage: "<old> <new>",
		Action: func(_ context.Context, cmd *cli.Command) error {
			oldName := strings.TrimSpace(cmd.Args().Get(0))
			newName := strings.TrimSpace(cmd.Args().Get(1))

			if oldName == "" || newName == "" {
				return fmt.Errorf("both old and new profile names are required")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			if err := renameProfile(service, oldName, newName); err != nil {
				return fmt.Errorf("renaming profile %q to %q: %w", oldName, newName, err)
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q renamed to %q.\n", oldName, newName)
			return nil
		},
	}
}

func newProfileCopyCommand() *cli.Command {
	return &cli.Command{
		Name:      "copy",
		Usage:     "Copy a profile",
		ArgsUsage: "<src> <dst>",
		Action: func(_ context.Context, cmd *cli.Command) error {
			srcName := strings.TrimSpace(cmd.Args().Get(0))
			dstName := strings.TrimSpace(cmd.Args().Get(1))

			if srcName == "" || dstName == "" {
				return fmt.Errorf("both source and destination profile names are required")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			if err := copyProfile(service, srcName, dstName); err != nil {
				return fmt.Errorf("copying profile %q to %q: %w", srcName, dstName, err)
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q copied to %q.\n", srcName, dstName)
			return nil
		},
	}
}

func newProfileListCommand() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List profiles",
		Action: func(_ context.Context, cmd *cli.Command) error {
			service, err := newProfileService()
			if err != nil {
				return err
			}

			profiles, err := listProfiles(service)
			if err != nil {
				return fmt.Errorf("listing profiles: %w", err)
			}

			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(commandWriter(cmd), "No profiles found.")
				return nil
			}

			_, _ = fmt.Fprintln(commandWriter(cmd), "Profiles:")
			for _, profileName := range profiles {
				_, _ = fmt.Fprintf(commandWriter(cmd), "  • %s\n", profileName)
			}

			return nil
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
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Create profile without confirmation",
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
				if errors.Is(pairsErr, profile.ErrEditorFailed) {
					return fmt.Errorf("failed to open editor; set $DEADENV_EDITOR, $VISUAL, or $EDITOR (e.g. nano): %w", pairsErr)
				}

				return pairsErr
			}

			if len(pairs) == 0 {
				if fromPath == "" {
					_, _ = fmt.Fprintln(commandWriter(cmd), "Creation cancelled: no variables found.")
					return nil
				}

				return fmt.Errorf("%w: no variables found in %q", profile.ErrEmptyContent, fromPath)
			}

			if err := printPairSummary(commandWriter(cmd), pairs); err != nil {
				return fmt.Errorf("printing key summary: %w", err)
			}

			if !cmd.Bool("yes") {
				ok, err := promptConfirm("Create profile with these keys?")
				if err != nil {
					return fmt.Errorf("confirming profile creation: %w", err)
				}
				if !ok {
					_, _ = fmt.Fprintln(commandWriter(cmd), "Creation cancelled.")
					return nil
				}
			}

			if err := service.Create(profileName, pairs); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Profile %q created with %d keys.\n", profileName, len(pairs))
			return nil
		},
	}
}
