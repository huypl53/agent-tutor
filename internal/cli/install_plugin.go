package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/huypl53/agent-tutor/internal/plugin"
)

func NewInstallPluginCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "install-plugin",
		Short: "Install Claude Code plugin and tutor instructions",
		Long:  "Installs slash commands (/atu:check, /atu:hint, /atu:explain) and appends tutor instructions to CLAUDE.md.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := plugin.Scope(scope)
			if s != plugin.ScopeLocal && s != plugin.ScopeGlobal {
				return fmt.Errorf("invalid scope %q: must be 'local' or 'global'", scope)
			}

			projectDir := "."
			if s == plugin.ScopeLocal {
				fmt.Println("Installing agent-tutor plugin locally...")
			} else {
				fmt.Println("Installing agent-tutor plugin globally...")
				projectDir = ""
			}

			if err := plugin.Install(projectDir, s); err != nil {
				return fmt.Errorf("install failed: %w", err)
			}

			if s == plugin.ScopeLocal {
				fmt.Println("  Plugin: .agent-tutor/plugin/")
				fmt.Println("  CLAUDE.md: .claude/CLAUDE.md (appended)")
			} else {
				fmt.Println("  Skills: ~/.claude/skills/atu-{check,hint,explain}/")
				fmt.Println("  CLAUDE.md: ~/.claude/CLAUDE.md (appended)")
			}
			fmt.Println("\nAvailable commands: /atu:check, /atu:hint, /atu:explain")
			return nil
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "local", "Installation scope: 'local' (this project) or 'global' (all projects)")
	return cmd
}
