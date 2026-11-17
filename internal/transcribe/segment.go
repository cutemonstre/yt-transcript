package transcribe

// Segment represents a transcript entry with timing in milliseconds.
type Segment struct {
	StartMs int64
	EndMs   int64
	Text    string
}

// ClampOverlaps adjusts segment boundaries so no segment overlaps the next.
func ClampOverlaps(segments []Segment) []Segment {
	for i := 0; i < len(segments)-1; i++ {
		if segments[i].EndMs > segments[i+1].StartMs {
			segments[i].EndMs = segments[i+1].StartMs
		}
	}
	return segments
}
