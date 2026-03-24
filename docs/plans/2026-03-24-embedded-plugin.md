# Embedded Plugin Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bundle a Claude Code plugin in the agent-tutor binary with `/atu:check`, `/atu:hint`, `/atu:explain` commands and CLAUDE.md-based tutor instructions. Users install via `agent-tutor install-plugin --scope local|global`.

**Architecture:** Plugin files are embedded in the Go binary via `//go:embed` under `internal/plugin/`. An `Install()` function extracts them to the target directory and appends tutor instructions to CLAUDE.md with sentinel comments. The `start` command auto-installs locally if the plugin is missing. An `uninstall-plugin` command cleanly removes both artifacts.

**Tech Stack:** Go `embed` package, `os` for file I/O, cobra for CLI commands

---

### Task 1: Create embedded plugin files

**Files:**
- Create: `internal/plugin/embed/.claude-plugin/plugin.json`
- Create: `internal/plugin/embed/commands/atu:check.md`
- Create: `internal/plugin/embed/commands/atu:hint.md`
- Create: `internal/plugin/embed/commands/atu:explain.md`

**Step 1: Create plugin.json**

Create `internal/plugin/embed/.claude-plugin/plugin.json`:

```json
{
  "name": "agent-tutor",
  "version": "0.1.0",
  "description": "Programming tutor skills for agent-tutor sessions",
  "author": { "name": "agent-tutor" },
  "commands": "./commands/"
}
```

**Step 2: Create atu:check.md**

Create `internal/plugin/embed/commands/atu:check.md`:

```markdown
---
name: atu:check
description: Comprehensive review of your recent coding activity with coaching feedback
---

Review the student's recent work by gathering all available context, then provide coaching feedback.

1. Call `get_recent_file_changes` to see what code was written or modified
2. Call `get_terminal_activity` to see commands run and any errors
3. Call `get_git_activity` to see commits and working tree status
4. Call `get_coaching_config` to check the student's level

Based on all gathered context, provide coaching feedback:
- Point out what the student did well
- Identify one or two areas for improvement (don't overwhelm)
- If there are errors, explain why they happened and how to fix them
- If the code works but could be improved, explain the idiomatic approach
- Tailor your language to the student's level (beginner vs experienced)
```

**Step 3: Create atu:hint.md**

Create `internal/plugin/embed/commands/atu:hint.md`:

```markdown
---
name: atu:hint
description: Quick nudge — one teaching point based on what you're currently doing
---

Give the student a brief, focused hint about their current work.

1. Call `get_student_context` to get a quick overview of recent activity
2. Call `get_coaching_config` to check the student's level

Based on the context, give exactly ONE teaching point:
- Keep it to 2-3 sentences
- Focus on the most impactful thing they could improve right now
- If they're doing well, say so briefly and don't force a teaching moment
- Frame it as a suggestion, not a correction
```

**Step 4: Create atu:explain.md**

Create `internal/plugin/embed/commands/atu:explain.md`:

```markdown
---
name: atu:explain
description: Explain the most recent error or terminal output in detail
---

Explain what just happened in the student's terminal.

1. Call `get_terminal_activity` to see recent terminal output
2. Call `get_coaching_config` to check the student's level

Find the most recent error or notable output and explain it:
- What the error means in plain language
- Why it happened (the root cause, not just the symptom)
- How to fix it, step by step
- For beginners: explain the underlying concept
- For experienced devs: focus on the specific fix and any non-obvious gotchas

If there are no errors in the recent output, explain what the last command did and whether the output looks correct.
```

**Step 5: Verify directory structure**

Run: `find internal/plugin/embed -type f | sort`
Expected:
```
internal/plugin/embed/.claude-plugin/plugin.json
internal/plugin/embed/commands/atu:check.md
internal/plugin/embed/commands/atu:hint.md
internal/plugin/embed/commands/atu:explain.md
```

**Step 6: Commit**

```bash
git add internal/plugin/embed/
git commit -m "feat: add embedded plugin files for Claude Code commands"
```

---

### Task 2: Create plugin install/uninstall package

**Files:**
- Create: `internal/plugin/plugin.go`
- Create: `internal/plugin/plugin_test.go`

**Step 1: Write tests for Install and Uninstall**

Create `internal/plugin/plugin_test.go`:

```go
package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallLocal(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify plugin files exist
	files := []string{
		".agent-tutor/plugin/.claude-plugin/plugin.json",
		".agent-tutor/plugin/commands/atu:check.md",
		".agent-tutor/plugin/commands/atu:hint.md",
		".agent-tutor/plugin/commands/atu:explain.md",
	}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	// Verify CLAUDE.md was created with sentinels
	claudeMD := filepath.Join(dir, ".claude", "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("missing BEGIN sentinel")
	}
	if !strings.Contains(content, "<!-- END AGENT-TUTOR -->") {
		t.Error("missing END sentinel")
	}
	if !strings.Contains(content, "get_student_context") {
		t.Error("missing MCP tool reference")
	}
}

func TestInstallLocalAppendsCLAUDEmd(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# My Project\n\nExisting content.\n"), 0o644)

	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	content := string(data)
	if !strings.HasPrefix(content, "# My Project") {
		t.Error("existing content was overwritten")
	}
	if !strings.Contains(content, "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("agent-tutor section not appended")
	}
}

func TestInstallLocalIdempotent(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	Install(dir, ScopeLocal)

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	count := strings.Count(string(data), "<!-- BEGIN AGENT-TUTOR -->")
	if count != 1 {
		t.Errorf("expected 1 sentinel block, got %d", count)
	}
}

func TestUninstallLocal(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	if err := Uninstall(dir, ScopeLocal); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Plugin dir removed
	pluginDir := filepath.Join(dir, ".agent-tutor", "plugin")
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("plugin directory should be removed")
	}

	// CLAUDE.md sentinel section removed
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if strings.Contains(string(data), "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("sentinel section should be removed")
	}
}

func TestUninstallPreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# My Project\n\nKeep this.\n"), 0o644)

	Install(dir, ScopeLocal)
	Uninstall(dir, ScopeLocal)

	data, _ := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	content := string(data)
	if !strings.Contains(content, "Keep this.") {
		t.Error("non-agent-tutor content was removed")
	}
}

func TestInstallGlobal(t *testing.T) {
	dir := t.TempDir()
	// Override home for test
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	defer os.Setenv("HOME", origHome)

	if err := Install("", ScopeGlobal); err != nil {
		t.Fatalf("Install global failed: %v", err)
	}

	// Verify skills exist
	skills := []string{
		".claude/skills/atu-check/SKILL.md",
		".claude/skills/atu-hint/SKILL.md",
		".claude/skills/atu-explain/SKILL.md",
	}
	for _, s := range skills {
		path := filepath.Join(dir, s)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", s, err)
		}
	}

	// Verify CLAUDE.md
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("global CLAUDE.md not created: %v", err)
	}
	if !strings.Contains(string(data), "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("missing sentinel in global CLAUDE.md")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/plugin/ -v`
Expected: FAIL (package doesn't exist yet)

**Step 3: Implement plugin.go**

Create `internal/plugin/plugin.go`:

```go
package plugin

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed embed
var pluginFS embed.FS

type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeGlobal Scope = "global"
)

const beginSentinel = "<!-- BEGIN AGENT-TUTOR -->"
const endSentinel = "<!-- END AGENT-TUTOR -->"

const claudeMDSection = `<!-- BEGIN AGENT-TUTOR -->
# Agent Tutor

You are a programming tutor. A student is working in a terminal pane next to you.
You have MCP tools to observe their work — use them to provide relevant coaching.

## MCP Tools Reference

| Tool | Purpose | When to use |
|------|---------|-------------|
| ` + "`get_student_context`" + ` | 5-minute activity summary (markdown) | Quick overview of what the student is doing |
| ` + "`get_recent_file_changes`" + ` | File changes with diffs | When reviewing code the student wrote |
| ` + "`get_terminal_activity`" + ` | Recent terminal output | When the student hits errors or runs commands |
| ` + "`get_git_activity`" + ` | Commits and status changes | When the student commits or has uncommitted work |
| ` + "`get_coaching_config`" + ` | Current intensity and level | Check before deciding how proactive to be |
| ` + "`set_coaching_intensity`" + ` | Change coaching mode | When the student asks to adjust coaching |

## Coaching Behavior

- **proactive**: After messages, check ` + "`get_student_context`" + ` for teachable moments. On ` + "`tutor_nudge`" + `, offer coaching.
- **on-demand**: Only use tutor tools when the student asks or uses ` + "`/atu:check`" + `.
- **silent**: Never coach unless explicitly asked.

## Teaching Style

- Explain the "why" not just the "what"
- One teaching point per interaction, not five
- For beginners: explain concepts, suggest resources
- For experienced devs: focus on idioms, best practices
- If the student is doing well, say nothing
<!-- END AGENT-TUTOR -->`

// Install extracts embedded plugin files and appends CLAUDE.md section.
func Install(projectDir string, scope Scope) error {
	switch scope {
	case ScopeLocal:
		return installLocal(projectDir)
	case ScopeGlobal:
		return installGlobal()
	default:
		return fmt.Errorf("unknown scope: %s", scope)
	}
}

// Uninstall removes plugin files and CLAUDE.md section.
func Uninstall(projectDir string, scope Scope) error {
	switch scope {
	case ScopeLocal:
		return uninstallLocal(projectDir)
	case ScopeGlobal:
		return uninstallGlobal()
	default:
		return fmt.Errorf("unknown scope: %s", scope)
	}
}

// PluginDir returns the local plugin directory path for a project.
func PluginDir(projectDir string) string {
	return filepath.Join(projectDir, ".agent-tutor", "plugin")
}

// IsInstalled checks if the plugin is installed locally in the project.
func IsInstalled(projectDir string) bool {
	_, err := os.Stat(filepath.Join(PluginDir(projectDir), ".claude-plugin", "plugin.json"))
	return err == nil
}

func installLocal(projectDir string) error {
	destDir := PluginDir(projectDir)

	// Extract embedded files to .agent-tutor/plugin/
	if err := extractEmbedded(destDir); err != nil {
		return fmt.Errorf("extracting plugin files: %w", err)
	}

	// Append to .claude/CLAUDE.md
	claudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := appendCLAUDEmd(claudeMD); err != nil {
		return fmt.Errorf("updating CLAUDE.md: %w", err)
	}

	return nil
}

func installGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}

	// Install each command as a global skill
	commands, err := fs.ReadDir(pluginFS, "embed/commands")
	if err != nil {
		return fmt.Errorf("reading embedded commands: %w", err)
	}

	for _, entry := range commands {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// "atu:check.md" -> "atu-check"
		skillName := strings.TrimSuffix(name, ".md")
		skillName = strings.ReplaceAll(skillName, ":", "-")

		skillDir := filepath.Join(home, ".claude", "skills", skillName)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return err
		}

		data, err := pluginFS.ReadFile("embed/commands/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), data, 0o644); err != nil {
			return err
		}
	}

	// Append to ~/.claude/CLAUDE.md
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	return appendCLAUDEmd(claudeMD)
}

func uninstallLocal(projectDir string) error {
	// Remove plugin directory
	pluginDir := PluginDir(projectDir)
	os.RemoveAll(pluginDir)

	// Remove CLAUDE.md section
	claudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	return removeCLAUDEmdSection(claudeMD)
}

func uninstallGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Remove skill directories
	for _, name := range []string{"atu-check", "atu-hint", "atu-explain"} {
		os.RemoveAll(filepath.Join(home, ".claude", "skills", name))
	}

	// Remove CLAUDE.md section
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	return removeCLAUDEmdSection(claudeMD)
}

func extractEmbedded(destDir string) error {
	return fs.WalkDir(pluginFS, "embed", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip "embed/" prefix to get relative path
		rel, _ := filepath.Rel("embed", path)
		if rel == "." {
			return nil
		}
		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		data, err := pluginFS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

func appendCLAUDEmd(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	existing, _ := os.ReadFile(path)
	content := string(existing)

	// Idempotent: if already present, replace it
	if strings.Contains(content, beginSentinel) {
		content = removeSentinelBlock(content)
	}

	// Append with a blank line separator
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content != "" {
		content += "\n"
	}
	content += claudeMDSection + "\n"

	return os.WriteFile(path, []byte(content), 0o644)
}

func removeCLAUDEmdSection(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := removeSentinelBlock(string(data))
	return os.WriteFile(path, []byte(content), 0o644)
}

func removeSentinelBlock(content string) string {
	beginIdx := strings.Index(content, beginSentinel)
	endIdx := strings.Index(content, endSentinel)
	if beginIdx < 0 || endIdx < 0 {
		return content
	}

	before := content[:beginIdx]
	after := content[endIdx+len(endSentinel):]

	// Clean up extra blank lines
	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	if before == "" {
		return after
	}
	if after == "" {
		return before + "\n"
	}
	return before + "\n\n" + after
}
```

**Step 4: Run tests**

Run: `go test ./internal/plugin/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/plugin/
git commit -m "feat: add plugin install/uninstall package with embedded files"
```

---

### Task 3: Add install-plugin and uninstall-plugin CLI commands

**Files:**
- Create: `internal/cli/install_plugin.go`
- Create: `internal/cli/uninstall_plugin.go`
- Modify: `cmd/agent-tutor/main.go`

**Step 1: Create install_plugin.go**

Create `internal/cli/install_plugin.go`:

```go
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
```

**Step 2: Create uninstall_plugin.go**

Create `internal/cli/uninstall_plugin.go`:

```go
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
```

**Step 3: Register commands in main.go**

In `cmd/agent-tutor/main.go`, add:

```go
root.AddCommand(cli.NewInstallPluginCmd())
root.AddCommand(cli.NewUninstallPluginCmd())
```

**Step 4: Build and verify**

Run: `go build ./... && go run ./cmd/agent-tutor install-plugin --help`
Expected: shows help with --scope flag

**Step 5: Commit**

```bash
git add internal/cli/install_plugin.go internal/cli/uninstall_plugin.go cmd/agent-tutor/main.go
git commit -m "feat: add install-plugin and uninstall-plugin CLI commands"
```

---

### Task 4: Update start command to auto-install and pass --plugin-dir

**Files:**
- Modify: `internal/cli/start.go`

**Step 1: Update start.go to auto-install plugin and pass --plugin-dir**

In `internal/cli/start.go`, add import for `plugin` package and update `runStart`:

```go
import (
	// ... existing imports ...
	"github.com/huypl53/agent-tutor/internal/plugin"
)
```

After loading config and before creating the tmux session, add auto-install:

```go
// Auto-install plugin if not present
pluginDir := plugin.PluginDir(projectDir)
if !plugin.IsInstalled(projectDir) {
	fmt.Println("Installing agent-tutor plugin...")
	if err := plugin.Install(projectDir, plugin.ScopeLocal); err != nil {
		return fmt.Errorf("auto-installing plugin: %w", err)
	}
}
```

Update the agent command to include `--plugin-dir`:

```go
agentCmd := fmt.Sprintf("%s --mcp-config '%s' --plugin-dir %q", cfg.Agent.Command, string(mcpJSON), pluginDir)
```

Update the user-facing message to reference `/atu:check` instead of `/check`:

```go
fmt.Printf("Type /atu:check in the agent to get feedback on your work.\n\n")
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: compiles clean

**Step 3: Commit**

```bash
git add internal/cli/start.go
git commit -m "feat: auto-install plugin on start, pass --plugin-dir to claude"
```

---

### Task 5: Update docs

**Files:**
- Modify: `README.md`
- Modify: `docs/architecture.md`

**Step 1: Update README.md**

Add plugin installation section after the Commands table:

```markdown
## Plugin Installation

Agent-tutor includes a Claude Code plugin with coaching slash commands. It is auto-installed on `agent-tutor start`, or you can install it manually:

```bash
# Install in current project (default)
agent-tutor install-plugin

# Install globally for all projects
agent-tutor install-plugin --scope global

# Remove
agent-tutor uninstall-plugin
```

### Slash Commands

| Command | Description |
|---------|-------------|
| `/atu:check` | Comprehensive review of recent coding activity |
| `/atu:hint` | Quick nudge — one teaching point |
| `/atu:explain` | Explain the most recent error or output |
```

Update the Commands table to include install-plugin and uninstall-plugin.

Update the `/check` reference to `/atu:check`.

**Step 2: Update architecture.md**

Add a section about the plugin system under Components:

```markdown
### Plugin (`internal/plugin`)

Embeds Claude Code plugin files via `//go:embed`. The `Install()` function extracts plugin files and appends a tutor instruction section to `.claude/CLAUDE.md` with `<!-- BEGIN AGENT-TUTOR -->` / `<!-- END AGENT-TUTOR -->` sentinel comments for clean uninstall.

Two scopes:
- **local**: Plugin in `.agent-tutor/plugin/`, instructions in `.claude/CLAUDE.md`
- **global**: Skills in `~/.claude/skills/atu-*/`, instructions in `~/.claude/CLAUDE.md`

The `start` command auto-installs locally if the plugin is not present, and passes `--plugin-dir` to the claude command.
```

**Step 3: Commit**

```bash
git add README.md docs/architecture.md
git commit -m "docs: add plugin installation and slash commands documentation"
```
