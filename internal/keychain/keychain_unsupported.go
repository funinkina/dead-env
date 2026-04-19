//go:build !darwin && !linux && !windows

package keychain

func New() (Store, error) {
	return nil, ErrNotImplemented
}
