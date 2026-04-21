package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/funinkina/deadenv/internal/history"
	"github.com/funinkina/deadenv/internal/keychain"
)

func (p *ProfileService) ListProfiles() ([]string, error) {
	if lister, ok := p.store.(keychain.ProfileLister); ok {
		profiles, err := lister.ListProfiles()
		if err != nil {
			return nil, fmt.Errorf("error listing profiles from keychain: %w", err)
		}

		return uniqueSortedStrings(profiles), nil
	}

	candidates, err := listProfilesFromHistory()
	if err != nil {
		return nil, fmt.Errorf("error listing profiles from history: %w", err)
	}

	profiles := make([]string, 0, len(candidates))
	for _, profileName := range candidates {
		keys, keyErr := p.ListKeys(profileName)
		if keyErr != nil {
			continue
		}
		if len(keys) == 0 {
			continue
		}

		profiles = append(profiles, profileName)
	}

	return uniqueSortedStrings(profiles), nil
}

func listProfilesFromHistory() ([]string, error) {
	historyDir, err := history.DefaultHistoryDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	profiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSpace(entry.Name())
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		profileName := strings.TrimSuffix(name, filepath.Ext(name))
		if strings.TrimSpace(profileName) == "" {
			continue
		}

		profiles = append(profiles, profileName)
	}

	return profiles, nil
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}

	sort.Strings(unique)
	return unique
}
