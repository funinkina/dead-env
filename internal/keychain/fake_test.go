package keychain

import (
	"errors"
	"reflect"
	"testing"
)

func TestFakeStoreReadMissingKeyReturnsErrKeyNotFound(t *testing.T) {
	store := NewFake()

	_, err := store.Read("deadenv/myapp", "API_KEY", "prompt")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Read() error = %v, want ErrKeyNotFound", err)
	}
}

func TestFakeStoreReadCapturesPrompt(t *testing.T) {
	store := NewFake()
	if err := store.Write("deadenv/myapp", "API_KEY", "secret"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	_, err := store.Read("deadenv/myapp", "API_KEY", "access reason")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(store.ReadCalls) != 1 {
		t.Fatalf("ReadCalls = %d, want 1", len(store.ReadCalls))
	}
	if store.ReadCalls[0].Prompt != "access reason" {
		t.Fatalf("Read prompt = %q, want %q", store.ReadCalls[0].Prompt, "access reason")
	}
}

func TestFakeStoreListProfilesReturnsSortedUniqueProfiles(t *testing.T) {
	store := NewFake()
	if err := store.Write("deadenv/beta", "API_KEY", "x"); err != nil {
		t.Fatalf("Write(beta) error = %v", err)
	}
	if err := store.Write("deadenv/alpha", "TOKEN", "y"); err != nil {
		t.Fatalf("Write(alpha) error = %v", err)
	}
	if err := store.Write("deadenv/alpha", "DATABASE_URL", "z"); err != nil {
		t.Fatalf("Write(alpha second key) error = %v", err)
	}
	if err := store.Write("not-deadenv-service", "IGNORED", "n/a"); err != nil {
		t.Fatalf("Write(invalid service) error = %v", err)
	}

	profiles, err := store.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	want := []string{"alpha", "beta"}
	if !reflect.DeepEqual(profiles, want) {
		t.Fatalf("ListProfiles() = %v, want %v", profiles, want)
	}
}
