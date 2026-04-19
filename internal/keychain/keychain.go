package keychain

type Store interface {
	Write(service, account, value string) error
	Delete(service, account string) error
	Read(service, account, prompt string) (string, error)
	List(service string) ([]string, error)
}

func New() Store

// Platform constructor
