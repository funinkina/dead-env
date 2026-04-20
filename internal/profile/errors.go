package profile

import "errors"

var (
	ErrNilStore = errors.New("profile service requires a keychain store")
)
