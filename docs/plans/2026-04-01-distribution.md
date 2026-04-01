# Distribution Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make agent-tutor installable by other developers via npm (`@huypl53/agent-tutor`) and Claude Code marketplace (same repo).

**Architecture:** Update package.json for scoped npm publish, fix plugin validation errors (hooks.json format, missing frontmatter), add marketplace.json to repo root, update README with all install paths.

**Tech Stack:** npm, Claude Code plugin system, git

---

### Task 1: Fix hooks.json format and add teaching skill frontmatter

The `claude plugin validate` found errors we must fix before distribution.

**Files:**
- Modify: `plugin/hooks/hooks.json`
- Modify: `plugin/skills/atu-guided-debugging/SKILL.md`
- Modify: `plugin/skills/atu-code-review-learning/SKILL.md`
- Modify: `plugin/skills/atu-problem-decomposition/SKILL.md`
- Modify: `plugin/skills/atu-dev-workflow/SKILL.md`

**Step 1: Fix hooks.json**

The current format is wrong. Claude Code expects `{ "hooks": { "PostToolUse": [...] } }` with `matcher` as a sibling of `hooks` inside each entry. Replace `plugin/hooks/hooks.json` with:

```json
{
  "description": "Advisory hooks that suggest tutoring commands on teachable moments",
  "hooks": {
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
}
```

**Step 2: Add frontmatter to 4 teaching skills**

Each teaching skill SKILL.md needs YAML frontmatter. Prepend to each file:

`plugin/skills/atu-guided-debugging/SKILL.md`:
```yaml
---
name: atu-guided-debugging
description: 4-phase debugging methodology — teach students to debug systematically
---
```

`plugin/skills/atu-code-review-learning/SKILL.md`:
```yaml
---
name: atu-code-review-learning
description: Teach students to review their own code with graduated checklists
---
```

`plugin/skills/atu-problem-decomposition/SKILL.md`:
```yaml
---
name: atu-problem-decomposition
description: Problem decomposition techniques for breaking down complex tasks
---
```

`plugin/skills/atu-dev-workflow/SKILL.md`:
```yaml
---
name: atu-dev-workflow
description: Development workflow habit coaching — commit often, test first, branch well
---
```

**Step 3: Validate**

```bash
claude plugin validate plugin/
```

Expected: 0 errors (warnings about author OK — we fix that in Task 2).

**Step 4: Commit**

```bash
git add plugin/
git commit -m "fix: correct hooks.json format and add teaching skill frontmatter"
```

---

### Task 2: Update package.json and plugin.json for publishing

**Files:**
- Modify: `package.json`
- Modify: `plugin/.claude-plugin/plugin.json`
- Modify: `bin/cli.js`
- Modify: `.gitignore`

**Step 1: Update package.json**

Replace the full `package.json`:

```json
{
  "name": "@huypl53/agent-tutor",
  "version": "0.2.0",
  "description": "Programming tutor plugin for coding agents (Claude Code, Codex CLI)",
  "license": "MIT",
  "author": "huypl53",
  "homepage": "https://github.com/huypl53/agent-tutor",
  "repository": {
    "type": "git",
    "url": "https://github.com/huypl53/agent-tutor.git"
  },
  "keywords": [
    "claude-code",
    "codex",
    "programming-tutor",
    "mcp",
    "coding-agent",
    "plugin"
  ],
  "bin": {
    "agent-tutor": "./bin/cli.js"
  },
  "files": [
    "bin/",
    "plugin/",
    "scripts/",
    "CLAUDE.md",
    "AGENTS.md"
  ],
  "publishConfig": {
    "access": "public"
  },
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.12.1",
    "chokidar": "^4.0.3",
    "commander": "^13.1.0",
    "zod": "^3.24.0"
  },
  "engines": {
    "node": ">=18"
  }
}
```

**Step 2: Update plugin.json**

Replace `plugin/.claude-plugin/plugin.json`:

```json
{
  "name": "agent-tutor",
  "version": "0.2.0",
  "description": "Programming tutor plugin for coding agents",
  "author": {
    "name": "huypl53",
    "url": "https://github.com/huypl53"
  },
  "homepage": "https://github.com/huypl53/agent-tutor",
  "mcpServers": {
    "agent-tutor": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/servers/tutoring-mcp.js"]
    }
  },
  "hooks": "./hooks/hooks.json",
  "skills": "./skills/"
}
```

**Step 3: Make CLI version dynamic**

In `bin/cli.js`, change line 66 from:

```javascript
  .version('0.2.0')
```

To:

```javascript
  .version(require('../package.json').version)
```

**Step 4: Update .gitignore**

Add `.npmrc` to `.gitignore`:

```
.worktrees
node_modules/
.npmrc
```

**Step 5: Validate again**

```bash
claude plugin validate plugin/
```

Expected: 0 errors, 0 warnings.

**Step 6: Commit**

```bash
git add package.json plugin/.claude-plugin/plugin.json bin/cli.js .gitignore
git commit -m "feat: prepare package.json and plugin.json for npm + marketplace publishing"
```

---

### Task 3: Add marketplace manifest

**Files:**
- Create: `.claude-plugin/marketplace.json`

**Step 1: Create marketplace.json**

Create `.claude-plugin/marketplace.json` at repo root:

```json
{
  "$schema": "https://anthropic.com/claude-code/marketplace.schema.json",
  "name": "agent-tutor-marketplace",
  "description": "Programming tutor plugin for coding agents",
  "owner": {
    "name": "huypl53",
    "url": "https://github.com/huypl53"
  },
  "plugins": [
    {
      "name": "agent-tutor",
      "description": "Programming tutor — coaching, debugging, code review via MCP tools and skills",
      "category": "development",
      "source": "./plugin",
      "homepage": "https://github.com/huypl53/agent-tutor"
    }
  ]
}
```

**Step 2: Validate marketplace**

```bash
claude plugin validate .claude-plugin/marketplace.json
```

Expected: Valid marketplace manifest.

**Step 3: Commit**

```bash
git add .claude-plugin/
git commit -m "feat: add marketplace manifest for Claude Code plugin discovery"
```

---

### Task 4: Update README with all install paths

**Files:**
- Modify: `README.md`

**Step 1: Rewrite installation section**

Replace the Installation and Quick Start sections in `README.md` with:

```markdown
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
```

**Step 2: Update CLI commands table**

Update the Commands section to use the scoped package name in examples where `npx` is used.

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README with marketplace + scoped npm install instructions"
```

---

### Task 5: Update architecture docs

**Files:**
- Modify: `docs/architecture.md`

**Step 1: Add Distribution section**

Add a section to `docs/architecture.md` after the "Key Design Decisions" section:

```markdown
## Distribution

Three install channels:

1. **Claude Code marketplace** — `claude plugin marketplace add github:huypl53/agent-tutor` then `claude plugin install agent-tutor`. Plugin auto-starts MCP server, loads skills, registers hooks.

2. **npm** — `npm install -g @huypl53/agent-tutor`. Use via `claude --plugin-dir $(agent-tutor plugin-dir)` or `agent-tutor install` to inject CLAUDE.md instructions.

3. **Codex CLI** — `npx @huypl53/agent-tutor install --agent codex` injects AGENTS.md. MCP server added manually via `codex mcp add`.

The marketplace manifest (`.claude-plugin/marketplace.json`) at repo root points to `plugin/` as a `git-subdir` source. This lets Claude Code install directly from the GitHub repo without npm.
```

**Step 2: Commit**

```bash
git add docs/architecture.md
git commit -m "docs: add distribution section to architecture"
```

---

### Task 6: Publish to npm and verify all channels

**Step 1: Login to npm**

```bash
npm adduser
```

Follow the prompts to authenticate.

**Step 2: Dry-run publish**

```bash
npm publish --dry-run
```

Verify: package name is `@huypl53/agent-tutor`, files list includes `bin/`, `plugin/`, `scripts/`, `CLAUDE.md`, `AGENTS.md`.

**Step 3: Publish**

```bash
npm publish
```

**Step 4: Verify npm install**

```bash
npx @huypl53/agent-tutor --help
npx @huypl53/agent-tutor plugin-dir
```

Expected: Shows help and prints plugin directory path.

**Step 5: Verify marketplace install**

```bash
claude plugin marketplace add github:huypl53/agent-tutor
claude plugin install agent-tutor
claude plugin list
```

Expected: agent-tutor appears in installed plugins list.

**Step 6: Push to GitHub**

```bash
git push
```
