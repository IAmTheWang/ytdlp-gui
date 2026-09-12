package binmanager

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFakeExe creates an empty (non-executable-content, that's fine since
// Locate only stats the path) file named exeName(name) inside dir.
func writeFakeExe(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, exeName(name))
	if err := os.WriteFile(path, []byte("fake"), 0o755); err != nil {
		t.Fatalf("failed to write fake executable %s: %v", path, err)
	}
	return path
}

func TestLocate_ConfiguredPathWins(t *testing.T) {
	dir := t.TempDir()
	configured := writeFakeExe(t, dir, "yt-dlp")

	got, found := Locate("yt-dlp", configured, nil)
	if !found || got != configured {
		t.Fatalf("Locate() = (%q, %v), want (%q, true)", got, found, configured)
	}
}

func TestLocate_ConfiguredPathMissingFallsThrough(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // isolate from whatever yt-dlp the real machine has on PATH
	fallbackDir := t.TempDir()
	fallback := writeFakeExe(t, fallbackDir, "yt-dlp")

	got, found := Locate("yt-dlp", filepath.Join(t.TempDir(), "does-not-exist.exe"), []string{fallbackDir})
	if !found || got != fallback {
		t.Fatalf("Locate() = (%q, %v), want fallback (%q, true)", got, found, fallback)
	}
}

func TestLocate_FindsOnPATH(t *testing.T) {
	dir := t.TempDir()
	want := writeFakeExe(t, dir, "yt-dlp")
	t.Setenv("PATH", dir)

	got, found := Locate("yt-dlp", "", nil)
	if !found {
		t.Fatalf("Locate() did not find binary on PATH")
	}
	// Compare via Abs to avoid case/short-path differences on Windows.
	gotAbs, _ := filepath.Abs(got)
	wantAbs, _ := filepath.Abs(want)
	if gotAbs != wantAbs {
		t.Fatalf("Locate() = %q, want %q", gotAbs, wantAbs)
	}
}

func TestLocate_FindsInFallbackDir(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir, nothing to find on PATH
	fallbackDir := t.TempDir()
	want := writeFakeExe(t, fallbackDir, "ffmpeg")

	got, found := Locate("ffmpeg", "", []string{t.TempDir(), fallbackDir})
	if !found || got != want {
		t.Fatalf("Locate() = (%q, %v), want (%q, true)", got, found, want)
	}
}

func TestLocate_NotFoundAnywhere(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	got, found := Locate("yt-dlp", "", []string{t.TempDir()})
	if found {
		t.Fatalf("Locate() = (%q, true), want not found", got)
	}
}

func TestCheck_ReportsBothBinaries(t *testing.T) {
	dir := t.TempDir()
	ytdlp := writeFakeExe(t, dir, "yt-dlp")
	t.Setenv("PATH", t.TempDir()) // keep PATH empty so only FallbackDirs matter

	status := Check(ProbeConfig{FallbackDirs: []string{dir}})

	if !status.YtDlpFound || status.YtDlpPath != ytdlp {
		t.Errorf("expected yt-dlp found at %q, got found=%v path=%q", ytdlp, status.YtDlpFound, status.YtDlpPath)
	}
	if status.FFmpegFound {
		t.Errorf("did not expect ffmpeg to be found, got path=%q", status.FFmpegPath)
	}
	if status.Ready() {
		t.Errorf("Ready() should be false when ffmpeg is missing")
	}
}
