package profile

import (
	"fmt"
	"sort"
	"strings"

	"funinkina/deadenv/internal/envPair"
)

const editorHeaderTemplate = `# deadenv profile: %s
# Edit environment variables below.
#
# Accepted formats:
#   KEY=VALUE
#   KEY = VALUE
#   KEY VALUE
#   export KEY=VALUE
#
# Lines starting with # are ignored.
# Save and close the editor to continue.

`

func EditorTemplate(profile string) string {
	label := profile
	if strings.TrimSpace(label) == "" {
		label = "new-profile"
	}

	return fmt.Sprintf(editorHeaderTemplate, label)
}

func SerializeEnvPairs(profile string, pairs []envPair.EnvPair) string {
	sortedPairs := append([]envPair.EnvPair(nil), pairs...)
	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[i].Key < sortedPairs[j].Key
	})

	var b strings.Builder
	b.WriteString(EditorTemplate(profile))

	for _, pair := range sortedPairs {
		b.WriteString(pair.Key)
		b.WriteString("=")
		b.WriteString(formatSerializedValue(pair.Value))
		b.WriteString("\n")
	}

	return b.String()
}

func formatSerializedValue(value string) string {
	if value == "" {
		return ""
	}

	if !needsDoubleQuotes(value) {
		return value
	}

	return `"` + escapeDoubleQuoted(value) + `"`
}

func needsDoubleQuotes(value string) bool {
	if strings.TrimSpace(value) != value {
		return true
	}

	if strings.Contains(value, " #") {
		return true
	}

	for _, r := range value {
		switch r {
		case ' ', '\t', '\n', '\r', '#', '"', '\\':
			return true
		}
	}

	return false
}

func escapeDoubleQuoted(value string) string {
	var b strings.Builder
	b.Grow(len(value))

	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\n':
			b.WriteString("\\n")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}
