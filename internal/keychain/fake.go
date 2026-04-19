package keychain

type FakeStore struct {
	data map[string]map[string]string // service -> account -> value
	Err  error
}

func NewFake() *FakeStore {
	return &FakeStore{data: make(map[string]map[string]string)}
}
