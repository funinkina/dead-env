package profile

import (
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
)

type ProfileService struct {
	store    keychain.Store
	recorder history.Recorder
}
