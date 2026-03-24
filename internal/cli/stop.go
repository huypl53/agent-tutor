package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStopCmd() *cobra.Command {
	var socket string
	cmd := &cobra.Command{
		Use:   "stop [project-dir]",
		Short: "Stop the tutoring session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := resolveProjectDir(args)
			if !cmd.Flags().Changed("socket") {
				if s := socketFromConfig(); s != "" {
					socket = s
				}
			}
			sessName := SessionName(projectDir)
			tm := tmux.New(sessName)
			tm.Socket = socket
			if !tm.HasSession() {
				return fmt.Errorf("no active agent-tutor session for %s", projectDir)
			}
			if err := tm.KillSession(); err != nil {
				return fmt.Errorf("killing session: %w", err)
			}
			fmt.Printf("Agent Tutor session stopped (%s).\n", sessName)
			return nil
		},
	}
	cmd.Flags().StringVar(&socket, "socket", "agent-tutor", "tmux socket name")
	return cmd
}

// socketFromConfig tries to load the socket name from the project config.
func socketFromConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return ""
	}
	return cfg.Tmux.Socket
}

// resolveProjectDir returns an absolute project directory from CLI args,
// falling back to the current working directory.
func resolveProjectDir(args []string) string {
	if len(args) > 0 {
		if abs, err := filepath.Abs(args[0]); err == nil {
			return abs
		}
	}
	dir, _ := os.Getwd()
	return dir
}
