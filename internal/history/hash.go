package history

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	saltFileName = ".salt"
	saltSize     = 32
)

// HashValue computes SHA-256(salt + value) using the default history directory.
// Salt is stored at ~/.config/deadenv/history/.salt and created once per installation.
func HashValue(value string) (string, error) {
	historyDir, err := DefaultHistoryDir()
	if err != nil {
		return "", err
	}
	return HashValueAt(historyDir, value)
}

func HashValueAt(historyDir, value string) (string, error) {
	if strings.TrimSpace(historyDir) == "" {
		return "", ErrInvalidHistoryDir
	}

	salt, err := loadOrCreateSalt(historyDir)
	if err != nil {
		return "", fmt.Errorf("loading history salt: %w", err)
	}

	payload := make([]byte, 0, len(salt)+len(value))
	payload = append(payload, salt...)
	payload = append(payload, value...)

	sum := sha256.Sum256(payload)

	for i := range payload {
		payload[i] = 0
	}

	return hex.EncodeToString(sum[:]), nil
}

func loadOrCreateSalt(historyDir string) ([]byte, error) {

	if err := os.MkdirAll(historyDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating history directory %q: %w", historyDir, err)
	}

	saltPath := filepath.Join(historyDir, saltFileName)

	if b, err := os.ReadFile(saltPath); err == nil {
		if len(b) != saltSize {
			return nil, fmt.Errorf("invalid salt size: got %d, want %d", len(b), saltSize)
		}

		return b, nil

	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading salt file %q: %w", saltPath, err)
	}

	// creating new salt atomically
	salt := make([]byte, saltSize)

	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating random salt: %w", err)
	}

	f, err := os.OpenFile(saltPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			b, readErr := os.ReadFile(saltPath)
			if readErr != nil {
				return nil, fmt.Errorf("reading concurrently-created salt file %q: %w", saltPath, readErr)
			}

			if len(b) != saltSize {
				return nil, fmt.Errorf("invalid concurrent salt size: got %d, want %d", len(b), saltSize)
			}

			return b, nil
		}

		return nil, fmt.Errorf("creating salt file %q: %w", saltPath, err)
	}

	defer f.Close()

	if _, err := f.Write(salt); err != nil {
		return nil, fmt.Errorf("writing salt file %q: %w", saltPath, err)
	}

	return salt, nil
}
