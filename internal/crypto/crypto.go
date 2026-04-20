package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"funinkina/deadenv/internal/envPair"

	"golang.org/x/crypto/argon2"
)

func Export(profile string, pairs []envPair.EnvPair, password, outPath string) error {
	if strings.TrimSpace(profile) == "" {
		return fmt.Errorf("profile cannot be empty")
	}

	if strings.TrimSpace(outPath) == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	salt, err := randomBytes(saltSize)
	if err != nil {
		return fmt.Errorf("generating export salt: %w", err)
	}
	defer zeroBytes(salt)

	nonce, err := randomBytes(nonceSize)
	if err != nil {
		return fmt.Errorf("generating export nonce: %w", err)
	}
	defer zeroBytes(nonce)

	passwordBytes := []byte(password)
	defer zeroBytes(passwordBytes)

	key := argon2.IDKey(passwordBytes, salt, kdfTime, kdfMemory, kdfThreads, kdfKeyLen)
	defer zeroBytes(key)

	plain, err := json.Marshal(pairs)
	if err != nil {
		return fmt.Errorf("encoding env pairs: %w", err)
	}
	defer zeroBytes(plain)

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("creating aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("creating gcm: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	envelope := exportEnvelope{
		Version:    currentVersion,
		Profile:    profile,
		CreatedAt:  time.Now().UTC(),
		KDF:        defaultKDFParams(salt),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}

	payload, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding export envelope: %w", err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(outPath, payload, 0o600); err != nil {
		return fmt.Errorf("writing export file %q: %w", outPath, err)
	}

	return nil
}

func Import(path, password string) ([]envPair.EnvPair, string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, "", fmt.Errorf("path cannot be empty")
	}

	if password == "" {
		return nil, "", fmt.Errorf("password cannot be empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading export file %q: %w", path, err)
	}

	var envelope exportEnvelope
	if err := json.Unmarshal(b, &envelope); err != nil {
		return nil, "", fmt.Errorf("%w: decoding export envelope: %v", ErrInvalidFormat, err)
	}

	if envelope.Version != currentVersion {
		return nil, "", fmt.Errorf("%w: got %d, want %d", ErrUnsupportedVersion, envelope.Version, currentVersion)
	}

	if envelope.KDF.Algorithm != kdfAlgorithmArgon2id {
		return nil, "", fmt.Errorf("%w: unsupported kdf algorithm %q", ErrInvalidFormat, envelope.KDF.Algorithm)
	}

	salt, err := decodeBase64Field("kdf.salt", envelope.KDF.Salt)
	if err != nil {
		return nil, "", err
	}
	defer zeroBytes(salt)

	nonce, err := decodeBase64Field("nonce", envelope.Nonce)
	if err != nil {
		return nil, "", err
	}
	defer zeroBytes(nonce)

	ciphertext, err := decodeBase64Field("ciphertext", envelope.Ciphertext)
	if err != nil {
		return nil, "", err
	}

	passwordBytes := []byte(password)
	defer zeroBytes(passwordBytes)

	key := argon2.IDKey(passwordBytes, salt, envelope.KDF.Time, envelope.KDF.Memory, envelope.KDF.Threads, kdfKeyLen)
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", fmt.Errorf("creating aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("creating gcm: %w", err)
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, "", ErrDecryptFailed
	}
	defer zeroBytes(plain)

	var pairs []envPair.EnvPair
	if err := json.Unmarshal(plain, &pairs); err != nil {
		return nil, "", fmt.Errorf("%w: decoding decrypted payload: %v", ErrInvalidFormat, err)
	}

	return pairs, envelope.Profile, nil
}

func defaultKDFParams(salt []byte) kdfParams {
	return kdfParams{
		Algorithm: kdfAlgorithmArgon2id,
		Salt:      base64.StdEncoding.EncodeToString(salt),
		Time:      kdfTime,
		Memory:    kdfMemory,
		Threads:   kdfThreads,
	}
}

func decodeBase64Field(name, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid %s encoding: %v", ErrInvalidFormat, name, err)
	}

	return decoded, nil
}

func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
