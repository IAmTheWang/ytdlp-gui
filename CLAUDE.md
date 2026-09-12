# CLAUDE.md

This file guides Claude Code (or any future contributor) working in this repository.

## What this is

A desktop GUI for [yt-dlp](https://github.com/yt-dlp/yt-dlp), built with **Wails v2** (Go backend + React/TypeScript frontend, packaged as a single native Windows executable). It wraps a locally-installed `yt-dlp` + `ffmpeg` — it does not bundle them.

## Architecture

```
main.go / app.go        Wails entry point + App struct (methods exposed to the frontend)
internal/ytdlp/         Pure functions: build yt-dlp CLI args, parse -J metadata JSON,
                         parse --progress-template output. No subprocess execution except fetch.go.
internal/binmanager/    Locates yt-dlp/ffmpeg: configured path -> PATH -> known fallback dirs.
internal/downloader/    Concurrency-limited worker pool that runs download jobs. Decoupled from
                         Wails (plain ProgressCallback) so it's unit-testable without a Wails runtime.
internal/config/        Settings persistence (JSON at os.UserConfigDir()/ytdlp-gui/config.json),
                         with Sanitize() filling in defaults for blank/invalid fields.
frontend/src/           React + TypeScript UI. wailsjs/ is auto-generated — never hand-edit it.
frontend/src/i18n/      Lightweight custom i18n (English/Japanese/Chinese), no external library.
```

`app.go`'s `App` struct is the only place that imports both `internal/downloader` and
`github.com/wailsapp/wails/v2/pkg/runtime` — it's the adapter that turns `downloader.JobState`
callbacks into `runtime.EventsEmit` calls. Keep it that way; don't let `internal/*` packages import
Wails directly, or they stop being testable without a running Wails app.

## Commands

```bash
# Backend
go build ./...
go vet ./...
go test ./...
gofmt -l .                       # should print nothing

# Frontend (run from frontend/)
npx tsc --noEmit                 # typecheck

# Whole app
wails dev                        # hot-reload dev mode; also serves the UI over plain HTTP
                                  # at the printed "Frontend DevServer URL" / dev server port —
                                  # useful for browser-based testing/automation since you can
                                  # call window.go.main.App.<Method>(...) directly from devtools.
wails build                      # produces build/bin/ytdlp-gui.exe
wails generate module            # regenerate frontend/wailsjs/* after changing any App method
                                  # signature or return type — do this every time or the frontend
                                  # bindings silently go stale.
```

There is no test runner configured for the frontend; correctness there is checked via `tsc` plus
manual/dev-server verification.

## Conventions

- **Go tests are colocated** (`*_test.go` next to the code) and are pure/fast — no real subprocess
  or network calls. `internal/downloader`'s tests use a fake `ProcessRunner` (see `manager_test.go`)
  instead of spawning yt-dlp.
- **Every `App` method that's called from the frontend needs `wails generate module` re-run**
  afterward; the generated TS in `frontend/wailsjs/` is not watched automatically for Go-side
  signature changes in all cases.
- Event payloads (e.g. `download:progress` carrying a `downloader.JobState`) are **not**
  auto-generated as TS types since they're never a bound method's return value — see
  `frontend/src/types/download.ts`, which is hand-maintained to mirror the Go struct.
- i18n: every user-facing string goes through `useI18n().t(key)` — see
  `frontend/src/i18n/translations.ts` for the key list and `en`/`ja`/`zh` dictionaries. Add a key to
  all three language objects, not just one (TypeScript enforces this via `typeof en`).

## Known gotchas (found the hard way — verify before trusting related assumptions)

- **yt-dlp's `--progress-template` fallback syntax is `%(fieldA,fieldB)d` (comma), not
  `%(fieldA|%(fieldB)d)d`.** The pipe syntax exists but only takes a *literal* default value after
  it (e.g. `%(field|0)d`), not another field reference. Getting this wrong makes yt-dlp error out on
  every real download while still passing anything that doesn't actually invoke a subprocess.
- **The `"download:"` prefix in the `--progress-template` value never appears in yt-dlp's actual
  stdout.** It's yt-dlp's own directive selecting which progress phase the template applies to
  (default is `download` even without it), and yt-dlp strips it before printing. Don't try to match
  on it when parsing lines — see `internal/ytdlp/progress.go`'s `ParseLine`, which instead validates
  the parsed JSON has `"status":"downloading"`.
- **Canceling a job must kill the whole process tree, not just yt-dlp's own PID.** yt-dlp spawns
  ffmpeg as a child for merging/transcoding; killing only the parent leaves ffmpeg orphaned, holding
  the unfinished output file. See `internal/downloader/runner.go`'s `killProcessTree` (Windows:
  `taskkill /F /T`).
- **`--flat-playlist -J` is safe to always use** for the initial metadata fetch: on a single (non-
  playlist) video URL it's a no-op and still returns full `formats`, while on an actual playlist URL
  it returns `_type: "playlist"` with flat `entries` (no per-entry formats) — confirmed against real
  yt-dlp output, not just docs. This is why the frontend always calls `FetchPlaylist`, never
  `FetchMetadata`, for the initial URL submit.
- **Binary detection order is configured path → PATH → known fallback dirs.** On the original dev
  machine this repo was built on, PATH held a stale pip-installed yt-dlp (2023.11.16) while a fresh
  standalone copy lived outside PATH — meaning the *default* behavior silently prefers the stale
  one unless the user sets an explicit override in Settings. If format lists look suspiciously short
  or extraction fails oddly, check Settings → yt-dlp path / the `bin-status` readout before assuming
  a code bug.
- **Storyboard/thumbnail-sprite formats** (`format_note: "storyboard"`, `ext: "mhtml"`) report
  `vcodec: "none"` *and* `acodec: "none"`. Filter these out before rendering a format picker, or
  they get misclassified as audio-only (see `FormatPicker.tsx`'s `downloadable` filter).
