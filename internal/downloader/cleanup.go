package downloader

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanupPartialFiles removes yt-dlp's leftover partial-download artifacts
// (.part, .ytdl, and fragment files like "video.f137.part-Frag12") from dir,
// but only ones modified at or after since. That guards against deleting
// unrelated partial files that predate this job, since we can't reliably
// predict yt-dlp's exact output filename ahead of time (it depends on the
// output template and the video's actual title).
func CleanupPartialFiles(dir string, since time.Time) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var removed []string
	for _, entry := range entries {
		if entry.IsDir() || !isPartialFile(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.ModTime().Before(since) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := os.Remove(path); err == nil {
			removed = append(removed, path)
		}
	}
	return removed, nil
}

func isPartialFile(name string) bool {
	return strings.HasSuffix(name, ".part") ||
		strings.HasSuffix(name, ".ytdl") ||
		strings.Contains(name, ".part-Frag")
}
