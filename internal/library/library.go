package library

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cutemonstre/yt-transcript/internal/ytdlp"
)

var idPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[?&]v=([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`youtu\.be/([A-Za-z0-9_-]{11})`),
	regexp.MustCompile(`/(?:shorts|live|embed)/([A-Za-z0-9_-]{11})`),
}

func VideoID(url string) (string, error) {
	for _, pattern := range idPatterns {
		if match := pattern.FindStringSubmatch(url); match != nil {
			return match[1], nil
		}
	}
	return "", fmt.Errorf("no video ID in %q", url)
}

func Find(dir, videoID string) (string, bool) {
	matches, _ := filepath.Glob(filepath.Join(dir, "*--"+videoID+".md"))
	if len(matches) == 0 {
		return "", false
	}
	abs, err := filepath.Abs(matches[0])
	if err == nil {
		return abs, true
	}
	return matches[0], true
}

func Write(dir string, meta ytdlp.Meta, doc []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, slug(meta.Title)+"--"+meta.ID+".md")
	tmp, err := os.CreateTemp(dir, ".yt-transcript-*.tmp")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return "", err
	}
	if _, err := tmp.Write(doc); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return "", err
	}

	oldMatches, _ := filepath.Glob(filepath.Join(dir, "*--"+meta.ID+".md"))
	for _, oldPath := range oldMatches {
		if oldPath != path {
			_ = os.Remove(oldPath)
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slug(s string) string {
	s = strings.ToLower(s)
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = strings.Trim(s[:80], "-")
	}
	if s == "" {
		s = "video"
	}
	return s
}
