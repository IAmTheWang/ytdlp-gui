<p align="center">
  <b>English</b> | <a href="README.ja.md">日本語</a> | <a href="README.zh.md">中文</a>
</p>

# yt-dlp GUI

A lightweight desktop GUI for [yt-dlp](https://github.com/yt-dlp/yt-dlp), built with
[Wails](https://wails.io) (Go backend + React/TypeScript frontend), packaged as a single native
Windows executable.

The GUI itself is available in **English, Japanese, and Chinese** (default: English), switchable
any time from Settings.

## Features

- Paste a video **or playlist** URL to fetch metadata (title, thumbnail, duration, available
  formats)
- Pick a quality/format for a single video, or check off which videos to grab from a playlist
- Concurrent download queue with live progress, speed, ETA, and per-item cancel
- Settings: download directory, output filename template, concurrency limit, proxy, cookie source
  (browser or `cookies.txt`), manual yt-dlp/ffmpeg path overrides, dark/light theme, GUI language

## Requirements

This app does **not** bundle `yt-dlp` or `ffmpeg` — install them separately and either put them on
your `PATH` or point Settings → yt-dlp/ffmpeg path at them directly:

- [yt-dlp](https://github.com/yt-dlp/yt-dlp/releases) (keep it updated — sites change often, and an
  old build will silently return fewer formats or fail to extract at all)
- [ffmpeg](https://ffmpeg.org/download.html) (needed to merge separate video/audio streams and to
  extract audio-only downloads)

## Development

Requires Go 1.25+, Node 22+, and the [Wails CLI](https://wails.io/docs/gettingstarted/installation):

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

```bash
# Run with hot reload
wails dev

# Run backend tests
go test ./...

# Typecheck the frontend
cd frontend && bunx tsc --noEmit

# Build a production executable -> build/bin/ytdlp-gui.exe
wails build
```

See [CLAUDE.md](CLAUDE.md) for architecture notes and gotchas discovered while building this.

## Project layout

```
app.go, main.go       Wails entry point / bound API exposed to the frontend
internal/ytdlp/       yt-dlp argument building + metadata/progress parsing
internal/binmanager/  Locates yt-dlp/ffmpeg on disk
internal/downloader/  Concurrency-limited download worker pool
internal/config/      Settings persistence
frontend/src/         React + TypeScript UI
```
