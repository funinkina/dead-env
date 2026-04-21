package cmd

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/funinkina/deadenv/internal/crypto"
	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/keychain"
	"github.com/funinkina/deadenv/internal/profile"
)

func TestExportCommandShellFormat(t *testing.T) {
	oldNewProfileService := newProfileService
	oldRenderExportOutput := renderExportOutput
	defer func() {
		newProfileService = oldNewProfileService
		renderExportOutput = oldRenderExportOutput
	}()

	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "B", "two words"); err != nil {
		t.Fatalf("Write(B) error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write(A) error = %v", err)
	}

	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	called := false
	renderExportOutput = func(pairs []envPair.EnvPair, format string) (string, error) {
		called = true
		if format != "shell" {
			t.Fatalf("format = %q, want %q", format, "shell")
		}

		wantPairs := []envPair.EnvPair{
			{Key: "A", Value: "1"},
			{Key: "B", Value: "two words"},
		}
		if !reflect.DeepEqual(pairs, wantPairs) {
			t.Fatalf("pairs = %#v, want %#v", pairs, wantPairs)
		}

		return "", nil
	}

	root := NewRootCommand()

	err := root.Run(context.Background(), []string{"deadenv", "export", "myapp", "--format", "shell"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !called {
		t.Fatal("renderExportOutput was not called")
	}
}

func TestExportCommandEncryptedFileFlow(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptExportPassword := promptExportPassword
	oldExportEncryptedProfile := exportEncryptedProfile
	defer func() {
		newProfileService = oldNewProfileService
		promptExportPassword = oldPromptExportPassword
		exportEncryptedProfile = oldExportEncryptedProfile
	}()

	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	passwordCalls := 0
	promptExportPassword = func(label string) (string, error) {
		passwordCalls++
		if label == "" {
			t.Fatal("prompt label is empty")
		}
		return "shared-pass", nil
	}

	called := false
	exportEncryptedProfile = func(profile string, pairs []envPair.EnvPair, password, outPath string) error {
		called = true
		if profile != "myapp" {
			t.Fatalf("profile = %q, want %q", profile, "myapp")
		}
		if password != "shared-pass" {
			t.Fatalf("password = %q, want %q", password, "shared-pass")
		}
		if outPath != "backup.deadenv" {
			t.Fatalf("outPath = %q, want %q", outPath, "backup.deadenv")
		}
		if len(pairs) != 1 || pairs[0].Key != "API_KEY" {
			t.Fatalf("pairs = %#v, want API_KEY pair", pairs)
		}
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "export", "myapp", "--out", "backup.deadenv"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !called {
		t.Fatal("exportEncryptedProfile was not called")
	}
	if passwordCalls != 2 {
		t.Fatalf("password prompts = %d, want 2", passwordCalls)
	}
}

func TestExportCommandRejectsOutAndFormatTogether(t *testing.T) {
	oldNewProfileService := newProfileService
	defer func() {
		newProfileService = oldNewProfileService
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "export", "myapp", "--out", "x.deadenv", "--format", "shell"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("error = %v, want conflict message", err)
	}
}

func TestExportCommandPasswordMismatch(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptExportPassword := promptExportPassword
	oldExportEncryptedProfile := exportEncryptedProfile
	defer func() {
		newProfileService = oldNewProfileService
		promptExportPassword = oldPromptExportPassword
		exportEncryptedProfile = oldExportEncryptedProfile
	}()

	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	calls := 0
	promptExportPassword = func(label string) (string, error) {
		calls++
		if calls == 1 {
			return "one", nil
		}
		return "two", nil
	}

	exportCalled := false
	exportEncryptedProfile = func(profile string, pairs []envPair.EnvPair, password, outPath string) error {
		exportCalled = true
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "export", "myapp", "--out", "x.deadenv"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "passwords do not match") {
		t.Fatalf("error = %v, want mismatch message", err)
	}
	if exportCalled {
		t.Fatal("exportEncryptedProfile should not be called on password mismatch")
	}
}

func TestExportCommandReturnsEncryptedExportError(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptExportPassword := promptExportPassword
	oldExportEncryptedProfile := exportEncryptedProfile
	defer func() {
		newProfileService = oldNewProfileService
		promptExportPassword = oldPromptExportPassword
		exportEncryptedProfile = oldExportEncryptedProfile
	}()

	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	promptExportPassword = func(label string) (string, error) {
		_ = label
		return "pass", nil
	}

	wantErr := errors.New("boom")
	exportEncryptedProfile = func(profile string, pairs []envPair.EnvPair, password, outPath string) error {
		_ = profile
		_ = pairs
		_ = password
		_ = outPath
		return wantErr
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "export", "myapp", "--out", "x.deadenv"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestDecryptErrorIsReturnedFromImportPath(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptImportPassword := promptImportPassword
	oldImportEncryptedProfile := importEncryptedProfile
	defer func() {
		newProfileService = oldNewProfileService
		promptImportPassword = oldPromptImportPassword
		importEncryptedProfile = oldImportEncryptedProfile
	}()

	newProfileService = func() (*profile.ProfileService, error) {
		return mustProfileService(t, keychain.NewFake()), nil
	}
	promptImportPassword = func(label string) (string, error) {
		_ = label
		return "pass", nil
	}
	importEncryptedProfile = func(path, password string) ([]envPair.EnvPair, string, error) {
		_ = path
		_ = password
		return nil, "", crypto.ErrDecryptFailed
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import", "backup.deadenv"})
	if !errors.Is(err, crypto.ErrDecryptFailed) {
		t.Fatalf("Run() error = %v, want ErrDecryptFailed", err)
	}
}
