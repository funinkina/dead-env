package profile

import (
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
)

type ProfileService struct {
	store    keychain.Store
	recorder history.Recorder
}

func NewProfileService(store keychain.Store, recorder history.Recorder) *ProfileService {
	
}

// helper function
func getServiceName(profile string) string {
	return "deadenv/" + profile
}
