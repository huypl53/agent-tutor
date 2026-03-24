package tmux

import (
	"os/exec"
	"strings"
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

func TestCapturePaneANSICmdArgs(t *testing.T) {
	m := &Manager{Session: "test-sess", Socket: "test-sock"}
	cmd := m.capturePaneANSICmd("0")
	got := strings.Join(cmd.Args[1:], " ")
	want := "-L test-sock capture-pane -t test-sess:0.0 -p -e"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResizePaneCmdArgs(t *testing.T) {
	m := &Manager{Session: "test-sess", Socket: "test-sock"}
	cmd := m.resizePaneCmd("0", 80, 24)
	got := strings.Join(cmd.Args[1:], " ")
	want := "-L test-sock resize-pane -t test-sess:0.0 -x 80 -y 24"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSendKeysRawCmdArgs(t *testing.T) {
	m := &Manager{Session: "test-sess", Socket: "test-sock"}
	cmd := m.sendKeysRawCmd("1", "ls", "-l")
	got := strings.Join(cmd.Args[1:], " ")
	want := "-L test-sock send-keys -t test-sess:0.1 ls -l"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
