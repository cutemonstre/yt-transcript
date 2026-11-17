package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cutemonstre/yt-transcript/internal/ytdlp"
)

func TestVideoID(t *testing.T) {
	want := "zjkBMFhNj_g"
	urls := []string{
		"https://www.youtube.com/watch?v=zjkBMFhNj_g",
		"https://youtu.be/zjkBMFhNj_g?t=20",
		"https://www.youtube.com/shorts/zjkBMFhNj_g",
		"https://www.youtube.com/live/zjkBMFhNj_g",
		"https://www.youtube.com/embed/zjkBMFhNj_g",
	}
	for _, url := range urls {
		got, err := VideoID(url)
		if err != nil || got != want {
			t.Errorf("VideoID(%q) = %q, %v; want %q", url, got, err, want)
		}
	}
	if _, err := VideoID("https://example.com/foo"); err == nil {
		t.Fatal("expected unsupported URL error")
	}
}

func TestWriteReplacesStaleTitleAndFindsAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	meta := ytdlp.Meta{ID: "zjkBMFhNj_g", Title: "Old title"}
	oldPath, err := Write(dir, meta, []byte("old"))
	if err != nil {
		t.Fatal(err)
	}
	meta.Title = "New title"
	newPath, err := Write(dir, meta, []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	if oldPath == newPath {
		t.Fatal("title change did not change path")
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("stale transcript still exists: %v", err)
	}
	if got, ok := Find(dir, meta.ID); !ok || got != newPath || !filepath.IsAbs(got) {
		t.Fatalf("Find = %q, %v; want absolute %q", got, ok, newPath)
	}
}
