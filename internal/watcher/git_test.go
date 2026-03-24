package watcher

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

func TestGitWatcherDetectsCommit(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v failed: %v", args, err)
		}
	}
	run("init")
	os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n"), 0o644)
	run("add", ".")
	run("commit", "-m", "initial")

	s := store.New()
	gw := NewGitWatcher(dir, 100*time.Millisecond, s)
	gw.recordState()
	initialHead := gw.lastHead

	os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n// changed\n"), 0o644)
	run("add", ".")
	run("commit", "-m", "second commit")

	gw.poll()

	if gw.lastHead == initialHead {
		t.Error("expected head to change")
	}

	events := s.GitEvents()
	if len(events) == 0 {
		t.Error("expected at least one git event")
	}
}
