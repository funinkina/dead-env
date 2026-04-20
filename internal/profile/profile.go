package profile

import (
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
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
}

func NewProfileService(store keychain.Store, recorder history.Recorder) *ProfileService {
	return &ProfileService{store, recorder}
}

// helper function
func getServiceName(profile string) string {
	return "deadenv/" + profile
}

func SetKey(profile, key, value string) error {
	service := getServiceName(profile)
	store.Write(service, key, value)
}

func hashValue(value string) string {
	for idx, char := range value {
		char = char + rune(idx)
	}
	return string(value)
}
