// Package binmanager locates the yt-dlp and ffmpeg executables this app
// depends on: a user-configured path first, then the system PATH, then a
// list of known fallback directories.
package binmanager

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// exeName appends the platform executable extension (".exe" on Windows) to
// a bare binary name like "yt-dlp".
func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// Locate resolves an executable named `name` (without extension), checking
// in order:
//  1. configuredPath, if non-empty and it points at an existing file
//  2. the system PATH
//  3. each directory in fallbackDirs, joined with the platform executable name
//
// It returns the resolved path and whether it was found.
func Locate(name string, configuredPath string, fallbackDirs []string) (string, bool) {
	if configuredPath != "" {
		if info, err := os.Stat(configuredPath); err == nil && !info.IsDir() {
			return configuredPath, true
		}
	}

	if p, err := exec.LookPath(name); err == nil {
		return p, true
	}

	fname := exeName(name)
	for _, dir := range fallbackDirs {
		candidate := filepath.Join(dir, fname)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
	}

	return "", false
}
