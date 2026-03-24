package store

import (
	"testing"
	"time"
)

func TestFileEvents(t *testing.T) {
	s := New()
	for i := 0; i < 150; i++ {
		s.AddFileEvent(FileEvent{
			Path:      "test.go",
			Change:    "modify",
			Diff:      "some diff",
			Timestamp: time.Now(),
		})
	}
	events := s.FileEvents()
	if len(events) != 100 {
		t.Errorf("expected 100 events (ring buffer cap), got %d", len(events))
	}
}

func TestTerminalEvents(t *testing.T) {
	s := New()
	s.AddTerminalEvent(TerminalEvent{
		Content:   "$ go build\nOK",
		Timestamp: time.Now(),
	})
	events := s.TerminalEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].Content != "$ go build\nOK" {
		t.Errorf("unexpected content: %s", events[0].Content)
	}
}

func TestGitEvents(t *testing.T) {
	s := New()
	s.AddGitEvent(GitEvent{
		Type:      "commit",
		Summary:   "feat: add thing",
		Timestamp: time.Now(),
	})
	events := s.GitEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestSummary(t *testing.T) {
	s := New()
	s.AddFileEvent(FileEvent{Path: "main.go", Change: "modify", Diff: "diff1", Timestamp: time.Now()})
	s.AddTerminalEvent(TerminalEvent{Content: "$ go test\nPASS", Timestamp: time.Now()})
	s.AddGitEvent(GitEvent{Type: "commit", Summary: "init", Timestamp: time.Now()})

	summary := s.Summary(time.Duration(0))
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
