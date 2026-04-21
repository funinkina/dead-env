package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/funinkina/deadenv/internal/keychain"
	"github.com/funinkina/deadenv/internal/profile"
)

func TestRunCommandRequiresProfile(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "run"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestRunCommandRequiresSeparator(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "run", "myapp", "env"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestRunCommandRequiresCommandAfterSeparator(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "run", "myapp", "--"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestRunCommandDelegatesToProfileRun(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRunProfileCommand := runProfileCommand
	defer func() {
		newProfileService = oldNewProfileService
		runProfileCommand = oldRunProfileCommand
	}()

	service := mustProfileService(t, keychain.NewFake())
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	called := false
	runProfileCommand = func(gotService *profile.ProfileService, profileName, commandName string, args []string, opts profile.RunOptions) error {
		called = true

		if gotService != service {
			t.Fatal("service mismatch")
		}
		if profileName != "myapp" {
			t.Fatalf("profileName = %q, want %q", profileName, "myapp")
		}
		if commandName != "echo" {
			t.Fatalf("commandName = %q, want %q", commandName, "echo")
		}
		if len(args) != 1 || args[0] != "hello" {
			t.Fatalf("args = %v, want [hello]", args)
		}
		if opts.Stdout == nil {
			t.Fatal("opts.Stdout should not be nil")
		}
		if opts.Stderr == nil {
			t.Fatal("opts.Stderr should not be nil")
		}

		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "run", "myapp", "--", "echo", "hello"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !called {
		t.Fatal("runProfileCommand was not called")
	}
}

func TestRunCommandReturnsUnderlyingError(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRunProfileCommand := runProfileCommand
	defer func() {
		newProfileService = oldNewProfileService
		runProfileCommand = oldRunProfileCommand
	}()

	service := mustProfileService(t, keychain.NewFake())
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	wantErr := errors.New("boom")
	runProfileCommand = func(_ *profile.ProfileService, _ string, _ string, _ []string, _ profile.RunOptions) error {
		return wantErr
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "run", "myapp", "--", "echo", "hello"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestParseRunArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantProfile string
		wantCommand string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:        "valid",
			args:        []string{"myapp", "--", "npm", "run", "dev"},
			wantProfile: "myapp",
			wantCommand: "npm",
			wantArgs:    []string{"run", "dev"},
		},
		{
			name:    "missing profile",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "missing separator",
			args:    []string{"myapp", "npm"},
			wantErr: true,
		},
		{
			name:    "missing command",
			args:    []string{"myapp", "--"},
			wantErr: true,
		},
		{
			name:    "too many profile args",
			args:    []string{"myapp", "extra", "--", "npm"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProfile, gotCommand, gotArgs, err := parseRunArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRunArgs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if gotProfile != tt.wantProfile {
				t.Fatalf("profile = %q, want %q", gotProfile, tt.wantProfile)
			}
			if gotCommand != tt.wantCommand {
				t.Fatalf("command = %q, want %q", gotCommand, tt.wantCommand)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", gotArgs, tt.wantArgs)
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Fatalf("args = %v, want %v", gotArgs, tt.wantArgs)
				}
			}
		})
	}
}
