package tui

import (
	"fmt"
	"time"

	"github.com/huypl53/agent-tutor/internal/tmux"
)

// PaneModel wraps a single tmux pane — it captures content via
// tmux.Manager.CapturePaneANSI, forwards keys via SendKeysRaw,
// and computes adaptive tick intervals.
type PaneModel struct {
	targetID       string
	label          string
	content        string
	lastActive     time.Time
	activeMs       int
	idleMs         int
	idleThresholdS int
	errCount       int
	tm             *tmux.Manager
}

// NewPaneModel creates a PaneModel for the given tmux pane.
func NewPaneModel(tm *tmux.Manager, paneID, label string, activeMs, idleMs, idleThresholdS int) *PaneModel {
	return &PaneModel{
		targetID:       paneID,
		label:          label,
		content:        "",
		lastActive:     time.Now(),
		activeMs:       activeMs,
		idleMs:         idleMs,
		idleThresholdS: idleThresholdS,
		tm:             tm,
	}
}

// Label returns the display label for this pane.
func (p *PaneModel) Label() string { return p.label }

// Content returns the last captured pane content.
func (p *PaneModel) Content() string { return p.content }

// TickInterval returns an adaptive polling interval based on how long
// the pane has been idle. Active panes poll faster; idle panes slower.
func (p *PaneModel) TickInterval() time.Duration {
	idle := time.Since(p.lastActive)
	if idle < 2*time.Second {
		return time.Duration(p.activeMs) * time.Millisecond
	}
	threshDuration := time.Duration(p.idleThresholdS) * time.Second
	if idle < threshDuration {
		return 100 * time.Millisecond
	}
	return time.Duration(p.idleMs) * time.Millisecond
}

// Capture fetches the current pane content via tmux. If the content
// changed, lastActive is updated to now.
func (p *PaneModel) Capture() error {
	content, err := p.tm.CapturePaneANSI(p.targetID)
	if err != nil {
		p.errCount++
		if p.errCount >= 5 {
			p.content = fmt.Sprintf("[tmux capture failed: %v]", err)
		}
		return err
	}
	p.errCount = 0
	if content != p.content {
		p.lastActive = time.Now()
		p.content = content
	}
	return nil
}

// MarkActive resets the idle timer to now.
func (p *PaneModel) MarkActive() { p.lastActive = time.Now() }

// SendKeys forwards raw keys to the underlying tmux pane.
func (p *PaneModel) SendKeys(keys ...string) error {
	return p.tm.SendKeysRaw(p.targetID, keys...)
}
