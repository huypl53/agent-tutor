# Distribution Design

**Goal:** Make agent-tutor installable by other developers via npm and Claude Code marketplace.

## Distribution Channels

### 1. npm Registry (`@huypl53/agent-tutor`)

Users install globally and use via CLI:

```bash
npm install -g @huypl53/agent-tutor
agent-tutor install                                    # inject into CLAUDE.md
claude --plugin-dir $(agent-tutor plugin-dir)          # or use as plugin directly
```

For Codex:
```bash
npx @huypl53/agent-tutor install --agent codex
codex mcp add agent-tutor -- node $(npx @huypl53/agent-tutor plugin-dir)/servers/tutoring-mcp.js
```

### 2. Claude Code Marketplace (same repo)

A `.claude-plugin/marketplace.json` at repo root lets users add this repo as a marketplace source:

```bash
claude plugin marketplace add github:huypl53/agent-tutor
claude plugin install agent-tutor
```

The plugin auto-starts the MCP server, loads skills, and registers hooks — no CLI needed.

## Changes Required

### package.json

- Rename `name` to `@huypl53/agent-tutor`
- Add `publishConfig.access: "public"` (required for scoped packages)
- Add `repository`, `author`, `homepage`, `keywords`

### plugin/.claude-plugin/plugin.json

- Add `author` object with name/email
- Add `homepage` URL
- Add `skills: "./skills/"` pointer for marketplace discovery

### .claude-plugin/marketplace.json (new)

Marketplace manifest at repo root:

```json
{
  "$schema": "https://anthropic.com/claude-code/marketplace.schema.json",
  "name": "agent-tutor-marketplace",
  "description": "Programming tutor plugin for coding agents",
  "owner": {
    "name": "huypl53",
    "email": ""
  },
  "plugins": [
    {
      "name": "agent-tutor",
      "description": "Programming tutor — coaching, debugging, code review via MCP tools and skills",
      "category": "development",
      "source": {
        "source": "git-subdir",
        "url": "https://github.com/huypl53/agent-tutor",
        "ref": "master"
      },
      "homepage": "https://github.com/huypl53/agent-tutor"
    }
  ]
}
```

### bin/cli.js

- Update `program.version()` to read from package.json dynamically instead of hardcoded

### README.md

- Update install instructions for `@huypl53/agent-tutor`
- Add marketplace install section
- Add Codex setup with scoped package name

### .gitignore

- Add `.npmrc` to prevent publishing credentials
