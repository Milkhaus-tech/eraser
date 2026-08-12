package main

import (
	"testing"
	"time"

	"github.com/eraser-privacy/eraser/internal/history"
)

func TestShouldSkipBroker(t *testing.T) {
	now := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	sent := func(daysAgo int) *history.Record {
		return &history.Record{Status: history.StatusSent, SentAt: now.AddDate(0, 0, -daysAgo)}
	}

	cases := []struct {
		name string
		last *history.Record
		days int
		want bool
	}{
		{"never contacted", nil, 180, false},
		{"sent yesterday", sent(1), 180, true},
		{"sent inside cooldown", sent(179), 180, true},
		{"sent past cooldown", sent(181), 180, false},
		{"previous send failed", &history.Record{Status: history.StatusFailed, SentAt: now.AddDate(0, 0, -1)}, 180, false},
		{"negative cooldown never resends", sent(9999), -1, true},
		{"zero cooldown always resends", sent(0), 0, false},
		{"missing timestamp retries", &history.Record{Status: history.StatusSent}, 180, false},
	}

	for _, c := range cases {
		if got := shouldSkipBroker(c.last, c.days, now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
