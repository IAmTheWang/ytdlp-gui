package ytdlp

import "testing"

func TestParseMetadataJSON_SingleVideo(t *testing.T) {
	raw := []byte(`{
		"id": "abc123",
		"title": "Some Title",
		"thumbnail": "https://example.com/thumb.jpg",
		"duration": 125.5,
		"formats": [
			{"format_id": "137", "ext": "mp4", "vcodec": "avc1", "acodec": "none", "filesize": 1048576},
			{"format_id": "140", "ext": "m4a", "vcodec": "none", "acodec": "mp4a", "filesize_approx": 2048000},
			{"format_id": "251", "ext": "webm", "vcodec": "none", "acodec": "opus", "tbr": 160.0}
		]
	}`)

	m, err := ParseMetadataJSON(raw)
	if err != nil {
		t.Fatalf("ParseMetadataJSON returned error: %v", err)
	}
	if m.ID != "abc123" || m.Title != "Some Title" || m.Duration != 125.5 {
		t.Errorf("unexpected metadata fields: %+v", m)
	}
	if m.IsPlaylist() {
		t.Errorf("single video metadata should not report IsPlaylist()")
	}
	if len(m.Formats) != 3 {
		t.Fatalf("expected 3 formats, got %d", len(m.Formats))
	}
}

func TestParseMetadataJSON_Playlist(t *testing.T) {
	raw := []byte(`{
		"_type": "playlist",
		"id": "PL123",
		"title": "My Playlist",
		"entries": [
			{"id": "v1", "title": "Video 1", "url": "https://example.com/v1", "duration": 60},
			{"id": "v2", "title": "Video 2", "url": "https://example.com/v2", "duration": 90}
		]
	}`)

	m, err := ParseMetadataJSON(raw)
	if err != nil {
		t.Fatalf("ParseMetadataJSON returned error: %v", err)
	}
	if !m.IsPlaylist() {
		t.Errorf("expected IsPlaylist() to be true for _type=playlist")
	}
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}
	if m.Entries[0].Title != "Video 1" {
		t.Errorf("unexpected first entry: %+v", m.Entries[0])
	}
}

func TestParseMetadataJSON_Malformed(t *testing.T) {
	if _, err := ParseMetadataJSON([]byte("not json")); err == nil {
		t.Errorf("expected an error for malformed JSON")
	}
}

func TestFormat_EstimatedSize(t *testing.T) {
	i64 := func(v int64) *int64 { return &v }
	f64 := func(v float64) *float64 { return &v }

	tests := []struct {
		name      string
		format    Format
		duration  float64
		wantBytes int64
		wantExact bool
	}{
		{
			name:      "exact filesize wins even if filesize_approx also present",
			format:    Format{Filesize: i64(1000), FilesizeApprox: i64(9999)},
			wantBytes: 1000,
			wantExact: true,
		},
		{
			name:      "falls back to filesize_approx when filesize is absent",
			format:    Format{FilesizeApprox: i64(2048000)},
			wantBytes: 2048000,
			wantExact: false,
		},
		{
			name:      "falls back to filesize_approx when filesize is zero",
			format:    Format{Filesize: i64(0), FilesizeApprox: i64(2048000)},
			wantBytes: 2048000,
			wantExact: false,
		},
		{
			name:      "falls back to tbr*duration when nothing else available",
			format:    Format{TBR: f64(160)}, // 160 kbit/s
			duration:  100,                   // 100s
			wantBytes: 160 * 100 * 1000 / 8,  // = 2,000,000 bytes
			wantExact: false,
		},
		{
			name:      "tbr present but duration unknown yields 0",
			format:    Format{TBR: f64(160)},
			duration:  0,
			wantBytes: 0,
			wantExact: false,
		},
		{
			name:      "nothing available yields 0/inexact",
			format:    Format{},
			wantBytes: 0,
			wantExact: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, gotExact := tt.format.EstimatedSize(tt.duration)
			if gotBytes != tt.wantBytes || gotExact != tt.wantExact {
				t.Errorf("EstimatedSize(%v) = (%d, %v), want (%d, %v)", tt.duration, gotBytes, gotExact, tt.wantBytes, tt.wantExact)
			}
		})
	}
}
