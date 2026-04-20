package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	gitLookPath   = exec.LookPath
	profileNameRE = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
)

type GitRecorder struct {
	repoDir    string
	gitBinary  string
	now        func() time.Time
	commitName string
	commitMail string
}

func NewGitRecorder(historyDir string) (*GitRecorder, error) {
	if strings.TrimSpace(historyDir) == "" {
		var err error
		historyDir, err = DefaultHistoryDir()
		if err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(historyDir) == "" {
		return nil, ErrInvalidHistoryDir
	}

	gitBin, err := gitLookPath("git")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGitNotFound, err)
	}

	if err := os.MkdirAll(historyDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating history directory %q: %w", historyDir, err)
	}

	r := &GitRecorder{
		repoDir:    historyDir,
		gitBinary:  gitBin,
		now:        time.Now,
		commitName: "deadenv",
		commitMail: "deadenv@local",
	}

	if err := r.ensureRepository(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *GitRecorder) Record(profile, operation, key, valueHash string) error {
	if err := validateProfileName(profile); err != nil {
		return err
	}

	if strings.TrimSpace(operation) == "" {
		return fmt.Errorf("operation cannot be empty")
	}

	if operation != OpDeleteProfile && strings.TrimSpace(key) == "" {
		return fmt.Errorf("key cannot be empty")
	}

	snapshotPath := filepath.Join(r.repoDir, profile+".json")

	snapshot, err := readProfileSnapshot(snapshotPath, profile)

	if err != nil {
		return err
	}

	if snapshot.Keys == nil {
		snapshot.Keys = make(map[string]KeySnapshot)
	}

	now := r.now().UTC()

	if operation == OpDeleteProfile {
		snapshot.DeletedAt = &now
	} else {
		snapshot.DeletedAt = nil
		snapshot.Keys[key] = KeySnapshot{
			Op:        operation,
			ValueHash: valueHash,
			UpdatedAt: now,
		}
	}

	if err := writeProfileSnapshot(snapshotPath, snapshot); err != nil {
		return err
	}

	if err := r.git("add", filepath.Base(snapshotPath)); err != nil {
		return err
	}

	msg := fmt.Sprintf("[%s] %s %s", profile, operation, key)
	if operation == OpDeleteProfile {
		msg = fmt.Sprintf("[%s] delete profile", profile)
	}

	if err := r.git(
		"-c", "commit.gpgsign=false",
		"-c", "user.name="+r.commitName,
		"-c", "user.email="+r.commitMail,
		"commit", "-m", msg,
	); err != nil {
		return err
	}

	return nil
}

func (r *GitRecorder) ensureRepository() error {
	gitDir := filepath.Join(r.repoDir, ".git")

	if st, err := os.Stat(gitDir); err == nil && st.IsDir() {
		return nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking git dir %q: %w", gitDir, err)
	}

	if err := r.git("init"); err != nil {
		return err
	}

	return nil
}

func (r *GitRecorder) git(args ...string) error {
	cmdArgs := append([]string{"-C", r.repoDir}, args...)
	cmd := exec.Command(r.gitBinary, cmdArgs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"%w: git %s: %v: %s",
			ErrGitCommandFailed,
			strings.Join(args, " "),
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func readProfileSnapshot(path, profile string) (ProfileSnapshot, error) {
	b, err := os.ReadFile(path)

	if errors.Is(err, fs.ErrNotExist) {
		return ProfileSnapshot{
			Profile: profile,
			Keys:    make(map[string]KeySnapshot),
		}, nil
	}

	if err != nil {
		return ProfileSnapshot{}, fmt.Errorf("reading history snapshot %q: %w", path, err)
	}

	var snap ProfileSnapshot

	if err := json.Unmarshal(b, &snap); err != nil {
		return ProfileSnapshot{}, fmt.Errorf("decoding history snapshot %q: %w", path, err)
	}

	if snap.Profile == "" {
		snap.Profile = profile
	}

	if snap.Keys == nil {
		snap.Keys = make(map[string]KeySnapshot)
	}

	return snap, nil
}

func writeProfileSnapshot(path string, snap ProfileSnapshot) error {
	b, err := json.MarshalIndent(snap, "", "  ")

	if err != nil {
		return fmt.Errorf("encoding history snapshot %q: %w", path, err)
	}

	b = append(b, '\n')

	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("writing history snapshot %q: %w", path, err)
	}

	return nil
}

func validateProfileName(profile string) error {
	if !profileNameRE.MatchString(profile) {
		return fmt.Errorf("%w: %q", ErrInvalidProfileName, profile)
	}

	return nil
}
