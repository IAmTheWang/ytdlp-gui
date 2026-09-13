<p align="center">
  <a href="README.md">English</a> | <b>日本語</b> | <a href="README.zh.md">中文</a>
</p>

# yt-dlp GUI

[yt-dlp](https://github.com/yt-dlp/yt-dlp) のための軽量デスクトップ GUI。
[Wails](https://wails.io)（Go バックエンド + React/TypeScript フロントエンド）で構築されており、
単一の Windows 実行ファイルとして配布できます。

GUI 自体は **英語・日本語・中国語** に対応しています（デフォルト：英語）。設定画面からいつでも切り替え可能です。

## 主な機能

- 動画または**プレイリスト**の URL を貼り付けてメタデータ（タイトル・サムネイル・再生時間・利用可能なフォーマット）を取得
- 単一動画は画質/フォーマットを選択、プレイリストはダウンロードする動画にチェックを入れて選択
- 進捗・速度・残り時間をリアルタイム表示する並列ダウンロードキュー（個別キャンセル対応）
- 設定：ダウンロード先フォルダ、出力ファイル名テンプレート、同時ダウンロード数、プロキシ、Cookie
  取得元（ブラウザ or `cookies.txt`）、yt-dlp/ffmpeg パスの手動指定、ダーク/ライトテーマ、GUI 言語

## 必要環境

このアプリは `yt-dlp` や `ffmpeg` を同梱していません。別途インストールし、`PATH` に通すか、設定画面の
yt-dlp/ffmpeg パス欄に直接指定してください。

- [yt-dlp](https://github.com/yt-dlp/yt-dlp/releases)（頻繁に更新されるため、最新版を維持してください。
  古いバージョンだと取得できるフォーマット数が減ったり、抽出自体に失敗したりします）
- [ffmpeg](https://ffmpeg.org/download.html)（映像/音声ストリームの結合や音声のみの抽出に必要）

## 開発

Go 1.25 以上、Node 22 以上、および [Wails CLI](https://wails.io/docs/gettingstarted/installation)
が必要です：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

```bash
# ホットリロード付きで起動
wails dev

# バックエンドのテストを実行
go test ./...

# フロントエンドの型チェック
cd frontend && bunx tsc --noEmit

# 本番用実行ファイルをビルド -> build/bin/ytdlp-gui.exe
wails build
```

アーキテクチャの詳細や開発中に判明した注意点は [CLAUDE.md](CLAUDE.md)（英語）を参照してください。

## プロジェクト構成

```
app.go, main.go       Wails のエントリポイント／フロントエンドに公開する API
internal/ytdlp/       yt-dlp の引数構築、メタデータ・進捗のパース
internal/binmanager/  yt-dlp/ffmpeg の実体を探索
internal/downloader/  並列数を制限したダウンロードワーカープール
internal/config/      設定の永続化
frontend/src/         React + TypeScript の UI
```
