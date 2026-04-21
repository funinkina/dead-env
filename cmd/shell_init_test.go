package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestInitCommandWithBashShell(t *testing.T) {
	root := NewRootCommand()

	err := root.Run(context.Background(), []string{"deadenv", "init", "--shell", "bash"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestInitCommandWithFishShell(t *testing.T) {
	root := NewRootCommand()

	err := root.Run(context.Background(), []string{"deadenv", "init", "--shell", "fish"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

}

func TestShellHookSnippetBash(t *testing.T) {
	snippet, err := shellHookSnippet(shellBash)
	if err != nil {
		t.Fatalf("shellHookSnippet() error = %v", err)
	}

	if !strings.Contains(snippet, "deadenv() {") {
		t.Fatalf("snippet missing deadenv function: %q", snippet)
	}
	if !strings.Contains(snippet, `eval "$(command deadenv export "$2")"`) {
		t.Fatalf("snippet missing export eval hook: %q", snippet)
	}
}

func TestShellHookSnippetFish(t *testing.T) {
	snippet, err := shellHookSnippet(shellFish)
	if err != nil {
		t.Fatalf("shellHookSnippet() error = %v", err)
	}

	if !strings.Contains(snippet, "function deadenv") {
		t.Fatalf("snippet missing deadenv fish function: %q", snippet)
	}
	if !strings.Contains(snippet, `command deadenv export "$argv[2]" --format=fish | source`) {
		t.Fatalf("snippet missing fish export pipeline: %q", snippet)
	}
}

func TestInitCommandRejectsUnsupportedShell(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "init", "--shell", "powershell"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestResolveShellName(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		envShell  string
		want      string
		wantErr   bool
	}{
		{name: "flag wins", flagValue: "zsh", envShell: "/bin/bash", want: "zsh"},
		{name: "detect from env", flagValue: "", envShell: "/bin/fish", want: "fish"},
		{name: "unsupported env shell", flagValue: "", envShell: "/bin/tcsh", wantErr: true},
		{name: "missing shell", flagValue: "", envShell: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveShellName(tt.flagValue, tt.envShell)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveShellName() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Fatalf("resolveShellName() = %q, want %q", got, tt.want)
			}
		})
	}
}
