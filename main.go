package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"funinkina/deadenv/cmd"
	"funinkina/deadenv/internal/crypto"
)

func main() {
	root := cmd.NewRootCommand()

	if err := root.Run(context.Background(), os.Args); err != nil {
		if errors.Is(err, crypto.ErrDecryptFailed) {
			_, _ = fmt.Fprintln(os.Stderr, "decryption failed — wrong password or file is corrupted")
			os.Exit(3)
		}

		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
