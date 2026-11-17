package transcribe

import "testing"

func TestClampOverlaps(t *testing.T) {
	segments := []Segment{
		{StartMs: 0, EndMs: 2000},
		{StartMs: 1500, EndMs: 3000},
	}
	got := ClampOverlaps(segments)
	if got[0].EndMs != 1500 || got[1].EndMs != 3000 {
		t.Fatalf("unexpected result: %#v", got)
	}
}
