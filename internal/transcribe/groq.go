package transcribe

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

// Transcribe sends audio to Groq Whisper and returns English segments with timestamps.
func Transcribe(ctx context.Context, apiKey, audioPath string) ([]Segment, error) {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = groqBaseURL
	client := openai.NewClientWithConfig(cfg)

	fmt.Fprint(os.Stderr, "Transcribing with Groq Whisper (whisper-large-v3)...")
	req := openai.AudioRequest{
		Model:    "whisper-large-v3",
		FilePath: audioPath,
		Format:   openai.AudioResponseFormatVerboseJSON,
		Language: "en",
		TimestampGranularities: []openai.TranscriptionTimestampGranularity{
			openai.TranscriptionTimestampGranularitySegment,
		},
		Temperature: 0.0,
	}

	var resp openai.AudioResponse
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			fmt.Fprintf(os.Stderr, " retrying in %s...", backoff)
			time.Sleep(backoff)
		}
		resp, err = client.CreateTranscription(ctx, req)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("groq transcription: %w", err)
	}
	fmt.Fprintln(os.Stderr, " done")

	segments := make([]Segment, 0, len(resp.Segments))
	for _, seg := range resp.Segments {
		if seg.Text == "" {
			continue
		}
		startMs := int64(math.Round(float64(seg.Start) * 1000))
		endMs := int64(math.Round(float64(seg.End) * 1000))
		if endMs <= startMs {
			continue
		}
		segments = append(segments, Segment{StartMs: startMs, EndMs: endMs, Text: seg.Text})
	}
	return ClampOverlaps(segments), nil
}
