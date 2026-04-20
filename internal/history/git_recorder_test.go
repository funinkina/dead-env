package history

import (
	"os/exec"
	"strings"
	"testing"
)

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
