package downloader

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeBehavior scripts what a fakeRunner does for one job: emit some lines,
// then optionally block on gate (or ctx cancellation) before returning err.
type fakeBehavior struct {
	lines []string
	err   error
	gate  chan struct{}
}

// fakeRunner is a ProcessRunner that never spawns a real subprocess, so
// Manager's scheduling/state-machine logic (queuing, concurrency limits,
// cancel-vs-fail) can be tested without yt-dlp installed.
type fakeRunner struct {
	mu        sync.Mutex
	behaviors map[string]*fakeBehavior
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{behaviors: make(map[string]*fakeBehavior)}
}

func (f *fakeRunner) set(jobID string, b fakeBehavior) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.behaviors[jobID] = &b
}

func (f *fakeRunner) Run(ctx context.Context, job Job, onLine func(string)) error {
	f.mu.Lock()
	b, ok := f.behaviors[job.ID]
	f.mu.Unlock()
	if !ok {
		return nil
	}

	for _, line := range b.lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		onLine(line)
	}

	if b.gate != nil {
		select {
		case <-b.gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return b.err
}

// recorder collects JobState updates in arrival order, safe for concurrent
// use by the Manager's job goroutines.
type recorder struct {
	mu     sync.Mutex
	states []JobState
}

func (r *recorder) record(s JobState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states = append(r.states, s)
}

func (r *recorder) forJob(jobID string) []JobState {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []JobState
	for _, s := range r.states {
		if s.JobID == jobID {
			out = append(out, s)
		}
	}
	return out
}

func (r *recorder) lastStatus(jobID string) (Status, bool) {
	states := r.forJob(jobID)
	if len(states) == 0 {
		return "", false
	}
	return states[len(states)-1].Status, true
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !cond() {
		t.Fatalf("condition not met within %s", timeout)
	}
}

func waitForStatus(t *testing.T, rec *recorder, jobID string, want Status) {
	t.Helper()
	waitFor(t, 2*time.Second, func() bool {
		got, ok := rec.lastStatus(jobID)
		return ok && got == want
	})
}

func TestManager_SuccessfulJob(t *testing.T) {
	runner := newFakeRunner()
	runner.set("job1", fakeBehavior{lines: []string{
		`{"status":"downloading","downloaded_bytes":50,"total_bytes":100,"speed":10.5,"eta":5}`,
		`[Merger] Merging formats into "out.mp4"`,
	}})
	rec := &recorder{}
	m := newManager(2, rec.record, runner)

	m.Enqueue(Job{ID: "job1"})
	waitForStatus(t, rec, "job1", StatusCompleted)

	states := rec.forJob("job1")
	var gotProgress, gotPostProcessing bool
	for _, s := range states {
		if s.Status == StatusDownloading && s.TotalBytes == 100 {
			gotProgress = true
			if s.Percent != 50 {
				t.Errorf("expected Percent=50, got %v", s.Percent)
			}
		}
		if s.Status == StatusPostProcessing {
			gotPostProcessing = true
			if !strings.Contains(s.Message, "Merger") {
				t.Errorf("expected post-processing message to carry the raw line, got %q", s.Message)
			}
		}
	}
	if states[0].Status != StatusQueued {
		t.Errorf("expected first status to be Queued, got %v", states[0].Status)
	}
	if !gotProgress {
		t.Errorf("expected a downloading progress event, got %+v", states)
	}
	if !gotPostProcessing {
		t.Errorf("expected a post-processing event, got %+v", states)
	}
}

func TestManager_FailedJob(t *testing.T) {
	runner := newFakeRunner()
	runner.set("job2", fakeBehavior{err: errors.New("boom")})
	rec := &recorder{}
	m := newManager(2, rec.record, runner)

	m.Enqueue(Job{ID: "job2"})
	waitForStatus(t, rec, "job2", StatusFailed)

	states := rec.forJob("job2")
	last := states[len(states)-1]
	if !strings.Contains(last.Message, "boom") {
		t.Errorf("expected failure message to contain %q, got %q", "boom", last.Message)
	}
}

func TestManager_CancelDuringRun(t *testing.T) {
	runner := newFakeRunner()
	gate := make(chan struct{}) // never closed; job only stops via ctx cancellation
	runner.set("job3", fakeBehavior{gate: gate})
	rec := &recorder{}
	m := newManager(2, rec.record, runner)

	m.Enqueue(Job{ID: "job3"})
	waitForStatus(t, rec, "job3", StatusDownloading)

	m.Cancel("job3")
	waitForStatus(t, rec, "job3", StatusCanceled)

	// Must not be reported as Failed even though the runner returned
	// context.Canceled as its error.
	for _, s := range rec.forJob("job3") {
		if s.Status == StatusFailed {
			t.Errorf("canceled job must never be reported as Failed, got %+v", s)
		}
	}
}

func TestManager_ConcurrencyLimit(t *testing.T) {
	runner := newFakeRunner()
	gateA := make(chan struct{})
	runner.set("jobA", fakeBehavior{gate: gateA})
	runner.set("jobB", fakeBehavior{}) // completes immediately once it gets to run
	rec := &recorder{}
	m := newManager(1, rec.record, runner)

	m.Enqueue(Job{ID: "jobA"})
	waitForStatus(t, rec, "jobA", StatusDownloading)

	m.Enqueue(Job{ID: "jobB"})
	time.Sleep(50 * time.Millisecond) // give jobB's goroutine a chance to (wrongly) start
	if got, _ := rec.lastStatus("jobB"); got != StatusQueued {
		t.Fatalf("expected jobB to still be Queued while concurrency=1 and jobA runs, got %v", got)
	}

	close(gateA)
	waitForStatus(t, rec, "jobA", StatusCompleted)
	waitForStatus(t, rec, "jobB", StatusCompleted)
}

func TestManager_CancelBeforeAcquiringSlot(t *testing.T) {
	runner := newFakeRunner()
	gateX := make(chan struct{})
	runner.set("jobX", fakeBehavior{gate: gateX})
	runner.set("jobY", fakeBehavior{})
	rec := &recorder{}
	m := newManager(1, rec.record, runner)

	m.Enqueue(Job{ID: "jobX"})
	waitForStatus(t, rec, "jobX", StatusDownloading)

	m.Enqueue(Job{ID: "jobY"})
	time.Sleep(50 * time.Millisecond) // let jobY's goroutine reach the semaphore wait

	m.Cancel("jobY")
	waitForStatus(t, rec, "jobY", StatusCanceled)

	for _, s := range rec.forJob("jobY") {
		if s.Status == StatusDownloading {
			t.Errorf("jobY was canceled while queued, should never reach Downloading, got %+v", rec.forJob("jobY"))
		}
	}

	close(gateX)
	waitForStatus(t, rec, "jobX", StatusCompleted)
}
