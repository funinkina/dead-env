package keychain

import (
	"errors"
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
