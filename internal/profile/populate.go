package profile

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/parser"
)

type editorRunner func(path string) error

var (
	editorLookPath = exec.LookPath
	defaultEditors = []string{"nano", "nvim", "vim", "vi"}
)

func FromFile(path string) ([]envPair.EnvPair, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading env file %q: %w", path, err)
	}

	pairs, err := parser.ParseEnvContent(string(b))
	if err != nil {
		return nil, fmt.Errorf("parsing env file %q: %w", path, err)
	}

	return pairs, nil
}

func FromEditor(initialContent string) ([]envPair.EnvPair, error) {
	return fromEditorWithRunner(initialContent, runConfiguredEditor)
}

func fromEditorWithRunner(initialContent string, run editorRunner) ([]envPair.EnvPair, error) {
	if run == nil {
		return nil, fmt.Errorf("editor runner is nil")
	}

	tmp, err := os.CreateTemp("", "deadenv-*.env")
	if err != nil {
		return nil, fmt.Errorf("creating temp editor file: %w", err)
	}

	path := tmp.Name()
	defer func() {
		_ = os.Remove(path)
	}()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("setting temp editor file permissions: %w", err)
	}

	if _, err := tmp.WriteString(initialContent); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("writing temp editor file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("closing temp editor file: %w", err)
	}

	if err := run(path); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEditorFailed, err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading edited content: %w", err)
	}

	pairs, err := parser.ParseEnvContent(string(b))
	if err != nil {
		return nil, fmt.Errorf("parsing edited content: %w", err)
	}

	return pairs, nil
}

func runConfiguredEditor(path string) error {
	name, args, err := editorCommand(path)
	if err != nil {
		return err
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running editor %q: %w", name, err)
	}

	return nil
}

func editorCommand(path string) (string, []string, error) {
	editor := resolveEditor()
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("no editor configured; set DEADENV_EDITOR, VISUAL, or EDITOR")
	}

	name := parts[0]
	args := append(parts[1:], path)

	return name, args, nil
}

func resolveEditor() string {
	if deadenvEditor := strings.TrimSpace(os.Getenv("DEADENV_EDITOR")); deadenvEditor != "" {
		return deadenvEditor
	}

	if visual := strings.TrimSpace(os.Getenv("VISUAL")); visual != "" {
		return visual
	}

	if editor := strings.TrimSpace(os.Getenv("EDITOR")); editor != "" {
		return editor
	}

	for _, editor := range defaultEditors {
		if _, err := editorLookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}
