package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStatusCmd() *cobra.Command {
	var socket string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current session info",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("socket") {
				if s := socketFromConfig(); s != "" {
					socket = s
				}
			}
			tm := tmux.New(sessionName)
			tm.Socket = socket
			if !tm.HasSession() {
				fmt.Println("No active agent-tutor session.")
				return nil
			}
			fmt.Println("Agent Tutor session is running.")
			return nil
		},
	}
	cmd.Flags().StringVar(&socket, "socket", "agent-tutor", "tmux socket name")
	return cmd
}
