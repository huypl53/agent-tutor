package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/plugin"
	"github.com/huypl53/agent-tutor/internal/tmux"
	"github.com/huypl53/agent-tutor/internal/tui"
)

func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [project-dir]",
		Short: "Start a tutoring session",
		Long:  "Set up tmux with side-by-side panes: your terminal + coding agent with tutor capabilities.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runStart,
	}
	cmd.Flags().Bool("tui", false, "Launch the bubbletea TUI instead of attaching via tmux")
	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	projectDir := resolveProjectDir(args)

	cfg, err := config.Load(projectDir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Auto-install plugin if not present
	pluginDir := plugin.PluginDir(projectDir)
	if !plugin.IsInstalled(projectDir) {
		fmt.Println("Installing agent-tutor plugin...")
		if err := plugin.Install(projectDir, plugin.ScopeLocal); err != nil {
			return fmt.Errorf("auto-installing plugin: %w", err)
		}
	}

	sessName := SessionName(projectDir)
	tm := tmux.New(sessName)
	tm.Socket = cfg.Tmux.Socket
	if tm.HasSession() {
		return fmt.Errorf("session %q already exists — run 'agent-tutor stop' first", sessName)
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
	mcpConfig := map[string]any{
		"mcpServers": map[string]any{
			"agent-tutor": map[string]any{
				"command": self,
				"args":    []string{"mcp", "--project-dir", projectDir, "--socket", cfg.Tmux.Socket, "--session", sessName},
			},
		},
	}
	mcpJSON, _ := json.Marshal(mcpConfig)
	agentCmd := fmt.Sprintf("%s --mcp-config '%s' --plugin-dir %q", cfg.Agent.Command, string(mcpJSON), pluginDir)
	if err := tm.SendKeys("1", agentCmd); err != nil {
		tm.KillSession()
		return fmt.Errorf("starting agent: %w", err)
	}

	fmt.Printf("Agent Tutor session started.\n")
	fmt.Printf("  Session: %s\n", sessName)
	fmt.Printf("  Project: %s\n", projectDir)
	fmt.Printf("  Agent: %s\n", cfg.Agent.Command)
	fmt.Printf("  Coaching: %s\n", cfg.GetIntensity())
	fmt.Printf("\nAttaching to tmux session...\n")
	fmt.Printf("Left pane: your terminal. Right pane: your coding agent.\n")
	fmt.Printf("Type /atu:check in the agent to get feedback on your work.\n\n")

	useTUI, _ := cmd.Flags().GetBool("tui")
	if useTUI {
		model := tui.New(tm, cfg, sessName)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		fmt.Printf("TUI detached. Session %q is still running.\n", sessName)
		return nil
	}

	attachCmd := fmt.Sprintf("tmux -L %q attach-session -t %s", cfg.Tmux.Socket, sessName)
	return syscall.Exec("/usr/bin/env", []string{"env", "bash", "-c", attachCmd}, os.Environ())
}
