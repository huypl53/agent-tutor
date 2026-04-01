# npm Plugin Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the Go binary with a pure npm package. Agent-tutor becomes a Claude Code plugin with a Node.js MCP server, skills, hooks, and CLAUDE.md/AGENTS.md instructions.

**Architecture:** npm package with `bin/cli.js` for install/uninstall, `plugin/` directory for Claude Code plugin system (auto-starts MCP server via `plugin.json`), and `scripts/` for hook helpers. MCP server uses `@modelcontextprotocol/sdk` with `chokidar` for file watching and `child_process` for git queries.

**Tech Stack:** Node.js, `@modelcontextprotocol/sdk`, `chokidar`, `commander` (CLI)

---

### Task 1: Scaffold npm package and delete Go code

**Files:**
- Create: `package.json`
- Create: `.gitignore` (update for node_modules)
- Delete: `cmd/`, `internal/`, `go.mod`, `go.sum`, `.worktrees/`

**Step 1: Delete all Go code and worktree**

```bash
cd /home/huypham/code/spare/agent-tutor
git worktree remove .worktrees/pure-plugin --force 2>/dev/null
git branch -D feature/pure-plugin-architecture 2>/dev/null
rm -rf cmd/ internal/ go.mod go.sum .worktrees/
```

**Step 2: Create package.json**

Create `package.json`:

```json
{
  "name": "agent-tutor",
  "version": "0.2.0",
  "description": "Programming tutor plugin for coding agents (Claude Code, Codex CLI)",
  "license": "MIT",
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
  "dependencies": {
    "@anthropic-ai/sdk": "^0.52.0",
    "@modelcontextprotocol/sdk": "^1.12.1",
    "chokidar": "^4.0.3",
    "commander": "^13.1.0"
  },
  "engines": {
    "node": ">=18"
  }
}
```

**Step 3: Update .gitignore**

Add `node_modules/` to `.gitignore` if not present.

**Step 4: Install dependencies**

```bash
npm install
```

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: replace Go with npm package scaffold"
```

---

### Task 2: Create plugin.json and hooks.json

**Files:**
- Create: `plugin/.claude-plugin/plugin.json`
- Create: `plugin/hooks/hooks.json`

**Step 1: Create plugin.json**

Create `plugin/.claude-plugin/plugin.json`:

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

**Step 2: Create hooks.json**

Create `plugin/hooks/hooks.json`:

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

**Step 3: Commit**

```bash
git add plugin/
git commit -m "feat: add plugin.json and hooks.json"
```

---

### Task 3: Move skills and hook scripts

**Files:**
- Move: `internal/plugin/embed/commands/*.md` → `plugin/skills/atu-*/SKILL.md`
- Move: `internal/plugin/embed/skills/` → `plugin/skills/`
- Move: `internal/plugin/embed/hooks/*.js` → `scripts/`

**Step 1: Create skill directories from command files**

Each command file (e.g., `atu-check.md`) becomes `plugin/skills/atu-check/SKILL.md`. The files need their `name:` frontmatter field to use colons (e.g., `atu:check`).

```bash
mkdir -p plugin/skills
for cmd in internal/plugin/embed/commands/*.md; do
  base=$(basename "$cmd" .md)        # e.g. atu-check
  mkdir -p "plugin/skills/$base"
  cp "$cmd" "plugin/skills/$base/SKILL.md"
done
```

**Step 2: Copy teaching skills**

```bash
cp -r internal/plugin/embed/skills/* plugin/skills/
```

**Step 3: Copy hook scripts**

```bash
mkdir -p scripts
cp internal/plugin/embed/hooks/*.js scripts/
```

**Step 4: Update skill files to remove stale MCP tool references**

In `plugin/skills/atu-check/SKILL.md`, `atu-debug/SKILL.md`, `atu-explain/SKILL.md`: replace references to `get_terminal_activity` with `get_student_context` or `get_recent_file_changes` as appropriate. (The worktree branch already has these fixes — use those versions.)

**Step 5: Verify all files are in place**

```bash
find plugin/skills -name "SKILL.md" | sort
ls scripts/*.js
```

Expected: 13 SKILL.md files (9 commands + 4 teaching skills), 2 JS scripts.

**Step 6: Commit**

```bash
git add plugin/skills/ scripts/
git commit -m "feat: migrate skills and hook scripts from Go embed"
```

---

### Task 4: Write the MCP server

**Files:**
- Create: `plugin/servers/tutoring-mcp.js`

**Step 1: Write the MCP server**

Create `plugin/servers/tutoring-mcp.js`. This is a Node.js MCP server using `@modelcontextprotocol/sdk` that provides 5 tools over stdio.

```javascript
#!/usr/bin/env node
'use strict';

const { McpServer } = require('@modelcontextprotocol/sdk/server/mcp.js');
const { StdioServerTransport } = require('@modelcontextprotocol/sdk/server/stdio.js');
const { z } = require('zod');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const chokidar = require('chokidar');

// --- Config ---

const CONFIG_PATH = '.agent-tutor/config.json';
const DEFAULT_CONFIG = { intensity: 'on-demand', level: 'auto' };

function loadConfig() {
  try {
    return { ...DEFAULT_CONFIG, ...JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8')) };
  } catch { return { ...DEFAULT_CONFIG }; }
}

function saveConfig(cfg) {
  fs.mkdirSync(path.dirname(CONFIG_PATH), { recursive: true });
  fs.writeFileSync(CONFIG_PATH, JSON.stringify(cfg, null, 2) + '\n');
}

// --- File watcher ring buffer ---

const MAX_FILE_EVENTS = 100;
const fileEvents = [];

function addFileEvent(evt) {
  fileEvents.push(evt);
  if (fileEvents.length > MAX_FILE_EVENTS) fileEvents.shift();
}

function git(cmd) {
  try { return execSync(`git ${cmd}`, { encoding: 'utf8', timeout: 5000 }).trim(); }
  catch { return ''; }
}

// --- Start file watcher ---

const FILE_PATTERNS = ['**/*.{js,ts,jsx,tsx,py,go,rs,java,rb,c,cpp,h,css,html,md,json,toml,yaml,yml}'];
const IGNORE_PATTERNS = ['**/node_modules/**', '**/.git/**', '**/vendor/**', '**/target/**', '**/.agent-tutor/**'];

const watcher = chokidar.watch(FILE_PATTERNS, {
  ignored: IGNORE_PATTERNS,
  persistent: true,
  ignoreInitial: true,
  awaitWriteFinish: { stabilityThreshold: 300 },
});

watcher.on('all', (event, filePath) => {
  let diff = '';
  if ((event === 'change' || event === 'add') && fs.existsSync(filePath)) {
    try { diff = git(`diff -- "${filePath}"`); } catch {}
  }
  addFileEvent({
    path: filePath,
    change: event === 'add' ? 'create' : event === 'unlink' ? 'delete' : 'modify',
    diff: diff.slice(0, 500),
    timestamp: new Date().toISOString(),
  });
});

// --- MCP Server ---

const server = new McpServer({
  name: 'agent-tutor',
  version: '0.2.0',
});

server.tool('get_student_context',
  'Get a summary of recent student activity including file changes and git operations',
  {},
  async () => {
    const recentFiles = fileEvents.slice(-20).map(e =>
      `- **${e.change}** \`${e.path}\``
    ).join('\n');

    const gitLog = git('log --oneline -10');
    const gitStatus = git('status --porcelain');
    const gitDiff = git('diff --stat');

    let summary = '';
    if (recentFiles) summary += `## File Changes\n\n${recentFiles}\n\n`;
    if (gitDiff) summary += `## Uncommitted Changes\n\n\`\`\`\n${gitDiff}\n\`\`\`\n\n`;
    if (gitLog) summary += `## Recent Commits\n\n\`\`\`\n${gitLog}\n\`\`\`\n\n`;
    if (gitStatus) summary += `## Working Tree\n\n\`\`\`\n${gitStatus}\n\`\`\`\n`;

    return { content: [{ type: 'text', text: summary || 'No recent activity.' }] };
  }
);

server.tool('get_recent_file_changes',
  'Get recent file changes with diffs',
  {},
  async () => {
    if (fileEvents.length === 0) {
      return { content: [{ type: 'text', text: 'No recent file changes.' }] };
    }
    const lines = fileEvents.slice(-30).map(e => {
      let s = `- ${e.change}: ${e.path}`;
      if (e.diff) s += `\n  \`\`\`\n  ${e.diff.slice(0, 200)}\n  \`\`\``;
      return s;
    });
    return { content: [{ type: 'text', text: lines.join('\n') }] };
  }
);

server.tool('get_git_activity',
  'Get recent git activity including commits and status changes',
  {},
  async () => {
    const log = git('log --oneline -10');
    const status = git('status --porcelain');
    let text = '';
    if (log) text += `## Recent Commits\n\n\`\`\`\n${log}\n\`\`\`\n\n`;
    if (status) text += `## Working Tree Status\n\n\`\`\`\n${status}\n\`\`\`\n`;
    return { content: [{ type: 'text', text: text || 'No recent git activity.' }] };
  }
);

server.tool('get_coaching_config',
  'Get the current coaching configuration (intensity and level)',
  {},
  async () => {
    const cfg = loadConfig();
    return { content: [{ type: 'text', text: `intensity: ${cfg.intensity}\nlevel: ${cfg.level}` }] };
  }
);

server.tool('set_coaching_intensity',
  'Set the coaching intensity level',
  { intensity: z.enum(['proactive', 'on-demand', 'silent']).describe('The coaching intensity level') },
  async ({ intensity }) => {
    const cfg = loadConfig();
    cfg.intensity = intensity;
    saveConfig(cfg);
    return { content: [{ type: 'text', text: `Coaching intensity set to: ${intensity}` }] };
  }
);

// --- Instructions ---

function buildInstructions() {
  const cfg = loadConfig();
  return `You are also a programming tutor. A student is using a coding agent and you observe their work via file changes and git operations.

Coaching intensity: ${cfg.intensity}
Student level: ${cfg.level}

When intensity is "proactive":
- After the student messages you, also check get_student_context for teachable moments
- Weave teaching naturally into your responses — don't lecture

When intensity is "on-demand" or "silent":
- Only use tutor tools when the student explicitly asks for feedback or uses /atu:check

Teaching style:
- Explain the "why" not just the "what"
- For beginners: explain concepts, suggest resources
- For experienced devs: focus on idioms, best practices, ecosystem conventions
- Be concise. One teaching point per interaction, not five.
- If the student is doing well, say nothing.`;
}

// --- Start ---

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch(err => {
  console.error('MCP server error:', err);
  process.exit(1);
});
```

**Step 2: Verify it starts**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | node plugin/servers/tutoring-mcp.js
```

Expected: JSON response with server capabilities.

**Step 3: Commit**

```bash
git add plugin/servers/
git commit -m "feat: add Node.js MCP server with 5 tutor tools"
```

---

### Task 5: Write the CLI installer

**Files:**
- Create: `bin/cli.js`

**Step 1: Write the CLI**

Create `bin/cli.js`:

```javascript
#!/usr/bin/env node
'use strict';

const { program } = require('commander');
const fs = require('fs');
const path = require('path');

const BEGIN = '<!-- BEGIN AGENT-TUTOR -->';
const END = '<!-- END AGENT-TUTOR -->';

function getInstructionsContent() {
  const mdPath = path.join(__dirname, '..', 'CLAUDE.md');
  return fs.readFileSync(mdPath, 'utf8');
}

function resolveClaudeMdPath(scope) {
  if (scope === 'global') {
    return path.join(require('os').homedir(), '.claude', 'CLAUDE.md');
  }
  return path.join(process.cwd(), '.claude', 'CLAUDE.md');
}

function resolveAgentsMdPath(scope) {
  if (scope === 'global') {
    return path.join(require('os').homedir(), 'AGENTS.md');
  }
  return path.join(process.cwd(), 'AGENTS.md');
}

function injectSection(filePath, section) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
  let content = '';
  try { content = fs.readFileSync(filePath, 'utf8'); } catch {}

  // Remove existing section (idempotent)
  if (content.includes(BEGIN)) {
    content = removeSection(content);
  }

  if (content && !content.endsWith('\n')) content += '\n';
  if (content) content += '\n';
  content += section + '\n';

  fs.writeFileSync(filePath, content);
}

function removeSection(content) {
  const beginIdx = content.indexOf(BEGIN);
  const endIdx = content.indexOf(END);
  if (beginIdx < 0 || endIdx < 0) return content;

  let before = content.slice(0, beginIdx).trimEnd();
  let after = content.slice(endIdx + END.length).trimStart();

  if (!before) return after;
  if (!after) return before + '\n';
  return before + '\n\n' + after;
}

function getPluginDir() {
  return path.join(__dirname, '..', 'plugin');
}

program
  .name('agent-tutor')
  .version('0.2.0')
  .description('Programming tutor plugin for coding agents');

program
  .command('install')
  .description('Install agent-tutor plugin')
  .option('--scope <scope>', 'local or global', 'local')
  .option('--agent <agent>', 'claude or codex', 'claude')
  .action((opts) => {
    const section = getInstructionsContent();

    if (opts.agent === 'codex') {
      const agentsPath = resolveAgentsMdPath(opts.scope);
      injectSection(agentsPath, section);
      console.log(`Injected tutor instructions into ${agentsPath}`);
      console.log(`\nAdd MCP server:\n  codex mcp add agent-tutor -- node ${path.join(getPluginDir(), 'servers', 'tutoring-mcp.js')}`);
    } else {
      const claudePath = resolveClaudeMdPath(opts.scope);
      injectSection(claudePath, section);
      console.log(`Injected tutor instructions into ${claudePath}`);
      console.log(`\nUse as plugin:\n  claude --plugin-dir ${getPluginDir()}`);
    }

    console.log('\nAvailable commands: /atu:check, /atu:hint, /atu:explain, /atu:debug, /atu:review, /atu:plan');
  });

program
  .command('uninstall')
  .description('Remove agent-tutor plugin')
  .option('--scope <scope>', 'local or global', 'local')
  .option('--agent <agent>', 'claude or codex', 'claude')
  .action((opts) => {
    const filePath = opts.agent === 'codex'
      ? resolveAgentsMdPath(opts.scope)
      : resolveClaudeMdPath(opts.scope);

    try {
      const content = fs.readFileSync(filePath, 'utf8');
      fs.writeFileSync(filePath, removeSection(content));
      console.log(`Removed tutor instructions from ${filePath}`);
    } catch {
      console.log('Nothing to uninstall.');
    }
  });

program
  .command('plugin-dir')
  .description('Print the plugin directory path (for --plugin-dir)')
  .action(() => {
    console.log(getPluginDir());
  });

program.parse();
```

**Step 2: Make executable**

```bash
chmod +x bin/cli.js
```

**Step 3: Test locally**

```bash
node bin/cli.js --help
node bin/cli.js plugin-dir
```

Expected: Shows help with install/uninstall/plugin-dir commands, prints plugin path.

**Step 4: Commit**

```bash
git add bin/
git commit -m "feat: add Node.js CLI installer (install/uninstall/plugin-dir)"
```

---

### Task 6: Write CLAUDE.md and AGENTS.md instruction files

**Files:**
- Create: `CLAUDE.md`
- Create: `AGENTS.md`

**Step 1: Write CLAUDE.md**

This is the tutor instruction section that gets injected. It's the same content as the Go `claudeMDSection` constant but as a standalone file. Remove the MCP Tools Reference table (tools are auto-discovered by the plugin system). Keep everything else: commands, teaching skills, coaching behavior, pedagogical principles, topic tracking, learning plan awareness, hook awareness, lesson auto-save.

**Step 2: Write AGENTS.md**

Same content as CLAUDE.md but formatted for Codex CLI. Can be identical content wrapped in the same sentinels.

**Step 3: Commit**

```bash
git add CLAUDE.md AGENTS.md
git commit -m "feat: add CLAUDE.md and AGENTS.md instruction files"
```

---

### Task 7: Update documentation

**Files:**
- Modify: `README.md`
- Modify: `docs/architecture.md`

**Step 1: Rewrite README.md**

Update for npm package:
- Installation: `npm install -g agent-tutor` or `npx agent-tutor install`
- Usage: `claude --plugin-dir $(npx agent-tutor plugin-dir)` or `npx agent-tutor install`
- Remove all Go references
- Show Codex CLI setup alongside Claude Code
- Update config section (JSON, not TOML)

**Step 2: Rewrite docs/architecture.md**

Update for new architecture:
- npm package structure
- Node.js MCP server (tools, file watcher, git queries)
- Plugin system integration (plugin.json, hooks.json)
- CLI installer (inject/remove CLAUDE.md sections)
- No Go, no ring buffer store (simplified), no polling loops

**Step 3: Commit**

```bash
git add README.md docs/architecture.md
git commit -m "docs: rewrite for npm plugin architecture"
```

---

### Task 8: Clean up old files and final verification

**Files:**
- Delete: remaining Go artifacts if any
- Delete: `docs/plans/2026-04-01-pure-plugin-architecture.md` (superseded)
- Delete: `docs/plans/2026-04-01-pure-plugin-architecture-design.md` (superseded)

**Step 1: Remove old plan files**

```bash
rm docs/plans/2026-04-01-pure-plugin-architecture.md
rm docs/plans/2026-04-01-pure-plugin-architecture-design.md
```

**Step 2: Verify plugin works with Claude Code**

```bash
# Verify MCP server starts
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | node plugin/servers/tutoring-mcp.js

# Verify CLI
node bin/cli.js --help
node bin/cli.js plugin-dir

# Verify install/uninstall in temp dir
tmpdir=$(mktemp -d)
cd "$tmpdir"
node /home/huypham/code/spare/agent-tutor/bin/cli.js install
cat .claude/CLAUDE.md | head -5
node /home/huypham/code/spare/agent-tutor/bin/cli.js uninstall
cd -
rm -rf "$tmpdir"

# Verify all skills exist
find plugin/skills -name "SKILL.md" | wc -l
# Expected: 13

# Verify hook scripts
ls scripts/*.js
# Expected: large-file-detect.js, error-pattern-detect.js
```

**Step 3: Commit any fixes**

```bash
git add -A
git commit -m "chore: clean up old files and verify"
```
