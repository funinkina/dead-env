package profile

import (
	"funinkina/deadenv/internal/keychain" // ✅ FIX: import keychain to use FakeStore
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
	// ✅ FIX: return keychain.Store instead of *FakeStore (interface-based)

	fakestore := keychain.NewFake()
	// ✅ FIX: correct constructor name (exported → NewFake, not newFake)

	fakerecorder := &FakeRecorder{}

	profileservice, _ := NewProfileService(fakestore, fakerecorder, nil)
	// ✅ FIX: must call constructor, not struct literal
	// also fixes hashValue being nil

	return profileservice, fakestore, fakerecorder
}

func TestSetKey(t *testing.T) {
	tests := []struct{
        name string
        input string 
        wantErr bool
    } {
		{
			name: "valid input",
			profile: "myapp",
			key: "API_KEY",
			value: "123",
			wantErr: false
		},
		{
			name: "empty profile",
			profile: "",
			key: "API_KEY",
			value: "123",
			wantErr: true
		},
		{
			name: "empty key",
			profile: "myapp",
			key: "",
			value: "123",
			wantErr: true
		}
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			profileservice, fakestore, fakerecorder := setup()
			err := profileservice.SetKey(tt.profile, tt.key, tt.value)
			if err != nil {
				t.Fatalf("SetKey failed: %v", err)
			}
			service := getServiceName(tt.profile)
			value, err := fakestore.Read(service, tt.key, "")
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			if value != tt.value {
				t.Fatalf("expected value %w, got '%s'", tt.value, value)
			}

			if len(fakerecorder.entries) != 1 {
				t.Fatalf("expected 1 entry in recorder, got %d", len(fakerecorder.entries))
			}

			entry := fakerecorder.entries[0]

			if entry.profile != tt.profile || entry.op != "set" || entry.key != tt.key {
				t.Fatalf("unexpected recorder entry: %+v", entry)
			}

			if entry.hash == "" {
				t.Fatalf("expected hash to be non-empty")
			}
		})
	}
}
