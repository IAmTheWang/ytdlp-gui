package downloader

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"ytdlp-gui/internal/ytdlp"
)

// jobHandle tracks the bits of in-flight state Manager needs beyond what
// gets reported through JobState: how to cancel it, whether it *was*
// canceled (so a context.Canceled error from the runner is reported as
// StatusCanceled rather than StatusFailed), and when it started (for
// scoping partial-file cleanup).
type jobHandle struct {
	cancel    context.CancelFunc
	canceled  atomic.Bool
	startedAt time.Time
}

// Manager runs download jobs through a concurrency-limited worker pool.
type Manager struct {
	mu         sync.Mutex
	sem        chan struct{}
	runner     ProcessRunner
	onProgress ProgressCallback
	jobs       map[string]*jobHandle
}

// NewManager creates a Manager that runs at most `concurrency` jobs at
// once (values below 1 are treated as 1), reporting progress through
// onProgress.
func NewManager(concurrency int, onProgress ProgressCallback) *Manager {
	return newManager(concurrency, onProgress, execRunner{})
}

// newManager is the constructor tests use to inject a fake ProcessRunner
// instead of spawning a real subprocess.
func newManager(concurrency int, onProgress ProgressCallback, runner ProcessRunner) *Manager {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Manager{
		sem:        make(chan struct{}, concurrency),
		runner:     runner,
		onProgress: onProgress,
		jobs:       make(map[string]*jobHandle),
	}
}

// SetConcurrency changes how many jobs may run at once. It only affects
// jobs that acquire a slot after this call.
func (m *Manager) SetConcurrency(n int) {
	if n < 1 {
		n = 1
	}
	m.mu.Lock()
	m.sem = make(chan struct{}, n)
	m.mu.Unlock()
}

func (m *Manager) semaphore() chan struct{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sem
}

// Enqueue registers a job and starts working on it in the background as
// soon as a concurrency slot is free.
func (m *Manager) Enqueue(job Job) {
	ctx, cancel := context.WithCancel(context.Background())
	h := &jobHandle{cancel: cancel, startedAt: time.Now()}

	m.mu.Lock()
	m.jobs[job.ID] = h
	m.mu.Unlock()

	m.emit(JobState{JobID: job.ID, Status: StatusQueued})
	go m.run(ctx, job, h)
}

// Cancel requests cancellation of a queued or running job. Canceling a job
// that has already finished (or doesn't exist) is a no-op.
func (m *Manager) Cancel(jobID string) {
	m.mu.Lock()
	h, ok := m.jobs[jobID]
	m.mu.Unlock()
	if !ok {
		return
	}
	h.canceled.Store(true)
	h.cancel()
}

func (m *Manager) run(ctx context.Context, job Job, h *jobHandle) {
	defer func() {
		m.mu.Lock()
		delete(m.jobs, job.ID)
		m.mu.Unlock()
	}()

	sem := m.semaphore()
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		m.emit(JobState{JobID: job.ID, Status: StatusCanceled})
		return
	}

	// The job may have been canceled in the brief window between
	// Enqueue and acquiring a slot.
	if ctx.Err() != nil {
		m.emit(JobState{JobID: job.ID, Status: StatusCanceled})
		return
	}

	m.emit(JobState{JobID: job.ID, Status: StatusDownloading})

	err := m.runner.Run(ctx, job, func(line string) {
		m.emit(toJobState(job.ID, ytdlp.ParseLine(line)))
	})

	switch {
	case h.canceled.Load():
		if job.CleanupPartialFiles && job.Options.OutputDir != "" {
			_, _ = CleanupPartialFiles(job.Options.OutputDir, h.startedAt)
		}
		m.emit(JobState{JobID: job.ID, Status: StatusCanceled})
	case err != nil:
		m.emit(JobState{JobID: job.ID, Status: StatusFailed, Message: err.Error()})
	default:
		m.emit(JobState{JobID: job.ID, Status: StatusCompleted})
	}
}

func toJobState(jobID string, event ytdlp.ProgressEvent) JobState {
	state := JobState{JobID: jobID}
	if event.Status == ytdlp.StatusDownloading {
		state.Status = StatusDownloading
		state.DownloadedBytes = event.DownloadedBytes
		state.TotalBytes = event.TotalBytes
		state.Percent = event.Percent()
		state.Speed = event.Speed
		state.ETA = event.ETA
	} else {
		state.Status = StatusPostProcessing
		state.Message = event.Message
	}
	return state
}

func (m *Manager) emit(state JobState) {
	if m.onProgress != nil {
		m.onProgress(state)
	}
}
