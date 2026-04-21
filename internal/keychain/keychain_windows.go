//go:build windows

package keychain

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/danieljoos/wincred"
)

const windowsCredentialBlobLimit = 2560

type windowsStore struct{}

func New() (Store, error) {
	return &windowsStore{}, nil
}

func (s *windowsStore) Write(service, account, value string) error {
	target, err := targetName(service, account)
	if err != nil {
		return err
	}

	blob := []byte(value)
	if len(blob) > windowsCredentialBlobLimit {
		return fmt.Errorf("writing credential %q: credential blob exceeds windows limit of %d bytes (got %d)", target, windowsCredentialBlobLimit, len(blob))
	}

	cred := wincred.NewGenericCredential(target)
	cred.Persist = wincred.PersistLocalMachine
	cred.CredentialBlob = blob

	if err := cred.Write(); err != nil {
		return fmt.Errorf("writing credential %q: %w", target, err)
	}

	return nil
}

func (s *windowsStore) Read(service, account, prompt string) (string, error) {
	_ = prompt

	target, err := targetName(service, account)
	if err != nil {
		return "", err
	}

	cred, err := wincred.GetGenericCredential(target)
	if err != nil {
		if errors.Is(err, wincred.ErrElementNotFound) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("reading credential %q: %w", target, err)
	}

	return string(cred.CredentialBlob), nil
}

func (s *windowsStore) Delete(service, account string) error {
	target, err := targetName(service, account)
	if err != nil {
		return err
	}

	cred, err := wincred.GetGenericCredential(target)
	if err != nil {
		if errors.Is(err, wincred.ErrElementNotFound) {
			return nil
		}
		return fmt.Errorf("loading credential %q for delete: %w", target, err)
	}

	if err := cred.Delete(); err != nil {
		if errors.Is(err, wincred.ErrElementNotFound) {
			return nil
		}
		return fmt.Errorf("deleting credential %q: %w", target, err)
	}

	return nil
}

func (s *windowsStore) List(service string) ([]string, error) {
	profile, err := profileFromService(service)
	if err != nil {
		return nil, err
	}

	creds, err := wincred.List()
	if err != nil {
		return nil, fmt.Errorf("listing windows credentials: %w", err)
	}

	prefix := servicePrefix + profile + "/"
	keys := make([]string, 0)
	for _, cred := range creds {
		if !strings.HasPrefix(cred.TargetName, prefix) {
			continue
		}

		key := strings.TrimPrefix(cred.TargetName, prefix)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys, nil
}

func (s *windowsStore) ListProfiles() ([]string, error) {
	creds, err := wincred.List()
	if err != nil {
		return nil, fmt.Errorf("listing windows credentials: %w", err)
	}

	profiles := make([]string, 0)
	seen := make(map[string]struct{})

	for _, cred := range creds {
		if !strings.HasPrefix(cred.TargetName, servicePrefix) {
			continue
		}

		remainder := strings.TrimPrefix(cred.TargetName, servicePrefix)
		profile, _, ok := strings.Cut(remainder, "/")
		if !ok {
			continue
		}

		profile = strings.TrimSpace(profile)
		if profile == "" {
			continue
		}

		if _, exists := seen[profile]; exists {
			continue
		}

		seen[profile] = struct{}{}
		profiles = append(profiles, profile)
	}

	sort.Strings(profiles)
	return profiles, nil
}
