package store

import (
	"strings"
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

func TestSummaryOmitsHeadersWhenAllEventsOlderThanCutoff(t *testing.T) {
	s := New()
	old := time.Now().Add(-2 * time.Hour)
	s.AddFileEvent(FileEvent{Path: "old.go", Change: "modify", Timestamp: old})
	s.AddTerminalEvent(TerminalEvent{Content: "old output", Timestamp: old})
	s.AddGitEvent(GitEvent{Type: "commit", Summary: "old commit", Timestamp: old})

	summary := s.Summary(1 * time.Hour)
	if summary != "" {
		t.Errorf("expected empty summary when all events are older than cutoff, got:\n%s", summary)
	}
}

func TestSummaryIncludesOnlyRecentEvents(t *testing.T) {
	s := New()
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()
	s.AddFileEvent(FileEvent{Path: "old.go", Change: "modify", Timestamp: old})
	s.AddFileEvent(FileEvent{Path: "new.go", Change: "create", Timestamp: recent})
	s.AddTerminalEvent(TerminalEvent{Content: "old output", Timestamp: old})
	s.AddGitEvent(GitEvent{Type: "commit", Summary: "old commit", Timestamp: old})

	summary := s.Summary(1 * time.Hour)
	if strings.Contains(summary, "old.go") {
		t.Error("summary should not contain old file event")
	}
	if !strings.Contains(summary, "new.go") {
		t.Error("summary should contain recent file event")
	}
	if strings.Contains(summary, "## Terminal Activity") {
		t.Error("summary should not contain Terminal Activity header when all terminal events are old")
	}
	if strings.Contains(summary, "## Git Activity") {
		t.Error("summary should not contain Git Activity header when all git events are old")
	}
}
