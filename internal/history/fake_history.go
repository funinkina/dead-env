package history

type FakeRecorder struct {
	Entries []HistoryEntry
}

func (f *FakeRecorder) Record(profile, operation, key, valueHash string) error {
	f.Entries = append(f.Entries, HistoryEntry{profile, operation, key, valueHash})
	return nil
}
