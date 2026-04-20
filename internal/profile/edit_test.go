package profile

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
)

type flakyStore struct {
	*keychain.FakeStore
	writeFailures  map[string]error
	deleteFailures map[string]error
}

func (s *flakyStore) Write(service, account, value string) error {
	if err, ok := s.writeFailures[account]; ok {
		return err
	}

	return s.FakeStore.Write(service, account, value)
}

func (s *flakyStore) Delete(service, account string) error {
	if err, ok := s.deleteFailures[account]; ok {
		return err
	}

	return s.FakeStore.Delete(service, account)
}

func TestEditNoChangesReturnsErrNoChanges(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Edit("myapp", EditOptions{
		Yes: true,
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			if !strings.Contains(initialContent, "A=1") {
				t.Fatalf("initial editor content does not include expected key/value: %q", initialContent)
			}

			return []envPair.EnvPair{{Key: "A", Value: "1"}}, nil
		},
	})

	if !errors.Is(err, ErrNoChanges) {
		t.Fatalf("Edit() error = %v, want ErrNoChanges", err)
	}

	if len(recorder.Entries) != 0 {
		t.Fatalf("history entries = %d, want 0", len(recorder.Entries))
	}
}

func TestEditModifySingleKey(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "B", "2"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	var out bytes.Buffer
	err = service.Edit("myapp", EditOptions{
		Yes: true,
		Out: &out,
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			_ = initialContent
			return []envPair.EnvPair{
				{Key: "A", Value: "10"},
				{Key: "B", Value: "2"},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	gotA, err := store.Read("deadenv/myapp", "A", "")
	if err != nil {
		t.Fatalf("Read(A) error = %v", err)
	}
	if gotA != "10" {
		t.Fatalf("A value = %q, want %q", gotA, "10")
	}

	gotB, err := store.Read("deadenv/myapp", "B", "")
	if err != nil {
		t.Fatalf("Read(B) error = %v", err)
	}
	if gotB != "2" {
		t.Fatalf("B value = %q, want %q", gotB, "2")
	}

	if len(recorder.Entries) != 1 {
		t.Fatalf("history entries = %d, want 1", len(recorder.Entries))
	}
	if recorder.Entries[0].Operation != history.OpSet || recorder.Entries[0].Key != "A" {
		t.Fatalf("history entry = %+v, want set A", recorder.Entries[0])
	}

	summary := out.String()
	if !strings.Contains(summary, "[modified] A") {
		t.Fatalf("summary does not include modified key label: %q", summary)
	}
	if strings.Contains(summary, "10") {
		t.Fatalf("summary leaked value: %q", summary)
	}
}

func TestEditRenameTreatsAsSetAndUnset(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "OLD_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Edit("myapp", EditOptions{
		Yes: true,
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			_ = initialContent
			return []envPair.EnvPair{{Key: "NEW_KEY", Value: "secret"}}, nil
		},
	})
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}

	if _, err := store.Read("deadenv/myapp", "OLD_KEY", ""); !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("OLD_KEY should be removed, got err=%v", err)
	}

	value, err := store.Read("deadenv/myapp", "NEW_KEY", "")
	if err != nil {
		t.Fatalf("Read(NEW_KEY) error = %v", err)
	}
	if value != "secret" {
		t.Fatalf("NEW_KEY value = %q, want %q", value, "secret")
	}

	if len(recorder.Entries) != 2 {
		t.Fatalf("history entries = %d, want 2", len(recorder.Entries))
	}
}

func TestEditEmptyContentReturnsErrEmptyContent(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, history.NewNoopRecorder(), fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Edit("myapp", EditOptions{
		Yes: true,
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			_ = initialContent
			return nil, nil
		},
	})

	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("Edit() error = %v, want ErrEmptyContent", err)
	}
}

func TestEditCancelledByUser(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Edit("myapp", EditOptions{
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			_ = initialContent
			return []envPair.EnvPair{{Key: "A", Value: "2"}}, nil
		},
		Confirm: func(message string) (bool, error) {
			_ = message
			return false, nil
		},
	})

	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("Edit() error = %v, want ErrCancelled", err)
	}

	value, readErr := store.Read("deadenv/myapp", "A", "")
	if readErr != nil {
		t.Fatalf("Read(A) error = %v", readErr)
	}
	if value != "1" {
		t.Fatalf("A value = %q, want %q", value, "1")
	}

	if len(recorder.Entries) != 0 {
		t.Fatalf("history entries = %d, want 0", len(recorder.Entries))
	}
}

func TestEditReturnsPartialApplyError(t *testing.T) {
	base := keychain.NewFake()
	if err := base.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := base.Write("deadenv/myapp", "B", "2"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	store := &flakyStore{
		FakeStore:      base,
		writeFailures:  map[string]error{"B": errors.New("boom")},
		deleteFailures: map[string]error{},
	}

	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Edit("myapp", EditOptions{
		Yes: true,
		FromEditor: func(initialContent string) ([]envPair.EnvPair, error) {
			_ = initialContent
			return []envPair.EnvPair{
				{Key: "A", Value: "10"},
				{Key: "B", Value: "20"},
			}, nil
		},
	})

	if !errors.Is(err, ErrApplyChanges) {
		t.Fatalf("Edit() error = %v, want ErrApplyChanges", err)
	}

	partial, ok := err.(*PartialApplyError)
	if !ok {
		t.Fatalf("Edit() error type = %T, want *PartialApplyError", err)
	}

	wantSucceeded := []string{"A"}
	if !reflect.DeepEqual(partial.Succeeded, wantSucceeded) {
		t.Fatalf("Succeeded = %v, want %v", partial.Succeeded, wantSucceeded)
	}

	if len(partial.Failed) != 1 || partial.Failed["B"] == nil {
		t.Fatalf("Failed = %v, want key B failure", partial.Failed)
	}

	valueA, err := base.Read("deadenv/myapp", "A", "")
	if err != nil {
		t.Fatalf("Read(A) error = %v", err)
	}
	if valueA != "10" {
		t.Fatalf("A value = %q, want %q", valueA, "10")
	}

	valueB, err := base.Read("deadenv/myapp", "B", "")
	if err != nil {
		t.Fatalf("Read(B) error = %v", err)
	}
	if valueB != "2" {
		t.Fatalf("B value = %q, want %q", valueB, "2")
	}

	if len(recorder.Entries) != 1 {
		t.Fatalf("history entries = %d, want 1", len(recorder.Entries))
	}
	if recorder.Entries[0].Key != "A" || recorder.Entries[0].Operation != history.OpSet {
		t.Fatalf("history entry = %+v, want set A", recorder.Entries[0])
	}
}
