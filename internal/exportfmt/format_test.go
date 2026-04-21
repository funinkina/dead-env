package exportfmt

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/funinkina/deadenv/internal/envPair"
)

func TestRenderShellSortsAndQuotesValues(t *testing.T) {
	got, err := Render([]envPair.EnvPair{
		{Key: "B", Value: "space value"},
		{Key: "A", Value: "it's=ok"},
	}, FormatShell)
	if err != nil {
		t.Fatalf("Render(shell) error = %v", err)
	}

	want := "export A='it'\\''s=ok'\nexport B='space value'\n"
	if got != want {
		t.Fatalf("Render(shell) = %q, want %q", got, want)
	}
}

func TestRenderFishSortsAndQuotesValues(t *testing.T) {
	got, err := Render([]envPair.EnvPair{
		{Key: "B", Value: "line1\nline2"},
		{Key: "A", Value: "it's=ok"},
	}, FormatFish)
	if err != nil {
		t.Fatalf("Render(fish) error = %v", err)
	}

	want := "set -gx A 'it'\\''s=ok'\nset -gx B 'line1\nline2'\n"
	if got != want {
		t.Fatalf("Render(fish) = %q, want %q", got, want)
	}
}

func TestRenderJSONProducesObject(t *testing.T) {
	got, err := Render([]envPair.EnvPair{
		{Key: "B", Value: "2"},
		{Key: "A", Value: "1"},
	}, FormatJSON)
	if err != nil {
		t.Fatalf("Render(json) error = %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("Unmarshal(json) error = %v", err)
	}

	want := map[string]string{"A": "1", "B": "2"}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("json map = %#v, want %#v", decoded, want)
	}

	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("json output should end with newline")
	}
}

func TestRenderDeduplicatesByLastValue(t *testing.T) {
	got, err := Render([]envPair.EnvPair{
		{Key: "A", Value: "old"},
		{Key: "A", Value: "new"},
	}, FormatShell)
	if err != nil {
		t.Fatalf("Render(shell) error = %v", err)
	}

	want := "export A='new'\n"
	if got != want {
		t.Fatalf("Render(shell) = %q, want %q", got, want)
	}
}

func TestRenderRejectsUnknownFormat(t *testing.T) {
	_, err := Render([]envPair.EnvPair{{Key: "A", Value: "1"}}, "toml")
	if err == nil {
		t.Fatal("Render() error = nil, want non-nil")
	}
}

func TestRenderRejectsEmptyKey(t *testing.T) {
	_, err := Render([]envPair.EnvPair{{Key: "", Value: "1"}}, FormatShell)
	if err == nil {
		t.Fatal("Render() error = nil, want non-nil")
	}
}
