//go:build !darwin || !cgo

package gokeychain

import "errors"

var errUnsupportedPlatform = errors.New("go-keychain wrapper is only available on darwin with cgo enabled")

func Write(service, account, value string) error {
	_ = service
	_ = account
	_ = value
	return errUnsupportedPlatform
}

func Read(service, account string) ([]byte, error) {
	_ = service
	_ = account
	return nil, errUnsupportedPlatform
}

func Delete(service, account string) error {
	_ = service
	_ = account
	return errUnsupportedPlatform
}

func List(service string) ([]string, error) {
	_ = service
	return nil, errUnsupportedPlatform
}

func IsItemNotFound(err error) bool {
	_ = err
	return false
}

func IsAuthDenied(err error) bool {
	_ = err
	return false
}
