package profile

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/funinkina/deadenv/internal/envPair"
	"github.com/funinkina/deadenv/internal/tui"
)

type EditOptions struct {
	Yes        bool
	Out        io.Writer
	FromEditor func(initialContent string) ([]envPair.EnvPair, error)
	Confirm    func(message string) (bool, error)
}

func (p *ProfileService) Edit(profile string, opts EditOptions) error {
	if profile == "" {
		return ErrProfileNameEmpty
	}

	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	if opts.FromEditor == nil {
		opts.FromEditor = FromEditor
	}

	if opts.Confirm == nil {
		opts.Confirm = tui.PromptConfirm
	}

	currentPairs, err := p.loadProfilePairs(profile)
	if err != nil {
		return err
	}

	content := SerializeEnvPairs(profile, currentPairs)
	editedPairs, err := opts.FromEditor(content)
	if err != nil {
		return fmt.Errorf("opening editor for profile %q: %w", profile, err)
	}

	if len(editedPairs) == 0 {
		return ErrEmptyContent
	}

	diff := DiffPairs(currentPairs, editedPairs)
	if len(diff.Added) == 0 && len(diff.Modified) == 0 && len(diff.Removed) == 0 {
		return ErrNoChanges
	}

	addedKeys := keysOf(diff.Added)
	modifiedKeys := keysOf(diff.Modified)
	removedKeys := keysOf(diff.Removed)

	if err := tui.PrintChangeSummary(opts.Out, addedKeys, modifiedKeys, removedKeys); err != nil {
		return fmt.Errorf("printing change summary: %w", err)
	}

	if !opts.Yes {
		ok, confirmErr := opts.Confirm("Apply these changes?")
		if confirmErr != nil {
			return fmt.Errorf("confirming profile edit: %w", confirmErr)
		}
		if !ok {
			return ErrCancelled
		}
	}

	applyErr := &PartialApplyError{
		Succeeded: make([]string, 0),
		Failed:    make(map[string]error),
	}

	for _, pair := range diff.Added {
		if err := p.SetKey(profile, pair.Key, pair.Value); err != nil {
			applyErr.Failed[pair.Key] = err
			continue
		}
		applyErr.Succeeded = append(applyErr.Succeeded, pair.Key)
	}

	for _, pair := range diff.Modified {
		if err := p.SetKey(profile, pair.Key, pair.Value); err != nil {
			applyErr.Failed[pair.Key] = err
			continue
		}
		applyErr.Succeeded = append(applyErr.Succeeded, pair.Key)
	}

	for _, pair := range diff.Removed {
		if err := p.UnsetKey(profile, pair.Key); err != nil {
			applyErr.Failed[pair.Key] = err
			continue
		}
		applyErr.Succeeded = append(applyErr.Succeeded, pair.Key)
	}

	if len(applyErr.Failed) > 0 {
		return applyErr
	}

	return nil
}

func (p *ProfileService) loadProfilePairs(profile string) ([]envPair.EnvPair, error) {
	keys, err := p.ListKeys(profile)
	if err != nil {
		return nil, fmt.Errorf("listing keys for profile %q: %w", profile, err)
	}

	pairs := make([]envPair.EnvPair, 0, len(keys))
	for _, key := range keys {
		value, readErr := p.GetKey(profile, key)
		if readErr != nil {
			return nil, fmt.Errorf("reading key %q from profile %q: %w", key, profile, readErr)
		}

		pairs = append(pairs, envPair.EnvPair{Key: key, Value: value})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})

	return pairs, nil
}

func keysOf(pairs []envPair.EnvPair) []string {
	keys := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		keys = append(keys, pair.Key)
	}

	return keys
}
