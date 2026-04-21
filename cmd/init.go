package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/profile"

	"github.com/urfave/cli/v3"
)

var globalConfigDir string
var globalNoHistory bool
var globalQuiet bool
var globalYes bool

var (
	newProfileService = defaultProfileService
	gitWarningPrinted bool
)

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "deadenv",
		Usage: "Dead simple and secure way to manage your .env",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Override config directory",
				Destination: &globalConfigDir,
			},
			&cli.BoolFlag{
				Name:        "no-history",
				Usage:       "Skip git history commit for this operation",
				Destination: &globalNoHistory,
			},
			&cli.BoolFlag{
				Name:        "quiet",
				Usage:       "Suppress informational output",
				Destination: &globalQuiet,
			},
			&cli.BoolFlag{
				Name:        "yes",
				Aliases:     []string{"y"},
				Usage:       "Skip confirmation prompts",
				Destination: &globalYes,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowRootCommandHelp(cmd)
		},
		Commands: []*cli.Command{
			NewProfileCommand(),
			NewSetCommand(),
			NewGetCommand(),
			NewUnsetCommand(),
			NewEditCommand(),
			NewExportCommand(),
			NewImportCommand(),
			NewRunCommand(),
			NewHistoryCommand(),
			NewInitCommand(),
		},
	}
}

func defaultProfileService() (*profile.ProfileService, error) {
	store, err := keychain.New()
	if err != nil {
		return nil, fmt.Errorf("initializing keychain backend: %w", err)
	}

	var recorder history.Recorder
	gitRecorder, err := history.NewGitRecorder("")
	if err != nil {
		if errors.Is(err, history.ErrGitNotFound) {
			recorder = history.NewNoopRecorder()
			if !gitWarningPrinted {
				_, _ = fmt.Fprintln(os.Stderr, "git not found — history tracking disabled. Install git to enable it.")
				gitWarningPrinted = true
			}
		} else {
			return nil, fmt.Errorf("initializing history recorder: %w", err)
		}
	} else {
		recorder = gitRecorder
	}

	service, err := profile.NewProfileService(store, recorder, nil)
	if err != nil {
		return nil, fmt.Errorf("initializing profile service: %w", err)
	}

	return service, nil
}

func commandWriter(cmd *cli.Command) io.Writer {
	if cmd != nil {
		if cmd.Writer != nil {
			return cmd.Writer
		}

		if root := cmd.Root(); root != nil && root.Writer != nil {
			return root.Writer
		}
	}

	return os.Stdout
}

func commandErrWriter(cmd *cli.Command) io.Writer {
	if cmd != nil {
		if cmd.ErrWriter != nil {
			return cmd.ErrWriter
		}

		if root := cmd.Root(); root != nil && root.ErrWriter != nil {
			return root.ErrWriter
		}
	}

	return os.Stderr
}
