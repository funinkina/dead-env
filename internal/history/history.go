package history

type Recorder interface {
	Record(profile, operation, key, valueHash string) error
}

const (
	OpSet   = "set"
	OpUnset = "unset"
)
