package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/config"
	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStopCmd() *cobra.Command {
	var socket string
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the tutoring session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("socket") {
				if s := socketFromConfig(); s != "" {
					socket = s
				}
			}
			tm := tmux.New(sessionName)
			tm.Socket = socket
			if !tm.HasSession() {
				return fmt.Errorf("no active agent-tutor session")
			}
			if err := tm.KillSession(); err != nil {
				return fmt.Errorf("killing session: %w", err)
			}
			fmt.Println("Agent Tutor session stopped.")
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
