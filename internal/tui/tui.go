package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"funinkina/deadenv/internal/envPair"
)

func PromptConfirm(message string) (bool, error) {
	return promptConfirmWithIO(os.Stdin, os.Stdout, message)
}

func promptConfirmWithIO(in io.Reader, out io.Writer, message string) (bool, error) {
	reader := bufio.NewReader(in)

	for {
		if _, err := fmt.Fprintf(out, "%s [y/N]: ", strings.TrimSpace(message)); err != nil {
			return false, err
		}

		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return false, err
		}

		answer := strings.ToLower(strings.TrimSpace(line))
		switch answer {
		case "", "n", "no":
			return false, nil
		case "y", "yes":
			return true, nil
		default:
			if _, writeErr := fmt.Fprintln(out, "Please answer with 'y' or 'n'."); writeErr != nil {
				return false, writeErr
			}
		}
	}
}

func PrintChangeSummary(out io.Writer, added, modified, removed []string) error {
	if out == nil {
		out = os.Stdout
	}

	if _, err := fmt.Fprintln(out, "Changes detected:"); err != nil {
		return err
	}

	sort.Strings(added)
	sort.Strings(modified)
	sort.Strings(removed)

	for _, key := range added {
		if _, err := fmt.Fprintf(out, "  [set]      %s\n", key); err != nil {
			return err
		}
	}

	for _, key := range modified {
		if _, err := fmt.Fprintf(out, "  [modified] %s\n", key); err != nil {
			return err
		}
	}

	for _, key := range removed {
		if _, err := fmt.Fprintf(out, "  [removed]  %s\n", key); err != nil {
			return err
		}
	}

	return nil
}

func PrintPairSummary(out io.Writer, pairs []envPair.EnvPair) error {
	if out == nil {
		out = os.Stdout
	}

	keys := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		keys = append(keys, pair.Key)
	}
	sort.Strings(keys)

	if _, err := fmt.Fprintf(out, "Parsed %d keys:\n", len(keys)); err != nil {
		return err
	}

	for _, key := range keys {
		if _, err := fmt.Fprintf(out, "  - %s\n", key); err != nil {
			return err
		}
	}

	return nil
}

func MaskValue(value string) string {
	if value == "" {
		return ""
	}

	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}

	suffix := value[len(value)-4:]
	return strings.Repeat("*", len(value)-4) + suffix
}
