package tui

import (
	"testing"
	"time"
)

func TestTickInterval(t *testing.T) {
	tests := []struct {
		name       string
		idle       time.Duration
		activeMs   int
		idleMs     int
		idleThresh int
		want       time.Duration
	}{
		{"active", 500 * time.Millisecond, 50, 200, 10, 50 * time.Millisecond},
		{"recent", 5 * time.Second, 50, 200, 10, 100 * time.Millisecond},
		{"idle", 15 * time.Second, 50, 200, 10, 200 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaneModel{
				lastActive:     time.Now().Add(-tt.idle),
				activeMs:       tt.activeMs,
				idleMs:         tt.idleMs,
				idleThresholdS: tt.idleThresh,
			}
			got := p.TickInterval()
			if got != tt.want {
				t.Errorf("TickInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaneLabelDefault(t *testing.T) {
	p := &PaneModel{label: "User Terminal"}
	if p.Label() != "User Terminal" {
		t.Errorf("Label() = %q, want %q", p.Label(), "User Terminal")
	}
}
