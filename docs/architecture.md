# Architecture

## Overview

```
┌─────────────────────────────────────────────────┐
│              Claude Code / Codex CLI             │
│                                                  │
│  ┌─────────────┐  ┌──────────┐  ┌────────────┐  │
│  │ plugin.json │  │  Skills  │  │ hooks.json │  │
│  │ (MCP start) │  │ (slash   │  │ (advisory  │  │
│  │             │  │  cmds)   │  │  hooks)    │  │
│  └──────┬──────┘  └──────────┘  └────────────┘  │
│         │ stdio                                  │
│  ┌──────▼──────────────────────┐                 │
│  │     MCP Server (Node.js)   │                 │
│  │  tutoring-mcp.js           │                 │
│  │  ┌─────────┐ ┌───────────┐ │                 │
│  │  │chokidar │ │ git CLI   │ │                 │
│  │  │(files)  │ │(commits)  │ │                 │
│  │  └─────────┘ └───────────┘ │                 │
│  └────────────────────────────┘                 │
└─────────────────────────────────────────────────┘
```

## Package Structure

```
agent-tutor/
├── bin/cli.js                    # CLI installer (install/uninstall/plugin-dir)
├── package.json                  # npm package manifest
├── CLAUDE.md                     # Tutor instructions (injected into projects)
├── AGENTS.md                     # Same instructions for Codex CLI
├── plugin/
│   ├── .claude-plugin/
│   │   └── plugin.json           # Plugin manifest (MCP server, hooks)
│   ├── hooks/
│   │   └── hooks.json            # PostToolUse hook definitions
│   ├── servers/
│   │   └── tutoring-mcp.js       # MCP server (5 tools, file watcher)
│   └── skills/
│       ├── atu-check/SKILL.md    # 9 slash command skills
│       ├── atu-debug/SKILL.md
│       ├── atu-decompose/SKILL.md
│       ├── atu-explain/SKILL.md
│       ├── atu-hint/SKILL.md
│       ├── atu-plan/SKILL.md
│       ├── atu-review/SKILL.md
│       ├── atu-save/SKILL.md
│       ├── atu-workflow/SKILL.md
│       ├── atu-code-review-learning/    # 4 teaching methodology skills
│       ├── atu-dev-workflow/
│       ├── atu-guided-debugging/
│       └── atu-problem-decomposition/
└── scripts/
    ├── error-pattern-detect.js   # Hook: detect error patterns in Bash output
    └── large-file-detect.js      # Hook: detect large files after Write/Edit
```

## Components

### MCP Server (`plugin/servers/tutoring-mcp.js`)

Node.js MCP server using `@modelcontextprotocol/sdk` over stdio transport. Provides 5 tools:

| Tool | Input | Description |
|------|-------|-------------|
| `get_student_context` | none | Summary of recent file changes, git status, and commits |
| `get_recent_file_changes` | none | File change events with diffs (up to 30 recent) |
| `get_git_activity` | none | Recent commits and working tree status |
| `get_coaching_config` | none | Current intensity and student level |
| `set_coaching_intensity` | `intensity` (enum) | Set to proactive, on-demand, or silent |

**File watcher:** Uses `chokidar` to watch source files (`*.{js,ts,py,go,rs,...}`), ignoring `node_modules`, `.git`, etc. Events are stored in a ring buffer (max 100 entries) with diffs captured via `git diff`.

**Git queries:** Uses `child_process.execSync` to run `git log`, `git status`, and `git diff` with a 5-second timeout.

**Config:** Read/written from `.agent-tutor/config.json` in the working directory.

### CLI Installer (`bin/cli.js`)

Commander-based CLI with three commands:

- **`install`** — Reads `CLAUDE.md` from the package, injects it into the target file wrapped in `<!-- BEGIN AGENT-TUTOR -->` / `<!-- END AGENT-TUTOR -->` sentinels. Idempotent (replaces existing section). Supports `--scope local|global` and `--agent claude|codex`.
- **`uninstall`** — Removes the sentinel-wrapped section from the target file.
- **`plugin-dir`** — Prints the absolute path to the `plugin/` directory (for `--plugin-dir` flag).

### Plugin Manifest (`plugin/.claude-plugin/plugin.json`)

Declares the MCP server and hooks for Claude Code's plugin system:

- **mcpServers.agent-tutor** — Starts `tutoring-mcp.js` via `node`, using `${CLAUDE_PLUGIN_ROOT}` for path resolution.
- **hooks** — Points to `hooks.json` for PostToolUse advisory hooks.

### Skills (`plugin/skills/`)

Two categories:

1. **Command skills** (9) — Thin dispatchers that call MCP tools and provide coaching templates. Each is a `SKILL.md` file under `plugin/skills/atu-<name>/`.

2. **Teaching methodology skills** (4) — Detailed pedagogical methodologies with `references/` subdirectories:
   - `atu-guided-debugging` — 4-phase debugging methodology
   - `atu-problem-decomposition` — Problem breakdown techniques
   - `atu-code-review-learning` — Graduated self-review checklists
   - `atu-dev-workflow` — Development workflow habit coaching

### Hook Scripts (`scripts/`)

PostToolUse advisory hooks that inject `additionalContext`:

- **`large-file-detect.js`** — After Write/Edit, checks if the file exceeds 200 lines. Suggests `/atu:decompose`.
- **`error-pattern-detect.js`** — After Bash, checks for error patterns (panic, FAIL, traceback, etc.). Suggests `/atu:debug` or `/atu:explain`.

### Instruction Files (`CLAUDE.md`, `AGENTS.md`)

The tutor instruction block injected into projects. Contains:
- Commands table and teaching skills mapping
- Coaching behavior rules (proactive/on-demand/silent)
- Pedagogical principles (ask-before-tell, one-point-per-interaction, praise-first)
- Topic tracking lifecycle and state file format
- Learning plan awareness and integration
- Hook awareness (how to handle `additionalContext`)
- Lesson auto-save instructions and template

## Data Flow

```
Student activity
    │
    ├─ file save ──────► chokidar watcher ──► ring buffer (100 events)
    └─ git commit ─────► git CLI queries   ──► on-demand via tool calls
                                                    │
                                                    ▼
                                            MCP tool responses
                                            (markdown summaries)
                                                    │
                                                    ▼
                                            Agent coaching response
```

## Key Design Decisions

1. **Pure npm package** — No Go binary, no tmux, no TUI. The coding agent IS the interface. Agent Tutor is a plugin that enhances the agent's behavior.

2. **Plugin system integration** — Uses Claude Code's `plugin.json` to auto-start the MCP server and register hooks. No manual MCP configuration needed.

3. **Chokidar file watcher** — Replaces fsnotify. Cross-platform, handles debouncing natively via `awaitWriteFinish`.

4. **Ring buffer (in-memory)** — File events stored in a simple array with shift-on-overflow. No persistent store needed since the agent queries recent context only.

5. **Git via child_process** — Simple `execSync` calls with timeouts. No git library dependency needed for the read-only queries used here.

6. **Instruction-driven features** — Topic tracking, learning plans, and lesson auto-save are all driven by instructions in CLAUDE.md rather than MCP tools. The agent's file writing capability handles all I/O.

7. **Sentinel-based injection** — `<!-- BEGIN/END AGENT-TUTOR -->` markers enable idempotent install/uninstall of the instruction block.
