package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/plugin"
)

func NewUninstallPluginCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "uninstall-plugin",
		Short: "Remove Claude Code plugin and tutor instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := plugin.Scope(scope)
			if s != plugin.ScopeLocal && s != plugin.ScopeGlobal {
				return fmt.Errorf("invalid scope %q: must be 'local' or 'global'", scope)
			}

			projectDir := "."
			if s == plugin.ScopeGlobal {
				projectDir = ""
			}

			if err := plugin.Uninstall(projectDir, s); err != nil {
				return fmt.Errorf("uninstall failed: %w", err)
			}

			fmt.Println("Agent-tutor plugin removed.")
			return nil
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "local", "Uninstall scope: 'local' or 'global'")
	return cmd
}
