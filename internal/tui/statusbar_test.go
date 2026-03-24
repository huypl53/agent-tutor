package tui

import (
	"strings"
	"testing"
)

func TestStatusBarRender(t *testing.T) {
	sb := NewStatusBar("agent-tutor-myproj-abc123", "on-demand")
	sb.SetActivePane(0, "User Terminal")
	rendered := sb.RenderPlain()
	if rendered == "" {
		t.Error("expected non-empty render")
	}
	if !strings.Contains(rendered, "myproj") {
		t.Errorf("expected session name in render, got: %s", rendered)
	}
	if !strings.Contains(rendered, "on-demand") {
		t.Errorf("expected coaching mode in render, got: %s", rendered)
	}
}

func TestStatusBarActivePaneSwitch(t *testing.T) {
	sb := NewStatusBar("sess", "silent")
	sb.SetActivePane(0, "User Terminal")
	r1 := sb.RenderPlain()
	sb.SetActivePane(1, "Claude Code")
	r2 := sb.RenderPlain()
	if r1 == r2 {
		t.Error("expected different renders for different active panes")
	}
}
