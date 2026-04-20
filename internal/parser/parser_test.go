package parser

import (
	env "funinkina/deadenv/internal/envPair"
	"reflect"
	"testing"
)

func TestParseEnvContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []env.EnvPair
		wantErr bool
	}{
		{
			name:  "basic",
			input: "KEY=VALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "spaces",
			input: "KEY = VALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "export",
			input: "export KEY=VALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "export with extra whitespace",
			input: "export   KEY=VALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "space delimiter",
			input: "KEY VALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "tab delimiter",
			input: "KEY\tVALUE",
			want:  []env.EnvPair{{Key: "KEY", Value: "VALUE"}},
		},
		{
			name:  "space delimiter preserves value spacing",
			input: "KEY   first  second",
			want:  []env.EnvPair{{Key: "KEY", Value: "first  second"}},
		},
		{
			name:  "equals in value are preserved",
			input: "API_KEY=abc123==",
			want:  []env.EnvPair{{Key: "API_KEY", Value: "abc123=="}},
		},
		{
			name:  "quoted",
			input: `KEY="hello world"`,
			want:  []env.EnvPair{{Key: "KEY", Value: "hello world"}},
		},
		{
			name:  "single quoted",
			input: `KEY='quoted==value'`,
			want:  []env.EnvPair{{Key: "KEY", Value: "quoted==value"}},
		},
		{
			name:  "empty value",
			input: "KEY=",
			want:  []env.EnvPair{{Key: "KEY", Value: ""}},
		},
		{
			name:  "explicit empty quoted value",
			input: `KEY=""`,
			want:  []env.EnvPair{{Key: "KEY", Value: ""}},
		},
		{
			name:  "escaped double quoted content",
			input: `KEY="has \"escaped\"\\path\n\tend"`,
			want:  []env.EnvPair{{Key: "KEY", Value: "has \"escaped\"\\path\n\tend"}},
		},
		{
			name:  "inline comment",
			input: "KEY=value # comment",
			want:  []env.EnvPair{{Key: "KEY", Value: "value"}},
		},
		{
			name:  "quoted hash is preserved",
			input: `KEY="value # not a comment"`,
			want:  []env.EnvPair{{Key: "KEY", Value: "value # not a comment"}},
		},
		{
			name:  "url hash",
			input: "KEY=http://test#anchor",
			want:  []env.EnvPair{{Key: "KEY", Value: "http://test#anchor"}},
		},
		{
			name:  "unmatched quote falls back to literal",
			input: `KEY="unterminated`,
			want:  []env.EnvPair{{Key: "KEY", Value: `"unterminated`}},
		},
		{
			name:    "invalid key",
			input:   "1KEY=value",
			wantErr: true,
		},
		{
			name: "multi-line",
			input: `
# comment
	A=1
B=2
`,
			want: []env.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
			},
		},
		{
			name:  "skips blank lines and comments",
			input: "\n# comment\nFOO=bar\n\n  # another comment\nBAR=baz",
			want: []env.EnvPair{
				{Key: "FOO", Value: "bar"},
				{Key: "BAR", Value: "baz"},
			},
		},
		{
			name:  "parses keys without explicit values",
			input: "EMPTY\nNAME first second",
			want: []env.EnvPair{
				{Key: "EMPTY", Value: ""},
				{Key: "NAME", Value: "first second"},
			},
		},
		{
			name:  "windows line endings",
			input: "KEY=VALUE\r\nSECOND=NEXT\r\n",
			want: []env.EnvPair{
				{Key: "KEY", Value: "VALUE"},
				{Key: "SECOND", Value: "NEXT"},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			got, err := ParseEnvContent(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
