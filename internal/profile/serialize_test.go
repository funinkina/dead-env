package profile

import (
	"strings"
	"testing"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/parser"
)

func TestEditorTemplateIncludesFormatGuide(t *testing.T) {
	tpl := EditorTemplate("myapp")

	if !strings.Contains(tpl, "deadenv profile: myapp") {
		t.Fatalf("EditorTemplate() missing profile header: %q", tpl)
	}

	wants := []string{
		"KEY=VALUE",
		"KEY = VALUE",
		"KEY VALUE",
		"export KEY=VALUE",
	}

	for _, want := range wants {
		if !strings.Contains(tpl, want) {
			t.Fatalf("EditorTemplate() missing format %q", want)
		}
	}
}

func TestSerializeEnvPairsRoundTripViaParser(t *testing.T) {
	pairs := []envPair.EnvPair{
		{Key: "API_KEY", Value: "abc123=="},
		{Key: "COMMENTED", Value: "value # keep"},
		{Key: "DATABASE_URL", Value: "postgres://localhost/app"},
		{Key: "ESCAPED", Value: "line1\nline2\tend"},
		{Key: "EMPTY", Value: ""},
		{Key: "QUOTED", Value: `has "quotes" and \\ slash`},
		{Key: "SPACE", Value: "hello world"},
	}

	content := SerializeEnvPairs("myapp", pairs)
	got, err := parser.ParseEnvContent(content)
	if err != nil {
		t.Fatalf("ParseEnvContent() error = %v", err)
	}

	if len(got) != len(pairs) {
		t.Fatalf("parsed pair count = %d, want %d", len(got), len(pairs))
	}

	wantByKey := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		wantByKey[pair.Key] = pair.Value
	}

	for _, pair := range got {
		wantValue, ok := wantByKey[pair.Key]
		if !ok {
			t.Fatalf("unexpected key in parsed output: %q", pair.Key)
		}
		if pair.Value != wantValue {
			t.Fatalf("value mismatch for key %q: got %q, want %q", pair.Key, pair.Value, wantValue)
		}
	}
}

func TestSerializeEnvPairsWritesSortedKeys(t *testing.T) {
	pairs := []envPair.EnvPair{
		{Key: "Z", Value: "1"},
		{Key: "A", Value: "2"},
		{Key: "M", Value: "3"},
	}

	content := SerializeEnvPairs("myapp", pairs)
	idxA := strings.Index(content, "A=2")
	idxM := strings.Index(content, "M=3")
	idxZ := strings.Index(content, "Z=1")

	if idxA == -1 || idxM == -1 || idxZ == -1 {
		t.Fatalf("serialized content missing expected keys:\n%s", content)
	}

	if !(idxA < idxM && idxM < idxZ) {
		t.Fatalf("keys are not sorted in output:\n%s", content)
	}
}
