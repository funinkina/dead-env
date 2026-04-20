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

func (r *GitRecorder) Log(profile, key string) ([]HistoryEntry, error) {
	if err := validateProfileName(profile); err != nil {
		return nil, err
	}

	out, err := r.gitOutput("log", "--format=%H%x00%cI%x00%s", "--", profile+".json")
	if err != nil {
		if isNoCommitsError(err) {
			return []HistoryEntry{}, nil
		}

		return nil, err
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []HistoryEntry{}, nil
	}

	lines := strings.Split(raw, "\n")
	entries := make([]HistoryEntry, 0, len(lines))

	for _, line := range lines {
		entry, include, err := r.parseLogLine(profile, key, line)
		if err != nil {
			return nil, err
		}
		if include {
			entries = append(entries, entry)
		}
	}

	return entries, nil
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
	_, err := r.gitOutput(args...)
	return err
}

func (r *GitRecorder) gitOutput(args ...string) ([]byte, error) {
	cmdArgs := append([]string{"-C", r.repoDir}, args...)
	cmd := exec.Command(r.gitBinary, cmdArgs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf(
			"%w: git %s: %v: %s",
			ErrGitCommandFailed,
			strings.Join(args, " "),
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return out, nil
}

func (r *GitRecorder) parseLogLine(profile, keyFilter, line string) (HistoryEntry, bool, error) {
	parts := strings.SplitN(line, "\x00", 3)
	if len(parts) != 3 {
		return HistoryEntry{}, false, fmt.Errorf("unexpected git log format for profile %q", profile)
	}

	timestamp, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return HistoryEntry{}, false, fmt.Errorf("parsing history timestamp %q: %w", parts[1], err)
	}

	operation, key, err := parseCommitSubject(profile, parts[2])
	if err != nil {
		return HistoryEntry{}, false, err
	}

	if keyFilter != "" && key != keyFilter {
		return HistoryEntry{}, false, nil
	}

	valueHash := ""
	if operation != OpDeleteProfile {
		valueHash, err = r.valueHashAtCommit(parts[0], profile, key)
		if err != nil {
			return HistoryEntry{}, false, err
		}
	}

	return HistoryEntry{
		Profile:   profile,
		Operation: operation,
		Key:       key,
		ValueHash: valueHash,
		Timestamp: timestamp,
	}, true, nil
}

func parseCommitSubject(profile, subject string) (string, string, error) {
	deleteSubject := fmt.Sprintf("[%s] delete profile", profile)
	if subject == deleteSubject {
		return OpDeleteProfile, "", nil
	}

	prefix := fmt.Sprintf("[%s] ", profile)
	if !strings.HasPrefix(subject, prefix) {
		return "", "", fmt.Errorf("unexpected history commit subject %q", subject)
	}

	rest := strings.TrimPrefix(subject, prefix)
	operation, key, ok := strings.Cut(rest, " ")
	if !ok || strings.TrimSpace(key) == "" {
		return "", "", fmt.Errorf("unexpected history commit subject %q", subject)
	}

	return operation, key, nil
}

func (r *GitRecorder) valueHashAtCommit(commitHash, profile, key string) (string, error) {
	out, err := r.gitOutput("show", commitHash+":"+profile+".json")
	if err != nil {
		return "", err
	}

	var snapshot ProfileSnapshot
	if err := json.Unmarshal(out, &snapshot); err != nil {
		return "", fmt.Errorf("decoding history snapshot for commit %q: %w", commitHash, err)
	}

	keySnapshot, ok := snapshot.Keys[key]
	if !ok {
		return "", fmt.Errorf("history snapshot for commit %q missing key %q", commitHash, key)
	}

	return keySnapshot.ValueHash, nil
}

func isNoCommitsError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "does not have any commits yet") ||
		(strings.Contains(msg, "your current branch") && strings.Contains(msg, "does not have any commits yet"))
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
