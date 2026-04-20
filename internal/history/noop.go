package history

// NoopRecorder is a safe fallback when git is unavailable
type NoopRecorder struct{}

func NewNoopRecorder() *NoopRecorder {
	return &NoopRecorder{}
}

func (n *NoopRecorder) Record(profile, operation, key, valueHash string) error {
	return nil
}
