<p align="center">
  <a href="README.md">English</a> | <a href="README.ja.md">日本語</a> | <b>中文</b>
</p>

# yt-dlp GUI

一个轻量级的 [yt-dlp](https://github.com/yt-dlp/yt-dlp) 桌面图形界面，使用 [Wails](https://wails.io)
构建（Go 后端 + React/TypeScript 前端），打包为单个 Windows 原生可执行文件。

界面本身支持**英文、日文、中文**（默认英文），可随时在设置页切换。

## 功能

- 粘贴视频或**播放列表**链接，获取元信息（标题、缩略图、时长、可用格式）
- 单个视频可选择清晰度/格式；播放列表可勾选要下载的视频
- 支持并发的下载队列，实时显示进度、速度、剩余时间，可单独取消
- 设置项：下载目录、输出文件名模板、并发数、代理、Cookie 来源（浏览器或 `cookies.txt`）、
  手动指定 yt-dlp/ffmpeg 路径、深色/浅色主题、界面语言

## 环境要求

本应用**不会**内置 `yt-dlp` 或 `ffmpeg`，需要自行安装后加入 `PATH`，或直接在设置页的
yt-dlp/ffmpeg 路径栏里指定：

- [yt-dlp](https://github.com/yt-dlp/yt-dlp/releases)（请保持更新——各网站经常变动，版本过旧会导致
  可获取的格式变少，甚至直接解析失败）
- [ffmpeg](https://ffmpeg.org/download.html)（用于合并音视频流、提取纯音频）

## 开发

需要 Go 1.25+、Node 22+ 以及 [Wails CLI](https://wails.io/docs/gettingstarted/installation)：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

```bash
# 热重载开发模式
wails dev

# 运行后端测试
go test ./...

# 前端类型检查
cd frontend && bunx tsc --noEmit

# 打包生产环境可执行文件 -> build/bin/ytdlp-gui.exe
wails build
```

架构说明与开发过程中踩过的坑见 [CLAUDE.md](CLAUDE.md)（英文）。

## 项目结构

```
app.go, main.go       Wails 入口 / 暴露给前端调用的 API
internal/ytdlp/       yt-dlp 参数构建、元数据与进度解析
internal/binmanager/  探测 yt-dlp/ffmpeg 的实际位置
internal/downloader/  带并发限制的下载 worker pool
internal/config/      设置持久化
frontend/src/         React + TypeScript 前端界面
```
