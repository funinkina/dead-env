package profile

import (
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

func (p *ProfileService) SetKey(profile, key, value string) error {
	service := getServiceName(profile)
	err := p.store.Write(service, key, value)
	if err != nil {
		return err
	}
	hash, err := p.hashValue(value)
	if err != nil {
		return err
	}
	err = p.recorder.Record(profile, "set", key, hash)
	if err != nil {
		return err
	}
	return nil

}
