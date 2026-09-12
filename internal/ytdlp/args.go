package ytdlp

// progressTemplate is the --progress-template value used for download
// progress reporting. It deliberately spells out each field (rather than
// using the blanket %(progress)j) so json.Unmarshal on the other end has a
// fixed, predictable shape.
//
//   - The "download:" lead-in is yt-dlp's own directive selecting which
//     progress phase this template applies to (it's the default even
//     without this prefix, but spelling it out is clearer); yt-dlp strips
//     it before printing, so it never appears in the actual output line.
//   - "%(a,b)d" is yt-dlp's field-fallback syntax (try a, else b) — used so
//     total_bytes falls back to total_bytes_estimate, since many formats
//     never report an exact size until the download finishes.
//   - "%(field|0)d" supplies a literal default of 0 for fields yt-dlp would
//     otherwise render as the bare word NA (not valid JSON) before it has
//     a value yet, e.g. eta/speed at the very start of a download.
const progressTemplate = `download:{"status":"downloading","downloaded_bytes":%(progress.downloaded_bytes|0)d,"total_bytes":%(progress.total_bytes,progress.total_bytes_estimate|0)d,"speed":%(progress.speed|0)d,"eta":%(progress.eta|0)d}`

// ProgressTemplateArgs returns the --newline/--progress-template flags
// shared by every download invocation.
func ProgressTemplateArgs() []string {
	return []string{"--newline", "--progress-template", progressTemplate}
}

// BuildMetadataArgs builds the argv (without the binary name) for fetching
// metadata as JSON. flatPlaylist should be true when the caller only wants
// a quick listing of playlist entries rather than full per-video metadata.
func BuildMetadataArgs(url string, flatPlaylist bool) []string {
	args := make([]string, 0, 3)
	if flatPlaylist {
		args = append(args, "--flat-playlist")
	}
	args = append(args, "-J", url)
	return args
}

// BuildDownloadArgs builds the argv (without the binary name) for a single
// download job.
func BuildDownloadArgs(opts DownloadOptions) []string {
	args := ProgressTemplateArgs()

	// Resume partial downloads and never re-download a file that's already
	// finished; both matter because network interruptions during a large
	// video are common and re-fetching from scratch wastes bandwidth.
	args = append(args, "-c", "--no-overwrites")
	// Sanitize titles containing characters Windows can't use in filenames
	// (| ? : " < > *) so yt-dlp/ffmpeg never fail trying to write/merge them.
	args = append(args, "--windows-filenames")

	if opts.Playlist {
		if opts.PlaylistItems != "" {
			args = append(args, "--playlist-items", opts.PlaylistItems)
		}
	} else {
		args = append(args, "--no-playlist")
	}

	if opts.FormatID != "" {
		args = append(args, "-f", opts.FormatID)
	}

	if opts.AudioOnly {
		args = append(args, "-x")
		if opts.AudioFormat != "" {
			args = append(args, "--audio-format", opts.AudioFormat)
		}
	}

	if opts.Subtitles {
		args = append(args, "--write-subs")
		if opts.SubLangs != "" {
			args = append(args, "--sub-langs", opts.SubLangs)
		}
		if opts.EmbedSubs {
			args = append(args, "--embed-subs")
		}
	}

	if opts.Proxy != "" {
		args = append(args, "--proxy", opts.Proxy)
	}

	args = append(args, cookieArgs(opts.Cookies)...)

	if opts.OutputDir != "" {
		args = append(args, "-P", opts.OutputDir)
	}
	if opts.OutputTemplate != "" {
		args = append(args, "-o", opts.OutputTemplate)
	}

	args = append(args, opts.ExtraArgs...)
	args = append(args, opts.URL)
	return args
}

func cookieArgs(c CookieSource) []string {
	switch c.Kind {
	case CookieSourceBrowser:
		if c.Browser == "" {
			return nil
		}
		return []string{"--cookies-from-browser", c.Browser}
	case CookieSourceFile:
		if c.FilePath == "" {
			return nil
		}
		return []string{"--cookies", c.FilePath}
	default:
		return nil
	}
}
