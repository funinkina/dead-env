package parser

import "strings"

func ParseEnvContent(input string) {
	lines := strings.Split(input, "\n")
	var output []EnvPair
	for idx, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		var key, value string
		plusIndex := strings.Index(line, "=")
		if plusIndex != -1 {
			key = strings.TrimSpace(line[:plusIndex])
			value = strings.TrimSpace(line[plusIndex+1:])
		} else {
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			if len(fields) == 1 {
				key = fields[0]
				value = ""
			} else {
				key = fields[0]
				value = strings.Join(fields[1:], " ")
			}
		}
		output = append(output, EnvPair{Key: key, Value: value})
	}
}

func parseValue() {

}

func isValidKey() {

}
