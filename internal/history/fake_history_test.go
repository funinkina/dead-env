package history

import "testing"

func TestFakeRecorderLogFiltersByProfileAndKey(t *testing.T) {
	recorder := &FakeRecorder{}

	if err := recorder.Record("myapp", OpSet, "API_KEY", "hash:one"); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := recorder.Record("myapp", OpUnset, "API_KEY", ""); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := recorder.Record("other", OpSet, "API_KEY", "hash:other"); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	entries, err := recorder.Log("myapp", "API_KEY")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Log() len = %d, want 2", len(entries))
	}
	for _, entry := range entries {
		if entry.Profile != "myapp" || entry.Key != "API_KEY" {
			t.Fatalf("unexpected entry: %+v", entry)
		}
	}
}
