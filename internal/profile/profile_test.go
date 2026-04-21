package profile

import (
	"errors"
	"reflect"
	"testing"

	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/history"
	"github.com/funinkina/deadenv/internal/keychain"
)

func fixedHash(value string) (string, error) {
	return "hash:" + value, nil
}

func TestNewProfileServiceRejectsNilStore(t *testing.T) {
	_, err := NewProfileService(nil, history.NewNoopRecorder(), fixedHash)
	if !errors.Is(err, ErrNilStore) {
		t.Fatalf("NewProfileService() error = %v, want ErrNilStore", err)
	}
}

func TestNewProfileServiceNilRecorderFallsBackToNoop(t *testing.T) {
	store := keychain.NewFake()

	service, err := NewProfileService(store, nil, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.SetKey("myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("SetKey() error = %v", err)
	}

	got, err := store.Read("deadenv/myapp", "API_KEY", "")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != "secret" {
		t.Fatalf("Read() = %q, want %q", got, "secret")
	}
}

func TestSetKeyStoresValueAndRecordsHistory(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.SetKey("myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("SetKey() error = %v", err)
	}

	got, err := store.Read("deadenv/myapp", "API_KEY", "")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != "secret" {
		t.Fatalf("Read() = %q, want %q", got, "secret")
	}

	if len(recorder.Entries) != 1 {
		t.Fatalf("recorder entries = %d, want 1", len(recorder.Entries))
	}
	if recorder.Entries[0].Operation != history.OpSet {
		t.Fatalf("entry.Operation = %q, want %q", recorder.Entries[0].Operation, history.OpSet)
	}
}

func TestGetKeyUsesAccessPrompt(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	got, err := service.GetKey("myapp", "API_KEY")
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if got != "secret" {
		t.Fatalf("GetKey() = %q, want %q", got, "secret")
	}
	if len(store.ReadCalls) != 1 {
		t.Fatalf("ReadCalls = %d, want 1", len(store.ReadCalls))
	}
	wantPrompt := `deadenv wants to access profile "myapp"`
	if store.ReadCalls[0].Prompt != wantPrompt {
		t.Fatalf("Read prompt = %q, want %q", store.ReadCalls[0].Prompt, wantPrompt)
	}
}

func TestCreateRejectsEmptyPairs(t *testing.T) {
	store := keychain.NewFake()
	service, err := NewProfileService(store, history.NewNoopRecorder(), fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.Create("myapp", nil)
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("Create() error = %v, want ErrEmptyContent", err)
	}
}

func TestCreateStoresAllPairsAndRecordsHistory(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	pairs := []envPair.EnvPair{
		{Key: "API_KEY", Value: "secret"},
		{Key: "DATABASE_URL", Value: "postgres://localhost/app"},
	}

	if err := service.Create("myapp", pairs); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	keys, err := store.List("deadenv/myapp")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("List() length = %d, want 2", len(keys))
	}
	if len(recorder.Entries) != 2 {
		t.Fatalf("recorder entries = %d, want 2", len(recorder.Entries))
	}
}

func TestUnsetRemovesKeyAndRecordsHistory(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.UnsetKey("myapp", "API_KEY"); err != nil {
		t.Fatalf("UnsetKey() error = %v", err)
	}

	_, err = store.Read("deadenv/myapp", "API_KEY", "")
	if !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read() error = %v, want ErrKeyNotFound", err)
	}
	if len(recorder.Entries) != 1 {
		t.Fatalf("recorder entries = %d, want 1", len(recorder.Entries))
	}
	if recorder.Entries[0].Operation != history.OpUnset {
		t.Fatalf("entry.Operation = %q, want %q", recorder.Entries[0].Operation, history.OpUnset)
	}
}

func TestDeleteRecordsDeleteProfileOperation(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}

	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "DATABASE_URL", "postgres://localhost/app"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.Delete("myapp"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	keys, err := store.List("deadenv/myapp")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("List() = %v, want empty slice", keys)
	}

	if len(recorder.Entries) != 1 {
		t.Fatalf("recorder entries = %d, want 1", len(recorder.Entries))
	}

	entry := recorder.Entries[0]
	if entry.Profile != "myapp" {
		t.Fatalf("entry.Profile = %q, want %q", entry.Profile, "myapp")
	}
	if entry.Operation != history.OpDeleteProfile {
		t.Fatalf("entry.Operation = %q, want %q", entry.Operation, history.OpDeleteProfile)
	}
	if entry.Key != "" {
		t.Fatalf("entry.Key = %q, want empty string", entry.Key)
	}
}

func TestCopyUsesAccessPromptAndRecordsCopiedKeys(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/source", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := store.Write("deadenv/source", "DATABASE_URL", "postgres://localhost/app"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.Copy("source", "dest"); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	wantPrompt := `deadenv wants to access profile "source"`
	if len(store.ReadCalls) != 2 {
		t.Fatalf("ReadCalls = %d, want 2", len(store.ReadCalls))
	}
	for _, call := range store.ReadCalls {
		if call.Prompt != wantPrompt {
			t.Fatalf("Read prompt = %q, want %q", call.Prompt, wantPrompt)
		}
	}

	value, err := store.Read("deadenv/dest", "API_KEY", "")
	if err != nil {
		t.Fatalf("Read(dest API_KEY) error = %v", err)
	}
	if value != "secret" {
		t.Fatalf("Read(dest API_KEY) = %q, want %q", value, "secret")
	}

	if len(recorder.Entries) != 2 {
		t.Fatalf("recorder entries = %d, want 2", len(recorder.Entries))
	}
}

func TestRenameCopiesDestinationAndDeletesSource(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/source", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.Rename("source", "dest"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	value, err := store.Read("deadenv/dest", "API_KEY", "")
	if err != nil {
		t.Fatalf("Read(dest) error = %v", err)
	}
	if value != "secret" {
		t.Fatalf("Read(dest) = %q, want %q", value, "secret")
	}

	_, err = store.Read("deadenv/source", "API_KEY", "")
	if !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read(source) error = %v, want ErrKeyNotFound", err)
	}

	if len(recorder.Entries) != 2 {
		t.Fatalf("recorder entries = %d, want 2", len(recorder.Entries))
	}
	if recorder.Entries[0].Operation != history.OpSet {
		t.Fatalf("first entry operation = %q, want %q", recorder.Entries[0].Operation, history.OpSet)
	}
	if recorder.Entries[1].Operation != history.OpDeleteProfile {
		t.Fatalf("second entry operation = %q, want %q", recorder.Entries[1].Operation, history.OpDeleteProfile)
	}
}

func TestGetPairsReturnsSortedPairs(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/myapp", "Z_KEY", "3"); err != nil {
		t.Fatalf("Write(Z_KEY) error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "A_KEY", "1"); err != nil {
		t.Fatalf("Write(A_KEY) error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "M_KEY", "2"); err != nil {
		t.Fatalf("Write(M_KEY) error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	got, err := service.GetPairs("myapp")
	if err != nil {
		t.Fatalf("GetPairs() error = %v", err)
	}

	want := []envPair.EnvPair{
		{Key: "A_KEY", Value: "1"},
		{Key: "M_KEY", Value: "2"},
		{Key: "Z_KEY", Value: "3"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetPairs() = %#v, want %#v", got, want)
	}
}

func TestReplaceProfileRemovesMissingAndSetsIncoming(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/myapp", "A", "old-a"); err != nil {
		t.Fatalf("Write(A) error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "B", "old-b"); err != nil {
		t.Fatalf("Write(B) error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	err = service.ReplaceProfile("myapp", []envPair.EnvPair{
		{Key: "B", Value: "new-b"},
		{Key: "C", Value: "new-c"},
	})
	if err != nil {
		t.Fatalf("ReplaceProfile() error = %v", err)
	}

	if _, err := store.Read("deadenv/myapp", "A", ""); !errors.Is(err, keychain.ErrKeyNotFound) {
		t.Fatalf("Read(A) error = %v, want ErrKeyNotFound", err)
	}

	valueB, err := store.Read("deadenv/myapp", "B", "")
	if err != nil {
		t.Fatalf("Read(B) error = %v", err)
	}
	if valueB != "new-b" {
		t.Fatalf("Read(B) = %q, want %q", valueB, "new-b")
	}

	valueC, err := store.Read("deadenv/myapp", "C", "")
	if err != nil {
		t.Fatalf("Read(C) error = %v", err)
	}
	if valueC != "new-c" {
		t.Fatalf("Read(C) = %q, want %q", valueC, "new-c")
	}

	if len(recorder.Entries) != 3 {
		t.Fatalf("history entries = %d, want 3", len(recorder.Entries))
	}

	gotOps := []string{
		recorder.Entries[0].Operation + ":" + recorder.Entries[0].Key,
		recorder.Entries[1].Operation + ":" + recorder.Entries[1].Key,
		recorder.Entries[2].Operation + ":" + recorder.Entries[2].Key,
	}
	wantOps := []string{"unset:A", "set:B", "set:C"}
	if !reflect.DeepEqual(gotOps, wantOps) {
		t.Fatalf("history ops = %v, want %v", gotOps, wantOps)
	}
}

func TestReplaceProfileWithEmptyPairsClearsExistingKeys(t *testing.T) {
	store := keychain.NewFake()
	recorder := &history.FakeRecorder{}
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write(A) error = %v", err)
	}
	if err := store.Write("deadenv/myapp", "B", "2"); err != nil {
		t.Fatalf("Write(B) error = %v", err)
	}

	service, err := NewProfileService(store, recorder, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	if err := service.ReplaceProfile("myapp", nil); err != nil {
		t.Fatalf("ReplaceProfile() error = %v", err)
	}

	keys, err := store.List("deadenv/myapp")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("keys = %v, want empty", keys)
	}

	if len(recorder.Entries) != 2 {
		t.Fatalf("history entries = %d, want 2", len(recorder.Entries))
	}
	for _, entry := range recorder.Entries {
		if entry.Operation != history.OpUnset {
			t.Fatalf("history operation = %q, want %q", entry.Operation, history.OpUnset)
		}
	}
}
