package profile

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"

	"funinkina/deadenv/internal/envPair"
	"funinkina/deadenv/internal/history"
	"funinkina/deadenv/internal/keychain"
)

func TestMergeRunEnvOverridesAndAppends(t *testing.T) {
	base := []string{
		"PATH=/bin",
		"KEEP=1",
		"DUP=old",
		"DUP=older",
	}

	pairs := []envPair.EnvPair{
		{Key: "DUP", Value: "new"},
		{Key: "APPENDED", Value: "yes"},
		{Key: "", Value: "ignored"},
	}

	got := mergeRunEnv(base, pairs)

	want := []string{
		"PATH=/bin",
		"KEEP=1",
		"DUP=new",
		"APPENDED=yes",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeRunEnv() = %v, want %v", got, want)
	}
}

func TestProfileRunReturnsCommandNotFoundExitCode(t *testing.T) {
	store := keychain.NewFake()
	if err := store.Write("deadenv/myapp", "A", "1"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	service, err := NewProfileService(store, &history.FakeRecorder{}, fixedHash)
	if err != nil {
		t.Fatalf("NewProfileService() error = %v", err)
	}

	oldRunExecCommand := runExecCommand
	runExecCommand = func(name string, args ...string) *exec.Cmd {
		_ = args
		return exec.Command("/definitely-not-existing-command-deadenv")
	}
	defer func() {
		runExecCommand = oldRunExecCommand
	}()

	err = service.Run("myapp", "missing", nil, RunOptions{BaseEnv: []string{"PATH=/bin"}})
	if err == nil {
		t.Fatal("Run() error = nil, want non-nil")
	}

	var exitErr *ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Run() error type = %T, want *ExitCodeError", err)
	}
	if exitErr.ExitCode() != runCommandNotFoundExitCode {
		t.Fatalf("ExitCode() = %d, want %d", exitErr.ExitCode(), runCommandNotFoundExitCode)
	}
}
