package ytdlp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ErrNoCaptions = errors.New("no captions available")

type Chapter struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Title     string  `json:"title"`
}

type Meta struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Channel     string    `json:"channel"`
	Uploader    string    `json:"uploader"`
	Duration    float64   `json:"duration"`
	UploadDate  string    `json:"upload_date"`
	ViewCount   int64     `json:"view_count"`
	Description string    `json:"description"`
	WebpageURL  string    `json:"webpage_url"`
	Chapters    []Chapter `json:"chapters"`
}

func (m Meta) DurationString() string {
	d := time.Duration(m.Duration) * time.Second
	if d >= time.Hour {
		return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

func Metadata(ctx context.Context, url string) (Meta, error) {
	if err := requireYTDLP(); err != nil {
		return Meta{}, err
	}
	out, err := exec.CommandContext(ctx, "yt-dlp", "--skip-download", "--dump-json", url).Output()
	if err != nil {
		return Meta{}, ytdlpErr("dump-json", err)
	}
	var meta Meta
	if err := json.Unmarshal(out, &meta); err != nil {
		return Meta{}, fmt.Errorf("parsing yt-dlp metadata: %w", err)
	}
	return meta, nil
}

func Captions(ctx context.Context, url, lang string) ([]byte, string, error) {
	if err := requireYTDLP(); err != nil {
		return nil, "", err
	}
	dir, err := os.MkdirTemp("", "yt_subs_*")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(dir)

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--skip-download",
		"--write-subs", "--write-auto-subs",
		"--sub-langs", lang,
		"--sub-format", "json3/vtt",
		"-o", filepath.Join(dir, "cap.%(ext)s"),
		url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("yt-dlp captions: %w\n%s", err, out)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "cap.*.json3"))
	if len(matches) == 0 {
		return nil, "", ErrNoCaptions
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, "", err
	}
	return raw, "youtube-captions", nil
}

func Audio(ctx context.Context, url string) (string, error) {
	if err := requireYTDLP(); err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp("", "yt_audio_*")
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-f", "bestaudio",
		"-x", "--audio-format", "mp3",
		"--postprocessor-args", "ExtractAudio:-ar 16000 -ac 1 -b:a 24k",
		"-o", filepath.Join(dir, "audio.%(ext)s"),
		url)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("yt-dlp audio: %w\n%s", err, out)
	}
	audioPath := filepath.Join(dir, "audio.mp3")
	if _, err := os.Stat(audioPath); err != nil {
		os.RemoveAll(dir)
		return "", errors.New("yt-dlp reported success but no audio.mp3 produced")
	}
	return audioPath, nil
}

func requireYTDLP() error {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return errors.New("yt-dlp is required; install it with: brew install yt-dlp")
	}
	return nil
}

func ytdlpErr(op string, err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("yt-dlp %s: %w\n%s", op, err, exitErr.Stderr)
	}
	return fmt.Errorf("yt-dlp %s: %w", op, err)
}
