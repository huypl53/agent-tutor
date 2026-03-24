package watcher

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

// GitWatcher polls git status and git log at a configurable interval.
// It detects commits (HEAD change) and status changes, storing GitEvents.
type GitWatcher struct {
	dir          string
	pollInterval time.Duration
	store        *store.Store
	mu           sync.Mutex
	lastHead     string
	lastStatus   string
	cancel       context.CancelFunc
	done         chan struct{}
}

// NewGitWatcher creates a new GitWatcher for the given directory.
func NewGitWatcher(dir string, pollInterval time.Duration, s *store.Store) *GitWatcher {
	return &GitWatcher{
		dir:          dir,
		pollInterval: pollInterval,
		store:        s,
	}
}

// Start begins polling in a background goroutine.
func (gw *GitWatcher) Start(ctx context.Context) error {
	ctx, gw.cancel = context.WithCancel(ctx)
	gw.done = make(chan struct{})
	gw.recordState()
	go gw.loop(ctx)
	return nil
}

// Stop cancels the polling loop and waits for it to finish.
func (gw *GitWatcher) Stop() error {
	if gw.cancel != nil {
		gw.cancel()
		<-gw.done
	}
	return nil
}

// loop runs poll at the configured interval until the context is cancelled.
func (gw *GitWatcher) loop(ctx context.Context) {
	defer close(gw.done)
	ticker := time.NewTicker(gw.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gw.poll()
		}
	}
}

// recordState captures the current HEAD and status as baseline.
func (gw *GitWatcher) recordState() {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.lastHead = gw.getHead()
	gw.lastStatus = gw.getStatus()
}

// poll checks for HEAD and status changes, recording GitEvents as needed.
func (gw *GitWatcher) poll() {
	head := gw.getHead()
	status := gw.getStatus()

	gw.mu.Lock()
	defer gw.mu.Unlock()

	if head != gw.lastHead && head != "" {
		msg := gw.getLastCommitMessage()
		diff := gw.getLastCommitDiff()
		gw.store.AddGitEvent(store.GitEvent{
			Type:      "commit",
			Summary:   msg,
			Diff:      diff,
			Timestamp: time.Now(),
		})
		gw.lastHead = head
	}

	if status != gw.lastStatus {
		gw.store.AddGitEvent(store.GitEvent{
			Type:      "status_change",
			Summary:   status,
			Timestamp: time.Now(),
		})
		gw.lastStatus = status
	}
}

// getHead returns the current HEAD commit hash.
func (gw *GitWatcher) getHead() string {
	out, err := exec.Command("git", "-C", gw.dir, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getStatus returns the porcelain status output.
func (gw *GitWatcher) getStatus() string {
	out, err := exec.Command("git", "-C", gw.dir, "status", "--porcelain").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getLastCommitMessage returns the subject of the last commit.
func (gw *GitWatcher) getLastCommitMessage() string {
	out, err := exec.Command("git", "-C", gw.dir, "log", "-1", "--pretty=%s").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getLastCommitDiff returns the stat diff of the last commit, truncated to 500 chars.
func (gw *GitWatcher) getLastCommitDiff() string {
	out, err := exec.Command("git", "-C", gw.dir, "diff", "HEAD~1..HEAD", "--stat").Output()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if len(s) > 500 {
		s = s[:497] + "..."
	}
	return s
}
