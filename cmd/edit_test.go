package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/profile"
)

func TestEditCommandRequiresProfileName(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "edit"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestEditCommandHandlesNoChangesAsSuccess(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRunEdit := runEdit
	defer func() {
		newProfileService = oldNewProfileService
		runEdit = oldRunEdit
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	called := false
	runEdit = func(service *profile.ProfileService, profileName string, opts profile.EditOptions) error {
		if service == nil {
			t.Fatal("service is nil")
		}
		if profileName != "myapp" {
			t.Fatalf("profileName = %q, want %q", profileName, "myapp")
		}
		if opts.Out == nil {
			t.Fatal("opts.Out is nil")
		}
		called = true

		return profile.ErrNoChanges
	}

	root := NewRootCommand()
	var out bytes.Buffer
	root.Writer = &out

	err := root.Run(context.Background(), []string{"deadenv", "edit", "myapp"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !called {
		t.Fatal("runEdit was not called")
	}
}

func TestEditCommandPassesYesFlag(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRunEdit := runEdit
	defer func() {
		newProfileService = oldNewProfileService
		runEdit = oldRunEdit
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	called := false
	runEdit = func(service *profile.ProfileService, profileName string, opts profile.EditOptions) error {
		_ = service
		_ = profileName
		called = true
		if !opts.Yes {
			t.Fatal("opts.Yes = false, want true")
		}

		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "edit", "myapp", "--yes"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !called {
		t.Fatal("runEdit was not called")
	}
}

func TestEditCommandReturnsUnderlyingError(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRunEdit := runEdit
	defer func() {
		newProfileService = oldNewProfileService
		runEdit = oldRunEdit
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	wantErr := errors.New("boom")
	runEdit = func(_ *profile.ProfileService, _ string, _ profile.EditOptions) error {
		return wantErr
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "edit", "myapp"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}
