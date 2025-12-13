package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the yapatype config directory
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "yapatype")
}

// ConfigPath returns the default config file path
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// StatePath returns the state file path for persisting runtime state
func StatePath() string {
	return filepath.Join(ConfigDir(), ".state")
}

// GetPlatform returns the current platform as lowercase string
func GetPlatform() string {
	if runtime.GOOS == "darwin" {
		return "darwin"
	}
	return "linux"
}
