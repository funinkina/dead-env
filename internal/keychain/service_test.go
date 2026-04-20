package keychain

import (
	"errors"
	"testing"
)

func TestProfileFromService(t *testing.T) {
	tests := []struct {
		name    string
		service string
		want    string
		wantErr bool
	}{
		{name: "valid", service: "deadenv/myapp-dev", want: "myapp-dev"},
		{name: "missing prefix", service: "myapp-dev", wantErr: true},
		{name: "uppercase profile", service: "deadenv/MyApp", wantErr: true},
		{name: "empty profile", service: "deadenv/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := profileFromService(tt.service)
			if (err != nil) != tt.wantErr {
				t.Fatalf("profileFromService() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("profileFromService() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetName(t *testing.T) {
	got, err := targetName("deadenv/myapp", "API_KEY")
	if err != nil {
		t.Fatalf("targetName() error = %v", err)
	}
	if got != "deadenv/myapp/API_KEY" {
		t.Fatalf("targetName() = %q, want %q", got, "deadenv/myapp/API_KEY")
	}
}

func TestProfileFromServiceReturnsErrInvalidService(t *testing.T) {
	_, err := profileFromService("myapp")
	if !errors.Is(err, ErrInvalidService) {
		t.Fatalf("profileFromService() error = %v, want ErrInvalidService", err)
	}
}

func TestTargetNameReturnsErrInvalidAccount(t *testing.T) {
	_, err := targetName("deadenv/myapp", "   ")
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("targetName() error = %v, want ErrInvalidAccount", err)
	}
}
