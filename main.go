package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"funinkina/deadenv/cmd"
	"funinkina/deadenv/internal/crypto"
	"funinkina/deadenv/internal/keychain"
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
