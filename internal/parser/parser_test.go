package parser

import (
	"reflect"
	"testing"
)

func TestParseEnvContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []EnvPair
		wantErr bool
	}{
		{
			name:  "basic",
			input: "KEY=VALUE",
			want:  []EnvPair{{"KEY", "VALUE"}},
		},
		{
			name:  "spaces",
			input: "KEY = VALUE",
			want:  []EnvPair{{"KEY", "VALUE"}},
		},
		{
			name:  "export",
			input: "export KEY=VALUE",
			want:  []EnvPair{{"KEY", "VALUE"}},
		},
		{
			name:  "quoted",
			input: `KEY="hello world"`,
			want:  []EnvPair{{"KEY", "hello world"}},
		},
		{
			name:  "inline comment",
			input: "KEY=value # comment",
			want:  []EnvPair{{"KEY", "value"}},
		},
		{
			name:  "url hash",
			input: "KEY=http://test#anchor",
			want:  []EnvPair{{"KEY", "http://test#anchor"}},
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
			want: []EnvPair{
				{"A", "1"},
				{"B", "2"},
			},
		},
		{
			name:  "skips blank lines and comments",
			input: "\n# comment\nFOO=bar\n\n  # another comment\nBAR=baz",
			want: []EnvPair{
				{Key: "FOO", Value: "bar"},
				{Key: "BAR", Value: "baz"},
			},
		},
		{
			name:  "parses keys without explicit values",
			input: "EMPTY\nNAME first second",
			want: []EnvPair{
				{Key: "EMPTY", Value: ""},
				{Key: "NAME", Value: "first second"},
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
