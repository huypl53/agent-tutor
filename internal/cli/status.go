package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypham/agent-tutor/internal/tmux"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current session info",
		RunE: func(cmd *cobra.Command, args []string) error {
			tm := tmux.New(sessionName)
			if !tm.HasSession() {
				fmt.Println("No active agent-tutor session.")
				return nil
			}
			fmt.Println("Agent Tutor session is running.")
			return nil
		},
	}
}
