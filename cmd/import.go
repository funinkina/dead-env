package cmd

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/funinkina/deadenv/internal/crypto"
	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/tui"

	"github.com/urfave/cli/v3"
)

var (
	importEncryptedProfile = crypto.Import
	promptImportPassword   = tui.PromptHidden
	promptImportConfirm    = tui.PromptConfirm
	printImportSummary     = tui.PrintPairSummary
)

func NewImportCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Import an encrypted .deadenv profile file",
		ArgsUsage: "<file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "as",
				Usage: "Override destination profile name",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Import without confirmation",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			filePath := strings.TrimSpace(cmd.Args().First())
			if filePath == "" {
				return fmt.Errorf("import file path is required")
			}

			service, err := newProfileService()
			if err != nil {
				return err
			}

			password, err := promptImportPassword("Sharing password")
			if err != nil {
				return fmt.Errorf("reading sharing password: %w", err)
			}

			pairs, sourceProfile, err := importEncryptedProfile(filePath, password)
			if err != nil {
				return err
			}

			targetProfile := strings.TrimSpace(cmd.String("as"))
			if targetProfile == "" {
				targetProfile = sourceProfile
			}
			if targetProfile == "" {
				return fmt.Errorf("profile name missing in import file; provide one with --as")
			}

			existingProfiles, listErr := service.ListProfiles()
			profileExists := false
			if listErr == nil {
				if slices.Contains(existingProfiles, targetProfile) {
					profileExists = true
				}
			}

			if err := printImportSummary(commandWriter(cmd), dedupePairsByKey(pairs)); err != nil {
				return fmt.Errorf("printing key summary: %w", err)
			}

			confirmMsg := "Import profile with these keys?"
			if profileExists {
				confirmMsg = fmt.Sprintf("Warning: profile %q already exists. Import will overwrite existing keys. %s", targetProfile, confirmMsg)
			}

			if !cmd.Bool("yes") {
				ok, err := promptImportConfirm(confirmMsg)
				if err != nil {
					return fmt.Errorf("confirming profile import: %w", err)
				}
				if !ok {
					_, _ = fmt.Fprintln(commandWriter(cmd), "Import cancelled.")
					return nil
				}
			}

			if err := service.ReplaceProfile(targetProfile, pairs); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(commandWriter(cmd), "Imported %d keys into profile %q.\n", len(dedupePairsByKey(pairs)), targetProfile)
			return nil
		},
	}
}

func dedupePairsByKey(pairs []envPair.EnvPair) []envPair.EnvPair {
	byKey := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		if pair.Key == "" {
			continue
		}
		byKey[pair.Key] = pair.Value
	}

	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]envPair.EnvPair, 0, len(keys))
	for _, key := range keys {
		result = append(result, envPair.EnvPair{Key: key, Value: byKey[key]})
	}

	return result
}
