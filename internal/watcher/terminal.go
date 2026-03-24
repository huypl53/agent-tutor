package watcher

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

// errorPatterns are compiled regexps for detecting errors in terminal output.
var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^error[:\s]`),
	regexp.MustCompile(`(?im)^fatal[:\s]`),
	regexp.MustCompile(`(?im)^panic[:\s]`),
	regexp.MustCompile(`(?i)FAIL[:\s]`),
	regexp.MustCompile(`(?i)traceback`),
	regexp.MustCompile(`(?i)exception[:\s]`),
}

// TerminalWatcher polls tmux capture-pane at a configurable interval,
// diffs against the previous capture to detect new activity, and stores
// terminal events.
type TerminalWatcher struct {
	session      string
	paneID       string
	pollInterval time.Duration
	store        *store.Store
	socket       string
	lastContent  string

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

// NewTerminalWatcher creates a TerminalWatcher for the given tmux session and pane.
func NewTerminalWatcher(session, paneID string, pollInterval time.Duration, s *store.Store, socket string) *TerminalWatcher {
	return &TerminalWatcher{
		session:      session,
		paneID:       paneID,
		pollInterval: pollInterval,
		store:        s,
		socket:       socket,
	}
}

// Start launches the polling goroutine. It satisfies the Watcher interface.
func (tw *TerminalWatcher) Start(ctx context.Context) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.cancel != nil {
		return fmt.Errorf("terminal watcher already started")
	}

	ctx, tw.cancel = context.WithCancel(ctx)
	tw.done = make(chan struct{})

	go tw.loop(ctx)
	return nil
}

// Stop cancels the polling goroutine and waits for it to finish.
func (tw *TerminalWatcher) Stop() error {
	tw.mu.Lock()
	cancel := tw.cancel
	done := tw.done
	tw.mu.Unlock()

	if cancel == nil {
		return nil
	}

	cancel()
	if done != nil {
		<-done
	}

	tw.mu.Lock()
	tw.cancel = nil
	tw.done = nil
	tw.mu.Unlock()

	return nil
}

// loop runs the ticker-based polling loop.
func (tw *TerminalWatcher) loop(ctx context.Context) {
	defer close(tw.done)

	ticker := time.NewTicker(tw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tw.poll(ctx)
		}
	}
}

// poll captures the current tmux pane content, diffs it against the last
// capture, and stores any new output as a TerminalEvent.
func (tw *TerminalWatcher) poll(ctx context.Context) {
	target := fmt.Sprintf("%s:%s", tw.session, tw.paneID)
	args := []string{}
	if tw.socket != "" {
		args = append(args, "-L", tw.socket)
	}
	args = append(args, "capture-pane", "-t", target, "-p", "-J")
	cmd := exec.CommandContext(ctx, "tmux", args...)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	content := string(out)
	d := tw.diff(tw.lastContent, content)
	tw.lastContent = content

	if d == "" {
		return
	}

	tw.store.AddTerminalEvent(store.TerminalEvent{
		Content:   d,
		HasError:  tw.hasError(d),
		Timestamp: time.Now(),
	})
}

// diff returns the new lines added to the terminal output. If the screen was
// cleared (new has fewer lines), all new content is returned.
func (tw *TerminalWatcher) diff(old, new string) string {
	if old == new {
		return ""
	}

	oldLines := strings.Split(strings.TrimRight(old, "\n"), "\n")
	newLines := strings.Split(strings.TrimRight(new, "\n"), "\n")

	// If old is empty, everything is new.
	if old == "" {
		return new
	}

	// Screen was cleared or scrolled: fewer lines means return all new content.
	if len(newLines) <= len(oldLines) {
		// But only if content actually changed (already checked above).
		return strings.TrimRight(new, "\n")
	}

	// New lines appended: return only the new portion.
	added := newLines[len(oldLines):]
	return strings.Join(added, "\n")
}

// hasError checks whether the content contains any known error pattern.
func (tw *TerminalWatcher) hasError(content string) bool {
	for _, p := range errorPatterns {
		if p.MatchString(content) {
			return true
		}
	}
	return false
}
