package profile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"funinkina/deadenv/internal/envPair"
)

const runCommandNotFoundExitCode = 127

type RunOptions struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	BaseEnv []string
}

type ExitCodeError struct {
	Code int
	Err  error
}

func (e *ExitCodeError) Error() string {
	if e == nil || e.Err == nil {
		return "process exited"
	}

	return e.Err.Error()
}

func (e *ExitCodeError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *ExitCodeError) ExitCode() int {
	if e == nil {
		return 1
	}

	return e.Code
}

var runExecCommand = exec.Command

func (p *ProfileService) Run(profileName, commandName string, args []string, opts RunOptions) error {
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		return ErrProfileNameEmpty
	}

	commandName = strings.TrimSpace(commandName)
	if commandName == "" {
		return fmt.Errorf("command is required")
	}

	pairs, err := p.GetPairs(profileName)
	if err != nil {
		return fmt.Errorf("loading profile %q: %w", profileName, err)
	}

	env := mergeRunEnv(opts.BaseEnv, pairs)

	cmd := runExecCommand(commandName, args...)
	cmd.Env = env
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			return &ExitCodeError{
				Code: runCommandNotFoundExitCode,
				Err:  fmt.Errorf("command not found: %s", commandName),
			}
		}

		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			return &ExitCodeError{
				Code: runCommandNotFoundExitCode,
				Err:  fmt.Errorf("command not found: %s", commandName),
			}
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ExitCodeError{
				Code: exitErr.ExitCode(),
				Err:  err,
			}
		}

		return err
	}

	return nil
}

func mergeRunEnv(baseEnv []string, pairs []envPair.EnvPair) []string {
	if baseEnv == nil {
		baseEnv = os.Environ()
	}

	ordered := make([]string, 0, len(baseEnv)+len(pairs))
	indexByKey := make(map[string]int, len(baseEnv)+len(pairs))

	for _, entry := range baseEnv {
		key, ok := envEntryKey(entry)
		if !ok {
			ordered = append(ordered, entry)
			continue
		}

		if idx, exists := indexByKey[key]; exists {
			ordered[idx] = entry
			continue
		}

		indexByKey[key] = len(ordered)
		ordered = append(ordered, entry)
	}

	sortedPairs := append([]envPair.EnvPair(nil), pairs...)
	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[i].Key < sortedPairs[j].Key
	})

	for _, pair := range sortedPairs {
		if strings.TrimSpace(pair.Key) == "" {
			continue
		}

		entry := pair.Key + "=" + pair.Value
		if idx, exists := indexByKey[pair.Key]; exists {
			ordered[idx] = entry
			continue
		}

		indexByKey[pair.Key] = len(ordered)
		ordered = append(ordered, entry)
	}

	return ordered
}

func envEntryKey(entry string) (string, bool) {
	key, _, ok := strings.Cut(entry, "=")
	if !ok || key == "" {
		return "", false
	}

	return key, true
}
