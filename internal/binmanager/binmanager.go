package binmanager

import (
	"os/exec"
	"strings"
)

// BinaryStatus reports where (and whether) the required external
// executables were found.
type BinaryStatus struct {
	YtDlpPath  string
	YtDlpFound bool

	FFmpegPath  string
	FFmpegFound bool
}

// Ready reports whether every required binary was located.
func (s BinaryStatus) Ready() bool {
	return s.YtDlpFound && s.FFmpegFound
}

// ProbeConfig carries user-configured overrides and extra search
// directories for Check.
type ProbeConfig struct {
	// YtDlpPath and FFmpegPath are user-configured overrides from settings;
	// empty means auto-detect.
	YtDlpPath  string
	FFmpegPath string
	// FallbackDirs are searched last, after the configured path and PATH.
	// This is where a known tools folder (e.g. the user's existing
	// yt-dlp.exe/ffmpeg.exe install) gets a chance even if it's not on PATH.
	FallbackDirs []string
}

// Check locates yt-dlp and ffmpeg per the detection order documented on
// Locate.
func Check(cfg ProbeConfig) BinaryStatus {
	var s BinaryStatus
	s.YtDlpPath, s.YtDlpFound = Locate("yt-dlp", cfg.YtDlpPath, cfg.FallbackDirs)
	s.FFmpegPath, s.FFmpegFound = Locate("ffmpeg", cfg.FFmpegPath, cfg.FallbackDirs)
	return s
}

// Version runs `<path> --version` and returns its trimmed first line. It's
// used to surface the yt-dlp/ffmpeg version in the UI, e.g. so a stale
// yt-dlp build (which breaks often as sites change) is visible at a glance.
func Version(path string) (string, error) {
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", err
	}
	line, _, _ := strings.Cut(string(out), "\n")
	return strings.TrimSpace(line), nil
}
