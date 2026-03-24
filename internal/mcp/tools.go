package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/huypham/agent-tutor/internal/config"
	"github.com/huypham/agent-tutor/internal/store"
)

// ToolHandlers wraps the store and config to provide MCP tool implementations.
type ToolHandlers struct {
	store  *store.Store
	config *config.Config
}

// NewToolHandlers creates a new ToolHandlers.
func NewToolHandlers(s *store.Store, cfg *config.Config) *ToolHandlers {
	return &ToolHandlers{store: s, config: cfg}
}

// GetStudentContext returns a summary of recent student activity.
func (h *ToolHandlers) GetStudentContext() string {
	return h.store.Summary(5 * time.Minute)
}

// GetRecentFileChanges formats recent file events with diffs.
func (h *ToolHandlers) GetRecentFileChanges() string {
	events := h.store.FileEvents()
	if len(events) == 0 {
		return "No recent file changes."
	}

	var b strings.Builder
	for _, e := range events {
		b.WriteString(fmt.Sprintf("- %s: %s", e.Change, e.Path))
		if e.Diff != "" {
			b.WriteString(fmt.Sprintf("\n  ```\n  %s\n  ```", e.Diff))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// GetTerminalActivity formats recent terminal events.
func (h *ToolHandlers) GetTerminalActivity() string {
	events := h.store.TerminalEvents()
	if len(events) == 0 {
		return "No recent terminal activity."
	}

	var b strings.Builder
	for _, e := range events {
		b.WriteString(fmt.Sprintf("[%s]\n%s\n\n", e.Timestamp.Format(time.TimeOnly), e.Content))
	}
	return b.String()
}

// GetGitActivity formats recent git events.
func (h *ToolHandlers) GetGitActivity() string {
	events := h.store.GitEvents()
	if len(events) == 0 {
		return "No recent git activity."
	}

	var b strings.Builder
	for _, e := range events {
		b.WriteString(fmt.Sprintf("- %s: %s\n", e.Type, e.Summary))
	}
	return b.String()
}

// GetCoachingConfig returns the current coaching intensity and level.
func (h *ToolHandlers) GetCoachingConfig() string {
	return fmt.Sprintf("intensity: %s\nlevel: %s", h.config.Tutor.Intensity, h.config.Tutor.Level)
}

// SetCoachingIntensity validates and sets the coaching intensity.
func (h *ToolHandlers) SetCoachingIntensity(intensity string) string {
	switch intensity {
	case "proactive", "on-demand", "silent":
		h.config.Tutor.Intensity = intensity
		return fmt.Sprintf("Coaching intensity set to: %s", intensity)
	default:
		return fmt.Sprintf("Invalid intensity %q. Must be one of: proactive, on-demand, silent", intensity)
	}
}
