package parser

import "testing"

func FuzzParseEnvContent(f *testing.F) {
	f.Add("KEY=VALUE\n")
	f.Add("KEY = VALUE\n")
	f.Add("export KEY=VALUE\n")
	f.Add("KEY value\n")

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = ParseEnvContent(data)
	})
}
