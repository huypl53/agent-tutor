package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/huypham/agent-tutor/internal/store"
)

func TestFileWatcherDetectsChange(t *testing.T) {
	dir := t.TempDir()
	s := store.New()

	fw, err := NewFileWatcher(dir, []string{"**/*.go"}, []string{}, s)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := fw.Start(ctx); err != nil {
		t.Fatalf("failed to start watcher: %v", err)
	}
	defer fw.Stop()

	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main\n"), 0o644)

	time.Sleep(500 * time.Millisecond)

	events := s.FileEvents()
	if len(events) == 0 {
		t.Error("expected at least one file event")
	}
}
