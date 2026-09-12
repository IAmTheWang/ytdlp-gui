package config

import (
	"os"
	"path/filepath"
	"testing"

	"ytdlp-gui/internal/ytdlp"
)

func TestLoad_MissingFileReturnsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error for a missing file: %v", err)
	}
	want := Default()
	if got != want {
		t.Errorf("Load() = %+v, want Default() %+v", got, want)
	}
}

func TestSaveThenLoad_RoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	want := Settings{
		DownloadDir:    `D:\Videos`,
		OutputTemplate: "%(id)s.%(ext)s",
		Concurrency:    5,
		Proxy:          "socks5://127.0.0.1:1080",
		YtDlpPath:      `E:\software\ytdlp-interface\yt-dlp.exe`,
		FFmpegPath:     `E:\software\ytdlp-interface\ffmpeg.exe`,
		Cookies:        ytdlp.CookieSource{Kind: ytdlp.CookieSourceBrowser, Browser: "chrome"},
		Theme:          "light",
		Language:       "ja",
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got != want {
		t.Errorf("round-tripped settings = %+v, want %+v", got, want)
	}
}

func TestLoad_SanitizesInvalidFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	// Simulate a hand-edited or partially-written file: a non-positive
	// concurrency, and everything else left blank.
	if err := os.WriteFile(path, []byte(`{"Concurrency": 0}`), 0o644); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	def := Default()
	if got.Concurrency != def.Concurrency {
		t.Errorf("expected Concurrency to fall back to default %d, got %d", def.Concurrency, got.Concurrency)
	}
	if got.DownloadDir != def.DownloadDir {
		t.Errorf("expected DownloadDir to fall back to default %q, got %q", def.DownloadDir, got.DownloadDir)
	}
	if got.Cookies.Kind != ytdlp.CookieSourceNone {
		t.Errorf("expected Cookies.Kind to fall back to %q, got %q", ytdlp.CookieSourceNone, got.Cookies.Kind)
	}
	if got.Language != def.Language {
		t.Errorf("expected Language to fall back to default %q, got %q", def.Language, got.Language)
	}
}

func TestLoad_MalformedJSONReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("failed to write test fixture: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Errorf("expected an error for malformed JSON")
	}
}

func TestDefaultPath_UnderUserConfigDir(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath returned error: %v", err)
	}
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Skip("os.UserConfigDir unavailable in this environment")
	}
	want := filepath.Join(userConfigDir, "ytdlp-gui", "config.json")
	if path != want {
		t.Errorf("DefaultPath() = %q, want %q", path, want)
	}
}
