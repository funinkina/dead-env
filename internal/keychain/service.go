package keychain

import (
	"fmt"
	"regexp"
	"strings"
)

const ServicePrefix = "deadenv/"

var profileNameRE = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func profileFromService(service string) (string, error) {
	if !strings.HasPrefix(service, ServicePrefix) {
		return "", fmt.Errorf("%w: expected %q prefix (got %q)", ErrInvalidService, ServicePrefix, service)
	}

	profile := strings.TrimPrefix(service, ServicePrefix)
	if !profileNameRE.MatchString(profile) {
		return "", fmt.Errorf("%w: invalid profile name in service %q", ErrInvalidService, service)
	}

	return profile, nil
}

func targetName(service, account string) (string, error) {
	profile, err := profileFromService(service)
	if err != nil {
		return "", err
	}

	if err := validateAccount(account); err != nil {
		return "", err
	}

	return ServicePrefix + profile + "/" + account, nil
}

func validateAccount(account string) error {
	if strings.TrimSpace(account) == "" {
		return ErrInvalidAccount
	}

	return nil
}
