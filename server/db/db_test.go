package db

import (
	"os"
	"testing"
	"time"
)

func TestNowUTC(t *testing.T) {
	tmp, err := os.CreateTemp("", "serversmith-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if err := Init(tmp.Name()); err != nil {
		t.Fatalf("init db: %v", err)
	}

	n := NowUTC()
	parsed, err := time.Parse(time.RFC3339, n)
	if err != nil {
		t.Fatalf("NowUTC returned invalid RFC3339: %q err=%v", n, err)
	}
	// Must be UTC zone
	if parsed.Location().String() != "UTC" {
		t.Errorf("expected UTC zone, got %s", parsed.Location())
	}
}

func TestBeijingConverters(t *testing.T) {
	// Test that Beijing time is exactly 8 hours ahead of UTC
	bj := time.Now().In(TZ)
	utc := time.Now().UTC()

	_, bjOffset := bj.Zone()
	_, utcOffset := utc.Zone()

	if bjOffset != 8*3600 {
		t.Errorf("expected Beijing offset +0800, got %d", bjOffset)
	}
	if utcOffset != 0 {
		t.Errorf("expected UTC offset +0000, got %d", utcOffset)
	}
}
