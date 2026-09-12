package ytdlp

import "encoding/json"

// ProgressStatus classifies a line of yt-dlp stdout output.
type ProgressStatus string

const (
	// StatusDownloading means the line carried a structured progress
	// update (see progressTemplate).
	StatusDownloading ProgressStatus = "downloading"
	// StatusPostProcessing means the line was plain text: yt-dlp's own
	// post-processing chatter (merging formats, extracting audio,
	// embedding subtitles, ...) or ffmpeg's log output. yt-dlp does not
	// emit structured progress for this phase, so these lines are
	// surfaced to the UI as a log message instead of a percentage.
	StatusPostProcessing ProgressStatus = "post_processing"
)

// ProgressEvent is one parsed line of yt-dlp stdout.
type ProgressEvent struct {
	Status          ProgressStatus
	DownloadedBytes int64
	TotalBytes      int64
	Speed           float64
	ETA             int64
	// Message holds the raw line text when Status == StatusPostProcessing.
	Message string
}

// Percent returns the completion percentage, or 0 if TotalBytes is unknown.
func (e ProgressEvent) Percent() float64 {
	if e.TotalBytes <= 0 {
		return 0
	}
	return float64(e.DownloadedBytes) / float64(e.TotalBytes) * 100
}

type rawProgressJSON struct {
	Status          string  `json:"status"`
	DownloadedBytes int64   `json:"downloaded_bytes"`
	TotalBytes      int64   `json:"total_bytes"`
	Speed           float64 `json:"speed"`
	ETA             int64   `json:"eta"`
}

// ParseLine interprets a single line of yt-dlp stdout. It never errors:
// anything that isn't a well-formed progress line (including malformed
// JSON, which can happen mid-stream if a line gets split, or plain text
// during the post-processing phase where yt-dlp/ffmpeg print human-
// readable log lines) is returned as a StatusPostProcessing event carrying
// the raw line, rather than aborting the whole parse.
//
// Note there is no literal prefix (like "download:") to strip here even
// though progressTemplate's CLI value has one: that prefix is yt-dlp's own
// routing directive selecting which progress phase the template applies
// to, and yt-dlp consumes it before printing — it never appears in the
// actual output line.
func ParseLine(line string) ProgressEvent {
	var raw rawProgressJSON
	if err := json.Unmarshal([]byte(line), &raw); err == nil && raw.Status == "downloading" {
		return ProgressEvent{
			Status:          StatusDownloading,
			DownloadedBytes: raw.DownloadedBytes,
			TotalBytes:      raw.TotalBytes,
			Speed:           raw.Speed,
			ETA:             raw.ETA,
		}
	}
	return ProgressEvent{Status: StatusPostProcessing, Message: line}
}
