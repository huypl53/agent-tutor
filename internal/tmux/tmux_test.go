package tmux

import (
	"testing"
)

func TestBuildCreateSessionCmd(t *testing.T) {
	m := New("test-session")
	cmd := m.createSessionCmd("/tmp/test")
	args := cmd.Args
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
