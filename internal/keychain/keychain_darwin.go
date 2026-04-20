//go:build darwin

package keychain

import (
	"fmt"
	"sort"

	"funinkina/deadenv/internal/keychain/gokeychain"
)

type darwinStore struct{}

func New() (Store, error) {
	return &darwinStore{}, nil
}

func (s *darwinStore) Write(service, account, value string) error {
	if _, err := targetName(service, account); err != nil {
		return err
	}

	if err := gokeychain.Write(service, account, value); err != nil {
		return darwinKeychainError("writing secret", err)
	}

	return nil
}

func (s *darwinStore) Read(service, account, prompt string) (string, error) {
	_ = prompt // go-keychain API does not expose a per-call localized reason.

	if _, err := targetName(service, account); err != nil {
		return "", err
	}

	secret, err := gokeychain.Read(service, account)
	if err != nil {
		return "", darwinKeychainError("reading secret", err)
	}
	if secret == nil {
		return "", ErrKeyNotFound
	}

	return string(secret), nil
}

func (s *darwinStore) Delete(service, account string) error {
	if _, err := targetName(service, account); err != nil {
		return err
	}

	err := gokeychain.Delete(service, account)
	if err == nil || gokeychain.IsItemNotFound(err) {
		return nil
	}

	return darwinKeychainError("deleting secret", err)
}

func (s *darwinStore) List(service string) ([]string, error) {
	if _, err := profileFromService(service); err != nil {
		return nil, err
	}

	accounts, err := gokeychain.List(service)
	if err != nil {
		return nil, darwinKeychainError("listing keys", err)
	}

	sort.Strings(accounts)
	return accounts, nil
}

func darwinKeychainError(action string, err error) error {
	switch {
	case gokeychain.IsItemNotFound(err):
		return ErrKeyNotFound
	case gokeychain.IsAuthDenied(err):
		return ErrAuthDenied
	}

	return fmt.Errorf("%s: %w", action, err)
}
