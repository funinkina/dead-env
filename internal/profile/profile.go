package profile

import (
	"fmt"
	"funinkina/deadenv/internal/envPair"
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
		recorder = history.NewNoopRecorder()
	}

	if hashValue == nil {
		if _, ok := recorder.(*history.NoopRecorder); ok {
			hashValue = noopHash
		} else {
			hashValue = history.HashValue
		}
	}
	return &ProfileService{
		store:     store,
		recorder:  recorder,
		hashValue: hashValue,
	}, nil
}

func getServiceName(profile string) string {
	return "deadenv/" + profile
}

func readPrompt(profile string) string {
	return fmt.Sprintf(`deadenv wants to access profile "%s"`, profile)
}

func noopHash(string) (string, error) {
	return "", nil
}

func (p *ProfileService) SetKey(profile, key, value string) error {
	if profile == "" {
		return ErrProfileNameEmpty
	}
	if key == "" {
		return ErrKeyEmpty
	}
	service := getServiceName(profile)
	err := p.store.Write(service, key, value)
	if err != nil {
		return fmt.Errorf("error writing key: %w", err)
	}
	hash, err := p.hashValue(value)
	if err != nil {
		return fmt.Errorf("error hashing value: %w", err)
	}
	err = p.recorder.Record(profile, history.OpSet, key, hash)
	if err != nil {
		return fmt.Errorf("error recording operation: %w", err)
	}
	return nil
}

func (p *ProfileService) UnsetKey(profile, key string) error {
	service := getServiceName(profile)
	err := p.store.Delete(service, key)
	if err != nil {
		return fmt.Errorf("error deleting key: %w", err)
	}
	// if key not exist code write
	err = p.recorder.Record(profile, history.OpUnset, key, "")
	if err != nil {
		return fmt.Errorf("error recording operation: %w", err)
	}
	return nil
}

func (p *ProfileService) GetKey(profile, key string) (string, error) {
	service := getServiceName(profile)
	prompt := readPrompt(profile)
	value, err := p.store.Read(service, key, prompt)
	if err != nil {
		return "", fmt.Errorf("error reading key: %w", err)
	}
	return value, nil
}

func (p *ProfileService) ListKeys(profile string) ([]string, error) {
	service := getServiceName(profile)
	keys, err := p.store.List(service)
	if err != nil {
		return nil, fmt.Errorf("error listing keys: %w", err)
	}
	return keys, nil
}

func (p *ProfileService) Create(profile string, pairs []envPair.EnvPair) error {
	if profile == "" {
		return ErrProfileNameEmpty
	}
	if len(pairs) == 0 {
		return ErrEmptyContent
	}
	service := getServiceName(profile)
	for _, pair := range pairs {
		if pair.Key == "" {
			return ErrKeyEmpty
		}
		err := p.store.Write(service, pair.Key, pair.Value)
		if err != nil {
			return fmt.Errorf("error writing key %s: %w", pair.Key, err)
		}

		hash, err := p.hashValue(pair.Value)
		if err != nil {
			return fmt.Errorf("error hashing value for key %s: %w", pair.Key, err)
		}

		err = p.recorder.Record(profile, history.OpSet, pair.Key, hash)
		if err != nil {
			return fmt.Errorf("error recording operation for key %s: %w", pair.Key, err)
		}
	}
	return nil
}

func (p *ProfileService) Delete(profile string) error {
	if profile == "" {
		return ErrProfileNameEmpty
	}
	service := getServiceName(profile)
	keys, err := p.store.List(service)
	if err != nil {
		return fmt.Errorf("error listing keys: %w", err)
	}
	for _, key := range keys {
		err = p.store.Delete(service, key)
		if err != nil {
			return fmt.Errorf("error deleting key: %w", err)
		}
	}
	err = p.recorder.Record(profile, history.OpDeleteProfile, "", "")
	if err != nil {
		return fmt.Errorf("error recording operation: %w", err)
	}
	return nil
}

func (p *ProfileService) Copy(srcProfile, dstProfile string) error {
	if srcProfile == "" || dstProfile == "" {
		return ErrProfileNameEmpty
	}
	srcService := getServiceName(srcProfile)
	dstService := getServiceName(dstProfile)
	prompt := readPrompt(srcProfile)
	keys, err := p.store.List(srcService)
	if err != nil {
		return fmt.Errorf("error listing keys: %w", err)
	}
	for _, key := range keys {
		value, err := p.store.Read(srcService, key, prompt)
		if err != nil {
			return fmt.Errorf("error reading key %s: %w", key, err)
		}
		err = p.store.Write(dstService, key, value)
		if err != nil {
			return fmt.Errorf("error copying key %s: %w", key, err)
		}

		hash, err := p.hashValue(value)
		if err != nil {
			return fmt.Errorf("error hashing value for key %s: %w", key, err)
		}

		err = p.recorder.Record(dstProfile, history.OpSet, key, hash)
		if err != nil {
			return fmt.Errorf("error recording operation for key %s: %w", key, err)
		}
	}
	return nil
}

func (p *ProfileService) Rename(srcProfile, dstProfile string) error {
	if srcProfile == "" || dstProfile == "" {
		return ErrProfileNameEmpty
	}
	err := p.Copy(srcProfile, dstProfile)
	if err != nil {
		return fmt.Errorf("error renaming profile: %w", err)
	}
	err = p.Delete(srcProfile)
	if err != nil {
		return fmt.Errorf("error renaming profile: %w", err)
	}
	return nil
}
