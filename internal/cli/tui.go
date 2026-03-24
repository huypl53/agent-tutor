package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/tmux"
	"github.com/huypl53/agent-tutor/internal/tui"
)

func NewTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui [project-dir]",
		Short: "Launch the TUI for an existing tutoring session",
		Long:  "Open a dual-pane terminal UI showing your terminal and coding agent. The tmux session must already be running (use 'agent-tutor start' first).",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTUI,
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	projectDir := resolveProjectDir(args)

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	sessName := SessionName(projectDir)
	tm := tmux.New(sessName)
	tm.Socket = cfg.Tmux.Socket

	if !tm.HasSession() {
		return fmt.Errorf("no session %q found — run 'agent-tutor start' first", sessName)
	}

	model := tui.New(tm, cfg, sessName)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	fmt.Printf("TUI detached. Session %q is still running.\n", sessName)
	fmt.Printf("Reattach with: agent-tutor tui %s\n", projectDir)
	fmt.Printf("Or use tmux:   tmux -L %q attach-session -t %s\n", cfg.Tmux.Socket, sessName)
	return nil
}
