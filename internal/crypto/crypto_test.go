package crypto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"funinkina/deadenv/internal/envPair"
)

func TestExportImportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	pairs := []envPair.EnvPair{
		{Key: "API_KEY", Value: "secret=="},
		{Key: "DATABASE_URL", Value: "postgres://localhost/app"},
	}

	if err := Export("myapp", pairs, "password123", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	gotPairs, gotProfile, err := Import(path, "password123")
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if gotProfile != "myapp" {
		t.Fatalf("profile = %q, want %q", gotProfile, "myapp")
	}

	if !reflect.DeepEqual(gotPairs, pairs) {
		t.Fatalf("pairs = %#v, want %#v", gotPairs, pairs)
	}
}

func TestImportWrongPasswordReturnsErrDecryptFailed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "good-pass", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	_, _, err := Import(path, "bad-pass")
	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("Import() error = %v, want ErrDecryptFailed", err)
	}
}

func TestImportTamperedCiphertextReturnsErrDecryptFailed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "password", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	env := mustReadEnvelope(t, path)
	ciphertext := mustDecodeBase64(t, env.Ciphertext)
	ciphertext[len(ciphertext)-1] ^= 0xFF
	env.Ciphertext = base64.StdEncoding.EncodeToString(ciphertext)
	mustWriteEnvelope(t, path, env)

	_, _, err := Import(path, "password")
	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("Import() error = %v, want ErrDecryptFailed", err)
	}
}

func TestImportTamperedNonceReturnsErrDecryptFailed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "password", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	env := mustReadEnvelope(t, path)
	nonce := mustDecodeBase64(t, env.Nonce)
	nonce[0] ^= 0x01
	env.Nonce = base64.StdEncoding.EncodeToString(nonce)
	mustWriteEnvelope(t, path, env)

	_, _, err := Import(path, "password")
	if !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("Import() error = %v, want ErrDecryptFailed", err)
	}
}

func TestExportUses0600Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "password", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %o, want %o", got, 0o600)
	}
}

func TestExportUsesRandomNonceAndCiphertext(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "one.deadenv")
	path2 := filepath.Join(dir, "two.deadenv")

	pairs := []envPair.EnvPair{{Key: "A", Value: "1"}}
	if err := Export("myapp", pairs, "password", path1); err != nil {
		t.Fatalf("Export(path1) error = %v", err)
	}
	if err := Export("myapp", pairs, "password", path2); err != nil {
		t.Fatalf("Export(path2) error = %v", err)
	}

	env1 := mustReadEnvelope(t, path1)
	env2 := mustReadEnvelope(t, path2)

	if env1.Nonce == env2.Nonce {
		t.Fatalf("nonce should differ between exports")
	}

	if env1.Ciphertext == env2.Ciphertext {
		t.Fatalf("ciphertext should differ between exports")
	}

	if env1.KDF.Salt == env2.KDF.Salt {
		t.Fatalf("salt should differ between exports")
	}
}

func TestImportUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "password", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	env := mustReadEnvelope(t, path)
	env.Version = currentVersion + 1
	mustWriteEnvelope(t, path, env)

	_, _, err := Import(path, "password")
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("Import() error = %v, want ErrUnsupportedVersion", err)
	}
}

func TestImportInvalidBase64ReturnsErrInvalidFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "myapp.deadenv")

	if err := Export("myapp", []envPair.EnvPair{{Key: "A", Value: "1"}}, "password", path); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	env := mustReadEnvelope(t, path)
	env.Ciphertext = "%%%not-base64%%%"
	mustWriteEnvelope(t, path, env)

	_, _, err := Import(path, "password")
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("Import() error = %v, want ErrInvalidFormat", err)
	}
}

func mustReadEnvelope(t *testing.T, path string) exportEnvelope {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var env exportEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return env
}

func mustWriteEnvelope(t *testing.T, path string, env exportEnvelope) {
	t.Helper()

	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	b = append(b, '\n')

	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func mustDecodeBase64(t *testing.T, encoded string) []byte {
	t.Helper()

	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	return b
}
