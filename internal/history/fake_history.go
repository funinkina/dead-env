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

func (f *FakeRecorder) Log(profile, key string) ([]HistoryEntry, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	entries := make([]HistoryEntry, 0, len(f.Entries))
	for _, entry := range f.Entries {
		if entry.Profile != profile {
			continue
		}
		if key != "" && entry.Key != key {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
