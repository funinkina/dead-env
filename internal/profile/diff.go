package profile

import (
	"sort"

	"github.com/funinkina/deadenv/internal/envPair"
)

type PairDiff struct {
	Added    []envPair.EnvPair
	Modified []envPair.EnvPair
	Removed  []envPair.EnvPair
}

func DiffPairs(oldPairs, newPairs []envPair.EnvPair) PairDiff {
	oldMap := pairsToMap(oldPairs)
	newMap := pairsToMap(newPairs)

	diff := PairDiff{
		Added:    make([]envPair.EnvPair, 0),
		Modified: make([]envPair.EnvPair, 0),
		Removed:  make([]envPair.EnvPair, 0),
	}

	for _, key := range sortedMapKeys(newMap) {
		newValue := newMap[key]
		oldValue, exists := oldMap[key]
		if !exists {
			diff.Added = append(diff.Added, envPair.EnvPair{Key: key, Value: newValue})
			continue
		}

		if oldValue != newValue {
			diff.Modified = append(diff.Modified, envPair.EnvPair{Key: key, Value: newValue})
		}
	}

	for _, key := range sortedMapKeys(oldMap) {
		if _, exists := newMap[key]; exists {
			continue
		}

		diff.Removed = append(diff.Removed, envPair.EnvPair{Key: key, Value: oldMap[key]})
	}

	return diff
}

func pairsToMap(pairs []envPair.EnvPair) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		m[pair.Key] = pair.Value
	}

	return m
}

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
