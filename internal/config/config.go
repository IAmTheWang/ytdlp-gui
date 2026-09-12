// Package config persists user settings (download directory, concurrency,
// binary paths, cookie source, proxy, theme) as JSON, and fills in sane
// defaults for anything a loaded file leaves blank or invalid.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"ytdlp-gui/internal/ytdlp"
)

// Settings is the full set of user-configurable options.
type Settings struct {
	DownloadDir    string
	OutputTemplate string
	Concurrency    int
	Proxy          string

	// YtDlpPath and FFmpegPath are manual overrides for binmanager's
	// auto-detection; empty means keep auto-detecting.
	YtDlpPath  string
	FFmpegPath string

	Cookies ytdlp.CookieSource

	// Theme is "dark", "light", or "system".
	Theme string

	// Language is the GUI display language: "en", "ja", or "zh".
	Language string
}

// Default returns the settings a fresh install starts with.
func Default() Settings {
	return Settings{
		DownloadDir:    defaultDownloadDir(),
		OutputTemplate: "%(title)s.%(ext)s",
		Concurrency:    3,
		Cookies:        ytdlp.CookieSource{Kind: ytdlp.CookieSourceNone},
		Theme:          "light",
		Language:       "en",
	}
}

// defaultDownloadDir is this user's preferred download location. It's a
// fixed path rather than os.UserHomeDir()+"Downloads" because that's where
// they keep all their downloads (this app included), consistent with
// knownToolDirs in app.go pointing at their existing yt-dlp/ffmpeg install.
func defaultDownloadDir() string {
	return `E:\download`
}

// Sanitize fills in defaults for any field left blank or invalid, so a
// hand-edited or partially-written config.json can't leave the app in a
// broken state (e.g. Concurrency <= 0 would otherwise wedge every job in
// the queue forever, since the worker pool would have zero slots).
func (s Settings) Sanitize() Settings {
	d := Default()
	if s.DownloadDir == "" {
		s.DownloadDir = d.DownloadDir
	}
	if s.OutputTemplate == "" {
		s.OutputTemplate = d.OutputTemplate
	}
	if s.Concurrency < 1 {
		s.Concurrency = d.Concurrency
	}
	if s.Cookies.Kind == "" {
		s.Cookies = d.Cookies
	}
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	if s.Language == "" {
		s.Language = d.Language
	}
	return s
}

// DefaultPath returns the config file location: <user config dir>/ytdlp-gui/config.json.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ytdlp-gui", "config.json"), nil
}

// Load reads settings from path. A missing file is not an error: it means
// first run, so Default() (sanitized) is returned instead.
func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, err
	}
	return s.Sanitize(), nil
}

// Save writes settings to path, creating its parent directory if needed.
func Save(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
