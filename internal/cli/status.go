package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/tmux"
)

func NewStatusCmd() *cobra.Command {
	var socket string
	cmd := &cobra.Command{
		Use:   "status [project-dir]",
		Short: "Show current session info",
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
				fmt.Printf("No active agent-tutor session for %s.\n", projectDir)
				return nil
			}
			fmt.Printf("Agent Tutor session is running (%s).\n", sessName)
			return nil
		},
	}
	cmd.Flags().StringVar(&socket, "socket", "agent-tutor", "tmux socket name")
	return cmd
}
