# Agent Tutor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI that creates a tmux-based tutoring environment where a coding agent coaches the user by observing their real-time coding activity via MCP tools.

**Architecture:** MCP server (stdio) with three watchers (file, terminal, git) feeding an in-memory context store. CLI orchestrates tmux layout and agent launch. System prompt injection makes the agent behave as a tutor.

**Tech Stack:** Go 1.22+, `github.com/modelcontextprotocol/go-sdk`, `github.com/spf13/cobra`, `github.com/fsnotify/fsnotify`, `github.com/pelletier/go-toml/v2`

---

### Task 1: Project Scaffolding & Go Module

**Files:**
- Create: `cmd/agent-tutor/main.go`
- Create: `go.mod`

**Step 1: Initialize Go module**

Run: `go mod init github.com/huypl53/agent-tutor`

**Step 2: Install core dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/fsnotify/fsnotify@latest
go get github.com/pelletier/go-toml/v2@latest
go get github.com/modelcontextprotocol/go-sdk@latest
```

**Step 3: Write minimal main.go**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "agent-tutor",
		Short: "A programming tutor that coaches you through your coding agent",
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 4: Verify it builds**

Run: `go build ./cmd/agent-tutor`
Expected: Binary `agent-tutor` created, no errors.

**Step 5: Commit**

```bash
git add cmd/ go.mod go.sum
git commit -m "feat: scaffold Go project with cobra CLI"
```

---

### Task 2: Config Package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Tutor.Intensity != "on-demand" {
		t.Errorf("expected intensity on-demand, got %s", cfg.Tutor.Intensity)
	}
	if cfg.Agent.Command != "claude" {
		t.Errorf("expected agent command claude, got %s", cfg.Agent.Command)
	}
	if cfg.Watchers.TerminalPollInterval != "2s" {
		t.Errorf("expected terminal poll 2s, got %s", cfg.Watchers.TerminalPollInterval)
	}
}

func TestLoadCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tutor.Intensity != "on-demand" {
		t.Errorf("expected default intensity, got %s", cfg.Tutor.Intensity)
	}
	// Verify file was created
	path := filepath.Join(dir, ".agent-tutor", "config.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestLoadExisting(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".agent-tutor")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
[tutor]
intensity = "proactive"
level = "beginner"

[agent]
command = "codex"
args = []

[watchers]
file_patterns = ["**/*.rs"]
ignore_patterns = [".git"]
terminal_poll_interval = "1s"
git_poll_interval = "3s"

[tmux]
layout = "vertical"
user_pane_size = 60
`), 0o644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Tutor.Intensity != "proactive" {
		t.Errorf("expected proactive, got %s", cfg.Tutor.Intensity)
	}
	if cfg.Agent.Command != "codex" {
		t.Errorf("expected codex, got %s", cfg.Agent.Command)
	}
	if cfg.Tmux.UserPaneSize != 60 {
		t.Errorf("expected 60, got %d", cfg.Tmux.UserPaneSize)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL — package doesn't exist yet.

**Step 3: Write implementation**

```go
package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Tutor    TutorConfig    `toml:"tutor"`
	Agent    AgentConfig    `toml:"agent"`
	Watchers WatcherConfig  `toml:"watchers"`
	Tmux     TmuxConfig     `toml:"tmux"`
}

type TutorConfig struct {
	Intensity string `toml:"intensity"`
	Level     string `toml:"level"`
}

type AgentConfig struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

type WatcherConfig struct {
	FilePatterns          []string `toml:"file_patterns"`
	IgnorePatterns        []string `toml:"ignore_patterns"`
	TerminalPollInterval  string   `toml:"terminal_poll_interval"`
	GitPollInterval       string   `toml:"git_poll_interval"`
}

type TmuxConfig struct {
	Layout       string `toml:"layout"`
	UserPaneSize int    `toml:"user_pane_size"`
}

func Default() *Config {
	return &Config{
		Tutor: TutorConfig{
			Intensity: "on-demand",
			Level:     "auto",
		},
		Agent: AgentConfig{
			Command: "claude",
			Args:    []string{},
		},
		Watchers: WatcherConfig{
			FilePatterns:         []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts", "**/*.rs"},
			IgnorePatterns:       []string{"node_modules", ".git", "vendor", "target"},
			TerminalPollInterval: "2s",
			GitPollInterval:      "5s",
		},
		Tmux: TmuxConfig{
			Layout:       "horizontal",
			UserPaneSize: 50,
		},
	}
}

func configPath(projectDir string) string {
	return filepath.Join(projectDir, ".agent-tutor", "config.toml")
}

func Load(projectDir string) (*Config, error) {
	path := configPath(projectDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := Save(projectDir, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(projectDir string, cfg *Config) error {
	path := configPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: All 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with TOML loading and defaults"
```

---

### Task 3: Context Store (Ring Buffer)

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/store_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/store/ -v`
Expected: FAIL — package doesn't exist.

**Step 3: Write implementation**

```go
package store

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type FileEvent struct {
	Path      string
	Change    string
	Diff      string
	Timestamp time.Time
}

type TerminalEvent struct {
	Content   string
	Timestamp time.Time
}

type GitEvent struct {
	Type      string
	Summary   string
	Diff      string
	Timestamp time.Time
}

type Store struct {
	mu             sync.RWMutex
	fileEvents     []FileEvent
	terminalEvents []TerminalEvent
	gitEvents      []GitEvent
}

const (
	maxFileEvents     = 100
	maxTerminalEvents = 50
	maxGitEvents      = 30
)

func New() *Store {
	return &Store{
		fileEvents:     make([]FileEvent, 0, maxFileEvents),
		terminalEvents: make([]TerminalEvent, 0, maxTerminalEvents),
		gitEvents:      make([]GitEvent, 0, maxGitEvents),
	}
}

func (s *Store) AddFileEvent(e FileEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.fileEvents) >= maxFileEvents {
		s.fileEvents = s.fileEvents[1:]
	}
	s.fileEvents = append(s.fileEvents, e)
}

func (s *Store) AddTerminalEvent(e TerminalEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.terminalEvents) >= maxTerminalEvents {
		s.terminalEvents = s.terminalEvents[1:]
	}
	s.terminalEvents = append(s.terminalEvents, e)
}

func (s *Store) AddGitEvent(e GitEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.gitEvents) >= maxGitEvents {
		s.gitEvents = s.gitEvents[1:]
	}
	s.gitEvents = append(s.gitEvents, e)
}

func (s *Store) FileEvents() []FileEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FileEvent, len(s.fileEvents))
	copy(out, s.fileEvents)
	return out
}

func (s *Store) TerminalEvents() []TerminalEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]TerminalEvent, len(s.terminalEvents))
	copy(out, s.terminalEvents)
	return out
}

func (s *Store) GitEvents() []GitEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]GitEvent, len(s.gitEvents))
	copy(out, s.gitEvents)
	return out
}

func (s *Store) Summary(since time.Duration) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cutoff time.Time
	if since > 0 {
		cutoff = time.Now().Add(-since)
	}

	var b strings.Builder

	// File changes
	var fileCount int
	for _, e := range s.fileEvents {
		if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
			continue
		}
		if fileCount == 0 {
			b.WriteString("## Recent File Changes\n")
		}
		fmt.Fprintf(&b, "- %s (%s): %s\n", e.Path, e.Change, truncate(e.Diff, 200))
		fileCount++
		if fileCount >= 10 {
			break
		}
	}

	// Terminal activity
	var termCount int
	for i := len(s.terminalEvents) - 1; i >= 0; i-- {
		e := s.terminalEvents[i]
		if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
			continue
		}
		if termCount == 0 {
			b.WriteString("\n## Recent Terminal Activity\n")
		}
		fmt.Fprintf(&b, "```\n%s\n```\n", truncate(e.Content, 500))
		termCount++
		if termCount >= 5 {
			break
		}
	}

	// Git activity
	var gitCount int
	for i := len(s.gitEvents) - 1; i >= 0; i-- {
		e := s.gitEvents[i]
		if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
			continue
		}
		if gitCount == 0 {
			b.WriteString("\n## Recent Git Activity\n")
		}
		fmt.Fprintf(&b, "- %s: %s\n", e.Type, e.Summary)
		gitCount++
		if gitCount >= 5 {
			break
		}
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

**Step 4: Run tests**

Run: `go test ./internal/store/ -v`
Expected: All 4 tests PASS.

**Step 5: Commit**

```bash
git add internal/store/
git commit -m "feat: add context store with ring buffer for watcher events"
```

---

### Task 4: Tmux Manager

**Files:**
- Create: `internal/tmux/tmux.go`
- Create: `internal/tmux/tmux_test.go`

**Step 1: Write the failing test**

```go
package tmux

import (
	"os/exec"
	"testing"
)

func tmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestBuildCreateSessionCmd(t *testing.T) {
	m := New("test-session")
	cmd := m.createSessionCmd("/tmp/test")
	args := cmd.Args
	// Verify tmux new-session command structure
	if args[0] != "tmux" {
		t.Errorf("expected tmux, got %s", args[0])
	}
	found := false
	for _, a := range args {
		if a == "test-session" {
			found = true
		}
	}
	if !found {
		t.Error("session name not found in command args")
	}
}

func TestBuildSplitPaneCmd(t *testing.T) {
	m := New("test-session")
	cmd := m.splitPaneCmd(50, "horizontal")
	found := false
	for _, a := range cmd.Args {
		if a == "-h" || a == "-v" {
			found = true
		}
	}
	if !found {
		t.Error("split direction flag not found")
	}
}

func TestBuildCapturePaneCmd(t *testing.T) {
	m := New("test-session")
	cmd := m.capturePaneCmd("0")
	found := false
	for _, a := range cmd.Args {
		if a == "capture-pane" {
			found = true
		}
	}
	if !found {
		t.Error("capture-pane not found in command args")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tmux/ -v`
Expected: FAIL — package doesn't exist.

**Step 3: Write implementation**

```go
package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

type Manager struct {
	Session string
}

func New(session string) *Manager {
	return &Manager{Session: session}
}

func (m *Manager) createSessionCmd(workDir string) *exec.Cmd {
	return exec.Command("tmux", "new-session", "-d", "-s", m.Session, "-c", workDir)
}

func (m *Manager) splitPaneCmd(sizePercent int, layout string) *exec.Cmd {
	flag := "-h" // horizontal split = side-by-side panes
	if layout == "vertical" {
		flag = "-v"
	}
	return exec.Command("tmux", "split-window", flag, "-t", m.Session, "-p", fmt.Sprintf("%d", sizePercent))
}

func (m *Manager) capturePaneCmd(paneID string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", m.Session, paneID)
	return exec.Command("tmux", "capture-pane", "-t", target, "-p", "-J")
}

func (m *Manager) CreateSession(workDir string) error {
	return m.createSessionCmd(workDir).Run()
}

func (m *Manager) SplitPane(sizePercent int, layout string) error {
	return m.splitPaneCmd(sizePercent, layout).Run()
}

func (m *Manager) SendKeys(paneID string, keys string) error {
	target := fmt.Sprintf("%s:%s", m.Session, paneID)
	return exec.Command("tmux", "send-keys", "-t", target, keys, "Enter").Run()
}

func (m *Manager) CapturePane(paneID string) (string, error) {
	cmd := m.capturePaneCmd(paneID)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n "), nil
}

func (m *Manager) KillSession() error {
	return exec.Command("tmux", "kill-session", "-t", m.Session).Run()
}

func (m *Manager) Attach() error {
	cmd := exec.Command("tmux", "attach-session", "-t", m.Session)
	cmd.Stdin = nil // Will be set by caller
	return cmd.Run()
}

func (m *Manager) HasSession() bool {
	err := exec.Command("tmux", "has-session", "-t", m.Session).Run()
	return err == nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/tmux/ -v`
Expected: All 3 tests PASS (they test command construction, not execution).

**Step 5: Commit**

```bash
git add internal/tmux/
git commit -m "feat: add tmux manager for session/pane management"
```

---

### Task 5: File Watcher

**Files:**
- Create: `internal/watcher/watcher.go`
- Create: `internal/watcher/file.go`
- Create: `internal/watcher/file_test.go`

**Step 1: Write the watcher interface and file watcher test**

`internal/watcher/watcher.go`:
```go
package watcher

import "context"

type Watcher interface {
	Start(ctx context.Context) error
	Stop() error
}
```

`internal/watcher/file_test.go`:
```go
package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
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

	// Create a file
	testFile := filepath.Join(dir, "test.go")
	os.WriteFile(testFile, []byte("package main\n"), 0o644)

	// Wait for event
	time.Sleep(500 * time.Millisecond)

	events := s.FileEvents()
	if len(events) == 0 {
		t.Error("expected at least one file event")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/watcher/ -v -run TestFileWatcher`
Expected: FAIL — types don't exist.

**Step 3: Write implementation**

`internal/watcher/file.go`:
```go
package watcher

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/huypl53/agent-tutor/internal/store"
)

type FileWatcher struct {
	dir      string
	patterns []string
	ignores  []string
	store    *store.Store
	watcher  *fsnotify.Watcher
}

func NewFileWatcher(dir string, patterns, ignores []string, s *store.Store) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		dir:      dir,
		patterns: patterns,
		ignores:  ignores,
		store:    s,
		watcher:  w,
	}, nil
}

func (fw *FileWatcher) Start(ctx context.Context) error {
	if err := fw.addDirs(); err != nil {
		return err
	}

	go fw.loop(ctx)
	return nil
}

func (fw *FileWatcher) Stop() error {
	return fw.watcher.Close()
}

func (fw *FileWatcher) addDirs() error {
	return filepath.Walk(fw.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			for _, ig := range fw.ignores {
				if base == ig {
					return filepath.SkipDir
				}
			}
			return fw.watcher.Add(path)
		}
		return nil
	})
}

func (fw *FileWatcher) loop(ctx context.Context) {
	debounce := make(map[string]time.Time)
	const debounceInterval = 300 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if !fw.matchesPattern(event.Name) {
				continue
			}
			now := time.Now()
			if last, exists := debounce[event.Name]; exists && now.Sub(last) < debounceInterval {
				continue
			}
			debounce[event.Name] = now

			change := "modify"
			if event.Has(fsnotify.Create) {
				change = "create"
			} else if event.Has(fsnotify.Remove) {
				change = "delete"
			}

			diff := fw.getDiff(event.Name)
			fw.store.AddFileEvent(store.FileEvent{
				Path:      fw.relativePath(event.Name),
				Change:    change,
				Diff:      diff,
				Timestamp: now,
			})
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (fw *FileWatcher) matchesPattern(path string) bool {
	rel := fw.relativePath(path)
	for _, p := range fw.patterns {
		if matched, _ := filepath.Match(filepath.Base(p), filepath.Base(rel)); matched {
			return true
		}
	}
	return false
}

func (fw *FileWatcher) relativePath(path string) string {
	rel, err := filepath.Rel(fw.dir, path)
	if err != nil {
		return path
	}
	return rel
}

func (fw *FileWatcher) getDiff(path string) string {
	cmd := exec.Command("git", "-C", fw.dir, "diff", "--", path)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 50 {
		lines = lines[:50]
	}
	return strings.Join(lines, "\n")
}
```

Note: `file.go` uses `os` — add `"os"` to imports.

**Step 4: Run tests**

Run: `go test ./internal/watcher/ -v -run TestFileWatcher`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/watcher/
git commit -m "feat: add file watcher with fsnotify and debouncing"
```

---

### Task 6: Terminal Watcher

**Files:**
- Create: `internal/watcher/terminal.go`
- Create: `internal/watcher/terminal_test.go`

**Step 1: Write the failing test**

```go
package watcher

import (
	"testing"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

func TestTerminalDiff(t *testing.T) {
	tw := &TerminalWatcher{}

	old := "$ ls\nfile1.go\nfile2.go\n"
	new := "$ ls\nfile1.go\nfile2.go\n$ go build\nerror: syntax\n"

	diff := tw.diff(old, new)
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	if !containsString(diff, "go build") {
		t.Errorf("expected diff to contain 'go build', got: %s", diff)
	}
}

func TestDetectErrorPattern(t *testing.T) {
	tw := &TerminalWatcher{}

	tests := []struct {
		content string
		isError bool
	}{
		{"$ go build\nOK", false},
		{"error: something failed", true},
		{"Error: cannot find module", true},
		{"FAIL: TestFoo", true},
		{"panic: runtime error", true},
		{"all tests passed", false},
	}

	for _, tt := range tests {
		got := tw.hasError(tt.content)
		if got != tt.isError {
			t.Errorf("hasError(%q) = %v, want %v", tt.content, got, tt.isError)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/watcher/ -v -run "TestTerminal|TestDetect"`
Expected: FAIL.

**Step 3: Write implementation**

```go
package watcher

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^error[:\s]`),
	regexp.MustCompile(`(?i)^fatal[:\s]`),
	regexp.MustCompile(`(?i)^panic[:\s]`),
	regexp.MustCompile(`(?i)FAIL[:\s]`),
	regexp.MustCompile(`(?i)traceback`),
	regexp.MustCompile(`(?i)exception[:\s]`),
}

type TerminalWatcher struct {
	session      string
	paneID       string
	pollInterval time.Duration
	store        *store.Store
	lastContent  string
}

func NewTerminalWatcher(session, paneID string, pollInterval time.Duration, s *store.Store) *TerminalWatcher {
	return &TerminalWatcher{
		session:      session,
		paneID:       paneID,
		pollInterval: pollInterval,
		store:        s,
	}
}

func (tw *TerminalWatcher) Start(ctx context.Context) error {
	go tw.loop(ctx)
	return nil
}

func (tw *TerminalWatcher) Stop() error {
	return nil
}

func (tw *TerminalWatcher) loop(ctx context.Context) {
	ticker := time.NewTicker(tw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tw.poll()
		}
	}
}

func (tw *TerminalWatcher) poll() {
	target := tw.session + ":" + tw.paneID
	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-J")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	content := strings.TrimRight(string(out), "\n ")

	if content == tw.lastContent {
		return
	}

	newContent := tw.diff(tw.lastContent, content)
	tw.lastContent = content

	if newContent == "" {
		return
	}

	tw.store.AddTerminalEvent(store.TerminalEvent{
		Content:   newContent,
		Timestamp: time.Now(),
	})
}

func (tw *TerminalWatcher) diff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	if len(newLines) <= len(oldLines) {
		// Screen was cleared or scrolled — return all new content
		if old != new {
			return new
		}
		return ""
	}

	diffLines := newLines[len(oldLines):]
	return strings.Join(diffLines, "\n")
}

func (tw *TerminalWatcher) hasError(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		for _, pat := range errorPatterns {
			if pat.MatchString(line) {
				return true
			}
		}
	}
	return false
}
```

**Step 4: Run tests**

Run: `go test ./internal/watcher/ -v -run "TestTerminal|TestDetect"`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/watcher/terminal.go internal/watcher/terminal_test.go
git commit -m "feat: add terminal watcher with tmux capture-pane polling"
```

---

### Task 7: Git Watcher

**Files:**
- Create: `internal/watcher/git.go`
- Create: `internal/watcher/git_test.go`

**Step 1: Write the failing test**

```go
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
	// Set up a temp git repo
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
	// Record initial state
	gw.recordState()
	initialHead := gw.lastHead

	// Make a new commit
	os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main\n// changed\n"), 0o644)
	run("add", ".")
	run("commit", "-m", "second commit")

	// Poll
	gw.poll()

	if gw.lastHead == initialHead {
		t.Error("expected head to change")
	}

	events := s.GitEvents()
	if len(events) == 0 {
		t.Error("expected at least one git event")
	}
}

func TestParseGitStatus(t *testing.T) {
	gw := &GitWatcher{}
	output := " M main.go\n?? newfile.go\nA  staged.go\n"
	files := gw.parseStatus(output)
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/watcher/ -v -run "TestGitWatcher|TestParseGitStatus"`
Expected: FAIL.

**Step 3: Write implementation**

```go
package watcher

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/huypl53/agent-tutor/internal/store"
)

type GitWatcher struct {
	dir          string
	pollInterval time.Duration
	store        *store.Store
	lastHead     string
	lastStatus   string
}

func NewGitWatcher(dir string, pollInterval time.Duration, s *store.Store) *GitWatcher {
	return &GitWatcher{
		dir:          dir,
		pollInterval: pollInterval,
		store:        s,
	}
}

func (gw *GitWatcher) Start(ctx context.Context) error {
	gw.recordState()
	go gw.loop(ctx)
	return nil
}

func (gw *GitWatcher) Stop() error {
	return nil
}

func (gw *GitWatcher) loop(ctx context.Context) {
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

func (gw *GitWatcher) recordState() {
	gw.lastHead = gw.getHead()
	gw.lastStatus = gw.getStatus()
}

func (gw *GitWatcher) poll() {
	head := gw.getHead()
	status := gw.getStatus()

	if head != gw.lastHead && head != "" {
		msg := gw.getLastCommitMessage()
		diff := gw.getLastCommitDiff()
		gw.store.AddGitEvent(store.GitEvent{
			Type:      "commit",
			Summary:   msg,
			Diff:      diff,
			Timestamp: time.Now(),
		})
	}

	if status != gw.lastStatus {
		gw.store.AddGitEvent(store.GitEvent{
			Type:      "status_change",
			Summary:   status,
			Timestamp: time.Now(),
		})
	}

	gw.lastHead = head
	gw.lastStatus = status
}

func (gw *GitWatcher) getHead() string {
	cmd := exec.Command("git", "-C", gw.dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (gw *GitWatcher) getStatus() string {
	cmd := exec.Command("git", "-C", gw.dir, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (gw *GitWatcher) getLastCommitMessage() string {
	cmd := exec.Command("git", "-C", gw.dir, "log", "-1", "--pretty=%s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (gw *GitWatcher) getLastCommitDiff() string {
	cmd := exec.Command("git", "-C", gw.dir, "diff", "HEAD~1..HEAD", "--stat")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	s := string(out)
	if len(s) > 500 {
		s = s[:500] + "..."
	}
	return s
}

func (gw *GitWatcher) parseStatus(output string) []string {
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if len(line) < 4 {
			continue
		}
		files = append(files, strings.TrimSpace(line[2:]))
	}
	return files
}
```

**Step 4: Run tests**

Run: `go test ./internal/watcher/ -v -run "TestGitWatcher|TestParseGitStatus"`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/watcher/git.go internal/watcher/git_test.go
git commit -m "feat: add git watcher for commit and status change detection"
```

---

### Task 8: MCP Server — Core & Tools

**Files:**
- Create: `internal/mcp/server.go`
- Create: `internal/mcp/tools.go`
- Create: `internal/mcp/prompt.go`
- Create: `internal/mcp/server_test.go`

**Step 1: Write the failing test**

```go
package mcp

import (
	"testing"
	"time"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/store"
)

func TestBuildInstructions(t *testing.T) {
	cfg := config.Default()
	instructions := BuildInstructions(cfg)

	if instructions == "" {
		t.Error("expected non-empty instructions")
	}
	if !containsString(instructions, "on-demand") {
		t.Errorf("expected instructions to contain intensity level, got: %s", instructions)
	}
	if !containsString(instructions, "programming tutor") {
		t.Errorf("expected instructions to mention tutor role, got: %s", instructions)
	}
}

func TestToolHandlers(t *testing.T) {
	s := store.New()
	cfg := config.Default()

	s.AddFileEvent(store.FileEvent{
		Path: "main.go", Change: "modify", Diff: "added func", Timestamp: time.Now(),
	})
	s.AddTerminalEvent(store.TerminalEvent{
		Content: "$ go test\nPASS", Timestamp: time.Now(),
	})
	s.AddGitEvent(store.GitEvent{
		Type: "commit", Summary: "init commit", Timestamp: time.Now(),
	})

	handlers := NewToolHandlers(s, cfg)

	t.Run("get_student_context", func(t *testing.T) {
		result := handlers.GetStudentContext()
		if result == "" {
			t.Error("expected non-empty context")
		}
	})

	t.Run("get_recent_file_changes", func(t *testing.T) {
		result := handlers.GetRecentFileChanges()
		if result == "" {
			t.Error("expected non-empty file changes")
		}
	})

	t.Run("get_terminal_activity", func(t *testing.T) {
		result := handlers.GetTerminalActivity()
		if result == "" {
			t.Error("expected non-empty terminal activity")
		}
	})

	t.Run("get_git_activity", func(t *testing.T) {
		result := handlers.GetGitActivity()
		if result == "" {
			t.Error("expected non-empty git activity")
		}
	})

	t.Run("get_coaching_config", func(t *testing.T) {
		result := handlers.GetCoachingConfig()
		if result == "" {
			t.Error("expected non-empty config")
		}
	})
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
```

Note: Add `"strings"` to imports.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/mcp/ -v`
Expected: FAIL.

**Step 3: Write prompt.go**

```go
package mcp

import (
	"fmt"

	"github.com/huypl53/agent-tutor/internal/config"
)

func BuildInstructions(cfg *config.Config) string {
	return fmt.Sprintf(`You are also a programming tutor. A student is working in a terminal pane next to you. You have tools to observe their work.

Coaching intensity: %s
Student level: %s

When intensity is "proactive":
- After the student messages you, also check get_student_context for teachable moments
- When you receive a tutor_nudge notification, call get_student_context and offer relevant coaching
- Weave teaching naturally into your responses — don't lecture

When intensity is "on-demand" or "silent":
- Only use tutor tools when the student explicitly asks for feedback or uses /check

Teaching style:
- Explain the "why" not just the "what"
- For beginners: explain concepts, suggest resources
- For experienced devs: focus on idioms, best practices, ecosystem conventions
- Be concise. One teaching point per interaction, not five.
- If the student is doing well, say nothing. Don't coach for the sake of coaching.`, cfg.Tutor.Intensity, cfg.Tutor.Level)
}
```

**Step 4: Write tools.go**

```go
package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/store"
)

type ToolHandlers struct {
	store *store.Store
	cfg   *config.Config
}

func NewToolHandlers(s *store.Store, cfg *config.Config) *ToolHandlers {
	return &ToolHandlers{store: s, cfg: cfg}
}

func (h *ToolHandlers) GetStudentContext() string {
	return h.store.Summary(5 * time.Minute)
}

func (h *ToolHandlers) GetRecentFileChanges() string {
	events := h.store.FileEvents()
	if len(events) == 0 {
		return "No recent file changes."
	}
	var b strings.Builder
	for i := len(events) - 1; i >= 0 && i >= len(events)-10; i-- {
		e := events[i]
		fmt.Fprintf(&b, "### %s (%s) at %s\n", e.Path, e.Change, e.Timestamp.Format(time.Kitchen))
		if e.Diff != "" {
			fmt.Fprintf(&b, "```diff\n%s\n```\n", e.Diff)
		}
	}
	return b.String()
}

func (h *ToolHandlers) GetTerminalActivity() string {
	events := h.store.TerminalEvents()
	if len(events) == 0 {
		return "No recent terminal activity."
	}
	var b strings.Builder
	for i := len(events) - 1; i >= 0 && i >= len(events)-5; i-- {
		e := events[i]
		fmt.Fprintf(&b, "At %s:\n```\n%s\n```\n\n", e.Timestamp.Format(time.Kitchen), e.Content)
	}
	return b.String()
}

func (h *ToolHandlers) GetGitActivity() string {
	events := h.store.GitEvents()
	if len(events) == 0 {
		return "No recent git activity."
	}
	var b strings.Builder
	for i := len(events) - 1; i >= 0 && i >= len(events)-5; i-- {
		e := events[i]
		fmt.Fprintf(&b, "- **%s**: %s\n", e.Type, e.Summary)
		if e.Diff != "" {
			fmt.Fprintf(&b, "  ```\n  %s\n  ```\n", e.Diff)
		}
	}
	return b.String()
}

func (h *ToolHandlers) GetCoachingConfig() string {
	return fmt.Sprintf("Intensity: %s\nLevel: %s", h.cfg.Tutor.Intensity, h.cfg.Tutor.Level)
}

func (h *ToolHandlers) SetCoachingIntensity(intensity string) string {
	switch intensity {
	case "silent", "on-demand", "proactive":
		h.cfg.Tutor.Intensity = intensity
		return fmt.Sprintf("Coaching intensity set to: %s", intensity)
	default:
		return fmt.Sprintf("Invalid intensity: %s. Must be silent, on-demand, or proactive.", intensity)
	}
}
```

**Step 5: Write server.go**

```go
package mcp

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/store"
)

type Server struct {
	server   *gomcp.Server
	handlers *ToolHandlers
	cfg      *config.Config
}

func NewServer(s *store.Store, cfg *config.Config) *Server {
	srv := gomcp.NewServer(&gomcp.Implementation{
		Name:    "agent-tutor",
		Version: "0.1.0",
	}, &gomcp.ServerOptions{
		Instructions: BuildInstructions(cfg),
	})

	handlers := NewToolHandlers(s, cfg)
	registerTools(srv, handlers)

	return &Server{
		server:   srv,
		handlers: handlers,
		cfg:      cfg,
	}
}

func registerTools(srv *gomcp.Server, h *ToolHandlers) {
	type EmptyInput struct{}

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_student_context",
		Description: "Get a summary of the student's recent coding activity including file changes, terminal commands, and git activity",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.GetStudentContext(), nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_recent_file_changes",
		Description: "Get detailed diffs of recently modified files",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.GetRecentFileChanges(), nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_terminal_activity",
		Description: "Get recent terminal commands and output from the student's terminal pane",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.GetTerminalActivity(), nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_git_activity",
		Description: "Get recent git commits, diffs, and branch information",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.GetGitActivity(), nil
	})

	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "get_coaching_config",
		Description: "Get the current tutoring configuration (intensity level, student skill level)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input EmptyInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.GetCoachingConfig(), nil
	})

	type SetIntensityInput struct {
		Intensity string `json:"intensity" jsonschema:"enum=silent,enum=on-demand,enum=proactive,description=The coaching intensity level"`
	}
	gomcp.AddTool(srv, &gomcp.Tool{
		Name:        "set_coaching_intensity",
		Description: "Change the coaching intensity level (silent, on-demand, or proactive)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input SetIntensityInput) (*gomcp.CallToolResult, string, error) {
		return nil, h.SetCoachingIntensity(input.Intensity), nil
	})
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &gomcp.StdioTransport{})
}
```

**Step 6: Run tests**

Run: `go test ./internal/mcp/ -v`
Expected: All tests PASS.

**Step 7: Commit**

```bash
git add internal/mcp/
git commit -m "feat: add MCP server with tutor tools and system prompt injection"
```

---

### Task 9: Trigger Engine

**Files:**
- Create: `internal/trigger/trigger.go`
- Create: `internal/trigger/trigger_test.go`

**Step 1: Write the failing test**

```go
package trigger

import (
	"testing"
	"time"
)

func TestRuleFires(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "git.commit",
		Threshold: 1,
		Cooldown:  1 * time.Second,
	})

	engine.Fire("git.commit")
	if len(fired) != 1 {
		t.Errorf("expected 1 fire, got %d", len(fired))
	}
}

func TestRuleCooldown(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "git.commit",
		Threshold: 1,
		Cooldown:  5 * time.Minute,
	})

	engine.Fire("git.commit")
	engine.Fire("git.commit") // should be suppressed
	if len(fired) != 1 {
		t.Errorf("expected 1 fire (cooldown), got %d", len(fired))
	}
}

func TestRuleThreshold(t *testing.T) {
	var fired []string
	callback := func(event string) {
		fired = append(fired, event)
	}

	engine := New(callback)
	engine.AddRule(Rule{
		Event:     "terminal.error_repeat",
		Threshold: 3,
		Cooldown:  1 * time.Second,
	})

	engine.Fire("terminal.error_repeat") // count=1, no fire
	engine.Fire("terminal.error_repeat") // count=2, no fire
	if len(fired) != 0 {
		t.Errorf("expected 0 fires before threshold, got %d", len(fired))
	}

	engine.Fire("terminal.error_repeat") // count=3, fire!
	if len(fired) != 1 {
		t.Errorf("expected 1 fire at threshold, got %d", len(fired))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/trigger/ -v`
Expected: FAIL.

**Step 3: Write implementation**

```go
package trigger

import (
	"sync"
	"time"
)

type Rule struct {
	Event     string
	Threshold int
	Cooldown  time.Duration
}

type ruleState struct {
	rule     Rule
	count    int
	lastFire time.Time
}

type Engine struct {
	mu       sync.Mutex
	rules    map[string]*ruleState
	callback func(event string)
}

func New(callback func(event string)) *Engine {
	return &Engine{
		rules:    make(map[string]*ruleState),
		callback: callback,
	}
}

func (e *Engine) AddRule(r Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[r.Event] = &ruleState{rule: r}
}

func (e *Engine) Fire(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, ok := e.rules[event]
	if !ok {
		return
	}

	state.count++

	if state.count < state.rule.Threshold {
		return
	}

	if !state.lastFire.IsZero() && time.Since(state.lastFire) < state.rule.Cooldown {
		return
	}

	state.count = 0
	state.lastFire = time.Now()

	go e.callback(event)
}
```

**Step 4: Run tests**

Run: `go test ./internal/trigger/ -v`
Expected: All 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/trigger/
git commit -m "feat: add proactive trigger engine with threshold and cooldown"
```

---

### Task 10: CLI `start` Command

**Files:**
- Create: `internal/cli/start.go`
- Modify: `cmd/agent-tutor/main.go`

**Step 1: Write start.go**

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	mcpserver "github.com/huypl53/agent-tutor/internal/mcp"
	"github.com/huypl53/agent-tutor/internal/store"
	"github.com/huypl53/agent-tutor/internal/tmux"
	"github.com/huypl53/agent-tutor/internal/trigger"
	"github.com/huypl53/agent-tutor/internal/watcher"
)

const sessionName = "agent-tutor"

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start [project-dir]",
		Short: "Start a tutoring session",
		Long:  "Set up tmux with side-by-side panes: your terminal + coding agent with tutor capabilities.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	projectDir, _ := os.Getwd()
	if len(args) > 0 {
		var err error
		projectDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	// Load or create config
	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Check if session already exists
	tm := tmux.New(sessionName)
	if tm.HasSession() {
		return fmt.Errorf("session %q already exists — run 'agent-tutor stop' first", sessionName)
	}

	// Create tmux session
	if err := tm.CreateSession(projectDir); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Split pane for the coding agent
	agentPaneSize := 100 - cfg.Tmux.UserPaneSize
	if err := tm.SplitPane(agentPaneSize, cfg.Tmux.Layout); err != nil {
		tm.KillSession()
		return fmt.Errorf("splitting pane: %w", err)
	}

	// Start the coding agent in the right pane with MCP server
	// The MCP server binary path is the same binary with "mcp" subcommand
	self, _ := os.Executable()
	mcpCmd := fmt.Sprintf("%s mcp --project-dir %s", self, projectDir)

	agentCmd := fmt.Sprintf("%s --mcp-server '%s'", cfg.Agent.Command, mcpCmd)
	if err := tm.SendKeys("1", agentCmd); err != nil {
		tm.KillSession()
		return fmt.Errorf("starting agent: %w", err)
	}

	fmt.Printf("Agent Tutor session started.\n")
	fmt.Printf("  Project: %s\n", projectDir)
	fmt.Printf("  Agent: %s\n", cfg.Agent.Command)
	fmt.Printf("  Coaching: %s\n", cfg.Tutor.Intensity)
	fmt.Printf("\nAttaching to tmux session...\n")
	fmt.Printf("Left pane: your terminal. Right pane: your coding agent.\n")
	fmt.Printf("Type /check in the agent to get feedback on your work.\n\n")

	// Attach to the session
	attachCmd := fmt.Sprintf("tmux attach-session -t %s", sessionName)
	return syscall.Exec("/usr/bin/env", []string{"env", "bash", "-c", attachCmd}, os.Environ())
}
```

**Step 2: Write MCP subcommand for running as MCP server process**

Create `internal/cli/mcp.go`:
```go
package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	mcpserver "github.com/huypl53/agent-tutor/internal/mcp"
	"github.com/huypl53/agent-tutor/internal/store"
	"github.com/huypl53/agent-tutor/internal/trigger"
	"github.com/huypl53/agent-tutor/internal/watcher"
)

func NewMCPCmd() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:    "mcp",
		Short:  "Run as MCP server (used internally)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP(projectDir)
		},
	}
	cmd.Flags().StringVar(&projectDir, "project-dir", ".", "Project directory to watch")
	return cmd
}

func runMCP(projectDir string) error {
	cfg, err := config.Load(projectDir)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	s := store.New()

	// Parse poll intervals
	termInterval, _ := time.ParseDuration(cfg.Watchers.TerminalPollInterval)
	if termInterval == 0 {
		termInterval = 2 * time.Second
	}
	gitInterval, _ := time.ParseDuration(cfg.Watchers.GitPollInterval)
	if gitInterval == 0 {
		gitInterval = 5 * time.Second
	}

	// Start watchers
	fw, err := watcher.NewFileWatcher(projectDir, cfg.Watchers.FilePatterns, cfg.Watchers.IgnorePatterns, s)
	if err == nil {
		fw.Start(ctx)
		defer fw.Stop()
	}

	tw := watcher.NewTerminalWatcher(sessionName, "0", termInterval, s)
	tw.Start(ctx)

	gw := watcher.NewGitWatcher(projectDir, gitInterval, s)
	gw.Start(ctx)

	// Set up trigger engine (for proactive mode)
	_ = trigger.New(func(event string) {
		// In future: send MCP notification to agent
		// For v1, triggers are logged but not yet wired to notifications
	})

	// Run MCP server over stdio
	srv := mcpserver.NewServer(s, cfg)
	return srv.Run(ctx)
}
```

**Step 3: Write stop and status commands**

Create `internal/cli/stop.go`:
```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the tutoring session",
		RunE: func(cmd *cobra.Command, args []string) error {
			tm := tmux.New(sessionName)
			if !tm.HasSession() {
				return fmt.Errorf("no active agent-tutor session")
			}
			if err := tm.KillSession(); err != nil {
				return fmt.Errorf("killing session: %w", err)
			}
			fmt.Println("Agent Tutor session stopped.")
			return nil
		},
	}
}
```

Create `internal/cli/status.go`:
```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current session info",
		RunE: func(cmd *cobra.Command, args []string) error {
			tm := tmux.New(sessionName)
			if !tm.HasSession() {
				fmt.Println("No active agent-tutor session.")
				return nil
			}
			fmt.Println("Agent Tutor session is running.")
			return nil
		},
	}
}
```

**Step 4: Update main.go to wire commands**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/cli"
)

func main() {
	root := &cobra.Command{
		Use:   "agent-tutor",
		Short: "A programming tutor that coaches you through your coding agent",
	}

	root.AddCommand(cli.NewStartCmd())
	root.AddCommand(cli.NewStopCmd())
	root.AddCommand(cli.NewStatusCmd())
	root.AddCommand(cli.NewMCPCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 5: Verify it builds**

Run: `go build ./cmd/agent-tutor`
Expected: Binary builds successfully.

**Step 6: Commit**

```bash
git add internal/cli/ cmd/agent-tutor/main.go
git commit -m "feat: add CLI commands (start, stop, status, mcp)"
```

---

### Task 11: Integration Wiring & Smoke Test

**Files:**
- Modify: various — fix any import issues
- Create: `internal/integration_test.go`

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass. Fix any compilation errors.

**Step 2: Build and verify help output**

Run:
```bash
go build -o agent-tutor ./cmd/agent-tutor
./agent-tutor --help
./agent-tutor start --help
./agent-tutor mcp --help
```
Expected: Help text displays for all commands.

**Step 3: Verify MCP server starts on stdio**

Run: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}' | ./agent-tutor mcp --project-dir .`
Expected: JSON response with server info and instructions field containing tutor prompt.

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: integration wiring and build verification"
```

---

### Task 12: README & Architecture Docs

**Files:**
- Modify: `README.md`
- Create: `docs/architecture.md`

**Step 1: Write README**

Update `README.md` with:
- Project description
- Installation: `go install github.com/huypl53/agent-tutor/cmd/agent-tutor@latest`
- Quick start: `agent-tutor start ~/myproject`
- Configuration section
- How it works (brief architecture)

**Step 2: Write architecture doc**

Create `docs/architecture.md` with:
- Architecture diagram (from design doc)
- Component descriptions
- Data flow explanation
- MCP tools reference
- Key implementation decisions and why

**Step 3: Commit**

```bash
git add README.md docs/architecture.md
git commit -m "docs: add README and architecture documentation"
```
