package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager wraps tmux CLI commands for session management.
type Manager struct {
	Session string
	Socket  string
}

// tmuxCmd builds an exec.Cmd with optional -L socket flag.
func (m *Manager) tmuxCmd(args ...string) *exec.Cmd {
	if m.Socket != "" {
		args = append([]string{"-L", m.Socket}, args...)
	}
	return exec.Command("tmux", args...)
}

// New creates a new tmux Manager for the given session name.
func New(session string) *Manager {
	return &Manager{Session: session}
}

func (m *Manager) createSessionCmd(workDir string) *exec.Cmd {
	return m.tmuxCmd("new-session", "-d", "-s", m.Session, "-c", workDir)
}

func (m *Manager) splitPaneCmd(sizePercent int, layout string) *exec.Cmd {
	flag := "-h"
	if layout == "vertical" {
		flag = "-v"
	}
	return m.tmuxCmd("split-window", flag, "-t", m.Session, "-l", fmt.Sprintf("%d%%", sizePercent))
}

func (m *Manager) capturePaneCmd(paneID string) *exec.Cmd {
	target := fmt.Sprintf("%s:0.%s", m.Session, paneID)
	return m.tmuxCmd("capture-pane", "-t", target, "-p", "-J")
}

func (m *Manager) sendKeysCmd(paneID string, keys string) *exec.Cmd {
	target := fmt.Sprintf("%s:0.%s", m.Session, paneID)
	return m.tmuxCmd("send-keys", "-t", target, keys, "Enter")
}

func (m *Manager) killSessionCmd() *exec.Cmd {
	return m.tmuxCmd("kill-session", "-t", m.Session)
}

func (m *Manager) hasSessionCmd() *exec.Cmd {
	return m.tmuxCmd("has-session", "-t", m.Session)
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

func (m *Manager) capturePaneANSICmd(paneID string) *exec.Cmd {
	target := fmt.Sprintf("%s:0.%s", m.Session, paneID)
	return m.tmuxCmd("capture-pane", "-t", target, "-p", "-e")
}

// CapturePaneANSI captures the contents of the specified pane with ANSI escape codes preserved.
func (m *Manager) CapturePaneANSI(paneID string) (string, error) {
	out, err := m.capturePaneANSICmd(paneID).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (m *Manager) resizePaneCmd(paneID string, width, height int) *exec.Cmd {
	target := fmt.Sprintf("%s:0.%s", m.Session, paneID)
	return m.tmuxCmd("resize-pane", "-t", target, "-x", fmt.Sprintf("%d", width), "-y", fmt.Sprintf("%d", height))
}

// ResizePane resizes the specified pane to the given width and height.
func (m *Manager) ResizePane(paneID string, width, height int) error {
	return m.resizePaneCmd(paneID, width, height).Run()
}

func (m *Manager) sendKeysRawCmd(paneID string, keys ...string) *exec.Cmd {
	target := fmt.Sprintf("%s:0.%s", m.Session, paneID)
	args := append([]string{"send-keys", "-t", target}, keys...)
	return m.tmuxCmd(args...)
}

// SendKeysRaw sends keystrokes to the specified pane without appending Enter.
func (m *Manager) SendKeysRaw(paneID string, keys ...string) error {
	return m.sendKeysRawCmd(paneID, keys...).Run()
}

// KillSession destroys the tmux session.
func (m *Manager) KillSession() error {
	return m.killSessionCmd().Run()
}

// HasSession returns true if the session exists.
func (m *Manager) HasSession() bool {
	return m.hasSessionCmd().Run() == nil
}
