package profile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveEditorPrefersDeadenvEditor(t *testing.T) {
	t.Setenv("DEADENV_EDITOR", "nano --tempfile")
	t.Setenv("VISUAL", "vim")
	t.Setenv("EDITOR", "vi")

	got := resolveEditor()
	if got != "nano --tempfile" {
		t.Fatalf("resolveEditor() = %q, want %q", got, "nano --tempfile")
	}
}

func TestResolveEditorPrefersVisualOverEditor(t *testing.T) {
	t.Setenv("DEADENV_EDITOR", "")
	t.Setenv("VISUAL", "nvim")
	t.Setenv("EDITOR", "nano")

	got := resolveEditor()
	if got != "nvim" {
		t.Fatalf("resolveEditor() = %q, want %q", got, "nvim")
	}
}

func TestResolveEditorFallsBackToFirstAvailableDefault(t *testing.T) {
	t.Setenv("DEADENV_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	oldLookPath := editorLookPath
	editorLookPath = func(file string) (string, error) {
		switch file {
		case "nano":
			return "/usr/bin/nano", nil
		default:
			return "", errors.New("not found")
		}
	}
	defer func() {
		editorLookPath = oldLookPath
	}()

	got := resolveEditor()
	if got != "nano" {
		t.Fatalf("resolveEditor() = %q, want %q", got, "nano")
	}
}

func TestResolveEditorReturnsEmptyWhenNoEditorFound(t *testing.T) {
	t.Setenv("DEADENV_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	oldLookPath := editorLookPath
	editorLookPath = func(file string) (string, error) {
		_ = file
		return "", errors.New("not found")
	}
	defer func() {
		editorLookPath = oldLookPath
	}()

	got := resolveEditor()
	if got != "" {
		t.Fatalf("resolveEditor() = %q, want empty string", got)
	}
}

func TestEditorCommandReturnsHelpfulErrorWhenNoEditorConfigured(t *testing.T) {
	t.Setenv("DEADENV_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	oldLookPath := editorLookPath
	editorLookPath = func(file string) (string, error) {
		_ = file
		return "", errors.New("not found")
	}
	defer func() {
		editorLookPath = oldLookPath
	}()

	_, _, err := editorCommand("/tmp/file")
	if err == nil {
		t.Fatal("editorCommand() error = nil, want non-nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "DEADENV_EDITOR") || !strings.Contains(msg, "VISUAL") || !strings.Contains(msg, "EDITOR") {
		t.Fatalf("editorCommand() error = %q, want env var guidance", msg)
	}
}

func TestFromFileParsesValidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.env")
	if err := os.WriteFile(path, []byte("A=1\nB=2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	pairs, err := FromFile(path)
	if err != nil {
		t.Fatalf("FromFile() error = %v", err)
	}

	if len(pairs) != 2 {
		t.Fatalf("pair count = %d, want 2", len(pairs))
	}
}

func TestFromFileMissingPathReturnsError(t *testing.T) {
	_, err := FromFile(filepath.Join(t.TempDir(), "missing.env"))
	if err == nil {
		t.Fatal("FromFile() error = nil, want non-nil")
	}
}

func TestFromEditorWithRunnerCreates0600FileAndCleansUp(t *testing.T) {
	var tempPath string
	var mode fs.FileMode

	runner := func(path string) error {
		tempPath = path
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		mode = info.Mode().Perm()

		return os.WriteFile(path, []byte("A=1\n"), 0o600)
	}

	pairs, err := fromEditorWithRunner("# seed\n", runner)
	if err != nil {
		t.Fatalf("fromEditorWithRunner() error = %v", err)
	}

	if mode != 0o600 {
		t.Fatalf("temp file mode = %o, want 600", mode)
	}

	if len(pairs) != 1 || pairs[0].Key != "A" || pairs[0].Value != "1" {
		t.Fatalf("pairs = %v, want [{A 1}]", pairs)
	}

	_, statErr := os.Stat(tempPath)
	if !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("temp file still exists or unexpected error: %v", statErr)
	}
}

func TestFromEditorWithRunnerReturnsErrEditorFailed(t *testing.T) {
	runner := func(path string) error {
		_ = path
		return errors.New("editor failed")
	}

	_, err := fromEditorWithRunner("", runner)
	if !errors.Is(err, ErrEditorFailed) {
		t.Fatalf("fromEditorWithRunner() error = %v, want ErrEditorFailed", err)
	}
}

func TestFromEditorWithRunnerParsesCommentsOnlyAsEmpty(t *testing.T) {
	runner := func(path string) error {
		return os.WriteFile(path, []byte("# only comments\n\n   # another\n"), 0o600)
	}

	pairs, err := fromEditorWithRunner("", runner)
	if err != nil {
		t.Fatalf("fromEditorWithRunner() error = %v", err)
	}

	if len(pairs) != 0 {
		t.Fatalf("pair count = %d, want 0", len(pairs))
	}
}
