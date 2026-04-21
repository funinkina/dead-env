package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"funinkina/deadenv/cmd"
	"funinkina/deadenv/internal/crypto"
	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/profile"
)

func main() {
	root := cmd.NewRootCommand()

	if err := root.Run(context.Background(), os.Args); err != nil {
		if errors.Is(err, crypto.ErrDecryptFailed) {
			_, _ = fmt.Fprintln(os.Stderr, "decryption failed — wrong password or file is corrupted")
			os.Exit(3)
		}

		if errors.Is(err, keychain.ErrAuthDenied) {
			_, _ = fmt.Fprintln(os.Stderr, "authentication denied")
			os.Exit(2)
		}

		if errors.Is(err, keychain.ErrKeyNotFound) {
			msg := formatKeyNotFoundError(err)
			_, _ = fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrProfileNameEmpty) {
			_, _ = fmt.Fprintln(os.Stderr, "profile name is required. Create one with: deadenv profile new <name>")
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrKeyEmpty) {
			_, _ = fmt.Fprintln(os.Stderr, "key name is required")
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrEmptyContent) {
			_, _ = fmt.Fprintln(os.Stderr, "no variables found")
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrNoChanges) {
			_, _ = fmt.Fprintln(os.Stderr, "no changes detected — profile unchanged")
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrEditorFailed) {
			_, _ = fmt.Fprintln(os.Stderr, "failed to open editor; set $DEADENV_EDITOR, $VISUAL, or $EDITOR (e.g. nano)")
			os.Exit(1)
		}

		if errors.Is(err, profile.ErrCancelled) {
			_, _ = fmt.Fprintln(os.Stderr, "operation cancelled")
			os.Exit(1)
		}

		type exitCoder interface {
			ExitCode() int
		}

		var ec exitCoder
		if errors.As(err, &ec) {
			os.Exit(ec.ExitCode())
		}

		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func formatKeyNotFoundError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "not found") {
		parts := strings.Split(msg, "\"")
		if len(parts) >= 2 {
			key := parts[1]
			rest := strings.Join(parts[2:], "\"")
			return fmt.Sprintf("key %q not found in profile %s. Set it with: deadenv set", key, rest)
		}
	}
	return msg
}
