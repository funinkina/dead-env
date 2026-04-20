package profile

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrNilStore         = errors.New("profile service requires a keychain store")
	ErrProfileNameEmpty = errors.New("profile name cannot be empty")
	ErrKeyEmpty         = errors.New("key cannot be empty")
	ErrEmptyContent     = errors.New("content cannot be empty")
	ErrNoChanges        = errors.New("no changes detected")
	ErrEditorFailed     = errors.New("editor failed")
	ErrApplyChanges     = errors.New("one or more changes failed to apply")
	ErrCancelled        = errors.New("operation cancelled")
)

type PartialApplyError struct {
	Succeeded []string
	Failed    map[string]error
}

func (e *PartialApplyError) Error() string {
	if e == nil {
		return ErrApplyChanges.Error()
	}

	failedKeys := make([]string, 0, len(e.Failed))
	for key := range e.Failed {
		failedKeys = append(failedKeys, key)
	}
	sort.Strings(failedKeys)

	var details []string
	for _, key := range failedKeys {
		details = append(details, fmt.Sprintf("%s: %v", key, e.Failed[key]))
	}

	succeeded := append([]string(nil), e.Succeeded...)
	sort.Strings(succeeded)

	if len(details) == 0 {
		return ErrApplyChanges.Error()
	}

	if len(succeeded) == 0 {
		return fmt.Sprintf("%v: failed keys: %s", ErrApplyChanges, strings.Join(details, "; "))
	}

	return fmt.Sprintf(
		"%v: succeeded keys: %s; failed keys: %s",
		ErrApplyChanges,
		strings.Join(succeeded, ", "),
		strings.Join(details, "; "),
	)
}

func (e *PartialApplyError) Unwrap() error {
	return ErrApplyChanges
}

func (e *PartialApplyError) FailedKeys() []string {
	if e == nil {
		return nil
	}

	keys := make([]string, 0, len(e.Failed))
	for key := range e.Failed {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
