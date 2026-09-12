package downloader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func touch(t *testing.T, dir, name string, mtime time.Time) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create %s: %v", path, err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("failed to set mtime on %s: %v", path, err)
	}
	return path
}

func TestCleanupPartialFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	since := now.Add(-1 * time.Minute)

	newPart := touch(t, dir, "video.mp4.part", now)
	newYtdl := touch(t, dir, "video.info.json.ytdl", now)
	newFrag := touch(t, dir, "video.f137.part-Frag12", now)
	oldPart := touch(t, dir, "old.mp4.part", now.Add(-1*time.Hour))
	keep := touch(t, dir, "keep.txt", now)

	removed, err := CleanupPartialFiles(dir, since)
	if err != nil {
		t.Fatalf("CleanupPartialFiles returned error: %v", err)
	}

	wantRemoved := map[string]bool{newPart: true, newYtdl: true, newFrag: true}
	if len(removed) != len(wantRemoved) {
		t.Fatalf("removed = %v, want exactly %v", removed, wantRemoved)
	}
	for _, path := range removed {
		if !wantRemoved[path] {
			t.Errorf("unexpectedly removed %q", path)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected %q to be deleted from disk", path)
		}
	}

	for _, path := range []string{oldPart, keep} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %q to still exist, got err: %v", path, err)
		}
	}
}

func TestCleanupPartialFiles_NonexistentDir(t *testing.T) {
	if _, err := CleanupPartialFiles(filepath.Join(t.TempDir(), "does-not-exist"), time.Now()); err == nil {
		t.Errorf("expected an error for a nonexistent directory")
	}
}
