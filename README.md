# Agent Tutor

A programming tutor plugin for coding agents (Claude Code, Codex CLI). Observes your work via file changes and git activity, then coaches you through MCP tools and slash commands.

## Installation

```bash
npm install -g agent-tutor
```

Requires Node.js 18+ and git on your PATH.

## Quick Start

### Claude Code (plugin mode)

```bash
claude --plugin-dir $(npx agent-tutor plugin-dir)
```

Or inject tutor instructions into your project:

```bash
npx agent-tutor install
```

### Codex CLI

```bash
npx agent-tutor install --agent codex
codex mcp add agent-tutor -- node $(npx agent-tutor plugin-dir)/servers/tutoring-mcp.js
```

## Commands

| Command | Description |
|---------|-------------|
| `agent-tutor install [--scope] [--agent]` | Install tutor instructions and show plugin setup |
| `agent-tutor uninstall [--scope] [--agent]` | Remove tutor instructions |
| `agent-tutor plugin-dir` | Print the plugin directory path (for `--plugin-dir`) |

Options:
- `--scope local|global` — local (default) writes to `.claude/CLAUDE.md`, global writes to `~/.claude/CLAUDE.md`
- `--agent claude|codex` — target agent (default: claude)

## Slash Commands

| Command | Description |
|---------|-------------|
| `/atu:check` | Comprehensive review of recent coding activity |
| `/atu:hint` | Quick nudge — one teaching point |
| `/atu:explain` | Explain the most recent error or output |
| `/atu:save` | Save a lesson to `./lessons/` for later review |
| `/atu:debug` | Guided debugging session (4-phase methodology) |
| `/atu:review` | Self-review coaching with graduated checklist |
| `/atu:decompose` | Problem decomposition coaching |
| `/atu:workflow` | Development workflow habit coaching |
| `/atu:plan` | Create a learning plan or show progress |

## Lesson Export

Agent Tutor saves structured lesson files to `./lessons/` in your project directory.

**On-demand:** Type `/atu:save goroutines` to explicitly save a lesson about a topic.

**Automatic:** Lessons are saved after `/atu:check` feedback and git commit coaching nudges.

Each lesson follows this structure:

    # Topic Title

    **Date:** 2026-03-24
    **Topic:** category
    **Trigger:** manual|check|commit|nudge

    ## What I Learned
    ## Code Example
    ## Key Takeaway
    ## Common Mistakes

Add `lessons/` to `.gitignore` to keep them local, or commit them to share.

## Topic Tracking

The tutor tracks what you're learning in `.agent-tutor/current-topic.md`. It records key moments (struggles, hints, breakthroughs) and saves a lesson when you move to a new topic. After `/clear` or `/compact`, it reads this file to recover context.

## Learning Plans

Create a structured learning path with `/atu:plan`:

```
/atu:plan Build a REST API           # creates a 4-8 step plan
/atu:plan                            # shows current progress
/atu:plan next                       # marks current step done, advances
```

Plans are stored in `.agent-tutor/learning-plan.md` and integrate with topic tracking.

## Configuration

Config is stored in `.agent-tutor/config.json`:

```json
{
  "intensity": "on-demand",
  "level": "auto"
}
```

### Coaching intensity levels

- **silent** — Never coaches unless you explicitly ask.
- **on-demand** — Only coaches when you ask or use `/atu:check`.
- **proactive** — Checks your context and offers coaching when it spots teachable moments.

Change intensity via MCP tool: the agent can call `set_coaching_intensity` with `proactive`, `on-demand`, or `silent`.

## How It Works

Agent Tutor is a Claude Code plugin with three components:

1. **MCP Server** (`plugin/servers/tutoring-mcp.js`) — Node.js server providing 5 tools over stdio: `get_student_context`, `get_recent_file_changes`, `get_git_activity`, `get_coaching_config`, `set_coaching_intensity`. Uses `chokidar` for file watching and `child_process` for git queries.

2. **Skills** (`plugin/skills/`) — 9 slash command skills and 4 teaching methodology skills with reference material.

3. **Hooks** (`plugin/hooks/hooks.json`) — PostToolUse advisory hooks that detect large files and error patterns, suggesting relevant coaching commands.

The CLI (`bin/cli.js`) handles install/uninstall of tutor instructions into CLAUDE.md or AGENTS.md.
