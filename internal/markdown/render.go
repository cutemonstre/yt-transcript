package markdown

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cutemonstre/yt-transcript/internal/transcribe"
	"github.com/cutemonstre/yt-transcript/internal/ytdlp"
)

func stamp(videoID string, ms int64) string {
	sec := ms / 1000
	return fmt.Sprintf("[%s](https://youtu.be/%s?t=%ds)", clock(sec), videoID, sec)
}

func clock(sec int64) string {
	if sec >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

func Render(meta ytdlp.Meta, segs []transcribe.Segment, source string, window time.Duration) []byte {
	var b strings.Builder
	fmt.Fprintln(&b, "---")
	fmt.Fprintf(&b, "video_id: %s\n", meta.ID)
	fmt.Fprintf(&b, "url: %s\n", meta.WebpageURL)
	fmt.Fprintf(&b, "title: %q\n", meta.Title)
	fmt.Fprintf(&b, "channel: %q\n", meta.Channel)
	fmt.Fprintf(&b, "upload_date: %s\n", isoDate(meta.UploadDate))
	fmt.Fprintf(&b, "duration: %s\n", meta.DurationString())
	fmt.Fprintf(&b, "views: %d\n", meta.ViewCount)
	fmt.Fprintf(&b, "transcript_source: %s\n", source)
	fmt.Fprintln(&b, "---")
	b.WriteString("\n")

	fmt.Fprintf(&b, "# %s\n\n", meta.Title)
	fmt.Fprintf(&b, "> **For agents:** every timestamp below is a deep link to that exact moment "+
		"(`https://youtu.be/%s?t=<seconds>s`). When answering questions about this video, cite the "+
		"relevant timestamp links so the reader can jump straight to the snippet instead of watching the whole video.\n\n", meta.ID)

	if len(meta.Chapters) > 0 {
		b.WriteString("## Chapters\n\n")
		for _, chapter := range meta.Chapters {
			fmt.Fprintf(&b, "- %s %s\n", stamp(meta.ID, int64(chapter.StartTime*1000)), chapter.Title)
		}
		b.WriteString("\n")
	}
	if description := strings.TrimSpace(meta.Description); description != "" {
		b.WriteString("## Description\n\n")
		b.WriteString(description)
		b.WriteString("\n\n")
	}
	b.WriteString("## Transcript\n\n")
	renderBlocks(&b, meta, segs, window)
	return []byte(b.String())
}

func renderBlocks(b *strings.Builder, meta ytdlp.Meta, segs []transcribe.Segment, window time.Duration) {
	chapters := append([]ytdlp.Chapter(nil), meta.Chapters...)
	sort.Slice(chapters, func(i, j int) bool { return chapters[i].StartTime < chapters[j].StartTime })
	nextChapter := 0
	windowMs := window.Milliseconds()

	var blockStart int64 = -1
	var block []string
	flush := func() {
		if len(block) == 0 {
			return
		}
		fmt.Fprintf(b, "**%s** %s\n\n", stamp(meta.ID, blockStart), strings.Join(block, " "))
		block = block[:0]
	}

	for _, segment := range segs {
		for nextChapter < len(chapters) && float64(segment.StartMs)/1000 >= chapters[nextChapter].StartTime {
			flush()
			fmt.Fprintf(b, "### %s %s\n\n", stamp(meta.ID, int64(chapters[nextChapter].StartTime*1000)), chapters[nextChapter].Title)
			nextChapter++
			blockStart = -1
		}
		if blockStart == -1 {
			blockStart = segment.StartMs
		}
		block = append(block, segment.Text)
		if segment.StartMs-blockStart >= windowMs {
			flush()
			blockStart = -1
		}
	}
	flush()
}

func isoDate(yyyymmdd string) string {
	if len(yyyymmdd) == 8 {
		return yyyymmdd[:4] + "-" + yyyymmdd[4:6] + "-" + yyyymmdd[6:]
	}
	return yyyymmdd
}
