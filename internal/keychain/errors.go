package keychain

import "errors"

var (
	ErrKeyNotFound      = errors.New("Error! Key not found.")
	ErrNotImplemented   = errors.New(("Not implemented yet..."))
	ErrProfileNameEmpty = errors.New("profile name cannot be empty")
	ErrKeyEmpty         = errors.New("key cannot be empty")
	ErrEmptyContent     = errors.New("content cannot be empty")
	ErrKeyWrite         = errors.New("error writing key")
	ErrListKeys         = errors.New("error listing keys")
	ErrRecordOperation  = errors.New("error recording operation")
)
