package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConfigDir returns the yapatype config directory
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "yapatype")
}

// ExpandPath resolves a config path to an absolute path
// expands a leading ~ and resolves relative paths against the
// working directory, executable directory, then config directory
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}

	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	for _, base := range relativeBaseDirs() {
		candidate := filepath.Join(base, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// nothing found yet, anchor to cwd for a stable value
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// relativeBaseDirs returns dirs to resolve relative config paths against
func relativeBaseDirs() []string {
	var dirs []string
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		dirs = append(dirs, dir)
		// follow symlinks so models resolve next to the real binary
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			if rdir := filepath.Dir(resolved); rdir != dir {
				dirs = append(dirs, rdir)
			}
		}
	}
	return append(dirs, ConfigDir())
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
