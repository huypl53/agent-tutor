#!/usr/bin/env node
'use strict';

const { McpServer } = require('@modelcontextprotocol/sdk/server/mcp.js');
const { StdioServerTransport } = require('@modelcontextprotocol/sdk/server/stdio.js');
const { z } = require('zod');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const chokidar = require('chokidar');
const { StateManager, TOPIC_STATUSES } = require('./state-manager');
const { ProjectScanner } = require('./project-scanner');

// --- Config ---

const CONFIG_PATH = path.resolve(process.cwd(), '.agent-tutor/config.json');
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

function gitDiffFile(filePath) {
  try {
    return require('child_process').execFileSync('git', ['diff', '--', filePath], { encoding: 'utf8', timeout: 5000 }).trim();
  } catch { return ''; }
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
    try { diff = gitDiffFile(filePath); } catch {}
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
  version: require('../../package.json').version,
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

// --- Learning State ---

const stateManager = new StateManager(process.cwd());
const projectScanner = new ProjectScanner(process.cwd());

// Run migration on startup
stateManager.migrateIfNeeded().catch(err => {
  console.error('Migration error:', err);
});

server.tool('create_topic',
  'Register a new learning topic the student is working on',
  {
    id: z.string().describe('URL-safe identifier for the topic (e.g. "async-await")'),
    title: z.string().describe('Human-readable topic title'),
    complexity: z.number().min(1).max(10).optional().describe('Estimated complexity 1-10'),
    dependencies: z.array(z.string()).optional().describe('IDs of prerequisite topics'),
  },
  async ({ id, title, complexity, dependencies }) => {
    try {
      const topic = await stateManager.createTopic({ id, title, complexity, dependencies });
      return { content: [{ type: 'text', text: JSON.stringify(topic, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('update_topic',
  'Update a learning topic: change status, add a moment, set complexity, or link a lesson',
  {
    id: z.string().describe('Topic ID to update'),
    status: z.enum(TOPIC_STATUSES).optional().describe('New status (must be a valid transition)'),
    moment: z.object({
      type: z.enum(['struggle', 'hint', 'breakthrough', 'practice']),
      detail: z.string(),
    }).optional().describe('A learning moment to record'),
    complexity: z.number().min(1).max(10).optional().describe('Updated complexity estimate'),
    lessonFile: z.string().optional().describe('Path to the saved lesson file'),
  },
  async ({ id, status, moment, complexity, lessonFile }) => {
    try {
      const topic = await stateManager.updateTopic(id, { status, moment, complexity, lessonFile });
      return { content: [{ type: 'text', text: JSON.stringify(topic, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('get_topic',
  'Get full details of a learning topic',
  { id: z.string().describe('Topic ID') },
  async ({ id }) => {
    const topic = await stateManager.getTopic(id);
    if (!topic) return { content: [{ type: 'text', text: `Topic "${id}" not found.` }] };
    return { content: [{ type: 'text', text: JSON.stringify(topic, null, 2) }] };
  }
);

server.tool('list_topics',
  'List all learning topics, optionally filtered by status',
  { status: z.enum(TOPIC_STATUSES).optional().describe('Filter by status') },
  async ({ status }) => {
    const topics = await stateManager.listTopics(status);
    return { content: [{ type: 'text', text: JSON.stringify(topics, null, 2) }] };
  }
);

server.tool('delete_topic',
  'Delete a learning topic by ID',
  { id: z.string().describe('Topic ID to delete') },
  async ({ id }) => {
    try {
      const topic = await stateManager.deleteTopic(id);
      return { content: [{ type: 'text', text: `Deleted topic "${topic.title}".` }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('get_topic_graph',
  'Get the topic dependency graph with nodes and edges',
  {},
  async () => {
    const graph = await stateManager.getTopicGraph();
    return { content: [{ type: 'text', text: JSON.stringify(graph, null, 2) }] };
  }
);

server.tool('create_plan',
  'Create a structured learning plan with ordered steps',
  {
    goal: z.string().describe('The learning goal'),
    steps: z.array(z.object({
      topicId: z.string(),
      order: z.number(),
    })).describe('Ordered steps referencing topic IDs'),
    force: z.boolean().optional().describe('Set to true to overwrite an existing plan'),
  },
  async ({ goal, steps, force }) => {
    try {
      const plan = await stateManager.createPlan({ goal, steps, force });
      return { content: [{ type: 'text', text: JSON.stringify(plan, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('update_plan',
  'Update learning plan: mark steps completed, add steps',
  {
    stepUpdates: z.array(z.object({
      topicId: z.string(),
      status: z.enum(['pending', 'active', 'mastered', 'skipped']).optional(),
      order: z.number().optional(),
      action: z.enum(['add']).optional().describe('Set to "add" to add a new step'),
    })).describe('Array of step updates'),
  },
  async ({ stepUpdates }) => {
    try {
      const plan = await stateManager.updatePlan(stepUpdates);
      return { content: [{ type: 'text', text: JSON.stringify(plan, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('get_plan',
  'Get the current learning plan with progress',
  {},
  async () => {
    const plan = await stateManager.getPlan();
    if (!plan) return { content: [{ type: 'text', text: 'No learning plan exists yet.' }] };
    return { content: [{ type: 'text', text: JSON.stringify(plan, null, 2) }] };
  }
);

server.tool('delete_plan',
  'Delete the current learning plan',
  {},
  async () => {
    try {
      const plan = await stateManager.deletePlan();
      return { content: [{ type: 'text', text: `Deleted plan "${plan.goal}".` }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('save_session',
  'Save current session context for recovery after /clear or /compact',
  {
    activeTopicId: z.string().describe('The currently active topic ID'),
    resumeContext: z.string().describe('Description of what the student was doing'),
  },
  async ({ activeTopicId, resumeContext }) => {
    const session = await stateManager.saveSession({ activeTopicId, resumeContext });
    return { content: [{ type: 'text', text: JSON.stringify(session, null, 2) }] };
  }
);

server.tool('restore_session',
  'Restore the last saved session context',
  {},
  async () => {
    const session = await stateManager.restoreSession();
    if (!session) return { content: [{ type: 'text', text: 'No saved session found.' }] };
    return { content: [{ type: 'text', text: JSON.stringify(session, null, 2) }] };
  }
);

server.tool('get_learning_summary',
  'Get an aggregate learning summary: topics by status, plan progress, recent moments',
  {},
  async () => {
    const summary = await stateManager.getLearningSummary();
    return { content: [{ type: 'text', text: JSON.stringify(summary, null, 2) }] };
  }
);

// --- Project Analysis ---

server.tool('scan_project',
  'Scan the project structure, detect type, parse manifests, identify entry points. Fast — no source reading.',
  {},
  async () => {
    try {
      const profile = projectScanner.scan();
      await stateManager.saveProjectProfile(profile);
      return { content: [{ type: 'text', text: JSON.stringify(profile, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('get_project_profile',
  'Get the stored project profile and list of analysis docs',
  {},
  async () => {
    const profile = await stateManager.getProjectProfile();
    if (!profile) return { content: [{ type: 'text', text: 'No project profile. Run scan_project first.' }] };
    const docs = await stateManager.listProjectDocs();
    return { content: [{ type: 'text', text: JSON.stringify({ ...profile, availableDocs: docs }, null, 2) }] };
  }
);

server.tool('save_project_doc',
  'Save a project analysis document (used by sub-agents during onboarding)',
  {
    name: z.string().describe('Document name without extension (e.g. "architecture", "api-contracts")'),
    content: z.string().describe('Markdown content of the analysis document'),
  },
  async ({ name, content }) => {
    try {
      const filePath = await stateManager.saveProjectDoc(name, content);
      return { content: [{ type: 'text', text: `Saved to ${filePath}` }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

// --- Start ---

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch(err => {
  console.error('MCP server error:', err);
  process.exit(1);
});
