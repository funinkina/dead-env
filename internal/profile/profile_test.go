package profile

import (
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
	"testing"
)

type Entry struct {
	profile string
	op      string
	key     string
	hash    string
}

type FakeRecorder struct {
	entries []Entry
}

func (r *FakeRecorder) Record(profile, op, key, hash string) error {
	r.entries = append(r.entries, Entry{profile, op, key, hash})
	return nil
}

func setup() (*ProfileService, keychain.Store, *FakeRecorder) {
	fakestore := keychain.NewFake()
	fakerecorder := &FakeRecorder{}

	profileservice, _ := NewProfileService(fakestore, fakerecorder, nil)
	// ⚠️ (optional improvement): you may want to check error instead of ignoring it

	return profileservice, fakestore, fakerecorder
}

func TestSetKey(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid input",
			profile: "myapp",
			key:     "API_KEY",
			value:   "123",
			wantErr: false,
		},
		{
			name:    "empty profile",
			profile: "",
			key:     "API_KEY",
			value:   "123",
			wantErr: true,
		},
		{
			name:    "empty key",
			profile: "myapp",
			key:     "",
			value:   "123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileservice, fakestore, fakerecorder := setup()

			err := profileservice.SetKey(tt.profile, tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			service := getServiceName(tt.profile)

			value, err := fakestore.Read(service, tt.key, "")
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			if value != tt.value {
				t.Fatalf("expected value %s, got '%s'", tt.value, value)
			}

			if len(fakerecorder.entries) != 1 {
				t.Fatalf("expected 1 entry in recorder, got %d", len(fakerecorder.entries))
			}

			entry := fakerecorder.entries[0]

			if entry.profile != tt.profile || entry.op != history.OpSet || entry.key != tt.key {
				t.Fatalf("unexpected recorder entry: %+v", entry)
			}

			if entry.hash == "" {
				t.Fatalf("expected hash to be non-empty")
			}
		})
	}
}
