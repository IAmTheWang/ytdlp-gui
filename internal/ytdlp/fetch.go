package ytdlp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FetchMetadata runs `<binPath> -J <url>` (or `--flat-playlist -J` when
// flatPlaylist is true) and parses the result. It's a thin wrapper around
// BuildMetadataArgs and ParseMetadataJSON, which carry the logic actually
// covered by unit tests; this just wires them to a real subprocess.
func FetchMetadata(ctx context.Context, binPath, url string, flatPlaylist bool) (*Metadata, error) {
	args := BuildMetadataArgs(url, flatPlaylist)
	cmd := exec.CommandContext(ctx, binPath, args...)
	// Force UTF-8 so titles/paths with non-ASCII characters don't get
	// mangled by whatever code page the system happens to default to.
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp: %s", msg)
	}
	return ParseMetadataJSON(out)
}
