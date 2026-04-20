package crypto

import "time"

const (
	currentVersion = 1

	kdfAlgorithmArgon2id = "argon2id"
	kdfTime              = uint32(2)
	kdfMemory            = uint32(128 * 1024)
	kdfThreads           = uint8(4)
	kdfKeyLen            = uint32(32)

	saltSize  = 32
	nonceSize = 12
)

type exportEnvelope struct {
	Version    int       `json:"version"`
	Profile    string    `json:"profile"`
	CreatedAt  time.Time `json:"created_at"`
	KDF        kdfParams `json:"kdf"`
	Nonce      string    `json:"nonce"`
	Ciphertext string    `json:"ciphertext"`
}

type kdfParams struct {
	Algorithm string `json:"algorithm"`
	Salt      string `json:"salt"`
	Time      uint32 `json:"time"`
	Memory    uint32 `json:"memory"`
	Threads   uint8  `json:"threads"`
}
