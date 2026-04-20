package tui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"funinkina/deadenv/internal/envPair"
)

func TestPromptHiddenWithReader(t *testing.T) {
	var out bytes.Buffer
	calledFD := -1

	got, err := promptHiddenWithReader(&out, "Sharing password", 42, func(fd int) ([]byte, error) {
		calledFD = fd
		return []byte("s3cr3t"), nil
	})
	if err != nil {
		t.Fatalf("promptHiddenWithReader() error = %v", err)
	}

	if got != "s3cr3t" {
		t.Fatalf("promptHiddenWithReader() = %q, want %q", got, "s3cr3t")
	}

	if calledFD != 42 {
		t.Fatalf("fd = %d, want %d", calledFD, 42)
	}

	if out.String() != "Sharing password: \n" {
		t.Fatalf("output = %q, want %q", out.String(), "Sharing password: \n")
	}
}

func TestPromptHiddenWithReaderDefaultsLabel(t *testing.T) {
	var out bytes.Buffer

	_, err := promptHiddenWithReader(&out, "   ", 7, func(fd int) ([]byte, error) {
		_ = fd
		return []byte("secret"), nil
	})
	if err != nil {
		t.Fatalf("promptHiddenWithReader() error = %v", err)
	}

	if out.String() != "Password: \n" {
		t.Fatalf("output = %q, want %q", out.String(), "Password: \n")
	}
}

func TestPromptHiddenWithReaderReturnsReaderError(t *testing.T) {
	var out bytes.Buffer
	wantErr := errors.New("boom")

	_, err := promptHiddenWithReader(&out, "Password", 1, func(fd int) ([]byte, error) {
		_ = fd
		return nil, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("promptHiddenWithReader() error = %v, want %v", err, wantErr)
	}

	if out.String() != "Password: \n" {
		t.Fatalf("output = %q, want %q", out.String(), "Password: \n")
	}
}

func TestPromptHiddenWithReaderRejectsNilReader(t *testing.T) {
	_, err := promptHiddenWithReader(&bytes.Buffer{}, "Password", 1, nil)
	if err == nil {
		t.Fatal("promptHiddenWithReader() error = nil, want non-nil")
	}
}

func TestPrintChangeSummaryIncludesCountsAndLabels(t *testing.T) {
	var out bytes.Buffer
	err := PrintChangeSummary(&out, []string{"A"}, []string{"B"}, []string{"C"})
	if err != nil {
		t.Fatalf("PrintChangeSummary() error = %v", err)
	}

	got := out.String()
	containsAll(t, got,
		"Changes detected:",
		"Added: 1  Modified: 1  Removed: 1",
		"[set]      A",
		"[modified] B",
		"[removed]  C",
	)
}

func TestPrintPairSummaryHidesValues(t *testing.T) {
	var out bytes.Buffer
	err := PrintPairSummary(&out, []envPair.EnvPair{{Key: "B", Value: "VALUE_TWO"}, {Key: "A", Value: "VALUE_ONE"}})
	if err != nil {
		t.Fatalf("PrintPairSummary() error = %v", err)
	}

	got := out.String()
	containsAll(t, got,
		"Parsed 2 keys:",
		"  - A",
		"  - B",
		"(values are hidden)",
	)

	if strings.Contains(got, "VALUE_ONE") || strings.Contains(got, "VALUE_TWO") {
		t.Fatalf("summary leaked values: %q", got)
	}
}

func TestPromptConfirmParsesYesAndNo(t *testing.T) {
	yesIn := strings.NewReader("yes\n")
	var yesOut bytes.Buffer
	yes, err := promptConfirmWithIO(yesIn, &yesOut, "Proceed?")
	if err != nil {
		t.Fatalf("promptConfirmWithIO(yes) error = %v", err)
	}
	if !yes {
		t.Fatal("promptConfirmWithIO(yes) = false, want true")
	}

	noIn := strings.NewReader("\n")
	var noOut bytes.Buffer
	no, err := promptConfirmWithIO(noIn, &noOut, "Proceed?")
	if err != nil {
		t.Fatalf("promptConfirmWithIO(no) error = %v", err)
	}
	if no {
		t.Fatal("promptConfirmWithIO(no) = true, want false")
	}
}

func TestPromptConfirmRepromptsOnInvalidInput(t *testing.T) {
	in := strings.NewReader("maybe\ny\n")
	var out bytes.Buffer
	ok, err := promptConfirmWithIO(in, &out, "Proceed?")
	if err != nil {
		t.Fatalf("promptConfirmWithIO() error = %v", err)
	}
	if !ok {
		t.Fatal("promptConfirmWithIO() = false, want true")
	}

	containsAll(t, out.String(), "Please answer with 'y' or 'n'.")
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "short", input: "abc", want: "***"},
		{name: "long", input: "secret1234", want: "******1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskValue(tt.input)
			if got != tt.want {
				t.Fatalf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func containsAll(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("output = %q, missing %q", got, want)
		}
	}
}
