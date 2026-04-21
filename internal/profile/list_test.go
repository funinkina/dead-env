package profile

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/funinkina/deadenv/internal/history"
	"github.com/funinkina/deadenv/internal/keychain"
)

type storeWithoutProfileLister struct {
	data map[string]map[string]string
	err  error
}

func newStoreWithoutProfileLister() *storeWithoutProfileLister {
	return &storeWithoutProfileLister{data: make(map[string]map[string]string)}
}

func (s *storeWithoutProfileLister) Write(service, account, value string) error {
	if s.err != nil {
		return s.err
	}
	if s.data[service] == nil {
		s.data[service] = make(map[string]string)
	}
	s.data[service][account] = value
	return nil
}

func (s *storeWithoutProfileLister) Read(service, account, prompt string) (string, error) {
	_ = prompt
	if s.err != nil {
		return "", s.err
	}
	value, ok := s.data[service][account]
	if !ok {
		return "", keychain.ErrKeyNotFound
	}
	return value, nil
}

func (s *storeWithoutProfileLister) Delete(service, account string) error {
	if s.err != nil {
		return s.err
	}
	if s.data[service] == nil {
		return nil
	}
	delete(s.data[service], account)
	return nil
}

func (s *storeWithoutProfileLister) List(service string) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	keys := make([]string, 0, len(s.data[service]))
	for key := range s.data[service] {
		keys = append(keys, key)
	}
	return keys, nil
}

func TestListProfilesUsesStoreProfileLister(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/beta", "API_KEY", "x"); err != nil {
		t.Fatalf("Write(beta) error = %v", err)
	}
	if err := store.Write("deadenv/alpha", "TOKEN", "y"); err != nil {
		t.Fatalf("Write(alpha) error = %v", err)
	}

	service, err := NewProfileService(store, &history.FakeRecorder{}, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	profiles, err := service.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(profiles, want) {
		t.Fatalf("ListProfiles() = %v, want %v", profiles, want)
	}
}

func TestListProfilesFallsBackToHistoryWhenStoreHasNoLister(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	historyDir := filepath.Join(base, "deadenv", "history")
	if err := os.MkdirAll(historyDir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "myapp.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(myapp.json) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "deleted.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(deleted.json) error = %v", err)
	}

	store := newStoreWithoutProfileLister()
	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write(myapp) error = %v", err)
	}

	service, err := NewProfileService(store, &history.FakeRecorder{}, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	profiles, err := service.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	want := []string{"myapp"}
	if !reflect.DeepEqual(profiles, want) {
		t.Fatalf("ListProfiles() = %v, want %v", profiles, want)
	}
}

func TestListProfilesFromHistoryMissingDirReturnsEmpty(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	store := newStoreWithoutProfileLister()
	service, err := NewProfileService(store, &history.FakeRecorder{}, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	profiles, err := service.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("ListProfiles() = %v, want empty", profiles)
	}
}

func TestListProfilesPropagatesStoreListerError(t *testing.T) {
	store := keychain.NewFake()
	store.Err = errors.New("boom")

	service, err := NewProfileService(store, &history.FakeRecorder{}, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	_, err = service.ListProfiles()
	if err == nil {
		t.Fatal("ListProfiles() error = nil, want non-nil")
	}
}
