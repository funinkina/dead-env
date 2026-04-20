package keychain

// Store defines the keychain operations required by deadenv.
type Store interface {
	// Write stores or overwrites a secret under the given service/account pair.
	Write(service, account, value string) error
	// Delete removes a secret for the given service/account pair.
	Delete(service, account string) error
	// Read retrieves a secret and may use prompt as the OS-native auth reason.
	Read(service, account, prompt string) (string, error)
	// List returns all accounts stored under a service.
	List(service string) ([]string, error)
}
