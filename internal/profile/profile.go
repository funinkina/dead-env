package profile

import (
	"deadenv/internal/fake_history"
	"deadenv/internal/history"
	"deadenv/internal/keychain"
)

type HashFunc func(value string) (string, error)

type ProfileService struct {
	store     keychain.Store
	recorder  history.Recorder
	hashValue HashFunc
}

func NewProfileService(
	store keychain.Store,
	recorder history.Recorder,
	hashValue HashFunc,
) (*ProfileService, error) {
	if store == nil {
		return nil, ErrNilStore
	}

	if recorder == nil {
		// return nil,
	}

	if hashValue == nil {
		hashValue = history.HashValue
	}
	return &ProfileService{
		store:     store,
		recorder:  recorder,
		hashValue: hashValue,
	}, nil
}

// helper function
func getServiceName(profile string) string {
	return "deadenv/" + profile
}

func SetKey(profile, key, value string, store keychain.Store) error {
	service := getServiceName(profile)
	err := store.Write(service, key, value)
	if err != nil {
		return err
	}
	hash := hashValue(value)
	fake_history.Record(profile, "set", key, hash)
	return nil

}

func hashValue(value string) string {
	for idx, char := range value {
		char = char + rune(idx)
	}
	return string(value)
}
