//go:build linux

package keychain

import (
	"errors"
	"testing"
)

func TestNewReturnsLinuxStore(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	linuxStore, ok := store.(*linuxStore)
	if !ok {
		t.Fatalf("New() store type = %T, want *linuxStore", store)
	}
	if linuxStore.newService == nil {
		t.Fatal("linuxStore.newService is nil")
	}
}

func TestLinuxSecretAttributes(t *testing.T) {
	attrs := linuxSecretAttributes("myapp", "API_KEY")

	if got := attrs[linuxSecretAttributeApplication]; got != linuxSecretApplicationName {
		t.Fatalf("application attribute = %q, want %q", got, linuxSecretApplicationName)
	}
	if got := attrs[linuxSecretAttributeProfile]; got != "myapp" {
		t.Fatalf("profile attribute = %q, want %q", got, "myapp")
	}
	if got := attrs[linuxSecretAttributeKey]; got != "API_KEY" {
		t.Fatalf("key attribute = %q, want %q", got, "API_KEY")
	}
}

func TestLinuxSecretServiceErrorMapsAuthDenied(t *testing.T) {
	err := linuxSecretServiceError("reading secret", errors.New("prompt dismissed"))
	if !errors.Is(err, ErrAuthDenied) {
		t.Fatalf("linuxSecretServiceError() error = %v, want ErrAuthDenied", err)
	}
}

func TestLinuxSecretServiceErrorMapsNotImplemented(t *testing.T) {
	err := linuxSecretServiceError("opening secret service", errors.New("failed to open dbus connection: no bus"))
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("linuxSecretServiceError() error = %v, want ErrNotImplemented", err)
	}
}

func TestLinuxSecretLabel(t *testing.T) {
	got := linuxSecretLabel("myapp", "API_KEY")
	if got != "deadenv myapp:API_KEY" {
		t.Fatalf("linuxSecretLabel() = %q, want %q", got, "deadenv myapp:API_KEY")
	}
}
