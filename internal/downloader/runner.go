package downloader

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ytdlp-gui/internal/ytdlp"
)

// ProcessRunner runs a job's subprocess to completion (or until ctx is
// canceled), calling onLine for every line of stdout. It's an interface
// so the Manager's scheduling/state-machine logic can be unit tested with
// a fake runner instead of a real yt-dlp subprocess.
type ProcessRunner interface {
	Run(ctx context.Context, job Job, onLine func(line string)) error
}

// execRunner is the real ProcessRunner, backed by os/exec.
type execRunner struct{}

func (execRunner) Run(ctx context.Context, job Job, onLine func(line string)) error {
	args := ytdlp.BuildDownloadArgs(job.Options)
	cmd := exec.CommandContext(ctx, job.BinPath, args...)

	// Force UTF-8 so titles/paths with non-ASCII characters don't get
	// mangled by whatever code page the system happens to default to.
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")

	// exec.CommandContext's default cancellation just calls
	// cmd.Process.Kill() on the yt-dlp process itself; if yt-dlp has
	// already spawned ffmpeg to merge/transcode, ffmpeg survives as an
	// orphan holding the unfinished output file open. Killing the whole
	// process tree instead avoids that.
	cmd.Cancel = func() error { return killProcessTree(cmd) }
	cmd.WaitDelay = 5 * time.Second

	var stderr strings.Builder
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		onLine(scanner.Text())
	}

	err = cmd.Wait()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// killProcessTree terminates cmd's process and everything it spawned. On
// Windows, Process.Kill() alone only kills the yt-dlp process itself; if
// it already launched ffmpeg to merge/transcode, ffmpeg would be left
// running and holding the unfinished output file open. taskkill /T walks
// the whole process tree instead.
func killProcessTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	}
	return cmd.Process.Kill()
}
