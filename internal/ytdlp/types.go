// Package ytdlp builds yt-dlp command-line arguments and parses its JSON
// output. Everything here is pure (no subprocess execution), so it can be
// unit tested without yt-dlp installed.
package ytdlp

// CookieSourceKind selects how yt-dlp should read cookies for sites that
// gate downloads behind a login/bot check.
type CookieSourceKind string

const (
	CookieSourceNone    CookieSourceKind = "none"
	CookieSourceBrowser CookieSourceKind = "browser"
	CookieSourceFile    CookieSourceKind = "file"
)

// CookieSource configures --cookies-from-browser / --cookies.
type CookieSource struct {
	Kind CookieSourceKind
	// Browser is e.g. "chrome", "edge", "firefox". Used when Kind == CookieSourceBrowser.
	Browser string
	// FilePath is a path to a cookies.txt file. Used when Kind == CookieSourceFile.
	FilePath string
}

// DownloadOptions describes a single download job in terms a caller (the
// downloader package) understands; BuildDownloadArgs turns it into the
// literal argv for yt-dlp.
type DownloadOptions struct {
	URL string

	// FormatID is passed to yt-dlp's -f flag, e.g. "137+140" or "best".
	// Empty means let yt-dlp pick its default.
	FormatID string

	OutputDir      string
	OutputTemplate string

	AudioOnly   bool
	AudioFormat string // e.g. "mp3"; only used when AudioOnly is true

	Subtitles bool
	SubLangs  string // e.g. "en.*,ja.*"
	EmbedSubs bool

	// Playlist controls whether a playlist URL downloads every entry
	// (true) or only the single video it points at (false -> --no-playlist).
	Playlist      bool
	PlaylistItems string // e.g. "1,3,5-7" -> --playlist-items

	Proxy   string
	Cookies CookieSource

	// ExtraArgs is an escape hatch appended verbatim before the URL.
	ExtraArgs []string
}
