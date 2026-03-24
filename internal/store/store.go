// Package store provides an in-memory ring buffer for watcher events.
package store

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	maxFileEvents     = 100
	maxTerminalEvents = 50
	maxGitEvents      = 30
	maxSummaryLen     = 8000
)

// FileEvent represents a file system change.
type FileEvent struct {
	Path      string
	Change    string // create, modify, delete
	Diff      string
	Timestamp time.Time
}

// TerminalEvent represents terminal output.
type TerminalEvent struct {
	Content   string
	Timestamp time.Time
}

// GitEvent represents a git operation.
type GitEvent struct {
	Type      string // commit, status_change
	Summary   string
	Diff      string
	Timestamp time.Time
}

// ringBuffer is a generic fixed-capacity ring buffer.
type ringBuffer[T any] struct {
	items []T
	head  int // next write position
	count int
	cap   int
}

func newRingBuffer[T any](capacity int) *ringBuffer[T] {
	return &ringBuffer[T]{
		items: make([]T, capacity),
		cap:   capacity,
	}
}

func (rb *ringBuffer[T]) add(item T) {
	rb.items[rb.head] = item
	rb.head = (rb.head + 1) % rb.cap
	if rb.count < rb.cap {
		rb.count++
	}
}

// snapshot returns all items in insertion order (oldest first).
func (rb *ringBuffer[T]) snapshot() []T {
	if rb.count == 0 {
		return nil
	}
	result := make([]T, rb.count)
	start := (rb.head - rb.count + rb.cap) % rb.cap
	for i := 0; i < rb.count; i++ {
		result[i] = rb.items[(start+i)%rb.cap]
	}
	return result
}

// Store holds recent watcher events in ring buffers.
type Store struct {
	mu        sync.RWMutex
	files     *ringBuffer[FileEvent]
	terminals *ringBuffer[TerminalEvent]
	gits      *ringBuffer[GitEvent]
}

// New creates a Store with default capacities.
func New() *Store {
	return &Store{
		files:     newRingBuffer[FileEvent](maxFileEvents),
		terminals: newRingBuffer[TerminalEvent](maxTerminalEvents),
		gits:      newRingBuffer[GitEvent](maxGitEvents),
	}
}

// AddFileEvent appends a file event.
func (s *Store) AddFileEvent(e FileEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files.add(e)
}

// AddTerminalEvent appends a terminal event.
func (s *Store) AddTerminalEvent(e TerminalEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.terminals.add(e)
}

// AddGitEvent appends a git event.
func (s *Store) AddGitEvent(e GitEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gits.add(e)
}

// FileEvents returns a snapshot of file events.
func (s *Store) FileEvents() []FileEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.files.snapshot()
}

// TerminalEvents returns a snapshot of terminal events.
func (s *Store) TerminalEvents() []TerminalEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.terminals.snapshot()
}

// GitEvents returns a snapshot of git events.
func (s *Store) GitEvents() []GitEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gits.snapshot()
}

// Summary returns a markdown-formatted summary of recent events.
// If since is 0, all events are included; otherwise only events newer than
// time.Now().Add(-since) are included.
func (s *Store) Summary(since time.Duration) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cutoff time.Time
	if since > 0 {
		cutoff = time.Now().Add(-since)
	}

	var b strings.Builder

	// File events
	files := s.files.snapshot()
	if len(files) > 0 {
		b.WriteString("## File Changes\n\n")
		for _, e := range files {
			if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s** `%s`", e.Change, e.Path))
			if e.Diff != "" {
				b.WriteString(fmt.Sprintf("\n  ```\n  %s\n  ```", truncate(e.Diff, 200)))
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	// Terminal events
	terms := s.terminals.snapshot()
	if len(terms) > 0 {
		b.WriteString("## Terminal Activity\n\n")
		for _, e := range terms {
			if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
				continue
			}
			b.WriteString(fmt.Sprintf("```\n%s\n```\n", truncate(e.Content, 300)))
		}
		b.WriteByte('\n')
	}

	// Git events
	gits := s.gits.snapshot()
	if len(gits) > 0 {
		b.WriteString("## Git Activity\n\n")
		for _, e := range gits {
			if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", e.Type, e.Summary))
		}
		b.WriteByte('\n')
	}

	return truncate(b.String(), maxSummaryLen)
}

// truncate cuts s to at most maxLen bytes, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
