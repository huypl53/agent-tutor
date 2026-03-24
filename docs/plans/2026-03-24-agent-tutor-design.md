# Agent Tutor — Design Document

## Problem

Learning a programming language is slow when done in isolation. Coding agents (Claude Code, Codex) are powerful assistants, but they solve problems *for* you rather than teaching you. There's no tool that turns a coding agent into a programming tutor that watches you work and coaches you in context.

## Solution

**agent-tutor** is a Go CLI that creates a side-by-side tmux environment where a coding agent acts as a programming tutor. The tutor observes the user's real-time coding activity (file changes, terminal commands, git workflow) and weaves coaching into the agent's normal responses.

## Core Concepts

- **Shadowing/coaching model** — The user works on their own projects; the agent watches and teaches in context
- **The agent IS the tutor** — No separate tutor UI. Coaching is woven into the coding agent's natural responses
- **MCP-first architecture** — The tutor is an MCP server that gives the agent observation tools and a tutor system prompt
- **All-in-one CLI** — `agent-tutor start` handles everything: tmux layout, watchers, MCP registration, agent launch
- **Adapts to user level** — Works for beginners learning their first language and experienced devs picking up a new one
- **Language-agnostic** — Relies on the LLM's knowledge, not language-specific tooling

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    tmux window                          │
│  ┌──────────────────────┐  ┌──────────────────────────┐ │
│  │    User Terminal      │  │   Coding Agent TUI       │ │
│  │                       │  │   (Claude Code/Codex)    │ │
│  │  - edits code         │  │                          │ │
│  │  - runs commands      │  │   Agent has tutor MCP    │ │
│  │  - uses git           │  │   tools available.       │ │
│  │                       │  │   System prompt shapes   │ │
│  │                       │  │   it as a tutor.         │ │
│  └──────────────────────┘  └──────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
        │                              ▲
        │ observes                     │ provides context via MCP tools
        ▼                              │
┌─────────────────────────────────────────────────────────┐
│              agent-tutor (Go binary)                    │
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ File Watcher │  │ Tmux Watcher │  │  Git Watcher  │  │
│  │ (fsnotify)   │  │ (capture-    │  │  (poll git    │  │
│  │              │  │  pane)       │  │   status/log) │  │
│  └──────┬──────┘  └──────┬───────┘  └──────┬────────┘  │
│         └────────────┬───┘───────────┘                  │
│                      ▼                                  │
│            ┌──────────────────┐                          │
│            │  Context Store   │  in-memory ring buffer   │
│            │  (observations)  │  of recent activity      │
│            └──────────────────┘                          │
│                      ▲                                  │
│                      │ read by                          │
│            ┌──────────────────┐                          │
│            │   MCP Server     │  stdio transport         │
│            │                  │  tools + system prompt    │
│            └──────────────────┘                          │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Config: coaching intensity, language prefs, etc │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### Key components

1. **CLI orchestrator** — `agent-tutor start [project-dir]` sets up tmux, starts watchers, launches agent with MCP config
2. **Three watchers** — file (fsnotify), terminal (tmux capture-pane polling), git (git status/log polling) — all feed into a shared in-memory context store
3. **MCP server** — runs over stdio, exposes tools the agent calls to get student context, injects a tutor system prompt
4. **Config** — TOML file for coaching intensity, user skill level, agent command

## MCP Tools

| Tool | Description | When called |
|------|-------------|-------------|
| `get_student_context` | Returns recent file changes, terminal activity, and git activity as a combined summary | On `/check` or proactive trigger |
| `get_recent_file_changes` | Returns diffs of recently modified files with timestamps | When agent needs to see what the user wrote |
| `get_terminal_activity` | Returns recent terminal commands and their output from the user's pane | When agent wants to see what the user ran |
| `get_git_activity` | Returns recent commits, diffs, branch info | When reviewing user's git workflow |
| `get_coaching_config` | Returns current coaching intensity, user skill level, language preferences | Agent checks this to calibrate responses |
| `set_coaching_intensity` | Allows user to change coaching level via natural language in chat | User requests change via chat |

## System Prompt Injection

Injected via MCP `instructions` field:

```
You are also a programming tutor. A student is working in a terminal pane
next to you. You have tools to observe their work.

Coaching intensity: {intensity}
Student level: {level}

When intensity is "proactive":
- After the student messages you, also check get_student_context for
  teachable moments
- When you receive a tutor_nudge notification, call get_student_context
  and offer relevant coaching
- Weave teaching naturally into your responses — don't lecture

When intensity is "on-demand" or "silent":
- Only use tutor tools when the student explicitly asks for feedback
  or uses /check

Teaching style:
- Explain the "why" not just the "what"
- For beginners: explain concepts, suggest resources
- For experienced devs: focus on idioms, best practices, ecosystem conventions
- Be concise. One teaching point per interaction, not five.
- If the student is doing well, say nothing. Don't coach for the sake of coaching.
```

## MCP Notifications (proactive triggers)

| Event | Trigger condition |
|-------|-------------------|
| `tutor_nudge/commit` | User made a git commit |
| `tutor_nudge/error_loop` | User has hit the same error 3+ times in terminal |
| `tutor_nudge/idle_after_error` | User stopped typing for 2+ min after an error |
| `tutor_nudge/session_summary` | User has been working 30+ min, periodic checkpoint |

## Coaching Intensity Levels

- **Silent** — Agent never coaches on its own. User must explicitly ask.
- **On-demand** — Agent waits for `/check` or direct questions about their code.
- **Proactive** — Agent receives MCP notifications on key events and offers coaching. Cooldowns prevent spam (e.g., no more than one nudge per 5 minutes).

## CLI Commands

```
agent-tutor start [project-dir]    # Set up tmux, start watchers, launch agent
agent-tutor stop                   # Tear down tmux session, stop watchers
agent-tutor config                 # Interactive config editor
agent-tutor status                 # Show current session info
```

### `agent-tutor start` flow

1. Read/create config at `[project-dir]/.agent-tutor/config.toml`
2. Create tmux session "agent-tutor"
3. Split into two panes:
   - Left: user shell, cd to project directory
   - Right: coding agent with MCP server registered
4. Start watchers (file, terminal, git)
5. MCP server begins listening on stdio (spawned by the agent)

### First-run experience

If no config exists, asks three questions:
1. Which coding agent? (claude / codex / custom)
2. Your experience level? (beginner / experienced / auto-detect)
3. Coaching intensity? (silent / on-demand / proactive)

## Config File

Location: `[project-dir]/.agent-tutor/config.toml`

```toml
[tutor]
intensity = "on-demand"     # silent | on-demand | proactive
level = "auto"              # auto | beginner | intermediate | advanced

[agent]
command = "claude"           # claude | codex | custom command
args = []

[watchers]
file_patterns = ["**/*.go", "**/*.py", "**/*.js"]
ignore_patterns = ["node_modules", ".git", "vendor"]
terminal_poll_interval = "2s"
git_poll_interval = "5s"

[tmux]
layout = "horizontal"       # horizontal | vertical
user_pane_size = 50          # percentage
```

## Context Store

In-memory ring buffer per watcher category:

- `FileEvents`: last 100 file change events (path, diff snippet, timestamp)
- `TerminalEvents`: last 50 captured terminal snapshots (commands + output)
- `GitEvents`: last 30 git events (commits, branch switches, diffs)

When `get_student_context` is called, the store assembles a summary truncated to ~2000 tokens max to avoid flooding the agent's context window.

### Watcher details

**File Watcher (fsnotify):** Watches project directory recursively, filtered by config patterns. Debounces rapid saves (300ms). Stores file path, change type, short diff (first 50 lines).

**Terminal Watcher (tmux capture-pane):** Polls user's tmux pane at configured interval. Diffs against previous capture. Parses commands, output, and error patterns. Detects error loops (same error 3+ times).

**Git Watcher (git poll):** Polls `git status` and `git log` at configured interval. Detects commits, staging, branch switches, merge conflicts.

### Proactive trigger engine

```go
type Rule struct {
    Event     string        // e.g. "git.commit", "terminal.error_repeat"
    Threshold int           // e.g. 3 for error_repeat
    Cooldown  time.Duration // don't fire again within this window
}
```

When a rule fires and intensity is "proactive," sends an MCP notification. Cooldown prevents spam.

## Project Structure

```
agent-tutor/
├── cmd/
│   └── agent-tutor/
│       └── main.go              # CLI entrypoint (cobra)
├── internal/
│   ├── cli/
│   │   ├── start.go             # start command — tmux setup, launch agent
│   │   ├── stop.go              # teardown
│   │   ├── config.go            # interactive config editor
│   │   └── status.go            # show session info
│   ├── config/
│   │   └── config.go            # TOML config read/write/defaults
│   ├── tmux/
│   │   └── tmux.go              # tmux session/pane management
│   ├── watcher/
│   │   ├── file.go              # fsnotify file watcher
│   │   ├── terminal.go          # tmux capture-pane poller
│   │   ├── git.go               # git status/log poller
│   │   └── watcher.go           # common interface
│   ├── store/
│   │   └── store.go             # in-memory ring buffer context store
│   ├── trigger/
│   │   └── trigger.go           # proactive rule engine
│   └── mcp/
│       ├── server.go            # MCP server (stdio transport)
│       ├── tools.go             # tool definitions & handlers
│       └── prompt.go            # system prompt builder
├── go.mod
├── go.sum
├── README.md
└── docs/
    └── architecture.md
```

## Dependencies

- `cobra` — CLI framework
- `fsnotify` — file system watching
- `pelletier/go-toml` — config parsing
- MCP Go SDK (e.g., `mark3labs/mcp-go`)

## Future considerations (not in v1)

- **Knowledge map** — Track concepts the user has encountered and their proficiency over time
- **Session history** — Persist coaching interactions for review
- **Hybrid approach (Approach 3)** — Add periodic background analysis if pure MCP proves insufficient for proactive coaching
- **Per-category coaching toggles** — Fine-grained control (e.g., enable idiom tips, disable git tips)
