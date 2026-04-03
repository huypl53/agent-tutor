# Agent Tutor — Development

This is a Claude Code plugin that tutors students. You are a **contributor**, not the tutor.

## Build & Test

```bash
npm test                    # run all tests (node:test)
npm install                 # install dependencies
```

## Project Structure

- `plugin/` — the plugin itself (MCP server, skills, hooks)
- `plugin/servers/tutoring-mcp.js` — MCP server (21 tools)
- `plugin/servers/state-manager.js` — learning state persistence (v2 schema with project support)
- `plugin/servers/project-scanner.js` — project type detection & manifest parsing
- `plugin/data/project-types.csv` — 14 project type definitions
- `plugin/templates/tutor-instructions.md` — tutor persona injected into student projects
- `bin/cli.js` — CLI for install/uninstall
- `test/` — tests (node:test)
- `docs/architecture.md` — architecture docs

## Key Conventions

- MCP tool handlers are thin shells over `StateManager` methods
- Atomic writes via write-to-temp + rename
- Topic status follows a state machine (see `VALID_TRANSITIONS` in state-manager.js)
- Skills at `plugin/skills/atu-*/SKILL.md` need YAML frontmatter

## Tutor Instructions

The tutor persona lives in `plugin/templates/tutor-instructions.md`. It gets injected into student projects via `agent-tutor install`. Do not put tutor instructions in this file.
