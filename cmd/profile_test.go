package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/profile"
)

func TestProfileNewFromFileCreatesProfile(t *testing.T) {
	oldNewProfileService := newProfileService
	oldLoadPairsFromFile := loadPairsFromFile
	oldPromptConfirm := promptConfirm
	oldPrintPairSummary := printPairSummary
	defer func() {
		newProfileService = oldNewProfileService
		loadPairsFromFile = oldLoadPairsFromFile
		promptConfirm = oldPromptConfirm
		printPairSummary = oldPrintPairSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	loadPairsFromFile = func(path string) ([]envPair.EnvPair, error) {
		if path != "app.env" {
			t.Fatalf("loadPairsFromFile path = %q, want %q", path, "app.env")
		}

		return []envPair.EnvPair{{Key: "A", Value: "1"}}, nil
	}
	promptConfirm = func(message string) (bool, error) {
		if message == "" {
			t.Fatal("prompt message is empty")
		}

		return true, nil
	}
	printPairSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		return nil
	}

	root := NewRootCommand()
	var out bytes.Buffer
	root.Writer = &out

	err := root.Run(context.Background(), []string{"deadenv", "profile", "new", "myapp", "--from", "app.env"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	value, err := store.Read("deadenv/myapp", "A", "")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if value != "1" {
		t.Fatalf("value = %q, want %q", value, "1")
	}
}

func TestProfileNewRequiresName(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "new"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestProfileNewCancelled(t *testing.T) {
	oldNewProfileService := newProfileService
	oldLoadPairsFromEditor := loadPairsFromEditor
	oldPromptConfirm := promptConfirm
	oldPrintPairSummary := printPairSummary
	defer func() {
		newProfileService = oldNewProfileService
		loadPairsFromEditor = oldLoadPairsFromEditor
		promptConfirm = oldPromptConfirm
		printPairSummary = oldPrintPairSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	loadPairsFromEditor = func(initialContent string) ([]envPair.EnvPair, error) {
		if initialContent == "" {
			t.Fatal("expected initial editor content")
		}

		return []envPair.EnvPair{{Key: "A", Value: "1"}}, nil
	}
	promptConfirm = func(message string) (bool, error) {
		_ = message
		return false, nil
	}
	printPairSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "new", "myapp"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	_, err = store.Read("deadenv/myapp", "A", "")
	if !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read() error = %v, want ErrKeyNotFound", err)
	}
}

func TestProfileNewEditorEmptyCancelsCleanly(t *testing.T) {
	oldNewProfileService := newProfileService
	oldLoadPairsFromEditor := loadPairsFromEditor
	oldPrintPairSummary := printPairSummary
	defer func() {
		newProfileService = oldNewProfileService
		loadPairsFromEditor = oldLoadPairsFromEditor
		printPairSummary = oldPrintPairSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	loadPairsFromEditor = func(initialContent string) ([]envPair.EnvPair, error) {
		_ = initialContent
		return []envPair.EnvPair{}, nil
	}
	printPairSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		t.Fatal("printPairSummary should not be called for empty parsed content")
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "new", "myapp"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	_, readErr := store.Read("deadenv/myapp", "A", "")
	if !errors.Is(readErr, keychain.ErrKeyNotFound) {
		t.Fatalf("Read() error = %v, want ErrKeyNotFound", readErr)
	}
}

func TestProfileNewYesSkipsConfirmation(t *testing.T) {
	oldNewProfileService := newProfileService
	oldLoadPairsFromFile := loadPairsFromFile
	oldPromptConfirm := promptConfirm
	oldPrintPairSummary := printPairSummary
	defer func() {
		newProfileService = oldNewProfileService
		loadPairsFromFile = oldLoadPairsFromFile
		promptConfirm = oldPromptConfirm
		printPairSummary = oldPrintPairSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	loadPairsFromFile = func(path string) ([]envPair.EnvPair, error) {
		_ = path
		return []envPair.EnvPair{{Key: "A", Value: "1"}}, nil
	}
	promptConfirm = func(message string) (bool, error) {
		t.Fatalf("promptConfirm should not be called when --yes is set (message=%q)", message)
		return false, nil
	}
	printPairSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "new", "myapp", "--from", "app.env", "--yes"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	value, err := store.Read("deadenv/myapp", "A", "")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if value != "1" {
		t.Fatalf("value = %q, want %q", value, "1")
	}
}

func TestProfileNewGivesEditorHint(t *testing.T) {
	oldNewProfileService := newProfileService
	oldLoadPairsFromEditor := loadPairsFromEditor
	defer func() {
		newProfileService = oldNewProfileService
		loadPairsFromEditor = oldLoadPairsFromEditor
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)

	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}
	loadPairsFromEditor = func(initialContent string) ([]envPair.EnvPair, error) {
		_ = initialContent
		return nil, profile.ErrEditorFailed
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "new", "myapp"})
	if !errors.Is(err, profile.ErrEditorFailed) {
		t.Fatalf("Run() error = %v, want ErrEditorFailed", err)
	}

	msg := err.Error()
	if !strings.Contains(msg, "$DEADENV_EDITOR") || !strings.Contains(msg, "$VISUAL") || !strings.Contains(msg, "$EDITOR") {
		t.Fatalf("error message = %q, want editor hint", msg)
	}
}

func TestProfileListUsesServiceListProfiles(t *testing.T) {
	oldNewProfileService := newProfileService
	oldListProfiles := listProfiles
	defer func() {
		newProfileService = oldNewProfileService
		listProfiles = oldListProfiles
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	called := false
	listProfiles = func(svc *profile.ProfileService) ([]string, error) {
		called = true
		if svc == nil {
			t.Fatal("service is nil")
		}
		return []string{"alpha", "beta"}, nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "list"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Fatal("listProfiles was not called")
	}
}

func TestProfileListAliasLs(t *testing.T) {
	oldNewProfileService := newProfileService
	oldListProfiles := listProfiles
	defer func() {
		newProfileService = oldNewProfileService
		listProfiles = oldListProfiles
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	called := false
	listProfiles = func(svc *profile.ProfileService) ([]string, error) {
		called = true
		_ = svc
		return []string{"alpha"}, nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "ls"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !called {
		t.Fatal("listProfiles was not called")
	}
}

func TestProfileListNoProfilesIsSuccess(t *testing.T) {
	oldNewProfileService := newProfileService
	oldListProfiles := listProfiles
	defer func() {
		newProfileService = oldNewProfileService
		listProfiles = oldListProfiles
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	listProfiles = func(_ *profile.ProfileService) ([]string, error) {
		return []string{}, nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "list"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestProfileListReturnsUnderlyingError(t *testing.T) {
	oldNewProfileService := newProfileService
	oldListProfiles := listProfiles
	defer func() {
		newProfileService = oldNewProfileService
		listProfiles = oldListProfiles
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	wantErr := errors.New("boom")
	listProfiles = func(_ *profile.ProfileService) ([]string, error) {
		return nil, wantErr
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "profile", "list"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func mustProfileService(t *testing.T, store keychain.Store) *profile.ProfileService {
	t.Helper()

	service, err := profile.NewProfileService(store, &history.FakeRecorder{}, func(value string) (string, error) {
		return "hash:" + value, nil
	})
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	return service
}
