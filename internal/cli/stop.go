package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypham/agent-tutor/internal/tmux"
)

func NewStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the tutoring session",
		RunE: func(cmd *cobra.Command, args []string) error {
			tm := tmux.New(sessionName)
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
}
