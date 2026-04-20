package profile

import (
	"deadenv/internal/fake_history"
	"deadenv/internal/history"
	"deadenv/internal/keychain"
)

type ProfileService struct {
	store    keychain.Store
	recorder history.Recorder
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
