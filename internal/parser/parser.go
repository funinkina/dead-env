package parser

import (
	"fmt"
	"regexp"
	"strings"

	env "funinkina/deadenv/internal/envPair"
)

var keyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func ParseEnvContent(input string) ([]env.EnvPair, error) {

	lines := strings.Split(input, "\n")
	var output []env.EnvPair

	for idx, line := range lines {
		lineNum := idx + 1
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}

		key, value, err := splitLine(line)

		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if !isValidKey(key) {
			return nil, fmt.Errorf("line %d: invalid key: %q", lineNum, key)
		}

		output = append(output, env.EnvPair{Key: key, Value: value})
	}

	return output, nil
}

func splitLine(line string) (string, string, error) {
	if before, after, ok := strings.Cut(line, "="); ok {
		key := strings.TrimSpace(before)
		value := parseValue(after)
		return key, value, nil
	}

	for idx, r := range line {
		if r != ' ' && r != '\t' {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := parseValue(line[idx+1:])
		return key, value, nil
	}

	return line, "", nil
}

func parseValue(val string) string {
	val = strings.TrimLeft(val, " \t")

	if len(val) == 0 {
		return ""
	}

	if val[0] == '"' || val[0] == '\'' {
		quote := val[0]
		last := strings.LastIndexByte(val[1:], quote)
		if last == -1 {
			return val // unmatched quote fallback
		}

		content := val[1 : last+1]

		if quote == '"' {
			content = unescapeDoubleQuoted(content)
		}

		return content
	}

	// strip inline comment
	if idx := strings.Index(val, " #"); idx != -1 {
		val = val[:idx]
	}

	return strings.TrimSpace(val)
}

func isValidKey(k string) bool {
	return keyRegex.MatchString(k)
}

func unescapeDoubleQuoted(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}

		switch s[i+1] {
		case '"':
			b.WriteByte('"')
			i++
		case '\\':
			b.WriteByte('\\')
			i++
		case 'n':
			b.WriteByte('\n')
			i++
		case 't':
			b.WriteByte('\t')
			i++
		default:
			b.WriteByte(s[i])
		}
	}

	return b.String()
}
