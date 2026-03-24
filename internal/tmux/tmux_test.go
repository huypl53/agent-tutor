package tmux

import (
	"os/exec"
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

func TestSocketAppearsInCommands(t *testing.T) {
	m := New("test-session")
	m.Socket = "test-sock"

	cmds := []*exec.Cmd{
		m.createSessionCmd("/tmp"),
		m.splitPaneCmd(50, "horizontal"),
		m.capturePaneCmd("0"),
		m.sendKeysCmd("0", "echo hi"),
		m.killSessionCmd(),
		m.hasSessionCmd(),
	}

	for _, cmd := range cmds {
		if cmd.Args[1] != "-L" || cmd.Args[2] != "test-sock" {
			t.Errorf("expected -L test-sock in args, got %v", cmd.Args)
		}
	}
}

func TestNoSocketOmitsFlag(t *testing.T) {
	m := New("test-session")
	cmd := m.createSessionCmd("/tmp")
	for _, a := range cmd.Args {
		if a == "-L" {
			t.Error("should not have -L flag when Socket is empty")
		}
	}
}
