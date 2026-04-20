package crypto

import "errors"

var (
	ErrDecryptFailed      = errors.New("decryption failed")
	ErrInvalidFormat      = errors.New("invalid export format")
	ErrUnsupportedVersion = errors.New("unsupported export format version")
)
