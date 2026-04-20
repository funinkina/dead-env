package keychain

import "sort"

type ReadCall struct {
	Service string
	Account string
	Prompt  string
}

type FakeStore struct {
	data      map[string]map[string]string // service -> account -> value
	Err       error
	ReadCalls []ReadCall
}

var _ Store = (*FakeStore)(nil)

func NewFake() *FakeStore {
	return &FakeStore{data: make(map[string]map[string]string)}
}

func (f *FakeStore) Write(service, account, value string) error {
	if f.Err != nil {
		return f.Err
	}

	if f.data[service] == nil {
		f.data[service] = make(map[string]string)
	}

	f.data[service][account] = value
	return nil
}

func (f *FakeStore) Read(service, account, prompt string) (string, error) {
	f.ReadCalls = append(f.ReadCalls, ReadCall{
		Service: service,
		Account: account,
		Prompt:  prompt,
	})

	if f.Err != nil {
		return "", f.Err
	}

	v, ok := f.data[service][account]

	if !ok {
		return "", ErrKeyNotFound
	}

	return v, nil
}

func (f *FakeStore) Delete(service, account string) error {
	if f.Err != nil {
		return f.Err
	}

	if f.data[service] == nil {
		return nil
	}

	delete(f.data[service], account)
	return nil
}

func (f *FakeStore) List(service string) ([]string, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	var keys []string

	for k := range f.data[service] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys, nil
}
