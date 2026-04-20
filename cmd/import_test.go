package cmd

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/keychain"
	"funinkina/deadenv/internal/profile"
)

func TestImportCommandSuccessWithProfileOverride(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptImportPassword := promptImportPassword
	oldPromptImportConfirm := promptImportConfirm
	oldImportEncryptedProfile := importEncryptedProfile
	oldPrintImportSummary := printImportSummary
	defer func() {
		newProfileService = oldNewProfileService
		promptImportPassword = oldPromptImportPassword
		promptImportConfirm = oldPromptImportConfirm
		importEncryptedProfile = oldImportEncryptedProfile
		printImportSummary = oldPrintImportSummary
	}()

	store := keychain.NewFake()
	if err := store.Write("deadenv/target", "OLD", "stale"); err != nil {
		t.Fatalf("Write(OLD) error = %v", err)
	}
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	promptImportPassword = func(label string) (string, error) {
		if label == "" {
			t.Fatal("password prompt label is empty")
		}
		return "pass", nil
	}
	promptImportConfirm = func(message string) (bool, error) {
		if message == "" {
			t.Fatal("confirm message is empty")
		}
		return true, nil
	}

	importEncryptedProfile = func(path, password string) ([]envPair.EnvPair, string, error) {
		if path != "backup.deadenv" {
			t.Fatalf("path = %q, want %q", path, "backup.deadenv")
		}
		if password != "pass" {
			t.Fatalf("password = %q, want %q", password, "pass")
		}
		return []envPair.EnvPair{
			{Key: "A", Value: "1"},
			{Key: "B", Value: "2"},
		}, "source", nil
	}

	printImportSummary = func(_ io.Writer, pairs []envPair.EnvPair) error {
		if len(pairs) != 2 {
			t.Fatalf("summary pair count = %d, want 2", len(pairs))
		}
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import", "backup.deadenv", "--as", "target"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := store.Read("deadenv/target", "OLD", ""); !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read(OLD) error = %v, want ErrKeyNotFound", err)
	}

	valueA, err := store.Read("deadenv/target", "A", "")
	if err != nil {
		t.Fatalf("Read(A) error = %v", err)
	}
	if valueA != "1" {
		t.Fatalf("Read(A) = %q, want %q", valueA, "1")
	}

}

func TestImportCommandCancelled(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptImportPassword := promptImportPassword
	oldPromptImportConfirm := promptImportConfirm
	oldImportEncryptedProfile := importEncryptedProfile
	oldPrintImportSummary := printImportSummary
	defer func() {
		newProfileService = oldNewProfileService
		promptImportPassword = oldPromptImportPassword
		promptImportConfirm = oldPromptImportConfirm
		importEncryptedProfile = oldImportEncryptedProfile
		printImportSummary = oldPrintImportSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	promptImportPassword = func(label string) (string, error) {
		_ = label
		return "pass", nil
	}
	promptImportConfirm = func(message string) (bool, error) {
		_ = message
		return false, nil
	}
	importEncryptedProfile = func(path, password string) ([]envPair.EnvPair, string, error) {
		_ = path
		_ = password
		return []envPair.EnvPair{{Key: "A", Value: "1"}}, "myapp", nil
	}
	printImportSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import", "backup.deadenv"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := store.Read("deadenv/myapp", "A", ""); !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read(A) error = %v, want ErrKeyNotFound", err)
	}
}

func TestImportCommandRequiresFilePath(t *testing.T) {
	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
}

func TestImportCommandRequiresProfileNameWhenMissingInEnvelope(t *testing.T) {
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
		return []envPair.EnvPair{{Key: "A", Value: "1"}}, "", nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import", "backup.deadenv"})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "--as") {
		t.Fatalf("error = %v, want --as guidance", err)
	}
}

func TestImportCommandUsesEnvelopeProfileWhenOverrideMissing(t *testing.T) {
	oldNewProfileService := newProfileService
	oldPromptImportPassword := promptImportPassword
	oldPromptImportConfirm := promptImportConfirm
	oldImportEncryptedProfile := importEncryptedProfile
	oldPrintImportSummary := printImportSummary
	defer func() {
		newProfileService = oldNewProfileService
		promptImportPassword = oldPromptImportPassword
		promptImportConfirm = oldPromptImportConfirm
		importEncryptedProfile = oldImportEncryptedProfile
		printImportSummary = oldPrintImportSummary
	}()

	store := keychain.NewFake()
	service := mustProfileService(t, store)
	newProfileService = func() (*profile.ProfileService, error) {
		return service, nil
	}

	promptImportPassword = func(label string) (string, error) {
		_ = label
		return "pass", nil
	}
	promptImportConfirm = func(message string) (bool, error) {
		_ = message
		return true, nil
	}
	importEncryptedProfile = func(path, password string) ([]envPair.EnvPair, string, error) {
		_ = path
		_ = password
		return []envPair.EnvPair{{Key: "A", Value: "1"}}, "from-file", nil
	}
	printImportSummary = func(_ io.Writer, _ []envPair.EnvPair) error {
		return nil
	}

	root := NewRootCommand()
	err := root.Run(context.Background(), []string{"deadenv", "import", "backup.deadenv"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	value, err := store.Read("deadenv/from-file", "A", "")
	if err != nil {
		t.Fatalf("Read(A) error = %v", err)
	}
	if value != "1" {
		t.Fatalf("Read(A) = %q, want %q", value, "1")
	}
}
