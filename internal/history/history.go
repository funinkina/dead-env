package history

type Recorder interface {
	Record(profile, operation, key, valueHash string) error
	Log(profile, key string) ([]HistoryEntry, error)
}

const (
	OpSet           = "set"
	OpUnset         = "unset"
	OpDeleteProfile = "delete-profile"
)
