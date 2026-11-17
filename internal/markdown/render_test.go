package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/cutemonstre/yt-transcript/internal/transcribe"
	"github.com/cutemonstre/yt-transcript/internal/ytdlp"
)

func TestRenderContract(t *testing.T) {
	meta := ytdlp.Meta{
		ID:          "zjkBMFhNj_g",
		Title:       "A: title",
		Channel:     "Channel",
		Duration:    65,
		UploadDate:  "20231122",
		WebpageURL:  "https://www.youtube.com/watch?v=zjkBMFhNj_g",
		Description: "A description",
		Chapters: []ytdlp.Chapter{
			{StartTime: 31, Title: "Second"},
			{StartTime: 0, Title: "First"},
		},
	}
	segments := []transcribe.Segment{
		{StartMs: 0, EndMs: 1000, Text: "zero"},
		{StartMs: 15000, EndMs: 16000, Text: "fifteen"},
		{StartMs: 31000, EndMs: 32000, Text: "thirty one"},
	}
	doc := string(Render(meta, segments, "youtube-captions", 30*time.Second))

	for _, want := range []string{
		"title: \"A: title\"",
		"upload_date: 2023-11-22",
		"transcript_source: youtube-captions",
		"- [0:00](https://youtu.be/zjkBMFhNj_g?t=0s) First",
		"### [0:31](https://youtu.be/zjkBMFhNj_g?t=31s) Second",
		"**[0:00](https://youtu.be/zjkBMFhNj_g?t=0s)** zero fifteen",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("rendered document missing %q", want)
		}
	}
	if strings.Index(doc, "### [0:00]") > strings.Index(doc, "### [0:31]") {
		t.Fatal("chapter headings are not chronological")
	}
}
