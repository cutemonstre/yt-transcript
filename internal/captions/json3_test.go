package captions

import "testing"

func TestParseSkipsFillersCleansTextAndClampsOverlaps(t *testing.T) {
	raw := []byte(`{"events":[
		{"tStartMs":0,"dDurationMs":5000},
		{"tStartMs":1000,"dDurationMs":3000,"segs":[{"utf8":"hello\n"},{"utf8":"world"}]},
		{"tStartMs":2000,"dDurationMs":1000,"aAppend":1,"segs":[{"utf8":"\n"}]},
		{"tStartMs":3500,"dDurationMs":2000,"segs":[{"utf8":"next line"}]}
	]}`)

	segments, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(segments))
	}
	if segments[0].Text != "hello world" {
		t.Fatalf("got text %q", segments[0].Text)
	}
	if segments[0].EndMs != 3500 {
		t.Fatalf("overlap was not clamped: end=%d", segments[0].EndMs)
	}
}

func TestParseRejectsInvalidJSON(t *testing.T) {
	if _, err := Parse([]byte(`{"events":`)); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
