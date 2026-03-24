package watcher

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/huypl53/agent-tutor/internal/store"
)

// FileWatcher watches a directory tree for file changes and records events.
type FileWatcher struct {
	dir      string
	patterns []string
	ignores  []string
	store    *store.Store
	watcher  *fsnotify.Watcher

	mu            sync.Mutex
	stopped       bool
	pendingTimers map[string]*time.Timer
}

// NewFileWatcher creates a FileWatcher for the given directory.
// patterns are glob patterns to include (e.g. "**/*.go").
// ignores are directory names to skip (e.g. ".git", "node_modules").
func NewFileWatcher(dir string, patterns, ignores []string, s *store.Store) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		dir:           dir,
		patterns:      patterns,
		ignores:       ignores,
		store:         s,
		watcher:       w,
		pendingTimers: make(map[string]*time.Timer),
	}, nil
}

// Start adds directories to the watcher and begins processing events.
func (fw *FileWatcher) Start(ctx context.Context) error {
	if err := fw.addDirs(); err != nil {
		return err
	}
	go fw.loop(ctx)
	return nil
}

// Stop closes the underlying fsnotify watcher.
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.stopped = true
	for name, t := range fw.pendingTimers {
		t.Stop()
		delete(fw.pendingTimers, name)
	}
	return fw.watcher.Close()
}

// addDirs walks the directory tree and adds each directory to the watcher,
// skipping ignored directories.
func (fw *FileWatcher) addDirs() error {
	return filepath.WalkDir(fw.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			name := d.Name()
			for _, ig := range fw.ignores {
				if name == ig {
					return filepath.SkipDir
				}
			}
			return fw.watcher.Add(path)
		}
		return nil
	})
}

// loop processes fsnotify events with debouncing until the context is cancelled.
func (fw *FileWatcher) loop(ctx context.Context) {
	const debounce = 300 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if !fw.matchesPattern(ev.Name) {
				continue
			}

			change := changeType(ev.Op)
			if change == "" {
				continue
			}

			fw.mu.Lock()
			if t, exists := fw.pendingTimers[ev.Name]; exists {
				t.Stop()
			}
			fw.pendingTimers[ev.Name] = time.AfterFunc(debounce, func() {
				fw.mu.Lock()
				if fw.stopped {
					fw.mu.Unlock()
					return
				}
				delete(fw.pendingTimers, ev.Name)
				fw.mu.Unlock()

				rel := fw.relativePath(ev.Name)
				diff := fw.getDiff(ev.Name)
				fw.store.AddFileEvent(store.FileEvent{
					Path:      rel,
					Change:    change,
					Diff:      diff,
					Timestamp: time.Now(),
				})
			})
			fw.mu.Unlock()

		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

// matchesPattern checks whether the file path matches any of the configured
// glob patterns. Patterns like "**/*.go" are treated as a base-name match
// against "*.<ext>".
func (fw *FileWatcher) matchesPattern(path string) bool {
	if len(fw.patterns) == 0 {
		return true
	}
	base := filepath.Base(path)
	for _, p := range fw.patterns {
		// Strip leading "**/" to get the base pattern.
		pat := p
		if i := strings.LastIndex(pat, "/"); i >= 0 {
			pat = pat[i+1:]
		}
		if matched, _ := filepath.Match(pat, base); matched {
			return true
		}
	}
	return false
}

// relativePath returns path relative to the watched directory.
func (fw *FileWatcher) relativePath(path string) string {
	rel, err := filepath.Rel(fw.dir, path)
	if err != nil {
		return path
	}
	return rel
}

// getDiff attempts to get a git diff for the file. Returns empty string on
// failure (e.g. not a git repo, or file is new/untracked).
func (fw *FileWatcher) getDiff(path string) string {
	cmd := exec.Command("git", "diff", "--", path)
	cmd.Dir = fw.dir
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	s := string(out)
	// Truncate large diffs.
	if len(s) > 2000 {
		s = s[:2000] + "\n... (truncated)"
	}
	return s
}

// changeType maps fsnotify operations to a human-readable change type.
func changeType(op fsnotify.Op) string {
	switch {
	case op.Has(fsnotify.Create):
		return "create"
	case op.Has(fsnotify.Write):
		return "modify"
	case op.Has(fsnotify.Remove), op.Has(fsnotify.Rename):
		return "delete"
	default:
		return ""
	}
}
