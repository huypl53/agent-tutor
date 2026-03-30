package mcp

import (
	"strings"
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
	if !strings.Contains(instructions, "on-demand") {
		t.Errorf("expected instructions to contain intensity level")
	}
	if !strings.Contains(instructions, "programming tutor") {
		t.Errorf("expected instructions to mention tutor role")
	}
	if !strings.Contains(instructions, "current-topic.md") {
		t.Errorf("expected instructions to contain topic tracking state file reference")
	}
	if !strings.Contains(instructions, "Topic tracking") {
		t.Errorf("expected instructions to contain topic tracking section")
	}
	if !strings.Contains(instructions, "learning-plan.md") {
		t.Errorf("expected instructions to contain learning plan reference")
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

	t.Run("set_coaching_intensity", func(t *testing.T) {
		result := handlers.SetCoachingIntensity("proactive")
		if !strings.Contains(result, "proactive") {
			t.Error("expected result to confirm proactive")
		}
		if cfg.GetIntensity() != "proactive" {
			t.Error("expected config intensity to be updated")
		}

		result = handlers.SetCoachingIntensity("invalid")
		if !strings.Contains(result, "Invalid") {
			t.Error("expected error for invalid intensity")
		}
	})
}
