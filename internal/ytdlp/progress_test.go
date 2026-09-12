package ytdlp

import "testing"

func TestParseLine_ValidProgressJSON(t *testing.T) {
	// No "download:" prefix here: that's yt-dlp's own routing directive on
	// the --progress-template CLI value, consumed before printing, so it
	// never actually appears in the stdout line (confirmed against a real
	// yt-dlp run).
	line := `{"status":"downloading","downloaded_bytes":512000,"total_bytes":1048576,"speed":102400.5,"eta":5}`
	got := ParseLine(line)

	if got.Status != StatusDownloading {
		t.Fatalf("expected StatusDownloading, got %v", got.Status)
	}
	if got.DownloadedBytes != 512000 || got.TotalBytes != 1048576 || got.ETA != 5 {
		t.Errorf("unexpected fields: %+v", got)
	}
	if got.Speed != 102400.5 {
		t.Errorf("unexpected speed: %v", got.Speed)
	}
	if pct := got.Percent(); pct <= 48.8 || pct >= 48.9 {
		t.Errorf("expected ~48.83%% complete, got %v", pct)
	}
}

func TestParseLine_MalformedJSONFallsBackToPostProcessing(t *testing.T) {
	// Simulates a truncated/corrupted line rather than crashing the parser.
	line := `{"status":"downloading","downloaded_bytes":`
	got := ParseLine(line)

	if got.Status != StatusPostProcessing {
		t.Fatalf("expected StatusPostProcessing for malformed JSON, got %v", got.Status)
	}
	if got.Message != line {
		t.Errorf("expected Message to preserve the raw line, got %q", got.Message)
	}
}

func TestParseLine_UnrelatedJSONFallsBackToPostProcessing(t *testing.T) {
	// Well-formed JSON that doesn't carry our expected status="downloading"
	// marker must not be mistaken for a progress update.
	line := `{"status":"something_else","downloaded_bytes":1}`
	got := ParseLine(line)

	if got.Status != StatusPostProcessing {
		t.Fatalf("expected StatusPostProcessing for unrelated JSON, got %v", got.Status)
	}
}

func TestParseLine_PostProcessingPlainText(t *testing.T) {
	lines := []string{
		`[Merger] Merging formats into "video.mp4"`,
		`[ExtractAudio] Destination: audio.mp3`,
		`[ffmpeg] Adding subtitle`,
		`frame=  120 fps=30 q=-1.0 size=    2048kB time=00:00:04.00 bitrate=4194.3kbits/s`,
	}
	for _, line := range lines {
		got := ParseLine(line)
		if got.Status != StatusPostProcessing {
			t.Errorf("ParseLine(%q).Status = %v, want StatusPostProcessing", line, got.Status)
		}
		if got.Message != line {
			t.Errorf("ParseLine(%q).Message = %q, want original line", line, got.Message)
		}
	}
}

func TestProgressEvent_PercentWithZeroTotal(t *testing.T) {
	e := ProgressEvent{Status: StatusDownloading, DownloadedBytes: 100, TotalBytes: 0}
	if pct := e.Percent(); pct != 0 {
		t.Errorf("expected 0%% when TotalBytes is unknown, got %v", pct)
	}
}
