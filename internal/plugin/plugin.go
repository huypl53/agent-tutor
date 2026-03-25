package plugin

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:embed
var pluginFS embed.FS

type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeGlobal Scope = "global"
)

const beginSentinel = "<!-- BEGIN AGENT-TUTOR -->"
const endSentinel = "<!-- END AGENT-TUTOR -->"

// hookGroup matches Claude Code's settings.json PostToolUse hook format.
type hookGroup struct {
	Matcher string    `json:"matcher"`
	Hooks   []hookCmd `json:"hooks"`
}

type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

const agentTutorHookMarker = ".agent-tutor/plugin/hooks/"

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

## Commands Available

| Command | Purpose |
|---------|---------|
| ` + "`/atu:check`" + ` | Comprehensive review of recent activity |
| ` + "`/atu:hint`" + ` | Quick one-point nudge |
| ` + "`/atu:explain`" + ` | Explain the most recent error |
| ` + "`/atu:save`" + ` | Save current session as a lesson |
| ` + "`/atu:debug`" + ` | Guided debugging session (4-phase methodology) |
| ` + "`/atu:review`" + ` | Self-review coaching (graduated checklist) |
| ` + "`/atu:decompose`" + ` | Problem decomposition coaching |
| ` + "`/atu:workflow`" + ` | Development workflow habit coaching |

## Teaching Skills

When these commands are invoked, load the methodology by reading the corresponding skill file:

- ` + "`/atu:debug`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-guided-debugging/SKILL.md`" + `
- ` + "`/atu:decompose`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-problem-decomposition/SKILL.md`" + `
- ` + "`/atu:review`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-code-review-learning/SKILL.md`" + `
- ` + "`/atu:workflow`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-dev-workflow/SKILL.md`" + `

For deeper reference material, read the ` + "`references/`" + ` subdirectory of each skill.

## Coaching Behavior

- **proactive**: After messages, check ` + "`get_student_context`" + ` for teachable moments. On ` + "`tutor_nudge`" + `, offer coaching.
- **on-demand**: Only use tutor tools when the student asks or uses ` + "`/atu:check`" + `.
- **silent**: Never coach unless explicitly asked.

## Pedagogical Principles

- **Ask questions before giving answers.** "What do you think this error means?" before explaining.
- **One teaching point per interaction.** Never overwhelm with five things at once.
- **Praise specific good behavior first.** Acknowledge what worked before suggesting improvements.
- **Match depth to student level.** Vocabulary and checklist depth from ` + "`get_coaching_config`" + `.
- **Never fix code silently in proactive mode.** Always explain what and why.
- **If the student is doing well, say nothing.** Silence is valid coaching.

## Hook Awareness

The project has advisory hooks that inject ` + "`additionalContext`" + ` when:
- A file exceeds 200 lines after a Write/Edit (suggests ` + "`/atu:decompose`" + `)
- An error pattern appears in terminal output after a Bash command (suggests ` + "`/atu:debug`" + ` or ` + "`/atu:explain`" + `)

When ` + "`additionalContext`" + ` mentions a teachable moment, incorporate it naturally into your next response.
Do not parrot the hook text verbatim — use it as a trigger for genuine teaching.

## Lesson Auto-Save

After giving coaching feedback in these situations, also save a lesson file to ` + "`./lessons/`" + `:
- After responding to ` + "`/atu:check`" + ` — save the coaching feedback as a lesson
- After a ` + "`tutor_nudge`" + ` triggered by a git commit — save what was learned in that commit
- Whenever you explain a non-trivial concept and it would be valuable for review

Write each lesson to ` + "`./lessons/YYYY-MM-DD-<topic-slug>.md`" + ` using this template:
Create the ` + "`./lessons/`" + ` directory if it does not exist.

    # <Topic Title>

    **Date:** YYYY-MM-DD
    **Topic:** <category>
    **Trigger:** <check|commit|nudge|manual>

    ## What I Learned
    <Clear explanation tailored to student level>

    ## Code Example
    <Relevant code with annotations>

    ## Key Takeaway
    <One sentence to remember>

    ## Common Mistakes
    <Pitfalls to avoid>

Do not duplicate — if a lesson file for the same topic already exists today, skip it.
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

	// Merge hooks into .claude/settings.json (project-level, not user-level)
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	hooksDir := filepath.Join(destDir, "hooks")
	if err := mergeHookSettings(settingsPath, hooksDir); err != nil {
		return fmt.Errorf("updating settings.json: %w", err)
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
		// Embedded filenames already use dashes (e.g. "atu-check.md")
		// since go:embed forbids colons. Skill dirs keep dashes.
		skillName := strings.TrimSuffix(name, ".md")

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

	// Install teaching skills as global skills
	skills, err := fs.ReadDir(pluginFS, "embed/skills")
	if err == nil {
		for _, skillEntry := range skills {
			if !skillEntry.IsDir() {
				continue
			}
			skillName := skillEntry.Name()
			skillDestDir := filepath.Join(home, ".claude", "skills", skillName)
			skillSrcDir := "embed/skills/" + skillName
			if err := fs.WalkDir(pluginFS, skillSrcDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(skillSrcDir, path)
				if rel == "." {
					return os.MkdirAll(skillDestDir, 0o755)
				}
				dest := filepath.Join(skillDestDir, rel)
				if d.IsDir() {
					return os.MkdirAll(dest, 0o755)
				}
				data, err := pluginFS.ReadFile(path)
				if err != nil {
					return err
				}
				return os.WriteFile(dest, data, 0o644)
			}); err != nil {
				return fmt.Errorf("installing skill %s: %w", skillName, err)
			}
		}
	}

	// Append to ~/.claude/CLAUDE.md
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	return appendCLAUDEmd(claudeMD)
}

func uninstallLocal(projectDir string) error {
	// Remove plugin directory
	pluginDir := PluginDir(projectDir)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("removing plugin directory: %w", err)
	}

	// Remove hook entries from .claude/settings.json
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	if err := removeHookSettings(settingsPath); err != nil {
		return fmt.Errorf("removing hook settings: %w", err)
	}

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
	for _, name := range []string{
		"atu-check", "atu-hint", "atu-explain", "atu-save",
		"atu-debug", "atu-review", "atu-decompose", "atu-workflow",
		"atu-guided-debugging", "atu-problem-decomposition",
		"atu-code-review-learning", "atu-dev-workflow",
	} {
		if err := os.RemoveAll(filepath.Join(home, ".claude", "skills", name)); err != nil {
			return fmt.Errorf("removing skill %s: %w", name, err)
		}
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
		// Embedded files use dashes (atu-check.md) because go:embed
		// forbids colons. Restore colons for Claude Code command names.
		rel = restoreColons(rel)
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

// restoreColons converts embedded filenames like "commands/atu-check.md"
// back to "commands/atu:check.md" for Claude Code command registration.
func restoreColons(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	// Only restore colons for command files directly under commands/
	if dir == "commands" && strings.HasPrefix(base, "atu-") && strings.HasSuffix(base, ".md") {
		base = "atu:" + strings.TrimPrefix(base, "atu-")
	}
	if dir == "." {
		return base
	}
	return filepath.Join(dir, base)
}

// mergeHookSettings merges agent-tutor hook entries into .claude/settings.json.
// Preserves all existing settings. Idempotent.
func mergeHookSettings(settingsPath, hooksAbsDir string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	// Read existing settings as raw JSON map to preserve unknown fields.
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	// Parse existing hooks section.
	hooks := make(map[string]json.RawMessage)
	if h, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(h, &hooks)
	}

	// Parse existing PostToolUse entries.
	var postToolUse []hookGroup
	if p, ok := hooks["PostToolUse"]; ok {
		_ = json.Unmarshal(p, &postToolUse)
	}

	// Remove any existing agent-tutor entries (idempotency).
	postToolUse = removeAgentTutorHookGroups(postToolUse)

	// Add our two hooks.
	postToolUse = append(postToolUse,
		hookGroup{
			Matcher: "Write|Edit",
			Hooks: []hookCmd{{
				Type:    "command",
				Command: "node " + filepath.Join(hooksAbsDir, "large-file-detect.js"),
			}},
		},
		hookGroup{
			Matcher: "Bash",
			Hooks: []hookCmd{{
				Type:    "command",
				Command: "node " + filepath.Join(hooksAbsDir, "error-pattern-detect.js"),
			}},
		},
	)

	// Marshal back, preserving other fields.
	ptu, err := json.Marshal(postToolUse)
	if err != nil {
		return err
	}
	hooks["PostToolUse"] = ptu
	hooksRaw, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	raw["hooks"] = hooksRaw

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0o644)
}

// removeAgentTutorHookGroups filters out any hook groups that reference agent-tutor hooks.
func removeAgentTutorHookGroups(groups []hookGroup) []hookGroup {
	var result []hookGroup
	for _, g := range groups {
		isAgentTutor := false
		for _, h := range g.Hooks {
			if strings.Contains(h.Command, agentTutorHookMarker) {
				isAgentTutor = true
				break
			}
		}
		if !isAgentTutor {
			result = append(result, g)
		}
	}
	return result
}

// removeHookSettings removes agent-tutor hook entries from .claude/settings.json.
func removeHookSettings(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // File doesn't exist — nothing to do
	}

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // Unparseable — leave as-is
	}

	hooks := make(map[string]json.RawMessage)
	if h, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(h, &hooks)
	}

	var postToolUse []hookGroup
	if p, ok := hooks["PostToolUse"]; ok {
		_ = json.Unmarshal(p, &postToolUse)
	}

	postToolUse = removeAgentTutorHookGroups(postToolUse)

	if len(postToolUse) == 0 {
		delete(hooks, "PostToolUse")
	} else {
		ptu, err := json.Marshal(postToolUse)
		if err != nil {
			return err
		}
		hooks["PostToolUse"] = ptu
	}

	if len(hooks) == 0 {
		delete(raw, "hooks")
	} else {
		hooksRaw, err := json.Marshal(hooks)
		if err != nil {
			return err
		}
		raw["hooks"] = hooksRaw
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0o644)
}
