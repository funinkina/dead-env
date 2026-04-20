package history

import "time"

type FakeRecorder struct {
	Err     error
	Entries []HistoryEntry
}

func (f *FakeRecorder) Record(profile, operation, key, valueHash string) error {
	if f.Err != nil {
		return f.Err
	}

	f.Entries = append(f.Entries, HistoryEntry{
		Profile:   profile,
		Operation: operation,
		Key:       key,
		ValueHash: valueHash,
		Timestamp: time.Now().UTC(),
	})

	return nil
}
