package history

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDirName     = "deadenv"
	historyDirName = "history"
)

func DefaultHistoryDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config dir: %w", err)
	}

	return filepath.Join(base, appDirName, historyDirName), nil
}
