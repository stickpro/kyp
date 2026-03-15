package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultDBPath returns the platform-specific default path to kyp.db
// and ensures the parent directory exists.
//
// Priority: XDG_DATA_HOME (Linux) → OS default data dir → ~/.local/share
func DefaultDBPath() (string, error) {
	dataDir, err := userDataDir()
	if err != nil {
		return "", fmt.Errorf("get user data dir: %w", err)
	}
	path := filepath.Join(dataDir, "kyp", "kyp.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("create db dir: %w", err)
	}
	return path, nil
}

// userDataDir returns the OS-specific directory for user application data:
//
//	Linux:   $XDG_DATA_HOME or ~/.local/share
//	macOS:   ~/Library/Application Support
//	Windows: %APPDATA%
func userDataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support"), nil
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return appdata, nil
		}
	}
	return filepath.Join(home, ".local", "share"), nil
}
