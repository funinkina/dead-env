package history

type Recorder interface {
	Record(profile, operation, key, valueHash string) error
}
