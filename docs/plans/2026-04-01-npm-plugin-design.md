# Design: Pure npm Plugin with Node.js MCP Server

**Date:** 2026-04-01
**Goal:** Replace Go binary with a pure npm package. Agent-tutor becomes a Claude Code plugin with a Node.js MCP server, slash commands (skills), hooks, and CLAUDE.md/AGENTS.md instructions.

## Package Structure

```
agent-tutor/
├── package.json                    # npm package, bin: "agent-tutor"
├── bin/
│   └── cli.js                      # install/uninstall CLI (Node.js)
├── plugin/
│   ├── .claude-plugin/
│   │   └── plugin.json             # declares mcpServers, skills, hooks
│   ├── servers/
│   │   └── tutoring-mcp.js         # Node.js MCP server
│   ├── skills/
│   │   ├── atu-check/SKILL.md
│   │   ├── atu-hint/SKILL.md
│   │   ├── atu-explain/SKILL.md
│   │   ├── atu-save/SKILL.md
│   │   ├── atu-debug/SKILL.md
│   │   ├── atu-review/SKILL.md
│   │   ├── atu-decompose/SKILL.md
│   │   ├── atu-workflow/SKILL.md
│   │   ├── atu-plan/SKILL.md
│   │   ├── atu-guided-debugging/
│   │   │   ├── SKILL.md
│   │   │   └── references/
│   │   ├── atu-problem-decomposition/
│   │   │   ├── SKILL.md
│   │   │   └── references/
│   │   ├── atu-code-review-learning/
│   │   │   ├── SKILL.md
│   │   │   └── references/
│   │   └── atu-dev-workflow/
│   │       ├── SKILL.md
│   │       └── references/
│   └── hooks/
│       └── hooks.json              # PostToolUse hooks config
├── scripts/
│   ├── large-file-detect.js
│   └── error-pattern-detect.js
├── CLAUDE.md                       # tutor instructions (injected into project)
├── AGENTS.md                       # same instructions for Codex CLI
└── README.md
```

## plugin.json

```json
{
  "name": "agent-tutor",
  "version": "0.2.0",
  "description": "Programming tutor plugin for coding agents",
  "mcpServers": {
    "agent-tutor": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/servers/tutoring-mcp.js"]
    }
  },
  "hooks": "./hooks/hooks.json"
}
```

## Install Mechanism

### Claude Code

Two modes:
1. **`--plugin-dir`** (dev/local): `claude --plugin-dir ./node_modules/agent-tutor/plugin`
2. **Plugin marketplace** (future): users install via plugin manager

CLI commands:
- `npx agent-tutor install` — injects CLAUDE.md section into `.claude/CLAUDE.md`
- `npx agent-tutor install --scope global` — injects into `~/.claude/CLAUDE.md`
- `npx agent-tutor install --agent codex` — injects into AGENTS.md + registers MCP
- `npx agent-tutor uninstall` — removes injected section

### Codex CLI

```bash
codex mcp add agent-tutor -- node ./node_modules/agent-tutor/plugin/servers/tutoring-mcp.js
```

Installer handles this with `--agent codex`.

## MCP Server (tutoring-mcp.js)

Node.js MCP server using `@modelcontextprotocol/sdk`.

### Tools

| Tool | Implementation |
|------|---------------|
| `get_student_context` | `git diff --stat` + `git log --oneline -5` + recent file events |
| `get_recent_file_changes` | `chokidar` file watcher, in-memory ring buffer |
| `get_git_activity` | `git log` + `git status --porcelain` (fresh each call) |
| `get_coaching_config` | Reads `.agent-tutor/config.json` |
| `set_coaching_intensity` | Writes `.agent-tutor/config.json` |

### Dependencies

- `@modelcontextprotocol/sdk` — MCP protocol
- `chokidar` — file watching

### Key Design Choices

- Git info gathered fresh on each tool call (no polling loop needed)
- Ring buffer only for file events (track rapid changes between calls)
- Config stored as `.agent-tutor/config.json` (no TOML dependency)
- Server started automatically by Claude Code plugin system

## hooks.json

```json
{
  "PostToolUse": [
    {
      "matcher": "Write|Edit",
      "hooks": [{
        "type": "command",
        "command": "node ${CLAUDE_PLUGIN_ROOT}/../scripts/large-file-detect.js"
      }]
    },
    {
      "matcher": "Bash",
      "hooks": [{
        "type": "command",
        "command": "node ${CLAUDE_PLUGIN_ROOT}/../scripts/error-pattern-detect.js"
      }]
    }
  ]
}
```

## What Gets Deleted

- `cmd/agent-tutor/` — Go main
- `internal/` — all Go packages
- `go.mod`, `go.sum`
- `.worktrees/` — current Go worktree

## What Gets Adapted

- Slash command markdown → `plugin/skills/` as SKILL.md files
- Teaching skills → `plugin/skills/` (same structure)
- Hook scripts → `scripts/`
- CLAUDE.md instructions → standalone file, MCP tool table removed (tools auto-discovered)

## Migration Summary

| Before | After |
|--------|-------|
| Go binary (`agent-tutor`) | npm package (`npx agent-tutor`) |
| Go MCP server via `--mcp-config` | Node.js MCP server via `plugin.json` |
| `go:embed` for plugin files | Files shipped directly in npm package |
| `go install` | `npm install -g agent-tutor` |
| TOML config | JSON config |
| 19 Go files, ~2363 lines | ~5 JS files, ~500 lines |

## Agents Supported

| Feature | Claude Code | Codex CLI |
|---------|------------|-----------|
| MCP server | Auto-start via plugin.json | `codex mcp add` |
| Slash commands | `plugin/skills/` | `~/.codex/prompts/` (copied by installer) |
| Instructions | CLAUDE.md | AGENTS.md |
| Hooks | hooks.json (17+ events) | Limited (3 events) |
