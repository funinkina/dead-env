package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/funinkina/deadenv/internal/envPair"

	"golang.org/x/term"
)

type passwordReader func(fd int) ([]byte, error)

func PromptConfirm(message string) (bool, error) {
	return promptConfirmWithIO(os.Stdin, os.Stdout, message)
}

func PromptHidden(label string) (string, error) {
	return promptHiddenWithReader(os.Stdout, label, int(os.Stdin.Fd()), term.ReadPassword)
}

func promptHiddenWithReader(out io.Writer, label string, fd int, readPassword passwordReader) (string, error) {
	if readPassword == nil {
		return "", fmt.Errorf("password reader is nil")
	}

	if out == nil {
		out = os.Stdout
	}

	label = strings.TrimSpace(label)
	if label == "" {
		label = "Password"
	}

	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return "", err
	}

	b, err := readPassword(fd)
	if _, writeErr := fmt.Fprintln(out); writeErr != nil && err == nil {
		return "", writeErr
	}
	if err != nil {
		return "", err
	}

	return string(b), nil
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

	if _, err := fmt.Fprintf(
		out,
		"  Added: %d  Modified: %d  Removed: %d\n",
		len(added),
		len(modified),
		len(removed),
	); err != nil {
		return err
	}

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

	if _, err := fmt.Fprintln(out, "(values are hidden)"); err != nil {
		return err
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
