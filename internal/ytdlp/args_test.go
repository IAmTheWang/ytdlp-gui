package ytdlp

import (
	"slices"
	"testing"
)

func contains(args []string, s string) bool {
	return slices.Contains(args, s)
}

func containsSeq(args []string, seq ...string) bool {
	if len(seq) == 0 {
		return true
	}
	for i := 0; i+len(seq) <= len(args); i++ {
		if slices.Equal(args[i:i+len(seq)], seq) {
			return true
		}
	}
	return false
}

func TestBuildDownloadArgs_DefaultsAlwaysPresent(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "https://example.com/v"})

	for _, want := range []string{"--newline", "--progress-template", "-c", "--no-overwrites", "--windows-filenames"} {
		if !contains(args, want) {
			t.Errorf("expected args to contain %q, got %v", want, args)
		}
	}

	if !containsSeq(args, "--progress-template", progressTemplate) {
		t.Errorf("expected --progress-template value to be the shared progressTemplate constant, got %v", args)
	}

	if args[len(args)-1] != "https://example.com/v" {
		t.Errorf("expected URL to be the last argument, got %v", args)
	}
}

func TestBuildDownloadArgs_PlaylistFalseAddsNoPlaylist(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", Playlist: false})
	if !contains(args, "--no-playlist") {
		t.Errorf("expected --no-playlist when Playlist is false, got %v", args)
	}
}

func TestBuildDownloadArgs_PlaylistTrueOmitsNoPlaylistAndAddsItems(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", Playlist: true, PlaylistItems: "1,3,5-7"})
	if contains(args, "--no-playlist") {
		t.Errorf("did not expect --no-playlist when Playlist is true, got %v", args)
	}
	if !containsSeq(args, "--playlist-items", "1,3,5-7") {
		t.Errorf("expected --playlist-items 1,3,5-7, got %v", args)
	}
}

func TestBuildDownloadArgs_FormatID(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", FormatID: "137+140"})
	if !containsSeq(args, "-f", "137+140") {
		t.Errorf("expected -f 137+140, got %v", args)
	}
}

func TestBuildDownloadArgs_AudioOnly(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", AudioOnly: true, AudioFormat: "mp3"})
	if !contains(args, "-x") {
		t.Errorf("expected -x for audio-only, got %v", args)
	}
	if !containsSeq(args, "--audio-format", "mp3") {
		t.Errorf("expected --audio-format mp3, got %v", args)
	}
}

func TestBuildDownloadArgs_AudioOnlyWithoutFormat(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", AudioOnly: true})
	if !contains(args, "-x") {
		t.Errorf("expected -x, got %v", args)
	}
	if contains(args, "--audio-format") {
		t.Errorf("did not expect --audio-format when AudioFormat is empty, got %v", args)
	}
}

func TestBuildDownloadArgs_Subtitles(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{
		URL:       "u",
		Subtitles: true,
		SubLangs:  "en.*,ja.*",
		EmbedSubs: true,
	})
	if !contains(args, "--write-subs") {
		t.Errorf("expected --write-subs, got %v", args)
	}
	if !containsSeq(args, "--sub-langs", "en.*,ja.*") {
		t.Errorf("expected --sub-langs en.*,ja.*, got %v", args)
	}
	if !contains(args, "--embed-subs") {
		t.Errorf("expected --embed-subs, got %v", args)
	}
}

func TestBuildDownloadArgs_Proxy(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", Proxy: "socks5://127.0.0.1:1080"})
	if !containsSeq(args, "--proxy", "socks5://127.0.0.1:1080") {
		t.Errorf("expected --proxy socks5://127.0.0.1:1080, got %v", args)
	}
}

func TestBuildDownloadArgs_CookiesBrowser(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{
		URL:     "u",
		Cookies: CookieSource{Kind: CookieSourceBrowser, Browser: "chrome"},
	})
	if !containsSeq(args, "--cookies-from-browser", "chrome") {
		t.Errorf("expected --cookies-from-browser chrome, got %v", args)
	}
}

func TestBuildDownloadArgs_CookiesFile(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{
		URL:     "u",
		Cookies: CookieSource{Kind: CookieSourceFile, FilePath: `E:\software\ytdlp-interface\cookies.txt`},
	})
	if !containsSeq(args, "--cookies", `E:\software\ytdlp-interface\cookies.txt`) {
		t.Errorf("expected --cookies <path>, got %v", args)
	}
}

func TestBuildDownloadArgs_CookiesNoneAddsNothing(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", Cookies: CookieSource{Kind: CookieSourceNone}})
	if contains(args, "--cookies") || contains(args, "--cookies-from-browser") {
		t.Errorf("did not expect cookie args when Kind is none, got %v", args)
	}
}

func TestBuildDownloadArgs_OutputDirAndTemplate(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{
		URL:            "u",
		OutputDir:      `C:\Users\carlos\Downloads`,
		OutputTemplate: "%(title)s.%(ext)s",
	})
	if !containsSeq(args, "-P", `C:\Users\carlos\Downloads`) {
		t.Errorf("expected -P <dir>, got %v", args)
	}
	if !containsSeq(args, "-o", "%(title)s.%(ext)s") {
		t.Errorf("expected -o <template>, got %v", args)
	}
}

func TestBuildDownloadArgs_ExtraArgsBeforeURL(t *testing.T) {
	args := BuildDownloadArgs(DownloadOptions{URL: "u", ExtraArgs: []string{"--verbose"}})
	if !containsSeq(args, "--verbose", "u") {
		t.Errorf("expected ExtraArgs immediately before the URL, got %v", args)
	}
}

func TestBuildMetadataArgs(t *testing.T) {
	if got := BuildMetadataArgs("u", false); !slices.Equal(got, []string{"-J", "u"}) {
		t.Errorf("BuildMetadataArgs(u, false) = %v, want [-J u]", got)
	}
	if got := BuildMetadataArgs("u", true); !slices.Equal(got, []string{"--flat-playlist", "-J", "u"}) {
		t.Errorf("BuildMetadataArgs(u, true) = %v, want [--flat-playlist -J u]", got)
	}
}
