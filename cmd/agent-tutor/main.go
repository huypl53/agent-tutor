package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/cli"
)

func main() {
	root := &cobra.Command{
		Use:   "agent-tutor",
		Short: "A programming tutor that coaches you through your coding agent",
	}

	root.AddCommand(cli.NewStartCmd())
	root.AddCommand(cli.NewStopCmd())
	root.AddCommand(cli.NewStatusCmd())
	root.AddCommand(cli.NewMCPCmd())
	root.AddCommand(cli.NewInstallPluginCmd())
	root.AddCommand(cli.NewUninstallPluginCmd())
	root.AddCommand(cli.NewTUICmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
