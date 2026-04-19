//go:build darwin

package keychain

func New() (Store, error) {
	return nil, ErrNotImplemented
}
