package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"ytdlp-gui/internal/binmanager"
	"ytdlp-gui/internal/config"
	"ytdlp-gui/internal/downloader"
	"ytdlp-gui/internal/ytdlp"
)

// knownToolDirs are extra locations worth checking for yt-dlp/ffmpeg beyond
// PATH and the user's configured path. This user keeps a standalone
// yt-dlp/ffmpeg pair alongside other download tools rather than on PATH,
// so it's a sensible default fallback.
var knownToolDirs = []string{
	`E:\software\ytdlp-interface`,
}

// App struct
type App struct {
	ctx        context.Context
	downloader *downloader.Manager

	settingsMu   sync.RWMutex
	settings     config.Settings
	settingsPath string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved so we can
// call the runtime methods. This is the one place that wires downloader's
// plain ProgressCallback to Wails' EventsEmit, so the downloader package
// itself never needs to know Wails exists.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	path, err := config.DefaultPath()
	if err != nil {
		log.Printf("config: could not determine settings path, using in-memory defaults: %v", err)
		a.settings = config.Default()
	} else {
		a.settingsPath = path
		settings, err := config.Load(path)
		if err != nil {
			log.Printf("config: failed to load %s, using defaults: %v", path, err)
			settings = config.Default()
		}
		a.settings = settings
	}

	a.downloader = downloader.NewManager(a.settings.Concurrency, func(state downloader.JobState) {
		wailsruntime.EventsEmit(ctx, "download:progress", state)
	})
}

// GetSettings returns the current settings.
func (a *App) GetSettings() config.Settings {
	a.settingsMu.RLock()
	defer a.settingsMu.RUnlock()
	return a.settings
}

// SaveSettings sanitizes, persists, and applies new settings — including
// live-updating the downloader's concurrency limit, so changes take effect
// without restarting the app.
func (a *App) SaveSettings(s config.Settings) error {
	s = s.Sanitize()

	a.settingsMu.Lock()
	a.settings = s
	a.settingsMu.Unlock()

	a.downloader.SetConcurrency(s.Concurrency)

	if a.settingsPath == "" {
		return fmt.Errorf("settings path unavailable; changes will not persist across restarts")
	}
	return config.Save(a.settingsPath, s)
}

// ChooseDownloadDir opens a native directory picker. Settings dialogs use
// this (and ChooseFile below) instead of an HTML file input, since a
// browser-sandboxed <input type="file"> can't report a real absolute path.
// title is supplied by the frontend so the dialog respects the current GUI
// language.
func (a *App) ChooseDownloadDir(title string) (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: title,
	})
}

// ChooseFile opens a native file picker, e.g. for pointing at a yt-dlp.exe,
// ffmpeg.exe, or cookies.txt.
func (a *App) ChooseFile(title string) (string, error) {
	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: title,
	})
}

// CheckBinaries probes for yt-dlp and ffmpeg: the user's configured
// override path first, then PATH, then knownToolDirs.
func (a *App) CheckBinaries() binmanager.BinaryStatus {
	settings := a.GetSettings()
	return binmanager.Check(binmanager.ProbeConfig{
		YtDlpPath:    settings.YtDlpPath,
		FFmpegPath:   settings.FFmpegPath,
		FallbackDirs: knownToolDirs,
	})
}

// resolveYtDlpPath locates yt-dlp or returns an error describing that it's
// missing, so callers can surface a single clear message to the user
// instead of a raw "executable file not found" error.
func (a *App) resolveYtDlpPath() (string, error) {
	status := a.CheckBinaries()
	if !status.YtDlpFound {
		return "", fmt.Errorf("yt-dlp not found on PATH or in the known tool directories; configure its path in settings")
	}
	return status.YtDlpPath, nil
}

// FetchMetadata fetches full metadata (title, thumbnail, duration,
// available formats) for a single video URL.
func (a *App) FetchMetadata(url string) (*ytdlp.Metadata, error) {
	binPath, err := a.resolveYtDlpPath()
	if err != nil {
		return nil, err
	}
	return ytdlp.FetchMetadata(a.ctx, binPath, url, false)
}

// FetchPlaylist fetches a quick, flat listing of a playlist's entries
// (id/title/url/duration only, not each entry's full format list) so the
// UI doesn't have to wait on every video's metadata just to show a list.
func (a *App) FetchPlaylist(url string) (*ytdlp.Metadata, error) {
	binPath, err := a.resolveYtDlpPath()
	if err != nil {
		return nil, err
	}
	return ytdlp.FetchMetadata(a.ctx, binPath, url, true)
}

// StartDownload enqueues a download job and returns its job ID. Progress
// updates for it arrive as "download:progress" events carrying a
// downloader.JobState. Any of opts.OutputDir/Proxy/Cookies left at their
// zero value are filled in from the current settings.
func (a *App) StartDownload(opts ytdlp.DownloadOptions) (string, error) {
	binPath, err := a.resolveYtDlpPath()
	if err != nil {
		return "", err
	}

	settings := a.GetSettings()
	if opts.OutputDir == "" {
		opts.OutputDir = settings.DownloadDir
	}
	if opts.OutputTemplate == "" {
		opts.OutputTemplate = settings.OutputTemplate
	}
	if opts.Proxy == "" {
		opts.Proxy = settings.Proxy
	}
	if opts.Cookies.Kind == "" {
		opts.Cookies = settings.Cookies
	}

	id := uuid.NewString()
	a.downloader.Enqueue(downloader.Job{
		ID:                  id,
		BinPath:             binPath,
		Options:             opts,
		CleanupPartialFiles: true,
	})
	return id, nil
}

// CancelDownload requests cancellation of a queued or running download.
func (a *App) CancelDownload(jobID string) {
	a.downloader.Cancel(jobID)
}
