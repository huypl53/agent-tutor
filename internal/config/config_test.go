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
	path := filepath.Join(dir, ".agent-tutor", "config.toml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestDefaultTUIConfig(t *testing.T) {
	cfg := Default()
	if cfg.TUI.Layout != "horizontal" {
		t.Errorf("expected layout horizontal, got %s", cfg.TUI.Layout)
	}
	if cfg.TUI.SplitRatio != 50 {
		t.Errorf("expected split_ratio 50, got %d", cfg.TUI.SplitRatio)
	}
	if cfg.TUI.FocusKey != "ctrl+space" {
		t.Errorf("expected focus_key ctrl+space, got %s", cfg.TUI.FocusKey)
	}
	if cfg.TUI.QuitKey != "ctrl+q" {
		t.Errorf("expected quit_key ctrl+q, got %s", cfg.TUI.QuitKey)
	}
	if cfg.TUI.Polling.ActiveMs != 50 {
		t.Errorf("expected active_ms 50, got %d", cfg.TUI.Polling.ActiveMs)
	}
	if cfg.TUI.Polling.IdleMs != 200 {
		t.Errorf("expected idle_ms 200, got %d", cfg.TUI.Polling.IdleMs)
	}
	if cfg.TUI.Polling.IdleThresholdS != 10 {
		t.Errorf("expected idle_threshold_s 10, got %d", cfg.TUI.Polling.IdleThresholdS)
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
