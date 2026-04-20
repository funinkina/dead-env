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

var (
	newProfileService = defaultProfileService
	gitWarningPrinted bool
)

func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:  "deadenv",
		Usage: "Dead simple and secure way to manage your .env",
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowRootCommandHelp(cmd)
		},
		Commands: []*cli.Command{
			NewProfileCommand(),
			NewEditCommand(),
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
	if cmd != nil && cmd.Writer != nil {
		return cmd.Writer
	}

	return os.Stdout
}

func commandErrWriter(cmd *cli.Command) io.Writer {
	if cmd != nil && cmd.ErrWriter != nil {
		return cmd.ErrWriter
	}

	return os.Stderr
}
