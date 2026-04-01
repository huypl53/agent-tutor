# Agent Tutor

A programming tutor plugin for coding agents (Claude Code, Codex CLI). Observes your work via file changes and git activity, then coaches you through MCP tools and slash commands.

## Installation

### Via Claude Code Marketplace (recommended)

```bash
claude plugin marketplace add github:huypl53/agent-tutor
claude plugin install agent-tutor
```

The plugin auto-starts the MCP server, loads skills, and registers hooks.

### Via npm

```bash
npm install -g @huypl53/agent-tutor
```

Then use as a Claude Code plugin:

```bash
claude --plugin-dir $(agent-tutor plugin-dir)
```

Or inject tutor instructions into your project:

```bash
agent-tutor install
```

### Codex CLI

```bash
npx @huypl53/agent-tutor install --agent codex
codex mcp add agent-tutor -- node $(npx @huypl53/agent-tutor plugin-dir)/servers/tutoring-mcp.js
```

Requires Node.js 18+ and git on your PATH.

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

## Learning State Management

All learning state is stored in `.agent-tutor/state.json` and managed via MCP tools:

- **Topic tracking** — Create and track learning topics with status progression (`introduced → practicing → struggling → breakthrough → mastered`). Each topic records moments (struggles, hints, breakthroughs) and links to saved lessons.
- **Topic dependency graph** — Topics can declare dependencies, forming a graph the tutor uses to suggest learning order and connect concepts.
- **Learning plans** — Structured multi-step plans with progress tracking. Steps reference topics and are marked as mastered/skipped as the student progresses.
- **Session recovery** — After `/clear` or `/compact`, the tutor calls `restore_session` to recover the active topic and context without asking the student to re-explain.
- **Auto-migration** — Existing `current-topic.md` and `learning-plan.md` files are automatically migrated to JSON on first load.

Create a structured learning path with `/atu:plan`:

```
/atu:plan Build a REST API           # creates a 4-8 step plan
/atu:plan                            # shows current progress
/atu:plan next                       # marks current step done, advances
```

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

1. **MCP Server** (`plugin/servers/tutoring-mcp.js`) — Node.js server providing 16 tools over stdio: 5 observation tools (file changes, git activity, coaching config) and 11 learning state tools (topic CRUD, plans, sessions, summaries). Uses `chokidar` for file watching and a `StateManager` layer for atomic JSON state operations.

2. **Skills** (`plugin/skills/`) — 9 slash command skills and 4 teaching methodology skills with reference material.

3. **Hooks** (`plugin/hooks/hooks.json`) — PostToolUse advisory hooks that detect large files and error patterns, suggesting relevant coaching commands.

The CLI (`bin/cli.js`) handles install/uninstall of tutor instructions into CLAUDE.md or AGENTS.md.
