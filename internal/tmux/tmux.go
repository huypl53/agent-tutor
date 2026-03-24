package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager wraps tmux CLI commands for session management.
type Manager struct {
	Session string
}

// New creates a new tmux Manager for the given session name.
func New(session string) *Manager {
	return &Manager{Session: session}
}

func (m *Manager) createSessionCmd(workDir string) *exec.Cmd {
	return exec.Command("tmux", "new-session", "-d", "-s", m.Session, "-c", workDir)
}

func (m *Manager) splitPaneCmd(sizePercent int, layout string) *exec.Cmd {
	flag := "-h"
	if layout == "vertical" {
		flag = "-v"
	}
	return exec.Command("tmux", "split-window", flag, "-t", m.Session, "-p", fmt.Sprintf("%d", sizePercent))
}

func (m *Manager) capturePaneCmd(paneID string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", m.Session, paneID)
	return exec.Command("tmux", "capture-pane", "-t", target, "-p", "-J")
}

func (m *Manager) sendKeysCmd(paneID string, keys string) *exec.Cmd {
	target := fmt.Sprintf("%s:%s", m.Session, paneID)
	return exec.Command("tmux", "send-keys", "-t", target, keys, "Enter")
}

func (m *Manager) killSessionCmd() *exec.Cmd {
	return exec.Command("tmux", "kill-session", "-t", m.Session)
}

func (m *Manager) hasSessionCmd() *exec.Cmd {
	return exec.Command("tmux", "has-session", "-t", m.Session)
}

// CreateSession creates a new detached tmux session in the given working directory.
func (m *Manager) CreateSession(workDir string) error {
	return m.createSessionCmd(workDir).Run()
}

// SplitPane splits the current pane in the session.
// layout should be "horizontal" or "vertical".
func (m *Manager) SplitPane(sizePercent int, layout string) error {
	return m.splitPaneCmd(sizePercent, layout).Run()
}

// SendKeys sends keystrokes to the specified pane.
func (m *Manager) SendKeys(paneID string, keys string) error {
	return m.sendKeysCmd(paneID, keys).Run()
}

// CapturePane captures the contents of the specified pane and returns it as a string.
func (m *Manager) CapturePane(paneID string) (string, error) {
	out, err := m.capturePaneCmd(paneID).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// KillSession destroys the tmux session.
func (m *Manager) KillSession() error {
	return m.killSessionCmd().Run()
}

// HasSession returns true if the session exists.
func (m *Manager) HasSession() bool {
	return m.hasSessionCmd().Run() == nil
}
