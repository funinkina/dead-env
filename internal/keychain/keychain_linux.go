//go:build linux

package keychain

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/keybase/go-keychain/secretservice"
)

const (
	linuxSecretAttributeApplication = "application"
	linuxSecretAttributeProfile     = "profile"
	linuxSecretAttributeKey         = "key"
	linuxSecretApplicationName      = "deadenv"
)

type linuxStore struct {
	newService func() (*secretservice.SecretService, error)
}

func New() (Store, error) {
	return &linuxStore{newService: secretservice.NewService}, nil
}

func (s *linuxStore) Write(service, account, value string) error {
	profile, err := profileFromService(service)
	if err != nil {
		return err
	}
	if err := validateAccount(account); err != nil {
		return err
	}

	svc, session, err := s.openSession()
	if err != nil {
		return linuxSecretServiceError("opening secret service session", err)
	}
	defer svc.CloseSession(session)

	secret, err := session.NewSecret([]byte(value))
	if err != nil {
		return linuxSecretServiceError("creating secret payload", err)
	}

	_, err = svc.CreateItem(
		secretservice.DefaultCollection,
		secretservice.NewSecretProperties(linuxSecretApplicationName+" "+profile+":"+account, linuxSecretAttributes(profile, account)),
		secret,
		secretservice.ReplaceBehaviorReplace,
	)
	if err != nil {
		return linuxSecretServiceError("writing secret", err)
	}

	return nil
}

func (s *linuxStore) Read(service, account, prompt string) (string, error) {
	_ = prompt // Secret Service does not expose a per-call prompt reason.

	profile, err := profileFromService(service)
	if err != nil {
		return "", err
	}
	if err := validateAccount(account); err != nil {
		return "", err
	}

	svc, session, err := s.openSession()
	if err != nil {
		return "", linuxSecretServiceError("opening secret service session", err)
	}
	defer svc.CloseSession(session)

	items, err := svc.SearchCollection(secretservice.DefaultCollection, linuxSecretAttributes(profile, account))
	if err != nil {
		return "", linuxSecretServiceError("reading secret", err)
	}
	if len(items) == 0 {
		return "", ErrKeyNotFound
	}

	secret, err := svc.GetSecret(items[0], *session)
	if err != nil {
		return "", linuxSecretServiceError("reading secret", err)
	}

	return string(secret), nil
}

func (s *linuxStore) Delete(service, account string) error {
	profile, err := profileFromService(service)
	if err != nil {
		return err
	}
	if err := validateAccount(account); err != nil {
		return err
	}

	svc, err := s.newService()
	if err != nil {
		return linuxSecretServiceError("opening secret service connection", err)
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, linuxSecretAttributes(profile, account))
	if err != nil {
		return linuxSecretServiceError("deleting secret", err)
	}

	for _, item := range items {
		if err := svc.DeleteItem(item); err != nil {
			return linuxSecretServiceError("deleting secret", err)
		}
	}

	return nil
}

func (s *linuxStore) List(service string) ([]string, error) {
	profile, err := profileFromService(service)
	if err != nil {
		return nil, err
	}

	svc, err := s.newService()
	if err != nil {
		return nil, linuxSecretServiceError("opening secret service connection", err)
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, linuxSecretProfileAttributes(profile))
	if err != nil {
		return nil, linuxSecretServiceError("listing secrets", err)
	}

	keys := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		attributes, attrErr := svc.GetAttributes(item)
		if attrErr != nil {
			return nil, linuxSecretServiceError("listing secrets", attrErr)
		}

		key := strings.TrimSpace(attributes[linuxSecretAttributeKey])
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys, nil
}

func (s *linuxStore) ListProfiles() ([]string, error) {
	svc, err := s.newService()
	if err != nil {
		return nil, linuxSecretServiceError("opening secret service connection", err)
	}

	items, err := svc.SearchCollection(secretservice.DefaultCollection, secretservice.Attributes{
		linuxSecretAttributeApplication: linuxSecretApplicationName,
	})
	if err != nil {
		return nil, linuxSecretServiceError("listing profiles", err)
	}

	profiles := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		attributes, attrErr := svc.GetAttributes(item)
		if attrErr != nil {
			return nil, linuxSecretServiceError("listing profiles", attrErr)
		}

		profile := strings.TrimSpace(attributes[linuxSecretAttributeProfile])
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

func (s *linuxStore) openSession() (*secretservice.SecretService, *secretservice.Session, error) {
	svc, err := s.newService()
	if err != nil {
		return nil, nil, err
	}

	session, err := svc.OpenSession(secretservice.AuthenticationDHAES)
	if err == nil {
		return svc, session, nil
	}

	session, fallbackErr := svc.OpenSession(secretservice.AuthenticationInsecurePlain)
	if fallbackErr != nil {
		return nil, nil, fmt.Errorf("opening session (dh failed: %v): %w", err, fallbackErr)
	}

	return svc, session, nil
}

func linuxSecretAttributes(profile, account string) secretservice.Attributes {
	return secretservice.Attributes{
		linuxSecretAttributeApplication: linuxSecretApplicationName,
		linuxSecretAttributeProfile:     profile,
		linuxSecretAttributeKey:         account,
	}
}

func linuxSecretProfileAttributes(profile string) secretservice.Attributes {
	return secretservice.Attributes{
		linuxSecretAttributeApplication: linuxSecretApplicationName,
		linuxSecretAttributeProfile:     profile,
	}
}

func linuxSecretServiceError(action string, err error) error {
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[secretservice.PromptDismissedError](err); ok {
		return fmt.Errorf("%w: %s", ErrAuthDenied, action)
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "dismissed"),
		strings.Contains(msg, "canceled"),
		strings.Contains(msg, "cancelled"),
		strings.Contains(msg, "denied"),
		strings.Contains(msg, "interaction not allowed"):
		return fmt.Errorf("%w: %s", ErrAuthDenied, action)
	case strings.Contains(msg, "failed to open dbus connection"):
		return fmt.Errorf("%w: %s: %v", ErrNotImplemented, action, err)
	default:
		return fmt.Errorf("%s: %w", action, err)
	}
}
