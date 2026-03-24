# Embedded Claude Code Plugin for Agent Tutor

## Goal

Bundle a Claude Code plugin inside the agent-tutor binary so that when Claude is launched in the tutoring session, it has real slash commands (`/atu:check`, `/atu:hint`, `/atu:explain`) and persistent tutor instructions via CLAUDE.md.

## Architecture

Two artifacts are installed into the target project:

1. **Plugin directory** (`.agent-tutor/plugin/`) — contains slash command skill files
2. **CLAUDE.md section** (appended to `.claude/CLAUDE.md`) — MCP tool reference + teaching guidelines

Plugin files are embedded in the Go binary via `//go:embed`. The `install-plugin` command extracts them. The `start` command auto-installs if missing.

## Plugin Structure

```
.agent-tutor/plugin/
├── .claude-plugin/
│   └── plugin.json
└── commands/
    ├── atu:check.md       # comprehensive review using all MCP tools
    ├── atu:hint.md        # light nudge based on recent activity
    └── atu:explain.md     # explain last error/terminal output
```

### plugin.json

```json
{
  "name": "agent-tutor",
  "version": "0.1.0",
  "description": "Programming tutor skills for agent-tutor sessions",
  "author": { "name": "agent-tutor" },
  "commands": "./commands/"
}
```

### /atu:check

Calls `get_recent_file_changes`, `get_terminal_activity`, `get_git_activity` and provides comprehensive coaching feedback on recent student activity.

### /atu:hint

Calls `get_student_context` and gives a brief nudge — one teaching point, no full review.

### /atu:explain

Calls `get_terminal_activity` and explains the most recent error or output in detail, tailored to student level.

## CLAUDE.md Section

Appended to `.claude/CLAUDE.md` with sentinel comments for clean uninstall:

```markdown
<!-- BEGIN AGENT-TUTOR -->
# Agent Tutor

You are a programming tutor. A student is working in a terminal pane next to you.
You have MCP tools to observe their work — use them to provide relevant coaching.

## MCP Tools Reference

| Tool | Purpose | When to use |
|------|---------|-------------|
| `get_student_context` | 5-minute activity summary (markdown) | Quick overview of what the student is doing |
| `get_recent_file_changes` | File changes with diffs | When reviewing code the student wrote |
| `get_terminal_activity` | Recent terminal output | When the student hits errors or runs commands |
| `get_git_activity` | Commits and status changes | When the student commits or has uncommitted work |
| `get_coaching_config` | Current intensity and level | Check before deciding how proactive to be |
| `set_coaching_intensity` | Change coaching mode | When the student asks to adjust coaching |

## Coaching Behavior

- **proactive**: After messages, check `get_student_context` for teachable moments. On `tutor_nudge`, offer coaching.
- **on-demand**: Only use tutor tools when the student asks or uses `/atu:check`.
- **silent**: Never coach unless explicitly asked.

## Teaching Style

- Explain the "why" not just the "what"
- One teaching point per interaction, not five
- For beginners: explain concepts, suggest resources
- For experienced devs: focus on idioms, best practices
- If the student is doing well, say nothing
<!-- END AGENT-TUTOR -->
```

## CLI Commands

### `agent-tutor install-plugin`

```
agent-tutor install-plugin [--scope local|global]
```

- `local` (default): writes plugin to `.agent-tutor/plugin/`, appends to `.claude/CLAUDE.md` in current project
- `global`: writes plugin to `~/.claude/skills/atu-check/`, `~/.claude/skills/atu-hint/`, `~/.claude/skills/atu-explain/`, appends to `~/.claude/CLAUDE.md`

### `agent-tutor uninstall-plugin`

```
agent-tutor uninstall-plugin [--scope local|global]
```

- Removes plugin directory
- Removes `<!-- BEGIN AGENT-TUTOR -->...<!-- END AGENT-TUTOR -->` section from CLAUDE.md

## Integration with `start`

The `start` command:
1. Checks if `.agent-tutor/plugin/` exists
2. If not, auto-runs local install (extract embedded files + append CLAUDE.md)
3. Passes `--plugin-dir .agent-tutor/plugin` to the claude command

```go
agentCmd := fmt.Sprintf("%s --mcp-config '%s' --plugin-dir %q",
    cfg.Agent.Command, string(mcpJSON), pluginDir)
```

## Embedding

Plugin files live in `internal/plugin/embed/` and are embedded via `//go:embed`:

```go
//go:embed embed/*
var pluginFS embed.FS
```

The `Install()` function walks the embedded FS and writes files to the target directory.

## Scope Behavior

| Scope | Plugin location | CLAUDE.md location | Auto-installed by `start` |
|-------|----------------|-------------------|--------------------------|
| local | `.agent-tutor/plugin/` | `.claude/CLAUDE.md` | Yes |
| global | `~/.claude/skills/atu-*/` | `~/.claude/CLAUDE.md` | No |
