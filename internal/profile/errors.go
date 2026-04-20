package profile

import "errors"

var (
	ErrNilStore         = errors.New("profile service requires a keychain store")
	ErrProfileNameEmpty = errors.New("profile name cannot be empty")
	ErrKeyEmpty         = errors.New("key cannot be empty")
	ErrEmptyContent     = errors.New("content cannot be empty")
)
