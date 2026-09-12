package ytdlp

import "encoding/json"

// Format is one entry of yt-dlp's "formats" array from -J output. Fields
// are pointers where yt-dlp commonly omits them, so a missing field is
// distinguishable from an explicit zero.
type Format struct {
	FormatID       string   `json:"format_id"`
	Ext            string   `json:"ext"`
	Resolution     string   `json:"resolution"`
	FormatNote     string   `json:"format_note"`
	VCodec         string   `json:"vcodec"`
	ACodec         string   `json:"acodec"`
	FPS            *float64 `json:"fps"`
	Filesize       *int64   `json:"filesize"`
	FilesizeApprox *int64   `json:"filesize_approx"`
	TBR            *float64 `json:"tbr"` // total bitrate, kbit/s
}

// EstimatedSize returns the best available byte-size estimate for this
// format, and whether it is exact. Some platforms (e.g. Bilibili, some
// YouTube streams) never report filesize, only filesize_approx or a
// bitrate; without this fallback chain the UI would show 0/N/A for those.
// durationSeconds is the parent video's duration, used to derive a size
// from bitrate when nothing else is available.
func (f Format) EstimatedSize(durationSeconds float64) (bytes int64, exact bool) {
	if f.Filesize != nil && *f.Filesize > 0 {
		return *f.Filesize, true
	}
	if f.FilesizeApprox != nil && *f.FilesizeApprox > 0 {
		return *f.FilesizeApprox, false
	}
	if f.TBR != nil && *f.TBR > 0 && durationSeconds > 0 {
		// tbr is in kbit/s: kbit/s * s / 8 bits-per-byte * 1000 bits-per-kbit
		return int64(*f.TBR * durationSeconds * 1000 / 8), false
	}
	return 0, false
}

// PlaylistEntry is one item of a --flat-playlist -J "entries" array.
type PlaylistEntry struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
}

// Metadata is the subset of yt-dlp -J output this app cares about. It
// covers both a single video (Formats populated) and a playlist (Type ==
// "playlist", Entries populated).
type Metadata struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Thumbnail string          `json:"thumbnail"`
	Duration  float64         `json:"duration"`
	Formats   []Format        `json:"formats"`
	Type      string          `json:"_type"`
	Entries   []PlaylistEntry `json:"entries"`
}

// IsPlaylist reports whether this metadata describes a playlist listing
// rather than a single video.
func (m Metadata) IsPlaylist() bool {
	return m.Type == "playlist"
}

// ParseMetadataJSON parses the raw output of `yt-dlp -J` (or
// `--flat-playlist -J`) into a Metadata value.
func ParseMetadataJSON(data []byte) (*Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
