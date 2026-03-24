package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/tmux"
)

const sessionName = "agent-tutor"

func NewStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start [project-dir]",
		Short: "Start a tutoring session",
		Long:  "Set up tmux with side-by-side panes: your terminal + coding agent with tutor capabilities.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	projectDir, _ := os.Getwd()
	if len(args) > 0 {
		var err error
		projectDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	tm := tmux.New(sessionName)
	tm.Socket = cfg.Tmux.Socket
	if tm.HasSession() {
		return fmt.Errorf("session %q already exists — run 'agent-tutor stop' first", sessionName)
	}

	if err := tm.CreateSession(projectDir); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	agentPaneSize := 100 - cfg.Tmux.UserPaneSize
	if err := tm.SplitPane(agentPaneSize, cfg.Tmux.Layout); err != nil {
		tm.KillSession()
		return fmt.Errorf("splitting pane: %w", err)
	}

	self, _ := os.Executable()
	mcpCmd := fmt.Sprintf("%s mcp --project-dir %q --socket %q", self, projectDir, cfg.Tmux.Socket)
	agentCmd := fmt.Sprintf("%s --mcp-server '%s'", cfg.Agent.Command, mcpCmd)
	if err := tm.SendKeys("1", agentCmd); err != nil {
		tm.KillSession()
		return fmt.Errorf("starting agent: %w", err)
	}

	fmt.Printf("Agent Tutor session started.\n")
	fmt.Printf("  Project: %s\n", projectDir)
	fmt.Printf("  Agent: %s\n", cfg.Agent.Command)
	fmt.Printf("  Coaching: %s\n", cfg.GetIntensity())
	fmt.Printf("\nAttaching to tmux session...\n")
	fmt.Printf("Left pane: your terminal. Right pane: your coding agent.\n")
	fmt.Printf("Type /check in the agent to get feedback on your work.\n\n")

	attachCmd := fmt.Sprintf("tmux -L %q attach-session -t %s", cfg.Tmux.Socket, sessionName)
	return syscall.Exec("/usr/bin/env", []string{"env", "bash", "-c", attachCmd}, os.Environ())
}
