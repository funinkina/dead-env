package history

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGitRecorderInitializesRepository(t *testing.T) {
	repoDir := t.TempDir()

	recorder, err := NewGitRecorder(repoDir)
	if err != nil {
		t.Fatalf("NewGitRecorder() error = %v", err)
	}
	if recorder == nil {
		t.Fatal("NewGitRecorder() = nil, want recorder")
	}

	if _, err := exec.Command("git", "-C", repoDir, "rev-parse", "--git-dir").CombinedOutput(); err != nil {
		t.Fatalf("git repo not initialized: %v", err)
	}
}

func TestRecordDeleteProfileAllowsEmptyKeyAndCommitsExpectedMessage(t *testing.T) {
	repoDir := t.TempDir()

	recorder, err := NewGitRecorder(repoDir)
	if err != nil {
		t.Fatalf("NewGitRecorder() error = %v", err)
	}

	if err := recorder.Record("myapp", OpSet, "API_KEY", "hash:secret"); err != nil {
		t.Fatalf("Record(set) error = %v", err)
	}
	if err := recorder.Record("myapp", OpDeleteProfile, "", ""); err != nil {
		t.Fatalf("Record(delete-profile) error = %v", err)
	}

	snapshot, err := readProfileSnapshot(repoDir+"/myapp.json", "myapp")
	if err != nil {
		t.Fatalf("readProfileSnapshot() error = %v", err)
	}
	if snapshot.DeletedAt == nil {
		t.Fatalf("DeletedAt = nil, want timestamp after delete-profile")
	}

	cmd := exec.Command("git", "-C", repoDir, "log", "-1", "--pretty=%s")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log error = %v: %s", err, strings.TrimSpace(string(out)))
	}

	got := strings.TrimSpace(string(out))
	want := "[myapp] delete profile"
	if got != want {
		t.Fatalf("git log subject = %q, want %q", got, want)
	}
}

func TestRecordRejectsEmptyKeyForNonDeleteOperations(t *testing.T) {
	repoDir := t.TempDir()

	recorder, err := NewGitRecorder(repoDir)
	if err != nil {
		t.Fatalf("NewGitRecorder() error = %v", err)
	}

	err = recorder.Record("myapp", OpSet, "", "hash:secret")
	if err == nil {
		t.Fatal("Record() error = nil, want error for empty key on non-delete operation")
	}
	if !strings.Contains(err.Error(), "key cannot be empty") {
		t.Fatalf("Record() error = %v, want key cannot be empty", err)
	}
}

func TestLogReturnsEntriesAndSupportsKeyFilter(t *testing.T) {
	repoDir := t.TempDir()

	recorder, err := NewGitRecorder(repoDir)
	if err != nil {
		t.Fatalf("NewGitRecorder() error = %v", err)
	}

	if err := recorder.Record("myapp", OpSet, "API_KEY", "hash:one"); err != nil {
		t.Fatalf("Record(set) error = %v", err)
	}
	if err := recorder.Record("myapp", OpUnset, "API_KEY", ""); err != nil {
		t.Fatalf("Record(unset) error = %v", err)
	}
	if err := recorder.Record("myapp", OpSet, "DATABASE_URL", "hash:two"); err != nil {
		t.Fatalf("Record(set other key) error = %v", err)
	}

	entries, err := recorder.Log("myapp", "")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("Log() len = %d, want 3", len(entries))
	}
	if entries[0].Key != "DATABASE_URL" || entries[0].Operation != OpSet || entries[0].ValueHash != "hash:two" {
		t.Fatalf("latest entry = %+v, want DATABASE_URL set hash:two", entries[0])
	}
	if entries[1].Key != "API_KEY" || entries[1].Operation != OpUnset || entries[1].ValueHash != "" {
		t.Fatalf("middle entry = %+v, want API_KEY unset", entries[1])
	}
	if entries[2].Key != "API_KEY" || entries[2].Operation != OpSet || entries[2].ValueHash != "hash:one" {
		t.Fatalf("oldest entry = %+v, want API_KEY set hash:one", entries[2])
	}

	filtered, err := recorder.Log("myapp", "API_KEY")
	if err != nil {
		t.Fatalf("Log(filtered) error = %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("Log(filtered) len = %d, want 2", len(filtered))
	}
	for _, entry := range filtered {
		if entry.Key != "API_KEY" {
			t.Fatalf("filtered entry key = %q, want API_KEY", entry.Key)
		}
	}
}

func TestLogReturnsEmptyForNoCommits(t *testing.T) {
	repoDir := t.TempDir()

	recorder, err := NewGitRecorder(repoDir)
	if err != nil {
		t.Fatalf("NewGitRecorder() error = %v", err)
	}

	entries, err := recorder.Log("myapp", "")
	if err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("Log() len = %d, want 0", len(entries))
	}
}

func TestNewGitRecorderReturnsErrGitNotFoundWhenGitMissing(t *testing.T) {
	oldLookPath := gitLookPath
	gitLookPath = func(file string) (string, error) {
		return "", errors.New("missing git")
	}
	defer func() {
		gitLookPath = oldLookPath
	}()

	_, err := NewGitRecorder(filepath.Join(t.TempDir(), "history"))
	if !errors.Is(err, ErrGitNotFound) {
		t.Fatalf("NewGitRecorder() error = %v, want ErrGitNotFound", err)
	}
}
