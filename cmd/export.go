package cmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"funinkina/deadenv/internal/crypto"
	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/exportfmt"
	"funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var (
	renderExportOutput     = exportfmt.Render
	exportEncryptedProfile = crypto.Export
	promptExportPassword   = tui.PromptHidden
)

func NewExportCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "Export profile values to stdout or an encrypted .deadenv file",
		ArgsUsage: "<profile>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "out",
				Usage: "Write encrypted profile export to file",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format for stdout export: shell|fish|json",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			profileName := strings.TrimSpace(cmd.Args().First())
			if profileName == "" {
				return fmt.Errorf("profile name is required")
			}

			outPath := strings.TrimSpace(cmd.String("out"))
			format := strings.TrimSpace(cmd.String("format"))

			if outPath != "" && format != "" {
				return fmt.Errorf("--out and --format cannot be used together")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			profiles, listErr := service.ListProfiles()
			if listErr == nil {
				found := slices.Contains(profiles, profileName)
				if !found {
					return fmt.Errorf("profile %q not found. Create it with: deadenv profile new %s", profileName, profileName)
				}
			}

			keys, listKeysErr := service.ListKeys(profileName)
			if listKeysErr != nil {
				return fmt.Errorf("exporting profile %q: %w", profileName, listKeysErr)
			}

			if len(keys) == 0 {
				return fmt.Errorf("profile %q has no keys", profileName)
			}

			pairs := make([]envPair.EnvPair, 0, len(keys))
			for _, key := range keys {
				val, getErr := service.GetKey(profileName, key)
				if getErr != nil {
					return fmt.Errorf("reading key %q from profile %q: %w", key, profileName, getErr)
				}
				pairs = append(pairs, envPair.EnvPair{Key: key, Value: val})
			}

			if outPath != "" {
				if !strings.HasSuffix(outPath, ".deadenv") {
					outPath = outPath + ".deadenv"
				}

				if err := exportProfileToFile(profileName, pairs, outPath); err != nil {
					return err
				}

				_, _ = fmt.Fprintf(commandWriter(cmd), "Exported profile %q to %q.\n", profileName, outPath)
				return nil
			}

			if format == "" {
				format = exportfmt.FormatShell
			}

			rendered, err := renderExportOutput(pairs, format)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprint(commandWriter(cmd), rendered)
			return nil
		},
	}
}

func exportProfileToFile(profileName string, pairs []envPair.EnvPair, outPath string) error {
	password, err := promptExportPassword("Sharing password")
	if err != nil {
		return fmt.Errorf("reading sharing password: %w", err)
	}

	confirm, err := promptExportPassword("Confirm sharing password")
	if err != nil {
		return fmt.Errorf("reading sharing password confirmation: %w", err)
	}

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	if err := exportEncryptedProfile(profileName, pairs, password, outPath); err != nil {
		return fmt.Errorf("exporting encrypted profile: %w", err)
	}

	return nil
}
