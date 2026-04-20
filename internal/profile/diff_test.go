package profile

import (
	"reflect"
	"testing"

	"funinkina/deadenv/internal/envPair"
)

func TestDiffPairs(t *testing.T) {
	tests := []struct {
		name      string
		oldPairs  []envPair.EnvPair
		newPairs  []envPair.EnvPair
		wantAdded []envPair.EnvPair
		wantMod   []envPair.EnvPair
		wantRem   []envPair.EnvPair
	}{
		{
			name: "no changes",
			oldPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
			},
			wantAdded: []envPair.EnvPair{},
			wantMod:   []envPair.EnvPair{},
			wantRem:   []envPair.EnvPair{},
		},
		{
			name: "added key",
			oldPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
			},
			wantAdded: []envPair.EnvPair{{Key: "B", Value: "2"}},
			wantMod:   []envPair.EnvPair{},
			wantRem:   []envPair.EnvPair{},
		},
		{
			name: "modified value",
			oldPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "A", Value: "2"},
			},
			wantAdded: []envPair.EnvPair{},
			wantMod:   []envPair.EnvPair{{Key: "A", Value: "2"}},
			wantRem:   []envPair.EnvPair{},
		},
		{
			name: "removed key",
			oldPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
			},
			wantAdded: []envPair.EnvPair{},
			wantMod:   []envPair.EnvPair{},
			wantRem:   []envPair.EnvPair{{Key: "B", Value: "2"}},
		},
		{
			name: "key rename treated as remove and add",
			oldPairs: []envPair.EnvPair{
				{Key: "OLD_KEY", Value: "secret"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "NEW_KEY", Value: "secret"},
			},
			wantAdded: []envPair.EnvPair{{Key: "NEW_KEY", Value: "secret"}},
			wantMod:   []envPair.EnvPair{},
			wantRem:   []envPair.EnvPair{{Key: "OLD_KEY", Value: "secret"}},
		},
		{
			name: "mixed changes",
			oldPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "2"},
				{Key: "C", Value: "3"},
			},
			newPairs: []envPair.EnvPair{
				{Key: "A", Value: "1"},
				{Key: "B", Value: "20"},
				{Key: "D", Value: "4"},
			},
			wantAdded: []envPair.EnvPair{{Key: "D", Value: "4"}},
			wantMod:   []envPair.EnvPair{{Key: "B", Value: "20"}},
			wantRem:   []envPair.EnvPair{{Key: "C", Value: "3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffPairs(tt.oldPairs, tt.newPairs)

			if !reflect.DeepEqual(got.Added, tt.wantAdded) {
				t.Fatalf("Added = %v, want %v", got.Added, tt.wantAdded)
			}
			if !reflect.DeepEqual(got.Modified, tt.wantMod) {
				t.Fatalf("Modified = %v, want %v", got.Modified, tt.wantMod)
			}
			if !reflect.DeepEqual(got.Removed, tt.wantRem) {
				t.Fatalf("Removed = %v, want %v", got.Removed, tt.wantRem)
			}
		})
	}
}
