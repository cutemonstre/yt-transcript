package captions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cutemonstre/yt-transcript/internal/transcribe"
)

type json3Doc struct {
	Events []struct {
		TStartMs    int64 `json:"tStartMs"`
		DDurationMs int64 `json:"dDurationMs"`
		AAppend     int   `json:"aAppend"`
		Segs        []struct {
			UTF8 string `json:"utf8"`
		} `json:"segs"`
	} `json:"events"`
}

// Parse converts YouTube json3 captions into flat transcript segments.
func Parse(raw []byte) ([]transcribe.Segment, error) {
	var doc json3Doc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("json3 parse: %w", err)
	}
	segs := make([]transcribe.Segment, 0, len(doc.Events))
	for _, event := range doc.Events {
		if event.AAppend == 1 {
			continue
		}
		var text strings.Builder
		for _, segment := range event.Segs {
			text.WriteString(segment.UTF8)
		}
		clean := strings.TrimSpace(strings.ReplaceAll(text.String(), "\n", " "))
		if clean == "" {
			continue
		}
		segs = append(segs, transcribe.Segment{
			StartMs: event.TStartMs,
			EndMs:   event.TStartMs + event.DDurationMs,
			Text:    clean,
		})
	}
	return transcribe.ClampOverlaps(segs), nil
}
