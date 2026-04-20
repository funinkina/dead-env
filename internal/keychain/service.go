package keychain

import (
	"fmt"
	"regexp"
	"strings"
)

const servicePrefix = "deadenv/"

var profileNameRE = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

func profileFromService(service string) (string, error) {
	if !strings.HasPrefix(service, servicePrefix) {
		return "", fmt.Errorf("invalid service %q: expected %q prefix", service, servicePrefix)
	}

	profile := strings.TrimPrefix(service, servicePrefix)
	if !profileNameRE.MatchString(profile) {
		return "", fmt.Errorf("invalid service %q: invalid profile name", service)
	}

	return profile, nil
}

func targetName(service, account string) (string, error) {
	profile, err := profileFromService(service)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(account) == "" {
		return "", fmt.Errorf("account cannot be empty")
	}

	return servicePrefix + profile + "/" + account, nil
}
