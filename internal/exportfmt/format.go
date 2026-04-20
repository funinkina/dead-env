package exportfmt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"funinkina/deadenv/internal/envPair"
)

const (
	FormatShell = "shell"
	FormatFish  = "fish"
	FormatJSON  = "json"
)

func Render(pairs []envPair.EnvPair, format string) (string, error) {
	pairsByKey := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		if pair.Key == "" {
			return "", fmt.Errorf("pair key cannot be empty")
		}
		pairsByKey[pair.Key] = pair.Value
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", FormatShell:
		return renderShell(pairsByKey), nil
	case FormatFish:
		return renderFish(pairsByKey), nil
	case FormatJSON:
		return renderJSON(pairsByKey)
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func renderShell(pairs map[string]string) string {
	keys := sortedKeys(pairs)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString("export ")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(shellQuote(pairs[key]))
		b.WriteByte('\n')
	}

	return b.String()
}

func renderFish(pairs map[string]string) string {
	keys := sortedKeys(pairs)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString("set -gx ")
		b.WriteString(key)
		b.WriteString(" ")
		b.WriteString(shellQuote(pairs[key]))
		b.WriteByte('\n')
	}

	return b.String()
}

func renderJSON(pairs map[string]string) (string, error) {
	b, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding json format: %w", err)
	}

	return string(append(b, '\n')), nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
