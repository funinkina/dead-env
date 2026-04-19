//go:build windows

package keychain

func New() (Store, error) {
	return nil, ErrNotImplemented
}
