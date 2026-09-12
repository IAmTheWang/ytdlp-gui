// Package downloader runs yt-dlp download jobs through a concurrency-limited
// worker pool, reporting progress through a plain callback rather than
// depending on Wails: the caller (app.go) is the only place that needs to
// know about runtime.EventsEmit, which keeps this package unit-testable
// without a Wails runtime context.
package downloader

import "ytdlp-gui/internal/ytdlp"

// Status is a job's lifecycle state.
type Status string

const (
	StatusQueued         Status = "queued"
	StatusDownloading    Status = "downloading"
	StatusPostProcessing Status = "post_processing"
	StatusCompleted      Status = "completed"
	StatusCanceled       Status = "canceled"
	StatusFailed         Status = "failed"
)

// JobState is one progress update for a job, delivered to a
// ProgressCallback.
type JobState struct {
	JobID           string
	Status          Status
	DownloadedBytes int64
	TotalBytes      int64
	Percent         float64
	Speed           float64
	ETA             int64
	// Message holds the post-processing log line (Status ==
	// StatusPostProcessing) or the failure reason (Status == StatusFailed).
	Message string
}

// Job describes a single download to run.
type Job struct {
	ID      string
	BinPath string
	Options ytdlp.DownloadOptions
	// CleanupPartialFiles, if true, removes leftover .part/.ytdl files from
	// Options.OutputDir (created since the job started) when the job is
	// canceled.
	CleanupPartialFiles bool
}

// ProgressCallback receives every JobState update. It must not block for
// long, since it's called synchronously from the job's goroutine.
type ProgressCallback func(JobState)
