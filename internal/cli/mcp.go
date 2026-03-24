package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/huypham/agent-tutor/internal/config"
	mcpserver "github.com/huypham/agent-tutor/internal/mcp"
	"github.com/huypham/agent-tutor/internal/store"
	"github.com/huypham/agent-tutor/internal/trigger"
	"github.com/huypham/agent-tutor/internal/watcher"
)

func NewMCPCmd() *cobra.Command {
	var projectDir string

	cmd := &cobra.Command{
		Use:    "mcp",
		Short:  "Run as MCP server (used internally)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP(projectDir)
		},
	}
	cmd.Flags().StringVar(&projectDir, "project-dir", ".", "Project directory to watch")
	return cmd
}

func runMCP(projectDir string) error {
	cfg, err := config.Load(projectDir)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	s := store.New()

	termInterval, _ := time.ParseDuration(cfg.Watchers.TerminalPollInterval)
	if termInterval == 0 {
		termInterval = 2 * time.Second
	}
	gitInterval, _ := time.ParseDuration(cfg.Watchers.GitPollInterval)
	if gitInterval == 0 {
		gitInterval = 5 * time.Second
	}

	fw, err := watcher.NewFileWatcher(projectDir, cfg.Watchers.FilePatterns, cfg.Watchers.IgnorePatterns, s)
	if err == nil {
		fw.Start(ctx)
		defer fw.Stop()
	}

	tw := watcher.NewTerminalWatcher(sessionName, "0", termInterval, s)
	tw.Start(ctx)
	defer tw.Stop()

	gw := watcher.NewGitWatcher(projectDir, gitInterval, s)
	gw.Start(ctx)
	defer gw.Stop()

	_ = trigger.New(func(event string) {
		// Future: send MCP notification
	})

	srv := mcpserver.NewServer(s, cfg)
	return srv.Run(ctx)
}
