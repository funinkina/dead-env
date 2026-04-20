//go:build darwin && cgo

package gokeychain

import goKeychain "github.com/keybase/go-keychain"

func Write(service, account, value string) error {
	item := goKeychain.NewGenericPassword(service, account, "", []byte(value), "")
	item.SetSynchronizable(goKeychain.SynchronizableNo)
	item.SetAccessible(goKeychain.AccessibleWhenUnlocked)

	if err := goKeychain.AddItem(item); err != nil {
		if err != goKeychain.ErrorDuplicateItem {
			return err
		}

		query := goKeychain.NewItem()
		query.SetSecClass(goKeychain.SecClassGenericPassword)
		query.SetService(service)
		query.SetAccount(account)

		update := goKeychain.NewItem()
		update.SetSecClass(goKeychain.SecClassGenericPassword)
		update.SetService(service)
		update.SetAccount(account)
		update.SetData([]byte(value))
		update.SetAccessible(goKeychain.AccessibleWhenUnlocked)

		return goKeychain.UpdateItem(query, update)
	}

	return nil
}

func Read(service, account string) ([]byte, error) {
	return goKeychain.GetGenericPassword(service, account, "", "")
}

func Delete(service, account string) error {
	item := goKeychain.NewItem()
	item.SetSecClass(goKeychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(account)

	return goKeychain.DeleteItem(item)
}

func List(service string) ([]string, error) {
	return goKeychain.GetGenericPasswordAccounts(service)
}

func IsItemNotFound(err error) bool {
	return err == goKeychain.ErrorItemNotFound
}

func IsAuthDenied(err error) bool {
	switch err {
	case goKeychain.ErrorUserCanceled, goKeychain.ErrorAuthFailed, goKeychain.ErrorInteractionNotAllowed:
		return true
	default:
		return false
	}
}
