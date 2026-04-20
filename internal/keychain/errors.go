package keychain

import "errors"

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrAuthDenied     = errors.New("auth denied")
	ErrNotImplemented = errors.New("keychain backend not implemented")
)
