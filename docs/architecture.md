# Architecture

## Overview

```
┌───────────────────────────────────────────────────────┐
│                      tmux session                     │
│  ┌───────────────────────┐  ┌──────────────────────┐  │
│  │   User Terminal       │  │   Coding Agent       │  │
│  │   (pane 0)            │  │   (pane 1)           │  │
│  │                       │  │                      │  │
│  │   Student works       │  │   claude             │  │
│  │   here                │  │   --mcp-config       │  │
│  │                       │  │   'agent-tutor mcp'  │  │
│  └───────────────────────┘  └──────────┬───────────┘  │
│                                        │              │
└────────────────────────────────────────┼──────────────┘
                                         │ stdio
                                ┌────────▼────────┐
                                │   MCP Server    │
                                │   (agent-tutor  │
                                │    mcp)         │
                                └────────┬────────┘
                                         │
                      ┌──────────────────┼──────────────────┐
                      │                  │                  │
             ┌────────▼───────┐ ┌────────▼───────┐ ┌────────▼───────┐
             │  FileWatcher   │ │  TermWatcher   │ │  GitWatcher    │
             │  (fsnotify)    │ │  (poll pane)   │ │  (poll git)    │
             └────────┬───────┘ └────────┬───────┘ └────────┬───────┘
                      │                  │                  │
                      └──────────────────┼──────────────────┘
                                         │
                                ┌────────▼────────┐
                                │  Context Store  │
                                │  (ring buffers) │
                                └────────┬────────┘
                                         │
                                ┌────────▼────────┐
                                │ Trigger Engine  │
                                └─────────────────┘
```

## Components

### CLI (`cmd/agent-tutor`, `internal/cli`)

Cobra-based CLI with six commands:

- **start** -- Creates tmux session, splits panes, auto-installs plugin if missing, launches agent with MCP server and `--plugin-dir`, then `syscall.Exec`s into `tmux attach-session`.
- **stop** -- Kills the tmux session.
- **status** -- Reports whether a session is running.
- **install-plugin** -- Extracts embedded plugin files and appends tutor instructions to CLAUDE.md. Supports `--scope local|global`.
- **uninstall-plugin** -- Removes plugin files and tutor instructions from CLAUDE.md. Supports `--scope local|global`.
- **mcp** (hidden) -- Spawned by the agent process. Creates the store, starts watchers, initializes trigger engine, and runs the MCP server on stdio. Handles SIGINT/SIGTERM.

### Plugin (`internal/plugin`)

Embeds Claude Code plugin files via `//go:embed all:embed` (the `all:` prefix is needed to include the `.claude-plugin` hidden directory). The `Install()` function extracts plugin files and appends a tutor instruction section to `.claude/CLAUDE.md` with `<!-- BEGIN AGENT-TUTOR -->` / `<!-- END AGENT-TUTOR -->` sentinel comments for clean uninstall.

Two scopes:
- **local**: Plugin in `.agent-tutor/plugin/`, instructions in `.claude/CLAUDE.md`
- **global**: Skills in `~/.claude/skills/atu-*/`, instructions in `~/.claude/CLAUDE.md`

The `start` command auto-installs locally if the plugin is not present, and passes `--plugin-dir` to the claude command.

Note: Embedded command files use dashes (`atu-check.md`) because Go's embed package forbids colons in filenames. A `restoreColons()` helper maps them back to colons (`atu:check.md`) during local extraction for Claude Code command registration.

### Config (`internal/config`)

TOML-based configuration loaded from `.agent-tutor/config.toml` in the project directory. Creates a default config on first run. Sections: `[tutor]`, `[agent]`, `[watchers]`, `[tmux]`.

### Context Store (`internal/store`)

In-memory store using generic ring buffers (`ringBuffer[T]`) for three event types:

| Event type | Capacity | Fields |
|------------|----------|--------|
| FileEvent | 100 | Path, Change, Diff, Timestamp |
| TerminalEvent | 50 | Content, HasError, Timestamp |
| GitEvent | 30 | Type, Summary, Timestamp |

Thread-safe via `sync.RWMutex`. The `Summary(since)` method produces a markdown-formatted summary of recent events, truncated to 8 KB.

### Tmux Manager (`internal/tmux`)

Wraps tmux CLI commands via `os/exec`. Methods: `CreateSession`, `SplitPane`, `SendKeys`, `CapturePane`, `KillSession`, `HasSession`. Command construction is separated from execution for testability.

### Watchers (`internal/watcher`)

All watchers implement the `Watcher` interface (`Start(ctx)`, `Stop()`).

- **FileWatcher** -- Uses `fsnotify` for recursive directory watching. Debounces rapid saves (300ms per file). Captures git diffs for modified files. Configurable file patterns and ignore lists.
- **TerminalWatcher** -- Polls `tmux capture-pane` at a configurable interval. Diffs against previous capture to detect new output. Detects error patterns (panic, FAIL, traceback, etc.).
- **GitWatcher** -- Polls `git rev-parse HEAD` and `git status --porcelain`. Detects new commits and working tree status changes.

### MCP Server (`internal/mcp`)

Implements MCP over stdio using the official Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`). Registers six tools and injects a tutor system prompt via server instructions.

### Trigger Engine (`internal/trigger`)

Rule-based event trigger with threshold and cooldown. When an event fires enough times within the cooldown window, it calls a callback (used for proactive coaching nudges). Thread-safe.

## Data flow

```
Student activity
    │
    ├─ file save ──────► FileWatcher ──► store.AddFileEvent()
    ├─ terminal output ► TermWatcher ──► store.AddTerminalEvent()
    └─ git commit ─────► GitWatcher  ──► store.AddGitEvent()
                                              │
                                              ▼
                                      Context Store (ring buffers)
                                              │
                         ┌────────────────────┼────────────────────┐
                         ▼                    ▼                    ▼
                  Trigger Engine        MCP tool calls       Summary()
                  (threshold/cooldown)  (agent queries)      (markdown)
                         │
                         ▼
                  tutor_nudge → agent coaches proactively
```

## MCP tools reference

| Tool | Input | Description |
|------|-------|-------------|
| `get_student_context` | none | 5-minute activity summary (markdown) |
| `get_recent_file_changes` | none | Recent file events with diffs |
| `get_terminal_activity` | none | Recent terminal output snapshots |
| `get_git_activity` | none | Recent commits and status changes |
| `get_coaching_config` | none | Current intensity and student level |
| `set_coaching_intensity` | `intensity` (string) | Set to proactive, on-demand, or silent |

## Key design decisions

1. **tmux-based layout** -- Uses tmux rather than a custom terminal multiplexer. This avoids reinventing terminal handling and lets the student use their normal shell. The `start` command `syscall.Exec`s into tmux so the user's process is fully replaced.

2. **MCP over stdio** -- The MCP server runs as a subprocess of the coding agent (via `--mcp-config`), communicating over stdin/stdout. This is the standard MCP transport and requires no network ports.

3. **Ring buffer store** -- Fixed-capacity ring buffers prevent unbounded memory growth. Capacities (100/50/30) are tuned so the agent sees enough recent context without being overwhelmed.

4. **Watcher separation** -- File, terminal, and git watchers are independent. Each has its own polling strategy: fsnotify events for files (instant), polling for terminal (2s default) and git (5s default).

5. **Trigger engine with cooldown** -- Prevents the agent from spamming coaching nudges. Rules have both a threshold (N events before firing) and a cooldown (minimum time between fires).

6. **Hidden `mcp` subcommand** -- The `agent-tutor mcp` command is hidden from help output since it is only meant to be invoked by the agent process, not by the user directly.

7. **Config auto-creation** -- If no `.agent-tutor/config.toml` exists, a default is written on first `start`. This gives users a file to edit without requiring a separate `init` command.

8. **Isolated tmux socket** -- Uses `tmux -L agent-tutor` to run in a separate tmux server. This prevents interference with the user's existing tmux sessions and enables parallel E2E testing. The socket name is configurable via `[tmux] socket` in config or `--socket` CLI flag.

## internal/tmux

The `tmux` package (`internal/tmux/tmux.go`) provides a `Manager` struct that wraps tmux CLI commands via `os/exec`.

### Design

- **Socket isolation**: The `Socket` field, when set, prepends `-L <socket>` to every tmux command via the `tmuxCmd()` helper. This runs commands against a dedicated tmux server instance.
- **Command builders** (unexported): `createSessionCmd`, `splitPaneCmd`, `capturePaneCmd`, `sendKeysCmd`, `killSessionCmd`, `hasSessionCmd` -- these construct `*exec.Cmd` values via `tmuxCmd()` without running them, making them easy to test without a real tmux server.
- **Public methods**: `CreateSession`, `SplitPane`, `SendKeys`, `CapturePane`, `KillSession`, `HasSession` -- these call the corresponding builder and execute the command.
- **Pane targeting**: Panes are addressed as `session:0.paneID` (window.pane format), not `session:paneID` (which would target a window).

### Testing approach

Tests verify command argument construction only (no tmux required). This is done by inspecting `cmd.Args` from the unexported builder methods.

## internal/store

The `store` package (`internal/store/store.go`) provides an in-memory context store using generic ring buffers for watcher events.

### Design

- **Generic ring buffer** (`ringBuffer[T]`): fixed-capacity circular buffer with `add()` and `snapshot()` methods. `snapshot()` returns items in insertion order (oldest first).
- **Three event types**: `FileEvent` (cap 100), `TerminalEvent` (cap 50), `GitEvent` (cap 30).
- **Thread safety**: all public methods use `sync.RWMutex` -- writers take exclusive lock, readers take shared lock.
- **`Summary(since time.Duration)`**: produces a markdown-formatted summary of recent events, optionally filtered by time window. Output is truncated to 8000 bytes.

### Testing approach

Tests verify ring buffer overflow behavior (150 inserts -> 100 returned), basic add/get round-trips for all event types, and non-empty Summary output.

## internal/watcher

### Watcher interface (`watcher.go`)

All watchers implement the `Watcher` interface:

```go
type Watcher interface {
    Start(ctx context.Context) error
    Stop() error
}
```

### TerminalWatcher (`terminal.go`)

Polls tmux `capture-pane` at a configurable interval, diffs against the previous capture, and stores new output as `TerminalEvent` in the store. Accepts an optional `socket` parameter to target an isolated tmux server via `-L`.

**Diff logic** (`diff(old, new string) string`):
- If content is identical, returns empty string.
- If new has more lines than old, returns only the appended lines.
- If screen was cleared (fewer or equal lines but different content), returns all new content.

**Error detection** (`hasError(content string) bool`): checks content against compiled regex patterns (case-insensitive):
- `^error[:\s]`, `^fatal[:\s]`, `^panic[:\s]` (line-start anchored)
- `FAIL[:\s]`, `traceback`, `exception[:\s]` (anywhere in content)

### FileWatcher (`file.go`)

Watches a project directory recursively for file changes using `github.com/fsnotify/fsnotify`, debounces rapid saves (300ms per file), and stores `FileEvent` records. Configurable file patterns and ignore lists.

### GitWatcher (`git.go`)

Polls `git status` and `git log` at a configurable interval to detect commits (HEAD change) and working tree status changes.

## internal/mcp

The `mcp` package implements an MCP server over stdio using the official Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`).

### SDK insights

- Server creation: `mcp.NewServer(&mcp.Implementation{Name, Version}, &mcp.ServerOptions{Instructions})`.
- Tool registration uses the generic `mcp.AddTool(server, &mcp.Tool{...}, handlerFunc)`.
- The `jsonschema` struct tag must be a plain description string, not `key=value` pairs. Using `WORD=` format causes a panic in the SDK's `ForType` parser.
- Stdio transport: `server.Run(ctx, &mcp.StdioTransport{})`.

## internal/cli

### Commands

- **`start [project-dir]`** (`start.go`): Loads config, auto-installs plugin if missing, creates tmux session, splits panes, sends agent command with `--mcp-config` and `--plugin-dir`, then `syscall.Exec`s into `tmux attach-session`.
- **`stop`** (`stop.go`): Kills the tmux session.
- **`status`** (`status.go`): Reports whether a session is running.
- **`install-plugin`** (`install_plugin.go`): Extracts embedded plugin files and appends tutor instructions to CLAUDE.md. `--scope local|global`.
- **`uninstall-plugin`** (`uninstall_plugin.go`): Removes plugin files and tutor instructions. `--scope local|global`.
- **`mcp`** (`mcp.go`): Hidden command. Creates store, starts watchers, initializes trigger engine, runs MCP server on stdio.

## internal/trigger

Rule-based event trigger. Each rule has an event name, threshold count, and cooldown duration. The `Fire(event)` method increments the counter and calls the callback when threshold is reached and cooldown has elapsed. Thread-safe via mutex.

## E2E Integration Tests (`internal/integration`)

End-to-end tests that run in an isolated tmux server (`-L agent-tutor-test`). Build-tagged with `//go:build integration` so they don't run during `go test ./...`.

### Tests

- **TestE2ESessionLifecycle** -- Verifies session creation and 2-pane layout.
- **TestE2EGoLearnerActivity** -- Simulates a Go learner writing buggy code, fixing it, building, running, and committing. Verifies the trigger engine detects the commit.

### Running

```bash
go test -tags integration ./internal/integration/ -v -timeout 60s
```
